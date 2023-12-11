package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base

	versionToInstall string
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("install", "Install the specified version of the CLI")
	c.CmdClause.Arg("version", "CLI release version to install (e.g. 10.8.0)").Required().StringVar(&c.versionToInstall)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	var downloadedBin string
	err = spinner.Process(fmt.Sprintf("Fetching release %s", c.versionToInstall), func(_ *text.SpinnerWrapper) error {
		downloadedBin, err = c.Globals.Versioners.CLI.DownloadVersion(c.versionToInstall)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"CLI version to install": c.versionToInstall,
			})
			return fmt.Errorf("error downloading release version %s: %w", c.versionToInstall, err)
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

	text.Success(out, "\nInstalled version %s.", c.versionToInstall)
	return nil
}
