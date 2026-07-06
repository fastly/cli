package zone

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v16/fastly"
	"github.com/fastly/go-fastly/v16/fastly/dns/v1/dnszones"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a DNS Zone.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	zoneID string
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Describe a DNS Zone").Alias("get")

	// Required.
	c.CmdClause.Flag("zone-id", "The zone ID to describe.").Required().StringVar(&c.zoneID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	z, err := dnszones.Get(context.TODO(), fc, &dnszones.GetInput{
		ZoneID: &c.zoneID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Zone ID": c.zoneID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, z); ok {
		return err
	}

	text.PrintDNSZone(out, "", z)
	return nil
}
