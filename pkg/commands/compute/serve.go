package compute

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bep/debounce"
	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
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
	includeSrc       cmd.OptionalBool
	lang             cmd.OptionalString
	name             cmd.OptionalString
	skipVerification cmd.OptionalBool
	timeout          cmd.OptionalInt

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
		return errors.ErrIncompatibleServeFlags
	}

	if !c.skipBuild {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
	}

	progress := text.NewProgress(out, c.Globals.Verbose())

	bin, err := getViceroy(progress, out, c.viceroyVersioner)
	if err != nil {
		return err
	}

	progress.Step("Running local server...")
	progress.Done()

	for {
		err = local(bin, c.file, c.addr, c.env.Value, c.watch, c.Globals.Verbose(), progress, out)
		if err != nil {
			if err != errors.ErrViceroyRestart {
				if err == errors.ErrSignalInterrupt || err == errors.ErrSignalKilled {
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
	if c.includeSrc.WasSet {
		c.build.IncludeSrc = c.includeSrc.Value
	}
	if c.lang.WasSet {
		c.build.Lang = c.lang.Value
	}
	if c.name.WasSet {
		c.build.PackageName = c.name.Value
	}
	if c.skipVerification.WasSet {
		c.build.SkipVerification = c.skipVerification.Value
	}
	if c.timeout.WasSet {
		c.build.Timeout = c.timeout.Value
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
func getViceroy(progress text.Progress, out io.Writer, versioner update.Versioner) (bin string, err error) {
	progress.Step("Checking latest Viceroy release...")

	defer func(progress text.Progress) {
		if err != nil {
			progress.Fail()
		}
	}(progress)

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

	var install bool

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		// We presume an error means Viceroy needs to be installed.
		install = true
	}

	latest, err := versioner.LatestVersion(context.Background())
	if err != nil {
		// When we have an error getting the latest version information for Viceroy
		// and the user doesn't have a pre-existing install of Viceroy, then we're
		// forced to return the error.
		if install {
			return bin, errors.RemediationError{
				Inner:       fmt.Errorf("error fetching latest version: %w", err),
				Remediation: errors.NetworkRemediation,
			}
		}
		return bin, nil
	}

	archiveFormat := ".tar.gz"
	asset := fmt.Sprintf(update.DefaultAssetFormat, versioner.BinaryName(), latest, runtime.GOOS, runtime.GOARCH, archiveFormat)
	versioner.SetAsset(asset)

	if install {
		err := installViceroy(progress, versioner, latest, bin)
		if err != nil {
			return bin, err
		}
	} else {
		version := strings.TrimSpace(string(stdoutStderr))
		err := updateViceroy(progress, version, out, versioner, latest, bin)
		if err != nil {
			return bin, err
		}
	}

	err = setBinPerms(bin)
	if err != nil {
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

	viceroyError := errors.RemediationError{
		Inner:       fmt.Errorf("a Viceroy version was not found"),
		Remediation: errors.BugRemediation,
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

		return errors.RemediationError{
			Inner:       fmt.Errorf("error reading current version: %w", err),
			Remediation: errors.BugRemediation,
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

// local spawns a subprocess that runs the compiled binary.
func local(bin, file, addr, env string, watch, verbose bool, progress text.Progress, out io.Writer) error {
	if env != "" {
		env = "." + env
	}

	wd, err := os.Getwd()
	if err != nil {
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
		go watchFiles(cmd, out, restart)
	}

	if err := cmd.Exec(); err != nil {
		e := strings.TrimSpace(err.Error())
		if strings.Contains(e, "interrupt") {
			return errors.ErrSignalInterrupt
		}
		if strings.Contains(e, "killed") {
			select {
			case <-restart:
				// NOTE: If we don't tell the signal monitoring channel to close, then
				// when we restart the viceroy binary we'll end up with N number of
				// listeners, and this will result in a "os: process already finished"
				// error when we do finally come to stop the serve command using the
				// Ctrl-C key, depending on how often the user makes file modifications
				// this could end up exhausting resources, so best to do a clean-up.
				// cmd.SignalCh <- syscall.SIGTERM
				return errors.ErrViceroyRestart
			case <-time.After(1 * time.Second):
				return errors.ErrSignalKilled
			}
		}
		return err
	}

	return nil
}

// watchFiles watches the 'src' directory and restarts the viceroy executable
// when changes are detected.
func watchFiles(cmd *fstexec.Streaming, out io.Writer, restart chan<- bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	debounced := debounce.New(1 * time.Second)
	eventHandler := func() {
		text.Info(out, "File system modified: restarting local server")
		text.Break(out)

		// NOTE: If we don't force watcher.Close() by pushing true into done
		// channel, then after one restart of the viceroy binary we'll discover we
		// get the error: "os: process already finished" which happens because the
		// watchFiles function is run in a goroutine and so it keeps running with
		// its copy of the fstexec.Streaming instance, and that is a process that
		// has indeed been 'killed'.
		done <- true

		fmt.Printf("\nKill process: %+v\n", cmd.Process.Pid)

		// NOTE: We can't just send `true` to the restart channel (which will not
		// only cause the signal listener to be closed but will also initiate the
		// process to be killed) as doing so will cause a deadlock. We need to kill
		// the process here to cause the `cmd.Exec()` from within `local()` to
		// finish and subsequently the termination error to be handled, where we'll
		// then be able to close the signal listener and return a
		// ErrViceroyRestart error to trigger the restart logic.
		//
		// The downside here is that we've already killed the viceroy process and
		// so when we stop the signal listener, itself will try to kill the process
		// and discover it has already been killed and return an error to say:
		// `os: process already finished`.
		err := cmd.Signal(os.Kill)
		if err != nil {
			log.Fatal(err)
		}

		restart <- true
	}

	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				debounced(eventHandler)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				text.Output(out, "error event while watching files: %v", err)
			}
		}
	}()

	err = watcher.Add("src")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
