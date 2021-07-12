package compute

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/update"
)

// ServeCommand produces and runs an artifact from files on the local disk.
type ServeCommand struct {
	cmd.Base
	manifest         manifest.Data
	build            *BuildCommand
	viceroyVersioner update.Versioner

	// Build fields
	name       cmd.OptionalString
	lang       cmd.OptionalString
	includeSrc cmd.OptionalBool
	force      cmd.OptionalBool

	// Viceroy fields
	env cmd.OptionalString
}

// NewServeCommand returns a usable command registered under the parent.
func NewServeCommand(parent cmd.Registerer, globals *config.Data, build *BuildCommand, viceroyVersioner update.Versioner) *ServeCommand {
	var c ServeCommand

	c.build = build
	c.viceroyVersioner = viceroyVersioner

	c.Globals = globals
	c.CmdClause = parent.Command("serve", "Build and run a Compute@Edge package locally")

	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Build flags
	c.CmdClause.Flag("name", "Package name").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("force", "Skip verification steps and force build").Action(c.force.Set).BoolVar(&c.force.Value)

	// Viceroy flags
	c.CmdClause.Flag("env", "The environment configuration to use (e.g. stage)").Action(c.env.Set).StringVar(&c.env.Value)

	return &c
}

// Exec implements the command interface.
func (c *ServeCommand) Exec(in io.Reader, out io.Writer) (err error) {
	// Reset the fields on the BuildCommand based on ServeCommand values.
	if c.name.WasSet {
		c.build.PackageName = c.name.Value
	}
	if c.lang.WasSet {
		c.build.Lang = c.lang.Value
	}
	if c.includeSrc.WasSet {
		c.build.IncludeSrc = c.includeSrc.Value
	}
	if c.force.WasSet {
		c.build.Force = c.force.Value
	}

	err = c.build.Exec(in, out)
	if err != nil {
		return err
	}

	text.Break(out)

	var progress text.Progress
	if c.Globals.Verbose() {
		progress = text.NewVerboseProgress(out)
	} else {
		progress = text.NewQuietProgress(out)
	}

	bin, err := getViceroy(c.Globals.Verbose(), progress, out, c.viceroyVersioner)
	if err != nil {
		return err
	}

	err = local(bin, out, c.env.Value, c.Globals.Verbose())
	if err != nil {
		return err
	}

	return nil
}

// getViceroy returns the path to the installed binary.
//
// NOTE: if Viceroy is installed then it is updated, otherwise download the
// latest version and install it in the same directory as the application
// configuration data.
func getViceroy(verbose bool, progress text.Progress, out io.Writer, versioner update.Versioner) (string, error) {
	progress.Step("Checking latest Viceroy release...")

	latest, err := versioner.LatestVersion(context.Background())
	if err != nil {
		progress.Fail()

		return "", errors.RemediationError{
			Inner:       fmt.Errorf("error fetching latest version: %w", err),
			Remediation: errors.NetworkRemediation,
		}
	}

	asset := fmt.Sprintf("%s_%s-%s", versioner.Binary(), runtime.GOOS, runtime.GOARCH)
	versioner.SetAsset(asset)

	bin := filepath.Join(InstallDir, versioner.Name())
	if verbose {
		text.Info(out, "Viceroy will be installed to: %s", bin)
		text.Break(out)
	}

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	cmd := exec.Command(bin, "--version")

	stdoutStderr, err := cmd.CombinedOutput()

	if err != nil {
		// We presume an error executing `viceroy --version` means it isn't installed.
		//
		// NOTE: we can't use exec.LookPath("viceroy") because PATH is unreliable
		// across OS platforms but also we actually install viceroy in the same
		// location as the application configuration, which means it wouldn't be
		// found looking up by the PATH env var.
		err := installViceroy(progress, versioner, latest, bin)
		if err != nil {
			return "", err
		}
	} else {
		version := strings.TrimSpace(string(stdoutStderr))
		err := updateViceroy(progress, version, out, versioner, latest, bin)
		if err != nil {
			return "", err
		}
	}

	progress.Done()

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
	text.Break(out)
	text.Info(out, "Current Viceroy version: %s", current)
	text.Break(out)

	if latest.GT(current) {
		text.Output(out, "Latest Viceroy version: %s", latest)

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

// local spawns a subprocess that runs the compiled binary.
func local(bin string, out io.Writer, env string, verbose bool) error {
	if env != "" {
		env = "." + env
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	manifest := filepath.Join(wd, fmt.Sprintf("fastly%s.toml", env))
	args := []string{"bin/main.wasm", "-C", manifest}

	if verbose {
		text.Output(out, "Manifest: %s", manifest)
	}

	cmd := fstexec.Streaming{
		Command: bin,
		Args:    args,
		Env:     os.Environ(),
		Output:  out,
	}
	cmd.MonitorSignals()

	if err := cmd.Exec(); err != nil {
		return err
	}

	return nil
}
