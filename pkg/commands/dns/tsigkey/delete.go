package tsigkey

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/dns/v1/tsigkeys"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete a TSIG key.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	tsigKeyID string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete a TSIG key").Alias("remove")

	// Required.
	c.CmdClause.Flag("tsig-key-id", "The TSIG key ID to delete.").Required().StringVar(&c.tsigKeyID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
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
		TSIGKeyID: &c.tsigKeyID,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TSIG Key ID": c.tsigKeyID,
		})
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{c.tsigKeyID, true}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted TSIG key (tsig-key-id: %s)", c.tsigKeyID)
	return nil
}
