package zone

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/dns/v1/dnszones"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a DNS Zone.
type DescribeCommand struct {
	argparser.Base

	// Required.
	ZoneID string `required:"" help:"The zone ID to describe."`

	// Optional.
	argparser.JSONOutput
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
		ZoneID: &c.ZoneID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Zone ID": c.ZoneID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, z); ok {
		return err
	}

	text.PrintDNSZone(out, "", z)
	return nil
}
