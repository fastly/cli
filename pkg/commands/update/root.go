package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/blang/semver"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("update", "Update the CLI to the latest version")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	var (
		current, latest semver.Version
		shouldUpdate    bool
	)

	err = spinner.Process("Updating versioning information", func(_ *text.SpinnerWrapper) error {
		current, latest, shouldUpdate = Check(revision.AppVersion, c.Globals.Versioners.CLI)
		return nil
	})
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

	var downloadedBin string
	err = spinner.Process("Fetching latest release", func(_ *text.SpinnerWrapper) error {
		downloadedBin, err = c.Globals.Versioners.CLI.DownloadLatest()
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Current CLI version": current,
				"Latest CLI version":  latest,
			})
			return fmt.Errorf("error downloading latest release: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	defer os.RemoveAll(downloadedBin)

	var currentBin string
	err = spinner.Process("Replacing binary", func(_ *text.SpinnerWrapper) error {
		execPath, err := os.Executable()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error determining executable path: %w", err)
		}

		currentBin, err = filepath.Abs(execPath)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Executable path": execPath,
			})
			return fmt.Errorf("error determining absolute target path: %w", err)
		}

		// Windows does not permit replacing a running executable, however it will
		// permit it if you first move the original executable. So we first move the
		// running executable to a new location, then we move the executable that we
		// downloaded to the same location as the original.
		// I've also tested this approach on nix systems and it works fine.
		//
		// Reference:
		// https://github.com/golang/go/issues/21997#issuecomment-331744930

		backup := currentBin + ".bak"
		if err := os.Rename(currentBin, backup); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Executable (source)":      downloadedBin,
				"Executable (destination)": currentBin,
			})
			return fmt.Errorf("error moving the current executable: %w", err)
		}

		if err = os.Remove(backup); err != nil {
			c.Globals.ErrLog.Add(err)
		}

		// Move the downloaded binary to the same location as the current executable.
		if err := os.Rename(downloadedBin, currentBin); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Executable (source)":      downloadedBin,
				"Executable (destination)": currentBin,
			})
			renameErr := err

			// Failing that we'll try to io.Copy downloaded binary to the current binary.
			if err := filesystem.CopyFile(downloadedBin, currentBin); err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]any{
					"Executable (source)":      downloadedBin,
					"Executable (destination)": currentBin,
				})
				return fmt.Errorf("error 'copying' latest binary in place: %w (following an error 'moving': %w)", err, renameErr)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	text.Success(out, "\nUpdated %s to %s.", currentBin, latest)
	return nil
}
