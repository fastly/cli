package compute

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bep/debounce"
	"github.com/blang/semver"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	ignore "github.com/sabhiram/go-gitignore"

	"github.com/fastly/cli/pkg/check"
	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

var viceroyError = fsterr.RemediationError{
	Inner:       fmt.Errorf("a Viceroy version was not found"),
	Remediation: fsterr.BugRemediation,
}

// ServeCommand produces and runs an artifact from files on the local disk.
type ServeCommand struct {
	cmd.Base
	manifest manifest.Data
	build    *BuildCommand
	av       github.AssetVersioner

	// Build fields
	includeSrc  cmd.OptionalBool
	lang        cmd.OptionalString
	packageName cmd.OptionalString
	timeout     cmd.OptionalInt

	// Serve fields
	addr                    string
	debug                   bool
	env                     cmd.OptionalString
	file                    string
	forceCheckViceroyLatest bool
	profileGuest            bool
	profileGuestDir         cmd.OptionalString
	skipBuild               bool
	viceroyBinPath          string
	watch                   bool
	watchDir                cmd.OptionalString
}

// NewServeCommand returns a usable command registered under the parent.
func NewServeCommand(parent cmd.Registerer, g *global.Data, build *BuildCommand, av github.AssetVersioner, m manifest.Data) *ServeCommand {
	var c ServeCommand

	c.build = build
	c.av = av

	c.Globals = g
	c.CmdClause = parent.Command("serve", "Build and run a Compute@Edge package locally")
	c.manifest = m

	c.CmdClause.Flag("addr", "The IPv4 address and port to listen on").Default("127.0.0.1:7676").StringVar(&c.addr)
	c.CmdClause.Flag("debug", "Run the server in Debug Adapter mode").Hidden().BoolVar(&c.debug)
	c.CmdClause.Flag("env", "The environment configuration to use (e.g. stage)").Action(c.env.Set).StringVar(&c.env.Value)
	c.CmdClause.Flag("file", "The Wasm file to run").Default("bin/main.wasm").StringVar(&c.file)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("package-name", "Package name").Action(c.packageName.Set).StringVar(&c.packageName.Value)
	c.CmdClause.Flag("profile-guest", "Profile the Wasm guest under Viceroy. View profiles at https://profiler.firefox.com/.").BoolVar(&c.profileGuest)
	c.CmdClause.Flag("profile-guest-dir", "The directory where the per-request profiles are saved to. Defaults to guest-profiles.").Action(c.profileGuestDir.Set).StringVar(&c.profileGuestDir.Value)
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.skipBuild)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").Action(c.timeout.Set).IntVar(&c.timeout.Value)
	c.CmdClause.Flag("viceroy-check", "Force the CLI to check for a newer version of the Viceroy binary").BoolVar(&c.forceCheckViceroyLatest)
	c.CmdClause.Flag("viceroy-path", "The path to a user installed version of the Viceroy binary").StringVar(&c.viceroyBinPath)
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

	if !c.skipBuild {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
	}

	text.Break(out)

	c.setBackendsWithDefaultOverrideHostIfMissing(out)

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	bin, err := GetViceroy(spinner, out, c.av, c.Globals, c.viceroyBinPath, c.forceCheckViceroyLatest)
	if err != nil {
		return err
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
			addr:            c.addr,
			bin:             bin,
			debug:           c.debug,
			env:             c.env.Value,
			errLog:          c.Globals.ErrLog,
			file:            c.file,
			out:             out,
			profileGuest:    c.profileGuest,
			profileGuestDir: c.profileGuestDir,
			restarted:       restart,
			verbose:         c.Globals.Verbose(),
			watch:           c.watch,
			watchDir:        c.watchDir,
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

	err := c.build.Exec(in, out)
	if err != nil {
		return err
	}

	return nil
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
						text.Info(out, "[local_server.backends.%s] (%s) is configured without an `override_host`. We will use %s as a default to help avoid any unexpected errors. See https://developer.fastly.com/reference/compute/fastly-toml/#local-server for more details.", k, backend.URL, u.Host)
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
// NOTE: if Viceroy is installed then it is updated, otherwise download the
// latest version and install it in the same directory as the application
// configuration data.
//
// In the case of a network failure we fallback to the latest installed version of the
// Viceroy binary as long as one is installed and has the correct permissions.
func GetViceroy(
	spinner text.Spinner,
	out io.Writer,
	av github.AssetVersioner,
	g *global.Data,
	viceroyBinPath string,
	forceCheckViceroyLatest bool,
) (bin string, err error) {
	if viceroyBinPath != "" {
		if g.Verbose() {
			text.Info(out, "Using user provided install of Viceroy via --viceroy-path flag: %s", viceroyBinPath)
		}
		return filepath.Abs(viceroyBinPath)
	}

	// Allows a user to use a version of Viceroy that is installed in the $PATH.
	if usePath := os.Getenv("FASTLY_VICEROY_USE_PATH"); checkViceroyEnvVar(usePath) {
		path, err := exec.LookPath("viceroy")
		if err != nil {
			return "", fmt.Errorf("failed to lookup viceroy binary in user $PATH (user has set $FASTLY_VICEROY_USE_PATH): %w", err)
		}
		if g.Verbose() {
			text.Info(out, "Using user provided install of Viceroy via $PATH lookup: %s", path)
		}
		return filepath.Abs(path)
	}

	bin = filepath.Join(InstallDir, av.BinaryName())

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
	c := exec.Command(bin, "--version")

	var installedVersion string

	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		g.ErrLog.Add(err)
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
	if v := av.RequestedVersion(); v != "" {
		versionToInstall = v

		if _, err := semver.Parse(versionToInstall); err != nil {
			return bin, fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to parse configured version as a semver: %w", err),
				Remediation: fmt.Sprintf("Ensure the fastly.toml `viceroy_version` value '%s' (under the [local_server] section) is a valid semver (https://semver.org/), e.g. `0.1.0`)", versionToInstall),
			}
		}
	}

	err = installViceroy(installedVersion, versionToInstall, forceCheckViceroyLatest, spinner, av, bin, g)
	if err != nil {
		g.ErrLog.Add(err)
		return bin, err
	}

	err = setBinPerms(bin)
	if err != nil {
		g.ErrLog.Add(err)
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

// InstallDir represents the directory where the Viceroy binary should be
// installed.
//
// NOTE: This is a package level variable as it makes testing the behaviour of
// the package easier because the test code can replace the value when running
// the test suite.
var InstallDir = func() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "fastly")
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".fastly")
	}
	panic("unable to deduce user config dir or user home dir")
}()

// installViceroy downloads the binary from GitHub.
//
// The logic flow is as follows:
//
// 1. Check if version to install is "latest"
// 2. If so, check the latest release matches the installed version.
// 3. If not latest, check the installed version matches the expected version.
func installViceroy(
	installedVersion, versionToInstall string,
	forceCheckViceroyLatest bool,
	spinner text.Spinner,
	av github.AssetVersioner,
	bin string,
	g *global.Data,
) error {
	var (
		err         error
		msg, tmpBin string
	)

	switch {
	case installedVersion == "": // Viceroy not installed
		if g.Verbose() {
			text.Info(g.Output, "Viceroy is not already installed, so we will install the %s version.", versionToInstall)
		}
		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Fetching Viceroy release: %s", versionToInstall)
		spinner.Message(msg + "...")

		if versionToInstall == "latest" {
			tmpBin, err = av.DownloadLatest()
		} else {
			tmpBin, err = av.DownloadVersion(versionToInstall)
		}
	case versionToInstall != "latest":
		if installedVersion == versionToInstall {
			if g.Verbose() {
				text.Info(g.Output, "Viceroy is already installed, and the installed version matches the required version (%s) in the fastly.toml file.", versionToInstall)
			}
			return nil
		}
		if g.Verbose() {
			text.Info(g.Output, "Viceroy is already installed, but the installed version (%s) doesn't match the required version (%s) specified in the fastly.toml file.", installedVersion, versionToInstall)
		}

		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Fetching Viceroy release: %s", versionToInstall)
		spinner.Message(msg + "...")

		tmpBin, err = av.DownloadVersion(versionToInstall)
	case versionToInstall == "latest":
		// Viceroy is already installed, so we check if the installed version matches the latest.
		// But we'll skip that check if the TTL for the Viceroy LastChecked hasn't expired.

		stale := g.Config.Viceroy.LastChecked != "" && g.Config.Viceroy.LatestVersion != "" && check.Stale(g.Config.Viceroy.LastChecked, g.Config.Viceroy.TTL)
		if !stale && !forceCheckViceroyLatest {
			if g.Verbose() {
				text.Info(g.Output, "Viceroy is installed but the CLI config (`fastly config`) shows the TTL, checking for a newer version, hasn't expired. To force a refresh, re-run the command with the `--viceroy-check` flag.\n\n")
			}
			return nil
		}

		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = "Checking latest Viceroy release"
		spinner.Message(msg + "...")

		// IMPORTANT: We declare separately so to shadow `err` from parent scope.
		var latestVersion string

		latestVersion, err = av.LatestVersion()
		if err != nil {
			err = fmt.Errorf("error fetching latest version: %w", err)
			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return spinner.WrapErr(err)
			}
			return fsterr.RemediationError{
				Inner:       err,
				Remediation: fsterr.NetworkRemediation,
			}
		}

		spinner.StopMessage(msg)
		err = spinner.Stop()
		if err != nil {
			return err
		}

		viceroyConfig := g.Config.Viceroy
		viceroyConfig.LatestVersion = latestVersion
		viceroyConfig.LastChecked = time.Now().Format(time.RFC3339)

		// Before attempting to write the config data back to disk we need to
		// ensure we reassign the modified struct which is a copy (not reference).
		g.Config.Viceroy = viceroyConfig

		err = g.Config.Write(g.ConfigPath)
		if err != nil {
			return err
		}

		if g.Verbose() {
			text.Info(g.Output, "The CLI config (`fastly config`) has been updated with the latest Viceroy version: %s", latestVersion)
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

		tmpBin, err = av.DownloadLatest()
	}

	if err != nil {
		err = fmt.Errorf("error downloading Viceroy release: %w", err)
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinner.WrapErr(err)
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
				return spinner.WrapErr(err)
			}
			return err
		}
	}

	spinner.StopMessage(msg)
	return spinner.Stop()
}

// setBinPerms ensures 0777 perms are set on the Viceroy binary.
func setBinPerms(bin string) error {
	// G302 (CWE-276): Expect file permissions to be 0600 or less
	// gosec flagged this:
	// Disabling as the file was not executable without it and we need all users
	// to be able to execute the binary.
	/* #nosec */
	err := os.Chmod(bin, 0o777)
	if err != nil {
		return fmt.Errorf("error setting executable permissions on Viceroy binary: %w", err)
	}
	return nil
}

// localOpts represents the inputs for `local()`.
type localOpts struct {
	addr            string
	bin             string
	debug           bool
	env             string
	errLog          fsterr.LogInterface
	file            string
	out             io.Writer
	profileGuest    bool
	profileGuestDir cmd.OptionalString
	restarted       bool
	verbose         bool
	watch           bool
	watchDir        cmd.OptionalString
}

// local spawns a subprocess that runs the compiled binary.
func local(opts localOpts) error {
	if opts.env != "" {
		opts.env = "." + opts.env
	}

	wd, err := os.Getwd()
	if err != nil {
		opts.errLog.Add(err)
		return err
	}

	manifestPath := filepath.Join(wd, fmt.Sprintf("fastly%s.toml", opts.env))

	// NOTE: Viceroy no longer displays errors unless in verbose mode.
	// This can cause confusion for customers: https://github.com/fastly/cli/issues/913
	// So regardless of CLI --verbose flag we'll always set verbose for Viceroy.
	args := []string{"-v", "-C", manifestPath, "--addr", opts.addr, opts.file}

	if opts.debug {
		args = append(args, "--debug")
	}

	if opts.profileGuest {
		directory := "guest-profiles"
		if opts.profileGuestDir.WasSet {
			directory = opts.profileGuestDir.Value
		}
		args = append(args, "--profile-guest="+directory)
		if opts.verbose {
			text.Info(opts.out, "Saving per-request profiles to %s.", directory)
		}
	}

	if opts.verbose {
		if opts.restarted {
			text.Break(opts.out)
		}
		text.Output(opts.out, "%s: %s", text.BoldYellow("Manifest"), manifestPath)
		text.Output(opts.out, "%s: %s", text.BoldYellow("Wasm binary"), opts.file)
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

	restart := make(chan bool)
	if opts.watch {
		root := "."
		if opts.watchDir.WasSet {
			root = opts.watchDir.Value
		}

		if opts.verbose {
			text.Info(opts.out, "Watching files for changes (using --watch-dir=%s). To ignore certain files, define patterns within a .fastlyignore config file (uses .fastlyignore from --watch-dir).", root)
		}

		gi := ignoreFiles(opts.watchDir)
		go watchFiles(root, gi, opts.verbose, s, opts.out, restart)
	}

	// NOTE: Once we run the viceroy executable, then it can be stopped by one of
	// two separate mechanisms:
	//
	// 1. File modification
	// 2. Explicit signal (SIGINT, SIGTERM etc).
	//
	// In the case of a signal (e.g. user presses Ctrl-c) the listener logic
	// inside of (*fstexec.Streaming).MonitorSignals() will call
	// (*fstexec.Streaming).Signal(signal os.Signal) to kill the process.
	//
	// In the case of a file modification the viceroy executable needs to first
	// be killed (handled by the watchFiles() function) and then we can stop the
	// signal listeners (handled below by sending a message to cmd.SignalCh).
	//
	// If we don't tell the signal listening channel to close, then the restart
	// of the viceroy executable will cause the user to end up with N number of
	// listeners. This will result in a "os: process already finished" error when
	// we do finally come to stop the `serve` command (e.g. user presses Ctrl-c).
	// How big an issue this is depends on how many file modifications a user
	// makes, because having lots of signal listeners could exhaust resources.
	if err := s.Exec(); err != nil {
		if !strings.Contains(err.Error(), "signal: ") {
			opts.errLog.Add(err)
		}
		e := strings.TrimSpace(err.Error())
		if strings.Contains(e, "interrupt") {
			return fsterr.ErrSignalInterrupt
		}
		if strings.Contains(e, "killed") {
			select {
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
func watchFiles(root string, gi *ignore.GitIgnore, verbose bool, s *fstexec.Streaming, out io.Writer, restart chan<- bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
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
		log.Fatal(err)
	}

	if verbose {
		text.Output(out, "%s", text.BoldYellow("Watching..."))
		text.Break(out)
		text.Output(out, buf.String())
		text.Break(out)
	}

	<-done
}

// ignoreFiles returns the specific ignore rules being respected.
//
// NOTE: We also ignore the .git directory.
func ignoreFiles(watchDir cmd.OptionalString) *ignore.GitIgnore {
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
