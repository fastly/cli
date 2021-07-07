package compute

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
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

	progress := text.NewQuietProgress(out)

	progress.Step("Checking latest viceroy release...")

	latest, err := c.viceroyVersioner.LatestVersion(context.Background())
	if err != nil {
		progress.Fail()

		return errors.RemediationError{
			Inner:       fmt.Errorf("error fetching latest version: %w", err),
			Remediation: errors.NetworkRemediation,
		}
	}

	dir, err := installDir()
	if err != nil {
		return errors.RemediationError{
			Inner:       err,
			Remediation: errors.BugRemediation,
		}
	}

	bin := filepath.Join(dir, c.viceroyVersioner.Name())
	cmd := exec.Command(bin, "--version")

	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		progress.Step("Fetching latest viceroy release...")

		tmp, err := c.viceroyVersioner.Download(context.Background(), latest)
		if err != nil {
			progress.Fail()
			return fmt.Errorf("error downloading latest viceroy release: %w", err)
		}

		if err := os.Rename(tmp, bin); err != nil {
			if err := filesystem.CopyFile(tmp, bin); err != nil {
				progress.Fail()
				return fmt.Errorf("error moving latest viceroy binary in place: %w", err)
			}
		}
	} else {
		progress.Step("Checking installed viceroy version...")

		var installedViceroyVersion string

		viceroyError := errors.RemediationError{
			Inner:       fmt.Errorf("viceroy version not found"),
			Remediation: errors.BugRemediation,
		}

		scanner := bufio.NewScanner(bytes.NewReader(stdoutStderr))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			// version output has the expected format: `viceroy 0.1.0`
			segs := strings.Split(scanner.Text(), " ")

			if len(segs) < 2 {
				return viceroyError
			}

			installedViceroyVersion = segs[1]
			break
		}

		if installedViceroyVersion == "" {
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
			text.Output(out, "Current viceroy version: %s", current)
			text.Output(out, "Latest viceroy version: %s", latest)

			progress.Step("Fetching latest viceroy release...")
			tmp, err := c.viceroyVersioner.Download(context.Background(), latest)
			if err != nil {
				progress.Fail()
				return fmt.Errorf("error downloading latest viceroy release: %w", err)
			}
			defer os.RemoveAll(tmp)

			progress.Step("Replacing viceroy binary...")

			if err := os.Rename(tmp, bin); err != nil {
				if err := filesystem.CopyFile(tmp, bin); err != nil {
					progress.Fail()
					return fmt.Errorf("error moving latest viceroy binary in place: %w", err)
				}
			}
		}
	}

	progress.Done()

	err = c.Local(bin, out)
	if err != nil {
		return err
	}

	return nil
}

func installDir() (string, error) {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "fastly"), nil
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".fastly"), nil
	}
	return "", fmt.Errorf("error locating directory to install viceroy")
}

// Local spawns a subprocess that runs the compiled binary.
func (c *ServeCommand) Local(bin string, out io.Writer) error {
	sig := make(chan os.Signal, 1)

	signals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	signal.Notify(sig, signals...)

	cmd := exec.Command(bin, "bin/main.wasm", "-C", "fastly.toml")
	cmd.Stdout = out
	cmd.Stderr = out

	go func(sig chan os.Signal, cmd *exec.Cmd) {
		<-sig
		signal.Stop(sig)

		err := cmd.Process.Signal(os.Kill)
		if err != nil {
			log.Fatal(err)
		}
	}(sig, cmd)

	err := cmd.Start()
	if err != nil {
		return err
	}

	cmd.Wait()

	return nil
}
