package aclentry

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Delete an ACL entry from a specified ACL").Alias("remove")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)
	c.CmdClause.Flag("id", "Alphanumeric string identifying an ACL Entry").Required().StringVar(&c.id)

	// Optional flags
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	cmd.Base

	aclID    string
	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if c.Globals.Verbose() {
		cmd.DisplayServiceID(serviceID, source, out)
	}
	if source == manifest.SourceUndefined {
		err := errors.ErrNoServiceID
		c.Globals.ErrLog.Add(err)
		return err
	}

	input := c.constructInput(serviceID)

	err := c.Globals.Client.DeleteACLEntry(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return err
	}

	text.Success(out, "Deleted ACL entry '%s' (service: %s)", input.ID, serviceID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput(serviceID string) *fastly.DeleteACLEntryInput {
	var input fastly.DeleteACLEntryInput

	input.ACLID = c.aclID
	input.ID = c.id
	input.ServiceID = serviceID

	return &input
}
