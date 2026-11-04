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

// CreateCommand calls the Fastly API to create a DNS Zone.
type CreateCommand struct {
	argparser.Base

	// Required.
	NameFlag string `name:"name" required:"" help:"The domain name for your zone. Must be in FQDN format."`

	// Optional.
	Description        argparser.OptionalString      `help:"A freeform descriptive note."`
	XfrPrimAddress     argparser.OptionalStringSlice `name:"primary-address" sep:"none" help:"An IPv4 address for the Primary DNS Server (repeatable)."`
	XfrPrimDescription argparser.OptionalStringSlice `name:"primary-description" sep:"none" help:"A description of the Primary DNS server (requires --primary-address, repeatable)."`
	XfrTSIGKeyID       argparser.OptionalString      `name:"inbound-tsig-key-id" help:"The ID of the TSIG key used to secure inbound zone transfers."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if strings.Contains(c.NameFlag, " ") {
		return fmt.Errorf("invalid --name value %q: zone names cannot contain spaces", c.NameFlag)
	}
	if len(c.NameFlag) > 255 {
		return fmt.Errorf("invalid --name value %q: zone names cannot exceed 255 characters", c.NameFlag)
	}

	if c.XfrPrimDescription.WasSet && !c.XfrPrimAddress.WasSet {
		return fmt.Errorf("--primary-description requires --primary-address")
	}
	if len(c.XfrPrimDescription.Value) > len(c.XfrPrimAddress.Value) {
		return fmt.Errorf("--primary-description cannot be provided more times than --primary-address")
	}

	// zoneType must be explicitly set to 'secondary' as it's the only possible API value.
	zoneType := "secondary"
	input := &dnszones.CreateInput{
		Name: &c.NameFlag,
		Type: &zoneType,
	}

	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	if c.XfrPrimAddress.WasSet {
		primaries := make([]dnszones.Primary, len(c.XfrPrimAddress.Value))
		for i, addr := range c.XfrPrimAddress.Value {
			primaries[i] = dnszones.Primary{Address: new(addr)}
			if c.XfrPrimDescription.WasSet && i < len(c.XfrPrimDescription.Value) {
				primaries[i].Description = new(c.XfrPrimDescription.Value[i])
			}
		}
		input.XfrConfigInbound = &dnszones.XfrConfigInboundInput{
			Primaries: primaries,
		}
	}

	if c.XfrTSIGKeyID.WasSet {
		if input.XfrConfigInbound == nil {
			input.XfrConfigInbound = &dnszones.XfrConfigInboundInput{}
		}
		input.XfrConfigInbound.InboundTSIGKeyID = fastly.NewNullable(c.XfrTSIGKeyID.Value)
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := dnszones.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Name": c.NameFlag,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	text.Success(out, "Created DNS zone '%s' (zone-id: %s)", *d.Name, *d.ID)
	return nil
}
