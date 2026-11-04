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

// DeleteCommand calls the Fastly API to delete a DNS Zone.
type DeleteCommand struct {
	argparser.Base

	// Required.
	ZoneID string `required:"" help:"The zone ID to delete."`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	err := dnszones.Delete(context.TODO(), fc, &dnszones.DeleteInput{
		ZoneID: &c.ZoneID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Zone ID": c.ZoneID,
		})
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{c.ZoneID, true}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted DNS zone (zone-id: %s)", c.ZoneID)
	return nil
}
