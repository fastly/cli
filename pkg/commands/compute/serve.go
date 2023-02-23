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
	"github.com/fastly/cli/pkg/check"
	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	ignore "github.com/sabhiram/go-gitignore"
)

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
	addr      string
	debug     bool
	env       cmd.OptionalString
	file      string
	skipBuild bool
	watch     bool
	watchDir  cmd.OptionalString
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
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.skipBuild)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").Action(c.timeout.Set).IntVar(&c.timeout.Value)
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

	c.setBackendsWithDefaultOverrideHostIfMissing(out)

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	bin, err := GetViceroy(spinner, out, c.av, c.Globals)
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

	for {
		err = local(bin, c.file, c.addr, c.env.Value, c.debug, c.watch, c.watchDir, c.Globals.Verbose(), out, c.Globals.ErrLog)
		if err != nil {
			if err != fsterr.ErrViceroyRestart {
				if err == fsterr.ErrSignalInterrupt || err == fsterr.ErrSignalKilled {
					text.Info(out, "Local server stopped")
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

	text.Break(out)

	return nil
}

// setBackendsWithDefaultOverrideHostIfMissing sets an override_host for any
// local_server.backends that is missing that property. The value will only be
// set if the URL defined uses a hostname (e.g. http://127.0.0.1/ won't) so we
// can set the override_host to match the hostname.
func (c *ServeCommand) setBackendsWithDefaultOverrideHostIfMissing(out io.Writer) {
	var missingOverrideHost bool

	for k, backend := range c.Globals.Manifest.File.LocalServer.Backends {
		if backend.OverrideHost == "" {
			if u, err := url.Parse(backend.URL); err == nil {
				segs := strings.Split(u.Host, ":") // avoid parsing IP with port
				if addr := net.ParseIP(segs[0]); addr != nil {
					// we have an IP
				} else {
					if c.Globals.Verbose() {
						text.Info(out, "[local_server.backends.%s] (%s) is configured without an `override_host`. We will use %s as a default to help avoid any unexpected errors. See https://developer.fastly.com/reference/compute/fastly-toml/#local-server for more details.", k, backend.URL, u.Host)
					}
					backend.OverrideHost = u.Host
					c.Globals.Manifest.File.LocalServer.Backends[k] = backend
					missingOverrideHost = true
				}
			}
		}
	}

	if missingOverrideHost && c.Globals.Verbose() {
		text.Break(out)
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
func GetViceroy(spinner text.Spinner, out io.Writer, av github.AssetVersioner, g *global.Data) (bin string, err error) {
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

	var install, checkUpdate bool

	stdoutStderr, err := c.CombinedOutput()
	if err != nil {
		g.ErrLog.Add(err)

		// We presume an error means Viceroy needs to be installed.
		install = true
	}

	viceroy := g.Config.Viceroy
	var latest semver.Version

	// Use latest_version from CLI app config if it's not stale.
	if viceroy.LastChecked != "" && viceroy.LatestVersion != "" && !check.Stale(viceroy.LastChecked, viceroy.TTL) {
		latest, err = semver.Parse(viceroy.LatestVersion)
		if err != nil {
			return bin, err
		}
	}

	// The latest_version value 0.0.0 means the property either has not been set
	// or is now stale and needs to be refreshed.
	if latest.String() == "0.0.0" {
		err := spinner.Start()
		if err != nil {
			return bin, err
		}
		msg := "Checking latest Viceroy release"
		spinner.Message(msg + "...")

		v, err := av.Version()
		if err != nil {
			g.ErrLog.Add(err)

			// When we have an error getting the latest version information for Viceroy
			// and the user doesn't have a pre-existing install of Viceroy, then we're
			// forced to return the error.
			if install {
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return bin, spinErr
				}

				return bin, fsterr.RemediationError{
					Inner:       fmt.Errorf("error fetching latest version: %w", err),
					Remediation: fsterr.NetworkRemediation,
				}
			}

			spinner.StopMessage(msg)
			err = spinner.Stop()
			if err != nil {
				return bin, err
			}
			return bin, nil
		}

		spinner.StopMessage(msg)
		err = spinner.Stop()
		if err != nil {
			return bin, err
		}

		// WARNING: This variable MUST shadow the parent scoped variable.
		latest, err = semver.Parse(v)
		if err != nil {
			return bin, err
		}

		viceroy.LatestVersion = latest.String()
		viceroy.LastChecked = time.Now().Format(time.RFC3339)

		// Before attempting to write the config data back to disk we need to
		// ensure we reassign the modified struct which is a copy (not reference).
		g.Config.Viceroy = viceroy

		err = g.Config.Write(g.Path)
		if err != nil {
			return bin, err
		}

		checkUpdate = true
	}

	if install {
		err := installViceroy(spinner, av, bin)
		if err != nil {
			g.ErrLog.Add(err)
			return bin, err
		}
	} else if checkUpdate {
		version := strings.TrimSpace(string(stdoutStderr))
		err := updateViceroy(spinner, version, out, av, latest, bin)
		if err != nil {
			g.ErrLog.Add(err)
			return bin, err
		}
	}

	err = setBinPerms(bin)
	if err != nil {
		g.ErrLog.Add(err)
		return bin, err
	}
	return bin, nil
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

// installViceroy downloads the latest release from GitHub.
func installViceroy(spinner text.Spinner, av github.AssetVersioner, bin string) error {
	err := spinner.Start()
	if err != nil {
		return err
	}
	msg := "Fetching latest Viceroy release"
	spinner.Message(msg + "...")

	tmpBin, err := av.Download()
	if err != nil {
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}
		return fmt.Errorf("error downloading latest Viceroy release: %w", err)
	}
	defer os.RemoveAll(tmpBin)

	if err := os.Rename(tmpBin, bin); err != nil {
		if err := filesystem.CopyFile(tmpBin, bin); err != nil {
			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return spinErr
			}
			return fmt.Errorf("error moving latest Viceroy binary in place: %w", err)
		}
	}

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return err
	}
	return nil
}

// updateViceroy checks if the currently installed version is out-of-date and
// downloads the latest release from GitHub.
func updateViceroy(
	spinner text.Spinner,
	version string,
	out io.Writer,
	av github.AssetVersioner,
	latest semver.Version,
	bin string,
) error {
	err := spinner.Start()
	if err != nil {
		return err
	}
	msg := "Checking installed Viceroy version"
	spinner.Message(msg + "...")

	viceroyError := fsterr.RemediationError{
		Inner:       fmt.Errorf("a Viceroy version was not found"),
		Remediation: fsterr.BugRemediation,
	}

	// version output has the expected format: `viceroy 0.1.0`
	segs := strings.Split(version, " ")

	if len(segs) < 2 {
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}
		return viceroyError
	}

	installedViceroyVersion := segs[1]
	if installedViceroyVersion == "" {
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}
		return viceroyError
	}

	current, err := semver.Parse(installedViceroyVersion)
	if err != nil {
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}

		return fsterr.RemediationError{
			Inner:       fmt.Errorf("error reading current version: %w", err),
			Remediation: fsterr.BugRemediation,
		}
	}

	if latest.GT(current) {
		text.Break(out)
		text.Break(out)
		text.Output(out, "Current Viceroy version: %s", current)
		text.Output(out, "Latest Viceroy version: %s", latest)
		text.Break(out)

		err := spinner.Start()
		if err != nil {
			return err
		}
		msg := "Fetching latest Viceroy release"
		spinner.Message(msg + "...")

		tmpBin, err := av.Download()
		if err != nil {
			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return spinErr
			}
			return fmt.Errorf("error downloading latest Viceroy release: %w", err)
		}
		defer os.RemoveAll(tmpBin)

		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = "Replacing Viceroy binary"
		spinner.Message(msg + "...")

		if err := os.Rename(tmpBin, bin); err != nil {
			if err := filesystem.CopyFile(tmpBin, bin); err != nil {
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return spinErr
				}
				return fmt.Errorf("error moving latest Viceroy binary in place: %w", err)
			}
		}
	}

	return nil
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

// local spawns a subprocess that runs the compiled binary.
func local(bin, file, addr, env string, debug, watch bool, watchDir cmd.OptionalString, verbose bool, out io.Writer, errLog fsterr.LogInterface) error {
	if env != "" {
		env = "." + env
	}

	wd, err := os.Getwd()
	if err != nil {
		errLog.Add(err)
		return err
	}

	manifestPath := filepath.Join(wd, fmt.Sprintf("fastly%s.toml", env))
	args := []string{"-C", manifestPath, "--addr", addr, file}

	if debug {
		args = append(args, "--debug")
	}

	if verbose {
		text.Break(out)
		text.Output(out, "Wasm file: %s", file)
		text.Output(out, "Manifest: %s", manifestPath)
	}

	s := &fstexec.Streaming{
		Args:        args,
		Command:     bin,
		Env:         os.Environ(),
		ForceOutput: true,
		Output:      out,
		SignalCh:    make(chan os.Signal, 1),
	}
	s.MonitorSignals()

	restart := make(chan bool)
	if watch {
		root := "."
		if watchDir.WasSet {
			root = watchDir.Value
		}

		if verbose {
			text.Info(out, "Watching files for changes (using --watch-dir=%s). To ignore certain files, define patterns within a .fastlyignore config file (uses .fastlyignore from --watch-dir).", root)
		}

		gi := ignoreFiles(watchDir)
		go watchFiles(root, gi, verbose, s, out, restart)
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
		errLog.Add(err)
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
			root = root + "/"
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
