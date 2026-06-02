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

// UpdateCommand calls the Fastly API to update a DNS Zone.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	zoneID string

	// Optional.
	description        argparser.OptionalString
	xfrTSIGKeyID       argparser.OptionalString
	xfrPrimAddress     argparser.OptionalStringSlice
	xfrPrimDescription argparser.OptionalStringSlice
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a DNS Zone")

	// Required.
	c.CmdClause.Flag("zone-id", "The zone ID to update.").Required().StringVar(&c.zoneID)

	// Optional.
	c.CmdClause.Flag("description", "A freeform descriptive note.").Action(c.description.Set).StringVar(&c.description.Value)
	c.CmdClause.Flag("primary-address", "An IPv4 address for the Primary DNS Server (repeatable).").Action(c.xfrPrimAddress.Set).StringsVar(&c.xfrPrimAddress.Value)
	c.CmdClause.Flag("primary-description", "A description of the Primary DNS server (requires --primary-address, repeatable).").Action(c.xfrPrimDescription.Set).StringsVar(&c.xfrPrimDescription.Value)
	c.CmdClause.Flag("inbound-tsig-key-id", "The ID of the TSIG key used to secure inbound zone transfers. Pass 'nil' to dissociate key.").Action(c.xfrTSIGKeyID.Set).StringVar(&c.xfrTSIGKeyID.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.xfrPrimDescription.WasSet && !c.xfrPrimAddress.WasSet {
		return fmt.Errorf("--primary-description requires --primary-address")
	}
	if len(c.xfrPrimDescription.Value) > len(c.xfrPrimAddress.Value) {
		return fmt.Errorf("--primary-description cannot be provided more times than --primary-address")
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &dnszones.UpdateInput{
		ZoneID: &c.zoneID,
	}

	if c.description.WasSet {
		if strings.TrimSpace(c.description.Value) == "" {
			input.Description = fastly.NullValue[string]()
		} else {
			input.Description = fastly.NewNullable(c.description.Value)
		}
	}

	if c.xfrPrimAddress.WasSet || c.xfrTSIGKeyID.WasSet {
		xfr := &dnszones.XfrConfigInboundInput{}

		if c.xfrPrimAddress.WasSet {
			primaries := make([]dnszones.Primary, len(c.xfrPrimAddress.Value))
			for i, addr := range c.xfrPrimAddress.Value {
				primaries[i] = dnszones.Primary{Address: fastly.ToPointer(addr)}
				if c.xfrPrimDescription.WasSet && i < len(c.xfrPrimDescription.Value) {
					primaries[i].Description = fastly.ToPointer(c.xfrPrimDescription.Value[i])
				}
			}
			xfr.Primaries = primaries
		}

		// We need to explicitly allow users to unset a given TSIG Key.
		if c.xfrTSIGKeyID.WasSet {
			if c.xfrTSIGKeyID.Value == "nil" {
				xfr.InboundTSIGKeyID = fastly.NullValue[string]()
			} else {
				xfr.InboundTSIGKeyID = fastly.NewNullable(c.xfrTSIGKeyID.Value)
			}
		}

		input.XfrConfigInbound = xfr
	}

	z, err := dnszones.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Zone ID": c.zoneID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, z); ok {
		return err
	}

	text.Success(out, "Updated DNS zone '%s' (zone-id: %s)", *z.Name, *z.ID)
	return nil
}
