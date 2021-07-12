package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
	cliVersioner   Versioner
	client         api.HTTPClient
	configFilePath string
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, configFilePath string, cliVersioner Versioner, client api.HTTPClient, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("update", "Update the CLI to the latest version")
	c.cliVersioner = cliVersioner
	c.client = client
	c.configFilePath = configFilePath
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	progress := text.NewQuietProgress(out)

	current, latest, shouldUpdate, err := Check(context.Background(), revision.AppVersion, c.cliVersioner)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error checking for latest version: %w", err)
	}

	text.Output(out, "Current version: %s", current)
	text.Output(out, "Latest version: %s", latest)
	if !shouldUpdate {
		text.Output(out, "No update required.")
		return nil
	}

	progress.Step("Fetching latest release...")
	latestPath, err := c.cliVersioner.Download(context.Background(), latest)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		progress.Fail()
		return fmt.Errorf("error downloading latest release: %w", err)
	}
	defer os.RemoveAll(latestPath)

	progress.Step("Replacing binary...")
	currentPath, err := os.Executable()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		progress.Fail()
		return fmt.Errorf("error determining executable path: %w", err)
	}

	currentPath, err = filepath.Abs(currentPath)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		progress.Fail()
		return fmt.Errorf("error determining absolute target path: %w", err)
	}

	if err := os.Rename(latestPath, currentPath); err != nil {
		if err := filesystem.CopyFile(latestPath, currentPath); err != nil {
			c.Globals.ErrLog.Add(err)
			progress.Fail()
			return fmt.Errorf("error moving latest binary in place: %w", err)
		}
	}

	progress.Done()

	text.Success(out, "Updated %s to %s.", currentPath, latest)
	return nil
}
