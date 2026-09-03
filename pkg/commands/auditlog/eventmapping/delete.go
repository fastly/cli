package eventmapping

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete an audit log event mapping.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	id string
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("delete", "Delete an audit log event mapping").Alias("remove")

	// Required.
	c.CmdClause.Flag("id", "The unique identifier of the event mapping").Required().StringVar(&c.id)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

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

	err := eventmappings.Delete(context.TODO(), fc, &eventmappings.DeleteInput{
		MappingID: &c.id,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Mapping ID": c.id,
		})
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			c.id,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted event mapping (id: %s)", c.id)
	return nil
}
