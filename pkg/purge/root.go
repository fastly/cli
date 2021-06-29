package purge

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base

	all      bool
	file     string
	key      string
	manifest manifest.Data
	soft     bool
	url      string
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.CmdClause = parent.Command("purge", "Remove an object from the Fastly cache")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Optional flags
	c.CmdClause.Flag("all", "Purge everything from a service").BoolVar(&c.all)
	c.CmdClause.Flag("file", "Purge a service with a line separated list of Surrogate Keys").StringVar(&c.file)
	c.CmdClause.Flag("key", "Purge a service of items tagged with a Surrogate Key").StringVar(&c.key)
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("soft", "A 'soft' purge marks the affected object as stale rather than making it inaccessible").BoolVar(&c.soft)
	c.CmdClause.Flag("url", "Purge an individual URL").StringVar(&c.url)

	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(in io.Reader, out io.Writer) error {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	// The URL purge API call doesn't require a Service ID.
	var serviceID string
	var source manifest.Source
	if c.url == "" {
		serviceID, source = c.manifest.ServiceID()
		if source == manifest.SourceUndefined {
			return errors.ErrNoServiceID
		}
	}

	if c.all {
		//
		return nil
	}

	if c.file != "" {
		//
		fmt.Printf("\n\n%+v\n\n", serviceID)
		return nil
	}

	if c.key != "" {
		//
		return nil
	}

	if c.url != "" {
		//
		return nil
	}

	return nil
}
