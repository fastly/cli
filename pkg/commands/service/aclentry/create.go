package aclentry

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *CreateCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required.
	ACLID string `name:"acl-id" required:"" help:"Alphanumeric string identifying a ACL"`

	// Optional.
	Comment     argparser.OptionalString        `help:"A freeform descriptive note"`
	IP          argparser.OptionalString        `help:"An IP address"`
	Negated     argparser.OptionalBool          `help:"Whether to negate the match"`
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	Subnet      argparser.OptionalInt           `help:"Number of bits for the subnet mask applied to the IP address"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := c.constructInput(serviceID)

	a, err := c.Globals.APIClient.CreateACLEntry(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	text.Success(out, "Created ACL entry '%s' (ip: %s, negated: %t, service: %s)", fastly.ToValue(a.EntryID), fastly.ToValue(a.IP), fastly.ToValue(a.Negated), fastly.ToValue(a.ServiceID))
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput(serviceID string) *fastly.CreateACLEntryInput {
	input := fastly.CreateACLEntryInput{
		ACLID:     c.ACLID,
		ServiceID: serviceID,
	}
	if c.IP.WasSet {
		input.IP = &c.IP.Value
	}
	if c.Comment.WasSet {
		input.Comment = &c.Comment.Value
	}
	if c.Negated.WasSet {
		input.Negated = new(fastly.Compatibool(c.Negated.Value))
	}
	if c.Subnet.WasSet {
		input.Subnet = &c.Subnet.Value
	}

	return &input
}
