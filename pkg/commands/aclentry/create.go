package aclentry

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Add an ACL entry to an ACL").Alias("add")

	// Required.
	c.CmdClause.Flag("acl-id", "Alphanumeric string identifying a ACL").Required().StringVar(&c.aclID)

	// Optional.
	c.CmdClause.Flag("comment", "A freeform descriptive note").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("ip", "An IP address").Action(c.ip.Set).StringVar(&c.ip.Value)
	c.CmdClause.Flag("negated", "Whether to negate the match").Action(c.negated.Set).BoolVar(&c.negated.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("subnet", "Number of bits for the subnet mask applied to the IP address").Action(c.subnet.Set).IntVar(&c.subnet.Value)

	return &c
}

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	aclID       string
	comment     argparser.OptionalString
	ip          argparser.OptionalString
	negated     argparser.OptionalBool
	serviceName argparser.OptionalServiceNameID
	subnet      argparser.OptionalInt
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := c.constructInput(serviceID)

	a, err := c.Globals.APIClient.CreateACLEntry(input)
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
		ACLID:     c.aclID,
		ServiceID: serviceID,
	}
	if c.ip.WasSet {
		input.IP = &c.ip.Value
	}
	if c.comment.WasSet {
		input.Comment = &c.comment.Value
	}
	if c.negated.WasSet {
		input.Negated = fastly.ToPointer(fastly.Compatibool(c.negated.Value))
	}
	if c.subnet.WasSet {
		input.Subnet = &c.subnet.Value
	}

	return &input
}
