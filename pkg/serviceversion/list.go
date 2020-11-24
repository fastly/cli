package serviceversion

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v2/fastly"
)

// ListCommand calls the Fastly API to list services.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListVersionsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Fastly service versions")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	versions, err := c.Globals.Client.ListVersions(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("NUMBER", "ACTIVE", "LAST EDITED (UTC)")
		for _, version := range versions {
			tw.AddLine(version.Number, version.Active, version.UpdatedAt.UTC().Format(common.TimeFormat))
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
