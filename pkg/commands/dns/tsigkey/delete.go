package tsigkey

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/dns/v1/tsigkeys"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete a TSIG key.
type DeleteCommand struct {
	argparser.Base

	// Required.
	TSIGKeyID string `required:"" help:"The TSIG key ID to delete."`

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

	err := tsigkeys.Delete(context.TODO(), fc, &tsigkeys.DeleteInput{
		TSIGKeyID: &c.TSIGKeyID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TSIG Key ID": c.TSIGKeyID,
		})
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{c.TSIGKeyID, true}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted TSIG key (tsig-key-id: %s)", c.TSIGKeyID)
	return nil
}
