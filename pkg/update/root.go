package update

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/version"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	common.Base
	versioner Versioner
	client    api.HTTPClient
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent common.Registerer, v Versioner, client api.HTTPClient) *RootCommand {
	var c RootCommand
	c.CmdClause = parent.Command("update", "Update the CLI to the latest version")
	c.versioner = v
	c.client = client
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	progress := text.NewQuietProgress(out)

	current, latest, shouldUpdate, err := Check(context.Background(), version.AppVersion, c.versioner)
	if err != nil {
		return fmt.Errorf("error checking for latest version: %w", err)
	}

	text.Output(out, "Current version: %s", current)
	text.Output(out, "Latest version: %s", latest)
	if !shouldUpdate {
		text.Output(out, "No update required.")
		return nil
	}

	progress.Step("Fetching latest release...")
	latestPath, err := c.versioner.Download(context.Background(), latest)
	if err != nil {
		progress.Fail()
		return fmt.Errorf("error downloading latest release: %w", err)
	}
	defer os.RemoveAll(latestPath)

	progress.Step("Replacing binary...")
	currentPath, err := os.Executable()
	if err != nil {
		progress.Fail()
		return fmt.Errorf("error determining executable path: %w", err)
	}

	currentPath, err = filepath.Abs(currentPath)
	if err != nil {
		progress.Fail()
		return fmt.Errorf("error determining absolute target path: %w", err)
	}

	if err := os.Rename(latestPath, currentPath); err != nil {
		if err := filesystem.CopyFile(latestPath, currentPath); err != nil {
			progress.Fail()
			return fmt.Errorf("error moving latest binary in place: %w", err)
		}
	}

	progress.Done()

	text.Success(out, "Updated %s to %s.", currentPath, latest)
	return nil
}
