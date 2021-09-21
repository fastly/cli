package aclentry

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v4/fastly"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data) *CreateCommand {
	var c CreateCommand
	c.CmdClause = parent.Command("create", "Add an ACL entry to an ACL").Alias("add")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	// Required flags
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)
	c.CmdClause.Flag("ip", "An IP address").Required().StringVar(&c.ip)

	// Optional flags
	c.CmdClause.Flag("comment", "A freeform descriptive note").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("negated", "Whether to negate the match").Action(c.negated.Set).BoolVar(&c.negated.Value)
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("subnet", "Number of bits for the subnet mask applied to the IP address").Action(c.subnet.Set).IntVar(&c.subnet.Value)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	cmd.Base

	aclID    string
	comment  cmd.OptionalString
	ip       string
	manifest manifest.Data
	negated  cmd.OptionalBool
	subnet   cmd.OptionalInt
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(in io.Reader, out io.Writer) error {
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

	a, err := c.Globals.Client.CreateACLEntry(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID": serviceID,
		})
		return err
	}

	text.Success(out, "Created ACL entry '%s' (ip: %s, service: %s)", a.ID, a.IP, a.ServiceID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string) *fastly.CreateACLEntryInput {
	var input fastly.CreateACLEntryInput

	input.ACLID = c.aclID
	input.IP = c.ip
	input.ServiceID = serviceID

	if c.comment.WasSet {
		input.Comment = c.comment.Value
	}
	if c.negated.WasSet {
		input.Negated = c.negated.Value
	}
	if c.subnet.WasSet {
		input.Subnet = c.subnet.Value
	}

	return &input
}
