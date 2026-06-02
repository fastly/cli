package zone

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/dns/v1/dnszones"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a DNS Zone.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	name string

	// Optional.
	description        argparser.OptionalString
	xfrTSIGKeyID       argparser.OptionalString
	xfrPrimAddress     argparser.OptionalStringSlice
	xfrPrimDescription argparser.OptionalStringSlice
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a DNS Zone").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The domain name for your zone. Must be in FQDN format.").Required().StringVar(&c.name)

	// Optional.
	c.CmdClause.Flag("description", "A freeform descriptive note.").Action(c.description.Set).StringVar(&c.description.Value)
	c.CmdClause.Flag("primary-address", "An IPv4 address for the Primary DNS Server (repeatable).").Action(c.xfrPrimAddress.Set).StringsVar(&c.xfrPrimAddress.Value)
	c.CmdClause.Flag("primary-description", "A description of the Primary DNS server (requires --primary-address, repeatable).").Action(c.xfrPrimDescription.Set).StringsVar(&c.xfrPrimDescription.Value)
	c.CmdClause.Flag("inbound-tsig-key-id", "The ID of the TSIG key used to secure inbound zone transfers.").Action(c.xfrTSIGKeyID.Set).StringVar(&c.xfrTSIGKeyID.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if strings.Contains(c.name, " ") {
		return fmt.Errorf("invalid --name value %q: zone names cannot contain spaces", c.name)
	}
	if len(c.name) > 255 {
		return fmt.Errorf("invalid --name value %q: zone names cannot exceed 255 characters", c.name)
	}

	// zoneType must be explicitly set to 'secondary' as it's the only possible API value.
	zoneType := "secondary"
	input := &dnszones.CreateInput{
		Name: &c.name,
		Type: &zoneType,
	}

	if c.description.WasSet {
		input.Description = &c.description.Value
	}

	if c.xfrPrimAddress.WasSet {
		primaries := make([]dnszones.Primary, len(c.xfrPrimAddress.Value))
		for i, addr := range c.xfrPrimAddress.Value {
			primaries[i] = dnszones.Primary{Address: fastly.ToPointer(addr)}
			if c.xfrPrimDescription.WasSet && i < len(c.xfrPrimDescription.Value) {
				primaries[i].Description = fastly.ToPointer(c.xfrPrimDescription.Value[i])
			}
		}
		input.XfrConfigInbound = &dnszones.XfrConfigInboundInput{
			Primaries: primaries,
		}
	}

	if c.xfrTSIGKeyID.WasSet {
		if input.XfrConfigInbound == nil {
			input.XfrConfigInbound = &dnszones.XfrConfigInboundInput{}
		}
		input.XfrConfigInbound.InboundTSIGKeyID = fastly.NewNullable(c.xfrTSIGKeyID.Value)
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := dnszones.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Name": c.name,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	text.Success(out, "Created DNS zone '%s' (zone-id: %s)", *d.Name, *d.ID)
	return nil
}
