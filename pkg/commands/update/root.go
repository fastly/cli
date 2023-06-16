package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/revision"
	fstruntime "github.com/fastly/cli/pkg/runtime"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	av             github.AssetVersioner
	configFilePath string
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, configFilePath string, av github.AssetVersioner, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("update", "Update the CLI to the latest version")
	c.av = av
	c.configFilePath = configFilePath
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	msg := "Updating versioning information"
	spinner.Message(msg + "...")

	current, latest, shouldUpdate := Check(revision.AppVersion, c.av)

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return err
	}

	text.Break(out)
	text.Output(out, "Current version: %s", current)
	text.Output(out, "Latest version: %s", latest)
	text.Break(out)

	if !shouldUpdate {
		text.Output(out, "No update required.")
		return nil
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	msg = "Fetching latest release"
	spinner.Message(msg + "...")

	tmpBin, err := c.av.DownloadLatest()
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Current CLI version": current,
			"Latest CLI version":  latest,
		})

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}

		return fmt.Errorf("error downloading latest release: %w", err)
	}
	defer os.RemoveAll(tmpBin)

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return err
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	msg = "Replacing binary"
	spinner.Message(msg + "...")

	execPath, err := os.Executable()
	if err != nil {
		c.Globals.ErrLog.Add(err)

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}

		return fmt.Errorf("error determining executable path: %w", err)
	}

	currentPath, err := filepath.Abs(execPath)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Executable path": execPath,
		})

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}

		return fmt.Errorf("error determining absolute target path: %w", err)
	}

	// Windows does not permit removing a running executable, however it will
	// permit renaming it! So we first rename the running executable and then we
	// move the executable that we downloaded to the same location as the
	// original executable (which is allowed since we first renamed the running
	// executable).
	//
	// Reference:
	// https://github.com/golang/go/issues/21997#issuecomment-331744930
	if fstruntime.Windows {
		if err := os.Rename(execPath, execPath+"~"); err != nil {
			c.Globals.ErrLog.Add(err)
			if err = os.Remove(execPath + "~"); err != nil {
				c.Globals.ErrLog.Add(err)
			}
		}
	}

	if err := os.Rename(tmpBin, currentPath); err != nil {
		if err := filesystem.CopyFile(tmpBin, currentPath); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Executable (source)":      tmpBin,
				"Executable (destination)": currentPath,
			})

			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return spinErr
			}

			return fmt.Errorf("error moving latest binary in place: %w", err)
		}
	}

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return err
	}

	text.Success(out, "Updated %s to %s.", currentPath, latest)
	return nil
}
