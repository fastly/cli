package zone

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/dns/v1/dnszones"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a DNS Zone.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ZoneID string `required:"" help:"The zone ID to update."`

	// Optional.
	Description        argparser.OptionalString      `help:"A freeform descriptive note."`
	XfrPrimAddress     argparser.OptionalStringSlice `name:"primary-address" sep:"none" help:"An IPv4 address for the Primary DNS Server (repeatable)."`
	XfrPrimDescription argparser.OptionalStringSlice `name:"primary-description" sep:"none" help:"A description of the Primary DNS server (requires --primary-address, repeatable)."`
	XfrTSIGKeyID       argparser.OptionalString      `name:"inbound-tsig-key-id" help:"The ID of the TSIG key used to secure inbound zone transfers. Pass 'nil' to dissociate key."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.XfrPrimDescription.WasSet && !c.XfrPrimAddress.WasSet {
		return fmt.Errorf("--primary-description requires --primary-address")
	}
	if len(c.XfrPrimDescription.Value) > len(c.XfrPrimAddress.Value) {
		return fmt.Errorf("--primary-description cannot be provided more times than --primary-address")
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &dnszones.UpdateInput{
		ZoneID: &c.ZoneID,
	}

	if c.Description.WasSet {
		if strings.TrimSpace(c.Description.Value) == "" {
			input.Description = fastly.NullValue[string]()
		} else {
			input.Description = fastly.NewNullable(c.Description.Value)
		}
	}

	if c.XfrPrimAddress.WasSet || c.XfrTSIGKeyID.WasSet {
		xfr := &dnszones.XfrConfigInboundInput{}

		if c.XfrPrimAddress.WasSet {
			primaries := make([]dnszones.Primary, len(c.XfrPrimAddress.Value))
			for i, addr := range c.XfrPrimAddress.Value {
				primaries[i] = dnszones.Primary{Address: new(addr)}
				if c.XfrPrimDescription.WasSet && i < len(c.XfrPrimDescription.Value) {
					primaries[i].Description = new(c.XfrPrimDescription.Value[i])
				}
			}
			xfr.Primaries = primaries
		}

		// We need to explicitly allow users to unset a given TSIG Key.
		if c.XfrTSIGKeyID.WasSet {
			if c.XfrTSIGKeyID.Value == "nil" {
				xfr.InboundTSIGKeyID = fastly.NullValue[string]()
			} else {
				xfr.InboundTSIGKeyID = fastly.NewNullable(c.XfrTSIGKeyID.Value)
			}
		}

		input.XfrConfigInbound = xfr
	}

	z, err := dnszones.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Zone ID": c.ZoneID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, z); ok {
		return err
	}

	text.Success(out, "Updated DNS zone '%s' (zone-id: %s)", *z.Name, *z.ID)
	return nil
}
