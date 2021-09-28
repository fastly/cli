package serviceversion

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/time"
	"github.com/fastly/go-fastly/v5/fastly"
)

// ListCommand calls the Fastly API to list services.
type ListCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ListVersionsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Fastly service versions")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, source, out)
	}
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	versions, err := c.Globals.Client.ListVersions(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("NUMBER", "ACTIVE", "LAST EDITED (UTC)")
		for _, version := range versions {
			tw.AddLine(version.Number, version.Active, version.UpdatedAt.UTC().Format(time.Format))
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Versions: %d\n", len(versions))
	for i, version := range versions {
		fmt.Fprintf(out, "\tVersion %d/%d\n", i+1, len(versions))
		text.PrintVersion(out, "\t\t", version)
	}
	fmt.Fprintln(out)

	return nil
}
