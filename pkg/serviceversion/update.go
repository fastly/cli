package serviceversion

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// UpdateCommand calls the Fastly API to update a service version.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data
	input    fastly.UpdateVersionInput

	comment common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("update", "Update a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of version you wish to update").Required().IntVar(&c.input.ServiceVersion)

	// TODO(integralist):
	// Make 'comment' field mandatory once we roll out a new release of Go-Fastly
	// which will hopefully have better/more correct consistency as far as which
	// fields are supposed to be optional and which should be 'required'.
	//
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}

	c.input.ServiceID = serviceID

	if !c.comment.WasSet {
		return fmt.Errorf("error parsing arguments: required flag --comment not provided")
	}

	if c.comment.WasSet {
		c.input.Comment = fastly.String(c.comment.Value)
	}

	v, err := c.Globals.Client.UpdateVersion(&c.input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated service %s version %d", v.ServiceID, c.input.ServiceVersion)
	return nil
}
