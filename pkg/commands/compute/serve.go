package compute

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bep/debounce"
	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/mitchellh/go-ps"
	ignore "github.com/sabhiram/go-gitignore"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/check"
	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	fstruntime "github.com/fastly/cli/pkg/runtime"
	"github.com/fastly/cli/pkg/text"
)

var viceroyError = fsterr.RemediationError{
	Inner:       fmt.Errorf("a Viceroy version was not found"),
	Remediation: fsterr.BugRemediation,
}

// ServeCommand produces and runs an artifact from files on the local disk.
type ServeCommand struct {
	argparser.Base
	build *BuildCommand

	// Build fields
	dir                   argparser.OptionalString
	includeSrc            argparser.OptionalBool
	lang                  argparser.OptionalString
	metadataDisable       argparser.OptionalBool
	metadataFilterEnvVars argparser.OptionalString
	metadataShow          argparser.OptionalBool
	packageName           argparser.OptionalString
	timeout               argparser.OptionalInt

	// Serve public fields (public for testing purposes)
	ForceCheckViceroyLatest bool
	ViceroyBinExtraArgs     string
	ViceroyBinPath          string
	ViceroyVersioner        github.AssetVersioner

	// Serve private fields
	addr                 string
	debug                bool
	enablePushpin        bool
	pushpinRunnerBinPath string
	pushpinProxyPort     string
	pushpinPublishPort   string
	env                  argparser.OptionalString
	file                 argparser.OptionalString
	profileGuest         bool
	profileGuestDir      argparser.OptionalString
	projectDir           string
	skipBuild            bool
	watch                bool
	watchDir             argparser.OptionalString
}

// NewServeCommand returns a usable command registered under the parent.
func NewServeCommand(parent argparser.Registerer, g *global.Data, build *BuildCommand) *ServeCommand {
	var c ServeCommand
	c.build = build
	c.Globals = g
	c.ViceroyVersioner = g.Versioners.Viceroy
	c.CmdClause = parent.Command("serve", "Build and run a Compute package locally")

	c.CmdClause.Flag("addr", "The IPv4 address and port to listen on").Default("127.0.0.1:7676").StringVar(&c.addr)
	c.CmdClause.Flag("debug", "Run the server in Debug Adapter mode").Hidden().BoolVar(&c.debug)
	c.CmdClause.Flag("dir", "Project directory to build (default: current directory)").Short('C').Action(c.dir.Set).StringVar(&c.dir.Value)
	c.CmdClause.Flag("env", "The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')").Action(c.env.Set).StringVar(&c.env.Value)
	c.CmdClause.Flag("file", "The Wasm file to run (causes build process to be skipped)").Action(c.file.Set).StringVar(&c.file.Value)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("metadata-disable", "Disable Wasm binary metadata annotations").Action(c.metadataDisable.Set).BoolVar(&c.metadataDisable.Value)
	c.CmdClause.Flag("metadata-filter-envvars", "Redact specified environment variables from [scripts.env_vars] using comma-separated list").Action(c.metadataFilterEnvVars.Set).StringVar(&c.metadataFilterEnvVars.Value)
	c.CmdClause.Flag("metadata-show", "Inspect the Wasm binary metadata").Action(c.metadataShow.Set).BoolVar(&c.metadataShow.Value)
	c.CmdClause.Flag("package-name", "Package name").Action(c.packageName.Set).StringVar(&c.packageName.Value)
	c.CmdClause.Flag("experimental-enable-pushpin", "Enable experimental Pushpin support for local testing of Fanout and WebSockets").BoolVar(&c.enablePushpin)
	c.CmdClause.Flag("pushpin-path", "The path to a user installed version of the Pushpin runner binary").StringVar(&c.pushpinRunnerBinPath)
	c.CmdClause.Flag("pushpin-proxy-port", "The port to run the Pushpin runner on. Overrides 'local_server.pushpin.proxy_port' from 'fastly.toml', and if not specified there, defaults to 7677.").StringVar(&c.pushpinProxyPort)
	c.CmdClause.Flag("pushpin-publish-port", "The port to run the Pushpin publish handler on. Overrides 'local_server.pushpin.publish_port' from 'fastly.toml', and if not specified there, defaults to 5561.").StringVar(&c.pushpinPublishPort)
	c.CmdClause.Flag("profile-guest", "Profile the Wasm guest under Viceroy (requires Viceroy 0.9.1 or higher). View profiles at https://profiler.firefox.com/.").BoolVar(&c.profileGuest)
	c.CmdClause.Flag("profile-guest-dir", "The directory where the per-request profiles are saved to. Defaults to guest-profiles.").Action(c.profileGuestDir.Set).StringVar(&c.profileGuestDir.Value)
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.skipBuild)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").Action(c.timeout.Set).IntVar(&c.timeout.Value)
	c.CmdClause.Flag("viceroy-args", "Additional arguments to pass to the Viceroy binary, separated by space").StringVar(&c.ViceroyBinExtraArgs)
	c.CmdClause.Flag("viceroy-check", "Force the CLI to check for a newer version of the Viceroy binary").BoolVar(&c.ForceCheckViceroyLatest)
	c.CmdClause.Flag("viceroy-path", "The path to a user installed version of the Viceroy binary").StringVar(&c.ViceroyBinPath)
	c.CmdClause.Flag("watch", "Watch for file changes, then rebuild project and restart local server").BoolVar(&c.watch)
	c.CmdClause.Flag("watch-dir", "The directory to watch files from (can be relative or absolute). Defaults to current directory.").Action(c.watchDir.Set).StringVar(&c.watchDir.Value)

	return &c
}

// Exec implements the command interface.
func (c *ServeCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if c.skipBuild && c.watch {
		return fsterr.ErrIncompatibleServeFlags
	}

	if runtime.GOARCH == "386" {
		return fsterr.RemediationError{
			Inner:       errors.New("this command doesn't support the '386' architecture"),
			Remediation: "Although the Fastly CLI supports '386', the `compute serve` command requires https://github.com/fastly/Viceroy which does not.",
		}
	}

	manifestFilename := EnvironmentManifest(c.env.Value)
	if c.env.Value != "" {
		if c.Globals.Verbose() {
			text.Info(out, EnvManifestMsg, manifestFilename, manifest.Filename)
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()
	manifestPath := filepath.Join(wd, manifestFilename)

	c.projectDir, err = ChangeProjectDirectory(c.dir.Value)
	if err != nil {
		return err
	}
	if c.projectDir != "" {
		if c.Globals.Verbose() {
			text.Info(out, ProjectDirMsg, c.projectDir)
		}
		manifestPath = filepath.Join(c.projectDir, manifestFilename)
	}

	wasmBinaryToRun := binWasmPath
	if c.file.WasSet {
		wasmBinaryToRun = c.file.Value
	}

	// We skip the build if explicitly told to with --skip-build but also when the
	// user sets --file to specify their own wasm binary to pass to Viceroy. This
	// is typically for users who compile a Wasm binary using an unsupported
	// programming language for the Fastly Compute platform.
	if !c.skipBuild && !c.file.WasSet {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
		text.Break(out)
	}

	c.setBackendsWithDefaultOverrideHostIfMissing(out)

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	// NOTE: We read again the manifest to catch a skip-build scenario.
	//
	// For example, a user runs `compute build` then `compute serve --skip-build`.
	// In that scenario our in-memory manifest could be invalid as the user might
	// have also called `compute serve --skip-build --env <...> --dir <...>`.
	//
	// If the user doesn't set --skip-build then `compute serve` will call
	// `compute build` and the logic there will update the manifest in-memory data
	// with the relevant project directory and environment manifest content.
	if c.skipBuild || c.file.WasSet {
		err := c.Globals.Manifest.File.Read(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to parse manifest '%s': %w", manifestPath, err)
		}
		c.ViceroyVersioner.SetRequestedVersion(c.Globals.Manifest.File.LocalServer.ViceroyVersion)
		if c.Globals.Verbose() {
			if c.skipBuild || c.file.WasSet {
				text.Break(out)
			}
			text.Info(out, "Fastly manifest set to: %s\n\n", manifestPath)
		}
	}

	bin, err := c.GetViceroy(spinner, out, manifestPath)
	if err != nil {
		return err
	}

	enablePushpin := c.enablePushpin || c.Globals.Manifest.File.LocalServer.Pushpin.EnablePushpin
	var pushpinCtx pushpinContext
	if enablePushpin {
		pushpinCtx, err = c.startPushpin(spinner, out)
		if err != nil {
			pushpinCtx.Close()
			return err
		}
		defer pushpinCtx.Close()
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	msg := "Running local server"
	spinner.Message(msg + "...")

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return err
	}

	if c.Globals.Verbose() {
		text.Break(out)
	}

	var restart bool
	for {
		err = local(localOpts{
			addr:             c.addr,
			bin:              bin,
			debug:            c.debug,
			errLog:           c.Globals.ErrLog,
			extraArgs:        c.ViceroyBinExtraArgs,
			manifestPath:     manifestPath,
			out:              out,
			profileGuest:     c.profileGuest,
			profileGuestDir:  c.profileGuestDir,
			pushpinProxyPort: pushpinCtx.proxyPort,
			restarted:        restart,
			verbose:          c.Globals.Verbose(),
			wasmBinPath:      wasmBinaryToRun,
			watch:            c.watch,
			watchDir:         c.watchDir,
		})
		if err != nil {
			if err != fsterr.ErrViceroyRestart {
				if err == fsterr.ErrSignalInterrupt || err == fsterr.ErrSignalKilled {
					text.Info(out, "\nLocal server stopped")
					return nil
				}
				return err
			}

			// Before restarting Viceroy we should rebuild.
			text.Break(out)
			err = c.Build(in, out)
			if err != nil {
				// NOTE: build errors at this point are going to be user related, so we
				// should display the error but keep watching the files so we can
				// rebuild successfully once the user has fixed the issues.
				fsterr.Deduce(err).Print(color.Error)
			}
			restart = true
		}
	}
}

// Build constructs and executes the build logic.
func (c *ServeCommand) Build(in io.Reader, out io.Writer) error {
	// Reset the fields on the BuildCommand based on ServeCommand values.
	if c.dir.WasSet {
		c.build.Flags.Dir = c.dir.Value
	}
	if c.env.WasSet {
		c.build.Flags.Env = c.env.Value
	}
	if c.includeSrc.WasSet {
		c.build.Flags.IncludeSrc = c.includeSrc.Value
	}
	if c.lang.WasSet {
		c.build.Flags.Lang = c.lang.Value
	}
	if c.packageName.WasSet {
		c.build.Flags.PackageName = c.packageName.Value
	}
	if c.timeout.WasSet {
		c.build.Flags.Timeout = c.timeout.Value
	}
	if c.metadataDisable.WasSet {
		c.build.MetadataDisable = c.metadataDisable.Value
	}
	if c.metadataFilterEnvVars.WasSet {
		c.build.MetadataFilterEnvVars = c.metadataFilterEnvVars.Value
	}
	if c.metadataShow.WasSet {
		c.build.MetadataShow = c.metadataShow.Value
	}
	if c.projectDir != "" {
		c.build.SkipChangeDir = true // we've already changed directory
	}
	return c.build.Exec(in, out)
}

// setBackendsWithDefaultOverrideHostIfMissing sets an override_host for any
// local_server.backends that is missing that property. The value will only be
// set if the URL defined uses a hostname (e.g. http://127.0.0.1/ won't) so we
// can set the override_host to match the hostname.
func (c *ServeCommand) setBackendsWithDefaultOverrideHostIfMissing(out io.Writer) {
	for k, backend := range c.Globals.Manifest.File.LocalServer.Backends {
		if backend.OverrideHost == "" {
			if u, err := url.Parse(backend.URL); err == nil {
				segs := strings.Split(u.Host, ":") // avoid parsing IP with port
				if ip := net.ParseIP(segs[0]); ip == nil {
					if c.Globals.Verbose() {
						text.Info(out, "[local_server.backends.%s] (%s) is configured without an `override_host`. We will use %s as a default to help avoid any unexpected errors. See https://www.fastly.com/documentation/reference/compute/fastly-toml#local-server for more details.", k, backend.URL, u.Host)
					}
					backend.OverrideHost = u.Host
					c.Globals.Manifest.File.LocalServer.Backends[k] = backend
				}
			}
		}
	}
}

// GetViceroy returns the path to the installed binary.
//
// If Viceroy is installed we either update it or pin it to the version defined
// in the fastly.toml [viceroy.viceroy_version]. Otherwise, if not installed, we
// install it in the same directory as the application configuration data.
//
// In the case of a network failure we fallback to the latest installed version of the
// Viceroy binary as long as one is installed and has the correct permissions.
func (c *ServeCommand) GetViceroy(spinner text.Spinner, out io.Writer, manifestPath string) (bin string, err error) {
	if c.ViceroyBinPath != "" {
		if c.Globals.Verbose() {
			text.Info(out, "Using user provided install of Viceroy via --viceroy-path flag: %s\n\n", c.ViceroyBinPath)
		}
		return filepath.Abs(c.ViceroyBinPath)
	}

	// Allows a user to use a version of Viceroy that is installed in the $PATH.
	if usePath := os.Getenv("FASTLY_VICEROY_USE_PATH"); checkViceroyEnvVar(usePath) {
		path, err := exec.LookPath("viceroy")
		if err != nil {
			return "", fmt.Errorf("failed to lookup viceroy binary in user $PATH (user has set $FASTLY_VICEROY_USE_PATH): %w", err)
		}
		if c.Globals.Verbose() {
			text.Info(out, "Using user provided install of Viceroy via $PATH lookup: %s\n\n", path)
		}
		return filepath.Abs(path)
	}

	bin = filepath.Join(github.InstallDir, c.ViceroyVersioner.BinaryName())

	// NOTE: When checking if Viceroy is installed we don't use
	// exec.LookPath("viceroy") because PATH is unreliable across OS platforms,
	// but also we actually install Viceroy in the same location as the
	// application configuration, which means it wouldn't be found looking up by
	// the PATH env var. We could pass the path for the application configuration
	// into exec.LookPath() but it's simpler to just execute the binary.
	//
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	// nosemgrep
	command := exec.Command(bin, "--version")

	var installedVersion string

	stdoutStderr, err := command.CombinedOutput()
	if err != nil {
		c.Globals.ErrLog.Add(err)
	} else {
		// Check the version output has the expected format: `viceroy 0.1.0`
		installedVersion = strings.TrimSpace(string(stdoutStderr))
		segs := strings.Split(installedVersion, " ")
		if len(segs) < 2 {
			return bin, viceroyError
		}
		installedVersion = segs[1]
	}

	// If the user hasn't explicitly set a Viceroy version, then we'll use
	// whatever the latest version is.
	versionToInstall := "latest"
	if v := c.ViceroyVersioner.RequestedVersion(); v != "" {
		versionToInstall = v

		if _, err := semver.Parse(versionToInstall); err != nil {
			return bin, fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to parse configured version as a semver: %w", err),
				Remediation: fmt.Sprintf("Ensure the %s `viceroy_version` value '%s' (under the [local_server] section) is a valid semver (https://semver.org/), e.g. `0.1.0`)", manifestPath, versionToInstall),
			}
		}
	}

	err = c.InstallViceroy(installedVersion, versionToInstall, manifestPath, bin, spinner)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return bin, err
	}

	err = github.SetBinPerms(bin)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return bin, err
	}
	return bin, nil
}

// checkViceroyEnvVar indicates if the CLI should use a Viceroy binary exposed
// on the user's $PATH.
func checkViceroyEnvVar(value string) bool {
	switch strings.ToUpper(value) {
	case "1", "TRUE":
		return true
	}
	return false
}

// InstallViceroy downloads the binary from GitHub.
//
// The logic flow is as follows:
//
// 1. Check if version to install is "latest"
// 2. If so, check the latest release matches the installed version.
// 3. If not latest, check the installed version matches the expected version.
func (c *ServeCommand) InstallViceroy(
	installedVersion, versionToInstall, manifestPath, bin string,
	spinner text.Spinner,
) error {
	var (
		err         error
		msg, tmpBin string
	)

	switch {
	case installedVersion == "": // Viceroy not installed
		if c.Globals.Verbose() {
			text.Info(c.Globals.Output, "Viceroy is not already installed, so we will install the %s version.\n\n", versionToInstall)
		}
		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Fetching Viceroy release: %s", versionToInstall)
		spinner.Message(msg + "...")

		if versionToInstall == "latest" {
			tmpBin, err = c.ViceroyVersioner.DownloadLatest()
		} else {
			tmpBin, err = c.ViceroyVersioner.DownloadVersion(versionToInstall)
		}
	case versionToInstall != "latest":
		if installedVersion == versionToInstall {
			if c.Globals.Verbose() {
				text.Info(c.Globals.Output, "Viceroy is already installed, and the installed version matches the required version (%s) in the %s file.\n\n", versionToInstall, manifestPath)
			}
			return nil
		}
		if c.Globals.Verbose() {
			text.Info(c.Globals.Output, "Viceroy is already installed, but the installed version (%s) doesn't match the required version (%s) specified in the %s file.\n\n", installedVersion, versionToInstall, manifestPath)
		}

		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Fetching Viceroy release: %s", versionToInstall)
		spinner.Message(msg + "...")

		tmpBin, err = c.ViceroyVersioner.DownloadVersion(versionToInstall)
	case versionToInstall == "latest":
		// Viceroy is already installed, so we check if the installed version matches the latest.
		// But we'll skip that check if the TTL for the Viceroy LastChecked hasn't expired.

		stale := check.Stale(c.Globals.Config.Viceroy.LastChecked, c.Globals.Config.Viceroy.TTL)
		if !stale && !c.ForceCheckViceroyLatest {
			if c.Globals.Verbose() {
				text.Info(c.Globals.Output, "Viceroy is installed but the CLI config (`fastly config`) shows the TTL, checking for a newer version, hasn't expired. To force a refresh, re-run the command with the `--viceroy-check` flag.\n\n")
			}
			return nil
		}

		// IMPORTANT: We declare separately so to shadow `err` from parent scope.
		var latestVersion string

		// NOTE: We won't stop the user because although we can't request the latest
		// version of the tool, the user may have a local version already installed.
		err = spinner.Process("Checking latest Viceroy release", func(_ *text.SpinnerWrapper) error {
			latestVersion, err = c.ViceroyVersioner.LatestVersion()
			if err != nil {
				return fsterr.RemediationError{
					Inner:       fmt.Errorf("error fetching latest version: %w", err),
					Remediation: fsterr.NetworkRemediation,
				}
			}
			return nil
		})
		if err != nil {
			return nil // short-circuit the rest of this function
		}

		viceroyConfig := c.Globals.Config.Viceroy
		viceroyConfig.LatestVersion = latestVersion
		viceroyConfig.LastChecked = time.Now().Format(time.RFC3339)

		// Before attempting to write the config data back to disk we need to
		// ensure we reassign the modified struct which is a copy (not reference).
		c.Globals.Config.Viceroy = viceroyConfig

		err = c.Globals.Config.Write(c.Globals.ConfigPath)
		if err != nil {
			return err
		}

		if c.Globals.Verbose() {
			text.Info(c.Globals.Output, "\nThe CLI config (`fastly config`) has been updated with the latest Viceroy version: %s\n\n", latestVersion)
		}

		if installedVersion != "" && installedVersion == latestVersion {
			return nil
		}

		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Fetching Viceroy release: %s", versionToInstall)
		spinner.Message(msg + "...")

		tmpBin, err = c.ViceroyVersioner.DownloadLatest()
	}

	// NOTE: The above `switch` needs to shadow the function-level `err` variable.
	if err != nil {
		err = fmt.Errorf("error downloading Viceroy release: %w", err)
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
		}
		return err
	}
	defer os.RemoveAll(tmpBin)

	if err := os.Rename(tmpBin, bin); err != nil {
		err = fmt.Errorf("failed to rename/move file: %w", err)
		if copyErr := filesystem.CopyFile(tmpBin, bin); copyErr != nil {
			err = fmt.Errorf("failed to copy file: %w (original error: %w)", copyErr, err)
			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
			}
			return err
		}
	}

	spinner.StopMessage(msg)
	return spinner.Stop()
}

// GetPushpinProxyPort returns the port to run the Pushpin proxy.
//
// The default value is 7677.
// It can be overridden by providing the --pushpin-proxy-port command-line parameter.
// If it is not found then `local_server.pushpin.proxy_port` in fastly.toml is also checked.
func (c *ServeCommand) GetPushpinProxyPort(out io.Writer) (uint16, error) {
	pushpinProxyPortStr := c.pushpinProxyPort
	var pushpinProxyPort uint16
	if pushpinProxyPortStr != "" {
		pushpinProxyPortInt, err := strconv.Atoi(pushpinProxyPortStr)
		if err != nil {
			return 0, fmt.Errorf("can't parse --pushpin-proxy-port value as a number: %s", pushpinProxyPortStr)
		}
		if pushpinProxyPortInt < 1 || pushpinProxyPortInt > 65535 {
			return 0, fmt.Errorf("--pushpin-proxy-port must be a number between 1 and 65535 (got: %d)", pushpinProxyPortInt)
		}
		pushpinProxyPort = uint16(pushpinProxyPortInt)
		if c.Globals.Verbose() {
			text.Info(out, "Using Pushpin proxy port from --pushpin-proxy-port flag: %d\n\n", pushpinProxyPort)
		}
		return pushpinProxyPort, nil
	}

	pushpinProxyPort = c.Globals.Manifest.File.LocalServer.Pushpin.PushpinProxyPort
	if pushpinProxyPort != 0 {
		if c.Globals.Verbose() {
			text.Info(out, "Using Pushpin proxy port via `local_server.pushpin.proxy_port` setting: %d\n\n", pushpinProxyPort)
		}
		return pushpinProxyPort, nil
	}

	pushpinProxyPort = 7677
	if c.Globals.Verbose() {
		text.Info(out, "Using default Pushpin proxy port %d\n\n", pushpinProxyPort)
	}
	return pushpinProxyPort, nil
}

// GetPushpinPublishPort returns the port to run the Pushpin publishing handler.
// The design of Pushpin opens four ports starting with this port, though the publishing
// handler itself runs on the specified port.
//
// The default value is 5561.
// It can be overridden by providing the --pushpin-publish-port command-line parameter.
// If it is not found then `local_server.pushpin.publish_port` in fastly.toml is also checked.
func (c *ServeCommand) GetPushpinPublishPort(out io.Writer) (uint16, error) {
	pushpinPublishPortStr := c.pushpinPublishPort
	var pushpinPublishPort uint16
	if pushpinPublishPortStr != "" {
		pushpinPublishPortInt, err := strconv.Atoi(pushpinPublishPortStr)
		if err != nil {
			return 0, fmt.Errorf("can't parse --pushpin-publish-port value as a number: %s", pushpinPublishPortStr)
		}
		if pushpinPublishPortInt < 1 || pushpinPublishPortInt > 65535 {
			return 0, fmt.Errorf("--pushpin-publish-port must be a number between 1 and 65535 (got: %d)", pushpinPublishPortInt)
		}
		pushpinPublishPort = uint16(pushpinPublishPortInt)
		if c.Globals.Verbose() {
			text.Info(out, "Using Pushpin publish handler port from --pushpin-publish-port flag: %d\n\n", pushpinPublishPort)
		}
		return pushpinPublishPort, nil
	}

	pushpinPublishPort = c.Globals.Manifest.File.LocalServer.Pushpin.PushpinPublishPort
	if pushpinPublishPort != 0 {
		if c.Globals.Verbose() {
			text.Info(out, "Using Pushpin publish handler port via `local_server.pushpin.publish_port` setting: %d\n\n", pushpinPublishPort)
		}
		return pushpinPublishPort, nil
	}

	pushpinPublishPort = 5561
	if c.Globals.Verbose() {
		text.Info(out, "Using default Pushpin publish handler port %d\n\n", pushpinPublishPort)
	}
	return pushpinPublishPort, nil
}

// GetPushpinRunner returns the path to the installed Pushpin binary.
//
// This value comes from searching the system path for `pushpin`
// It can be overridden by providing the --pushpin-path command-line parameter.
// If it is not found then `local_server.pushpin.pushpin_path` in fastly.toml is also checked.
func (c *ServeCommand) GetPushpinRunner(out io.Writer) (bin string, err error) {
	pushpinRunnerBinPath := c.pushpinRunnerBinPath
	if pushpinRunnerBinPath != "" {
		if c.Globals.Verbose() {
			text.Info(out, "Using user provided install of Pushpin runner via --pushpin-path flag: %s\n\n", pushpinRunnerBinPath)
		}
		return filepath.Abs(pushpinRunnerBinPath)
	}

	pushpinRunnerBinPath = c.Globals.Manifest.File.LocalServer.Pushpin.PushpinPath
	if pushpinRunnerBinPath != "" {
		if c.Globals.Verbose() {
			text.Info(out, "Using user provided install of Pushpin runner via `local_server.pushpin.pushpin_path` setting: %s\n\n", pushpinRunnerBinPath)
		}
		return filepath.Abs(pushpinRunnerBinPath)
	}

	if c.Globals.Verbose() {
		text.Info(out, "No --pushpin-path provided, attempting to find 'pushpin' in your PATH...")
	}
	pushpinRunnerBinPath, err = exec.LookPath("pushpin")
	if err != nil {
		return "", fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to find 'pushpin' in your $PATH"),
			Remediation: "Pushpin support was enabled via --enable-experimental-pushpin, but the 'pushpin' binary could not be found in your $PATH. Please install Pushpin (see: https://pushpin.org/docs/install/) or provide a path to the binary using the --pushpin-path flag.",
		}
	}

	if c.Globals.Verbose() {
		text.Info(out, "Found Pushpin runner via $PATH lookup: %s\n\n", pushpinRunnerBinPath)
	}
	return filepath.Abs(pushpinRunnerBinPath)
}

// BuildPushpinRoutes builds a slice of strings based on the backends
// defined in the manifest's backend section.
func (c *ServeCommand) BuildPushpinRoutes() []string {
	var routes []string
	for name, backend := range c.Globals.Manifest.File.LocalServer.Backends {

		// The target should be a URL
		u, err := url.Parse(backend.URL)
		if err != nil {
			// This is unlikely as we parse it elsewhere, but good to be safe.
			// We'll just skip this backend if the URL is invalid.
			continue
		}

		// Route Rule:
		// 1. `id=<backend_name>`: Match requests whose Pushpin-Route header equals the backend name.
		rules := fmt.Sprintf("id=%s", name)

		// 2. A backend may have a path component. If it does, then it will be prepended during forwarding.
		forwardPrefix := strings.TrimSuffix(u.Path, "/")
		if forwardPrefix != "" {
			rules += fmt.Sprintf(",replace_beg=%s", forwardPrefix)
		}

		// Target:
		target := u.Host
		// 1. `over_http`: Enable WebSocket-over-HTTP
		target += ",over_http"
		// 2. `ssl`: If backend is https
		if u.Scheme == "https" {
			target += ",ssl"
		}
		// 3. `host`: If the backend has an override_host.
		if backend.OverrideHost != "" {
			target += fmt.Sprintf(",host=%s", backend.OverrideHost)
		}

		// The final route format
		routeArg := fmt.Sprintf("%s %s", rules, target)
		routes = append(routes, routeArg)
	}

	return routes
}

func formatPushpinLog(line string) (string, string) {
	level := "INFO"
	msg := line

	if strings.HasPrefix(line, "[ERR]") || strings.HasPrefix(line, "[WARN]") ||
		strings.HasPrefix(line, "[INFO]") || strings.HasPrefix(line, "[DEBUG]") {
		parts := strings.SplitN(line, " ", 4)
		if len(parts) == 4 {
			level = strings.Trim(parts[0], "[]")
			if level == "ERR" {
				level = "ERROR"
			}
			msg = parts[3]
		}
	}

	// Return as-is if it doesn't match pattern
	return level, "[Pushpin] " + msg
}

// pushpinContext contains information about the instance of Pushpin that is
// executed when enabled.
type pushpinContext struct {
	pushpinRunnerBin string
	pushpinRunDir    string
	pushpinLogDir    string
	routesFilePath   string
	proxyPort        uint16
	publishPort      uint16
	cleanup          func()
}

// Close ends Pushpin if it's running by calling the registered cleanup function.
func (c *pushpinContext) Close() {
	if c.cleanup != nil {
		c.cleanup()
	}
}

// pushpinConfTemplate is a template used by buildPushpinConf.
//
//go:embed pushpin.conf.template
var pushpinConfTemplate string

// buildPushpinConf builds a temporary pushpin.conf file that contains everything that covers our needs.
func (c *pushpinContext) buildPushpinConf() string {
	pullPort := c.publishPort + 1
	subPort := c.publishPort + 2
	repPort := c.publishPort + 3
	return fmt.Sprintf(
		pushpinConfTemplate,
		c.pushpinRunDir,
		c.pushpinLogDir,
		c.routesFilePath,
		c.proxyPort,
		c.publishPort,
		pullPort,
		subPort,
		repPort,
	)
}

// startPushpin starts Pushpin based on the configuration provided by the
// command line and/or fastly.toml. The cleanup function on the returned pushpinContext
// needs to eventually be called by the caller to shut down Pushpin.
func (c *ServeCommand) startPushpin(spinner text.Spinner, out io.Writer) (pushpinContext, error) {
	pushpinCtx := pushpinContext{}

	var cleanup func()

	err := spinner.Start()
	if err != nil {
		return pushpinCtx, err
	}
	msg := "Running local Pushpin"
	spinner.Message(msg + "...")

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return pushpinCtx, err
	}

	pwd, _ := os.Getwd()
	pushpinCtx.pushpinLogDir = filepath.Join(pwd, "pushpin-logs")

	pushpinCtx.proxyPort, err = c.GetPushpinProxyPort(out)
	if err != nil {
		return pushpinCtx, err
	}
	pushpinCtx.publishPort, err = c.GetPushpinPublishPort(out)
	if err != nil {
		return pushpinCtx, err
	}
	pushpinCtx.pushpinRunnerBin, err = c.GetPushpinRunner(out)
	if err != nil {
		return pushpinCtx, err
	}
	if c.Globals.Verbose() {
		text.Info(out, "Local pushpin proxy port: %d", pushpinCtx.proxyPort)
		text.Info(out, "Local pushpin publisher port: %d", pushpinCtx.publishPort)
		text.Info(out, "Local pushpin other reserved ports: %d - %d", pushpinCtx.publishPort+1, pushpinCtx.publishPort+3)
		text.Info(out, "Local pushpin runner: %s", pushpinCtx.pushpinRunnerBin)
	}

	// Pushpin is configured with the following.
	// - A conf file that sets up the parameters of the instance. In our case, we:
	//   - set the runtime temporary files directory
	//   - set the log output directory
	//   - enable "pushpin-route" header for routing
	//   - set the message size (64k) to match Fanout
	//   - set the publishing addr and port
	//   - path to the routes file to use
	// - A routes file that sets up the routes. In our case, we:
	//   - wires up a backend name (id) to the server host
	//   - if the backend sets an override host, then we set thatt
	//   - if the backend enables HTTPS, then we enable that
	//   - if the backend has a path prefix, then we set that up
	//   - enables WebSocket-over-HTTP

	// The runtime temporary directory, as well as the conf file and routes file
	// are set up and torn down along with fastly compute serve.
	var pushpinInstanceID uint32
	for {
		p := make([]byte, 4)
		_, _ = rand.Read(p)
		pushpinInstanceID = binary.BigEndian.Uint32(p)
		if pushpinInstanceID != 0 {
			break
		}
	}

	pushpinRunDir := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("pushpin-%08x", pushpinInstanceID),
	)
	if c.Globals.Verbose() {
		text.Break(out)
		text.Info(out, "Pushpin temporary runtime directory is %s", pushpinRunDir)
	}

	pushpinConfFilePath := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("pushpin-%08x.conf", pushpinInstanceID),
	)
	pushpinRoutesFilePath := filepath.Join(
		os.TempDir(),
		fmt.Sprintf("pushpin-routes-%08x", pushpinInstanceID),
	)

	if c.Globals.Verbose() {
		text.Info(out, "Writing config file to %s...", pushpinConfFilePath)
	}

	pushpinConfContents := pushpinCtx.buildPushpinConf()
	err = os.WriteFile(pushpinConfFilePath, []byte(pushpinConfContents), 0o600)
	if err != nil {
		return pushpinCtx, fmt.Errorf("error writing config file %s: %w", pushpinConfFilePath, err)
	}

	if c.Globals.Verbose() {
		text.Info(out, "Writing routes file to %s...", pushpinRoutesFilePath)
	}
	pushpinRoutesContents := strings.Join(c.BuildPushpinRoutes(), "\n") + "\n"
	err = os.WriteFile(pushpinRoutesFilePath, []byte(pushpinRoutesContents), 0o600)
	if err != nil {
		return pushpinCtx, fmt.Errorf("error writing routes file %s: %w", pushpinRoutesFilePath, err)
	}

	if c.Globals.Verbose() {
		text.Info(out, "Starting local Pushpin...")
		text.Break(out)
	}

	args := []string{
		fmt.Sprintf("--config=%s", pushpinConfFilePath),
		"--verbose",
	}

	// Set up a context that can be canceled (prevent rogue Pushpin process)
	var pushpinCmd *exec.Cmd
	ctx, cancel := context.WithCancel(context.Background())

	var once sync.Once
	cleanup = func() {
		once.Do(func() {
			if pushpinCmd != nil && pushpinCmd.Process != nil {
				if c.Globals.Verbose() {
					text.Output(out, "shutting down Pushpin")
				}
				killProcessTree(pushpinCmd.Process.Pid)
			}
			if c.Globals.Verbose() {
				text.Output(out, "removing %s", pushpinRunDir)
			}
			_ = os.RemoveAll(pushpinRunDir)
			if c.Globals.Verbose() {
				text.Output(out, "deleting %s", pushpinConfFilePath)
			}
			_ = os.Remove(pushpinConfFilePath)
			if c.Globals.Verbose() {
				text.Output(out, "deleting %s", pushpinRoutesFilePath)
			}
			_ = os.Remove(pushpinRoutesFilePath)
			cancel()
		})
	}

	// Also allow other forms of termination to perform cleanups
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		<-sigCh
		cleanup()
	}()

	// gosec flagged this:
	// G204: Subprocess launched with a potential tainted input or cmd arguments
	// Disabling as we control this command.
	// #nosec
	// nosemgrep
	pushpinCmd = exec.CommandContext(ctx, pushpinCtx.pushpinRunnerBin, args...)
	pushpinCmd.Stderr = out
	stdout, err := pushpinCmd.StdoutPipe()
	if err != nil {
		return pushpinCtx, fmt.Errorf("failed to capture Pushpin stdout: %w", err)
	}

	// Start Pushpin
	if c.Globals.Verbose() {
		text.Info(out, "Starting Pushpin runner in the background...")
		text.Output(out, "%s: %s", text.BoldYellow("Pushpin command"), strings.Join(pushpinCmd.Args, " "))
	}
	if err := pushpinCmd.Start(); err != nil {
		return pushpinCtx, fmt.Errorf("failed to start Pushpin runner: %w", err)
	}

	// Monitor output from Pushpin
	// 1. convert output and log it
	// 2. wait for a timeout after a startup message
	startupError := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			// Successful if timeout passes after seeing 'started'
			if strings.HasSuffix(line, "started") {
				go func() {
					time.Sleep(1000 * time.Millisecond)
					startupError <- nil
				}()
			}

			level, msg := formatPushpinLog(line)
			if level != "DEBUG" || c.Globals.Verbose() {
				text.Output(out, "%s  %s %s", time.Now().UTC().Format("2006-01-02T15:04:05.000000Z"), level, msg)
			}
		}

		if err := scanner.Err(); err != nil {
			startupError <- fmt.Errorf("error reading Pushpin output: %w", err)
		} else {
			startupError <- fmt.Errorf("process Pushpin terminated")
		}
	}()

	// Startup error
	err = <-startupError
	if err != nil {
		return pushpinCtx, fsterr.RemediationError{
			Inner:       err,
			Remediation: fmt.Sprintf("A process may already be running on port %d.", pushpinCtx.proxyPort),
		}
	}

	if c.Globals.Verbose() {
		text.Info(out, "Local Pushpin started on port %d.", pushpinCtx.proxyPort)
		text.Break(out)
	}

	return pushpinCtx, nil
}

// localOpts represents the inputs for `local()`.
type localOpts struct {
	addr             string
	bin              string
	debug            bool
	errLog           fsterr.LogInterface
	extraArgs        string
	manifestPath     string
	out              io.Writer
	profileGuest     bool
	profileGuestDir  argparser.OptionalString
	pushpinProxyPort uint16
	restarted        bool
	verbose          bool
	wasmBinPath      string
	watch            bool
	watchDir         argparser.OptionalString
}

// local spawns a subprocess that runs the compiled binary.
func local(opts localOpts) error {
	// NOTE: Viceroy no longer displays errors unless in verbose mode.
	// This can cause confusion for customers: https://github.com/fastly/cli/issues/913
	// So regardless of CLI --verbose flag we'll always set verbose for Viceroy.
	args := []string{"-v", "-C", opts.manifestPath, "--addr", opts.addr, opts.wasmBinPath}

	if opts.debug {
		args = append(args, "--debug")
	}

	if opts.profileGuest {
		directory := "guest-profiles"
		if opts.profileGuestDir.WasSet {
			directory = opts.profileGuestDir.Value
		}
		args = append(args, "--profile=guest,"+directory)
		if opts.verbose {
			text.Info(opts.out, "Saving per-request profiles to %s.", directory)
		}
	}

	if opts.pushpinProxyPort != 0 {
		args = append(args, fmt.Sprintf("--local-pushpin-proxy-port=%d", opts.pushpinProxyPort))
	}

	if opts.extraArgs != "" {
		extraArgs := strings.Split(opts.extraArgs, " ")
		args = append(args, extraArgs...)
	}

	if opts.verbose {
		if opts.restarted {
			text.Break(opts.out)
		}
		text.Output(opts.out, "%s: %s", text.BoldYellow("Manifest"), opts.manifestPath)
		text.Output(opts.out, "%s: %s", text.BoldYellow("Wasm binary"), opts.wasmBinPath)
		text.Output(opts.out, "%s: %s", text.BoldYellow("Viceroy command"), strings.Join(args, " "))
		text.Output(opts.out, "%s: %s", text.BoldYellow("Viceroy binary"), opts.bin)

		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
		// Disabling as we trust the source of the variable.
		// #nosec
		// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
		c := exec.Command(opts.bin, "--version")
		if output, err := c.Output(); err == nil {
			text.Output(opts.out, "%s: %s", text.BoldYellow("Viceroy version"), string(output))
		}
		text.Info(opts.out, "Listening on http://%s", opts.addr)
		if opts.watch {
			text.Break(opts.out)
		}
	}

	s := &fstexec.Streaming{
		Args:        args,
		Command:     opts.bin,
		Env:         os.Environ(),
		ForceOutput: true,
		Output:      opts.out,
		SignalCh:    make(chan os.Signal, 1),
	}
	s.MonitorSignals()

	failure := make(chan error)
	restart := make(chan bool)
	if opts.watch {
		root := "."
		if opts.watchDir.WasSet {
			root = opts.watchDir.Value
		}

		if opts.verbose {
			text.Info(opts.out, "Watching files for changes (using --watch-dir=%s). To ignore certain files, define patterns within a .fastlyignore config file (uses .fastlyignore from --watch-dir).\n\n", root)
		}

		gi := ignoreFiles(opts.watchDir)
		go watchFiles(root, gi, opts.verbose, s, opts.out, restart, failure)
	}

	// NOTE: The viceroy executable can be stopped by one of three mechanisms.
	//
	// 1. File modification
	// 2. Explicit signal (SIGINT, SIGTERM etc).
	// 3. Irrecoverable error (i.e. error watching files).
	//
	// In the case of a signal (e.g. user presses Ctrl-c) the listener logic
	// inside of (*fstexec.Streaming).MonitorSignals() will call
	// (*fstexec.Streaming).Signal(signal os.Signal) to kill the process.
	//
	// In the case of a file modification the viceroy executable needs to first
	// be killed (handled by the watchFiles() function) and then we can stop the
	// signal listeners (handled below by sending a message to argparser.SignalCh).
	//
	// If we don't tell the signal listening channel to close, then the restart
	// of the viceroy executable will cause the user to end up with N number of
	// listeners. This will result in a "os: process already finished" error when
	// we do finally come to stop the `serve` command (e.g. user presses Ctrl-c).
	// How big an issue this is depends on how many file modifications a user
	// makes, because having lots of signal listeners could exhaust resources.
	//
	// When there is an error setting up the watching of files, if we error we
	// need to signal the error using a channel as watching files happens
	// asynchronously in a goroutine. We also need to be able to signal the
	// viceroy process to be killed, and we do that using `s.Signal(os.Kill)` from
	// within the relevant error handling blocks in `watchFiles`, where upon the
	// below `select` statement will pull the error message from the `failure`
	// channel and return it to the user. If we fail to kill the Viceroy process
	// then we still want to pull an error from the `failure` channel and so we
	// have a separate `select` statement to check for any initial errors prior to
	// the Viceroy executable starting and an error occurring in `watchFiles`.
	select {
	case asyncErr := <-failure:
		s.SignalCh <- syscall.SIGTERM
		return asyncErr
	case <-time.After(1 * time.Second):
		// no-op: allow logic to flow to starting up Viceroy executable.
	}

	if err := s.Exec(); err != nil {
		errPrefix := "signal: "
		errKilled := "killed"
		if fstruntime.Windows {
			errPrefix = "exit status"
			errKilled = errPrefix + " 1"
		}

		if !strings.Contains(err.Error(), errPrefix) {
			opts.errLog.Add(err)
		}
		e := strings.TrimSpace(err.Error())
		if strings.Contains(e, "interrupt") {
			return fsterr.ErrSignalInterrupt
		}
		if strings.Contains(e, errKilled) {
			select {
			case asyncErr := <-failure:
				s.SignalCh <- syscall.SIGTERM
				return asyncErr
			case <-restart:
				s.SignalCh <- syscall.SIGTERM
				return fsterr.ErrViceroyRestart
			case <-time.After(1 * time.Second):
				return fsterr.ErrSignalKilled
			}
		}
		return err
	}

	return nil
}

// watchFiles watches the language source directory and restarts the viceroy
// executable when changes are detected.
func watchFiles(root string, gi *ignore.GitIgnore, verbose bool, s *fstexec.Streaming, out io.Writer, restart chan<- bool, failure chan<- error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		signalErr := s.Signal(os.Kill)
		if signalErr != nil {
			failure <- fmt.Errorf("failed to stop Viceroy executable while trying to create a fsnotify.Watcher: %w: %w", signalErr, err)
			return
		}
		failure <- fmt.Errorf("failed to create a fsnotify.Watcher: %w", err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	debounced := debounce.New(1 * time.Second)
	eventHandler := func(modifiedFile string, _ fsnotify.Op) {
		// NOTE: We avoid describing the file operation (e.g. created, modified,
		// deleted, renamed etc) rather than checking the fsnotify.Op iota/enum type
		// because the output can be confusing depending on the application used to
		// edit a file.
		//
		// For example, modifying a file in Vim might cause the file to be
		// temporarily copied/renamed and this can cause the watcher to report an
		// existing file has been 'created' or 'renamed' when from a user's
		// perspective the file already exists and was only modified.
		text.Break(out)
		text.Output(out, "%s Restarting local server (%s)", text.BoldGreen("✓"), modifiedFile)

		// NOTE: We force closing the watcher by pushing true into a done channel.
		// We do this because if we didn't, then we'd get an error after one
		// restart of the viceroy executable: "os: process already finished".
		//
		// This error happens happens because the compute.watchFiles() function is
		// run in a goroutine and so it will keep running with a copy of the
		// fstexec.Streaming command instance that wraps a process which has
		// already been terminated.
		done <- true

		// NOTE: To be able to force both the current viceroy process signal listener
		// to close, and to restart the viceroy executable, we need to kill the
		// process and also send 'true' to a restart channel.
		//
		// If we only sent a message to the restart channel, but didn't terminate
		// the process, then we'd end up in a deadlock because we wouldn't be able
		// to take a message from the restart channel inside the local() function
		// because we need to have the process terminate first in order for us to
		// execute the flushing of channel messages.
		//
		// When we stop the signal listener it will internally try to kill the
		// process and discover it has already been killed and return an error:
		// `os: process already finished`. This is why we don't do error handling
		// within (*fstexec.Streaming).MonitorSignalsAsync() as the process could
		// well be killed already when a user is doing local development with the
		// --watch flag. The obvious downside to this logic flow is that if the
		// user is running `compute serve` just to validate the program once, then
		// there might be an unhandled error when they press Ctrl-c to stop the
		// serve command from blocking their terminal. That said, this is unlikely
		// and is a low risk concern.
		err := s.Signal(os.Kill)
		if err != nil {
			failure <- fmt.Errorf("failed to stop Viceroy executable while trying to restart the process: %w", err)
			return
		}

		restart <- true
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				debounced(func() {
					eventHandler(event.Name, event.Op)
				})
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				text.Output(out, "error event while watching files: %v", err)
			}
		}
	}()

	var buf bytes.Buffer

	// Walk all directories and files starting from the project's root directory.
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error configuring watching for file changes: %w", err)
		}
		// If there's no ignore file, we'll default to watching all directories
		// within the specified top-level directory.
		//
		// NOTE: Watching a directory implies watching all files within the root of
		// the directory. This means we don't need to call Add(path) for each file.
		if gi == nil && entry.IsDir() {
			watchFile(path, watcher, verbose, &buf)
		}
		if gi != nil && !entry.IsDir() && !gi.MatchesPath(path) {
			// If there is an ignore file, we avoid watching directories and instead
			// will only add files that don't match the exclusion patterns defined.
			watchFile(path, watcher, verbose, &buf)
		}
		return nil
	})
	if err != nil {
		signalErr := s.Signal(os.Kill)
		if signalErr != nil {
			failure <- fmt.Errorf("failed to stop Viceroy executable while trying to walk directory tree for watching files: %w: %w", signalErr, err)
			return
		}
		failure <- fmt.Errorf("failed to walk directory tree for watching files: %w", err)
		return
	}

	if verbose {
		text.Output(out, "%s\n\n", text.BoldYellow("Watching..."))
		fmt.Fprintln(out, buf.String()) // IMPORTANT: Avoid text.Output() as it fails to render with large buffer.
		text.Break(out)
	}

	<-done
}

// ignoreFiles returns the specific ignore rules being respected.
//
// NOTE: We also ignore the .git directory.
func ignoreFiles(watchDir argparser.OptionalString) *ignore.GitIgnore {
	var patterns []string

	root := ""
	if watchDir.WasSet {
		root = watchDir.Value
		if !strings.HasPrefix(root, "/") {
			root += "/"
		}
	}

	fastlyIgnore := root + ".fastlyignore"

	// NOTE: Using a loop to allow for future ignore files to be respected.
	for _, file := range []string{fastlyIgnore} {
		patterns = append(patterns, readIgnoreFile(file)...)
	}

	patterns = append(patterns, ".git/")

	return ignore.CompileIgnoreLines(patterns...)
}

// readIgnoreFile reads path and splits content into lines.
//
// NOTE: If there's an error reading the given path, then we'll return an empty
// string slice so that the caller can continue to function as expected.
func readIgnoreFile(path string) (lines []string) {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as the input is either provided by our own package or in the
	// case of identifying the user's global git ignore we need to read it from
	// their global git configuration.
	/* #nosec */
	bs, err := os.ReadFile(path)
	if err != nil {
		return lines
	}
	return strings.Split(string(bs), "\n")
}

func watchFile(path string, watcher *fsnotify.Watcher, verbose bool, out io.Writer) {
	absolute, err := filepath.Abs(path)
	if err != nil && verbose {
		text.Warning(out, "Unable to convert '%s' to an absolute path", path)
		return
	}

	err = watcher.Add(absolute)
	if err != nil {
		text.Output(out, "%s %s", text.BoldRed("✗"), absolute)
	} else if verbose {
		text.Output(out, "%s", absolute)
	}
}

func killProcessTree(pid int) {
	processes, err := ps.Processes()
	if err != nil {
		log.Printf("failed to list processes: %v", err)
		return
	}

	var children []int
	for _, p := range processes {
		if p.PPid() == pid {
			children = append(children, p.Pid())
		}
	}

	for _, child := range children {
		killProcessTree(child)
	}

	_ = killProcess(pid)
}
