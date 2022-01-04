package compute

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
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
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fsnotify/fsnotify"
)

// ServeCommand produces and runs an artifact from files on the local disk.
type ServeCommand struct {
	cmd.Base
	manifest         manifest.Data
	build            *BuildCommand
	viceroyVersioner update.Versioner

	// Build fields
	acceptCustomBuild cmd.OptionalBool
	includeSrc        cmd.OptionalBool
	lang              cmd.OptionalString
	name              cmd.OptionalString
	skipVerification  cmd.OptionalBool
	timeout           cmd.OptionalInt

	// Serve fields
	addr      string
	env       cmd.OptionalString
	file      string
	skipBuild bool
	watch     bool
}

// NewServeCommand returns a usable command registered under the parent.
func NewServeCommand(parent cmd.Registerer, globals *config.Data, build *BuildCommand, viceroyVersioner update.Versioner, data manifest.Data) *ServeCommand {
	var c ServeCommand

	c.build = build
	c.viceroyVersioner = viceroyVersioner

	c.Globals = globals
	c.CmdClause = parent.Command("serve", "Build and run a Compute@Edge package locally")
	c.manifest = data

	c.CmdClause.Flag("accept-custom-build", "Do not prompt when project manifest defines [scripts.build]").Action(c.acceptCustomBuild.Set).BoolVar(&c.acceptCustomBuild.Value)
	c.CmdClause.Flag("addr", "The IPv4 address and port to listen on").Default("127.0.0.1:7676").StringVar(&c.addr)
	c.CmdClause.Flag("env", "The environment configuration to use (e.g. stage)").Action(c.env.Set).StringVar(&c.env.Value)
	c.CmdClause.Flag("file", "The Wasm file to run").Default("bin/main.wasm").StringVar(&c.file)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("name", "Package name").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.skipBuild)
	c.CmdClause.Flag("skip-verification", "Skip verification steps and force build").Action(c.skipVerification.Set).BoolVar(&c.skipVerification.Value)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").Action(c.timeout.Set).IntVar(&c.timeout.Value)
	c.CmdClause.Flag("watch", "Watch for file changes, then rebuild project and restart local server").BoolVar(&c.watch)

	return &c
}

// Exec implements the command interface.
func (c *ServeCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if c.skipBuild && c.watch {
		return fsterr.ErrIncompatibleServeFlags
	}

	if !c.skipBuild {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
	}

	progress := text.ResetProgress(out, c.Globals.Verbose())

	bin, err := getViceroy(progress, out, c.viceroyVersioner, c.Globals)
	if err != nil {
		return err
	}

	progress.Step("Running local server...")
	progress.Done()

	srcDir := sourceDirectory(c.lang, c.manifest.File.Language, c.watch, out)

	for {
		err = local(bin, srcDir, c.file, c.addr, c.env.Value, c.watch, c.Globals.Verbose(), progress, out, c.Globals.ErrLog)
		if err != nil {
			if err != fsterr.ErrViceroyRestart {
				if err == fsterr.ErrSignalInterrupt || err == fsterr.ErrSignalKilled {
					text.Info(out, "Local server stopped")
					return nil
				}
				return err
			}

			// Before restarting Viceroy we should rebuild.
			err = c.Build(in, out)
			if err != nil {
				return err
			}
		}
	}
}

// Build constructs and executes the build logic.
func (c *ServeCommand) Build(in io.Reader, out io.Writer) error {
	// Reset the fields on the BuildCommand based on ServeCommand values.
	if c.acceptCustomBuild.WasSet {
		c.build.Flags.AcceptCustomBuild = c.acceptCustomBuild.Value
	}
	if c.includeSrc.WasSet {
		c.build.Flags.IncludeSrc = c.includeSrc.Value
	}
	if c.lang.WasSet {
		c.build.Flags.Lang = c.lang.Value
	}
	if c.name.WasSet {
		c.build.Flags.PackageName = c.name.Value
	}
	if c.skipVerification.WasSet {
		c.build.Flags.SkipVerification = c.skipVerification.Value
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

// getViceroy returns the path to the installed binary.
//
// NOTE: if Viceroy is installed then it is updated, otherwise download the
// latest version and install it in the same directory as the application
// configuration data.
//
// In the case of a network failure we fallback to the latest installed version of the
// Viceroy binary as long as one is installed and has the correct permissions.
func getViceroy(progress text.Progress, out io.Writer, versioner update.Versioner, cfg *config.Data) (bin string, err error) {
	defer func() {
		if err != nil {
			progress.Fail()
		}
	}()

	bin = filepath.Join(InstallDir, versioner.Binary())

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
	cmd := exec.Command(bin, "--version")

	var install, checkUpdate bool

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		cfg.ErrLog.Add(err)

		// We presume an error means Viceroy needs to be installed.
		install = true
	}

	viceroy := cfg.File.Viceroy
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
		progress.Step("Checking latest Viceroy release...")

		latest, err = versioner.LatestVersion(context.Background())
		if err != nil {
			cfg.ErrLog.Add(err)

			// When we have an error getting the latest version information for Viceroy
			// and the user doesn't have a pre-existing install of Viceroy, then we're
			// forced to return the error.
			if install {
				return bin, fsterr.RemediationError{
					Inner:       fmt.Errorf("error fetching latest version: %w", err),
					Remediation: fsterr.NetworkRemediation,
				}
			}
			return bin, nil
		}

		viceroy.LatestVersion = latest.String()
		viceroy.LastChecked = time.Now().Format(time.RFC3339)

		// Before attempting to write the config data back to disk we need to
		// ensure we reassign the modified struct which is a copy (not reference).
		cfg.File.Viceroy = viceroy

		err := cfg.File.Write(cfg.Path)
		if err != nil {
			return bin, err
		}

		checkUpdate = true
	}

	archiveFormat := ".tar.gz"
	asset := fmt.Sprintf(update.DefaultAssetFormat, versioner.BinaryName(), latest, runtime.GOOS, runtime.GOARCH, archiveFormat)
	versioner.SetAsset(asset)

	if install {
		err := installViceroy(progress, versioner, latest, bin)
		if err != nil {
			cfg.ErrLog.Add(err)
			return bin, err
		}
	} else if checkUpdate {
		version := strings.TrimSpace(string(stdoutStderr))
		err := updateViceroy(progress, version, out, versioner, latest, bin)
		if err != nil {
			cfg.ErrLog.Add(err)
			return bin, err
		}
	}

	err = setBinPerms(bin)
	if err != nil {
		cfg.ErrLog.Add(err)
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
func installViceroy(progress text.Progress, versioner update.Versioner, latest semver.Version, bin string) error {
	progress.Step("Fetching latest Viceroy release...")

	tmp, err := versioner.Download(context.Background(), latest)
	if err != nil {
		progress.Fail()
		return fmt.Errorf("error downloading latest Viceroy release: %w", err)
	}
	defer os.RemoveAll(tmp)

	if err := os.Rename(tmp, bin); err != nil {
		if err := filesystem.CopyFile(tmp, bin); err != nil {
			progress.Fail()
			return fmt.Errorf("error moving latest Viceroy binary in place: %w", err)
		}
	}

	return nil
}

// updateViceroy checks if the currently installed version is out-of-date and
// downloads the latest release from GitHub.
func updateViceroy(progress text.Progress, version string, out io.Writer, versioner update.Versioner, latest semver.Version, bin string) error {
	progress.Step("Checking installed Viceroy version...")

	var installedViceroyVersion string

	viceroyError := fsterr.RemediationError{
		Inner:       fmt.Errorf("a Viceroy version was not found"),
		Remediation: fsterr.BugRemediation,
	}

	// version output has the expected format: `viceroy 0.1.0`
	segs := strings.Split(version, " ")

	if len(segs) < 2 {
		progress.Fail()
		return viceroyError
	}

	installedViceroyVersion = segs[1]

	if installedViceroyVersion == "" {
		progress.Fail()
		return viceroyError
	}

	current, err := semver.Parse(installedViceroyVersion)
	if err != nil {
		progress.Fail()

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

		tmp, err := versioner.Download(context.Background(), latest)
		progress.Step("Fetching latest Viceroy release...")
		if err != nil {
			progress.Fail()
			return fmt.Errorf("error downloading latest Viceroy release: %w", err)
		}
		defer os.RemoveAll(tmp)

		progress.Step("Replacing Viceroy binary...")

		if err := os.Rename(tmp, bin); err != nil {
			if err := filesystem.CopyFile(tmp, bin); err != nil {
				progress.Fail()
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
	err := os.Chmod(bin, 0777)
	if err != nil {
		return fmt.Errorf("error setting executable permissions on Viceroy binary: %w", err)
	}
	return nil
}

// sourceDirectory identifies the source code directory for the given language.
func sourceDirectory(flag cmd.OptionalString, lang string, watch bool, out io.Writer) string {
	if flag.WasSet {
		lang = flag.Value
	}
	lang = strings.ToLower(strings.TrimSpace(lang))

	defaultDir := "src"

	switch lang {
	case "assemblyscript":
		return ASSourceDirectory
	case "javascript":
		return JSSourceDirectory
	case "rust":
		return RustSourceDirectory
	}
	if watch {
		text.Info(out, "The --watch flag defaults to watching file modifications in a ./src directory.")
	}
	return defaultDir
}

// local spawns a subprocess that runs the compiled binary.
func local(bin, srcDir, file, addr, env string, watch, verbose bool, progress text.Progress, out io.Writer, errLog fsterr.LogInterface) error {
	if env != "" {
		env = "." + env
	}

	wd, err := os.Getwd()
	if err != nil {
		errLog.Add(err)
		return err
	}

	manifest := filepath.Join(wd, fmt.Sprintf("fastly%s.toml", env))
	args := []string{"-C", manifest, "--addr", addr, file}

	if verbose {
		text.Output(out, "Wasm file: %s", file)
		text.Output(out, "Manifest: %s", manifest)
	}

	cmd := &fstexec.Streaming{
		Args:     args,
		Command:  bin,
		Env:      os.Environ(),
		Output:   out,
		SignalCh: make(chan os.Signal, 1),
	}
	cmd.MonitorSignals()

	text.Break(out)

	restart := make(chan bool)
	if watch {
		go watchFiles(srcDir, cmd, out, restart)
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
	if err := cmd.Exec(); err != nil {
		errLog.Add(err)
		e := strings.TrimSpace(err.Error())
		if strings.Contains(e, "interrupt") {
			return fsterr.ErrSignalInterrupt
		}
		if strings.Contains(e, "killed") {
			select {
			case <-restart:
				cmd.SignalCh <- syscall.SIGTERM
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
func watchFiles(dir string, cmd *fstexec.Streaming, out io.Writer, restart chan<- bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	debounced := debounce.New(1 * time.Second)
	eventHandler := func(modifiedFile string, op fsnotify.Op) {
		action := "modified"
		switch op {
		case fsnotify.Create:
			action = "created"
		case fsnotify.Remove:
			action = "removed"
		case fsnotify.Rename:
			action = "renamed"
		}

		text.Info(out, "Restarting: %s has been %s", modifiedFile, action)
		text.Break(out)

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
		err := cmd.Signal(os.Kill)
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

	filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}
		if entry.IsDir() {
			err = watcher.Add(path)
			if err != nil {
				log.Fatal(err)
			}
		}
		return nil
	})

	text.Info(out, "Watching ./%s for changes", dir)
	text.Break(out)
	<-done
}
