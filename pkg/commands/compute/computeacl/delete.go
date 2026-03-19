package computeacl

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/go-fastly/v13/fastly/computeacls"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// APIDeleteFunc defines the type of 'mock' function which can be provided
// by tests to to replace the function from go-fastly. The signature
// must exactly match the corresponding function in go-fastly.
type APIDeleteFunc func(context.Context, *fastly.Client, *computeacls.DeleteInput) error

// DeleteCommand calls the Fastly API to delete a compute ACL.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	apiHook APIDeleteFunc

	// Required.
	id string
}

// SetHook allows a test to supply a 'mock' function to replace the
// function from go-fastly, and satisfies the
// argparser.HookableCommand interface.
func (c *DeleteCommand) SetHook(f APIDeleteFunc) {
	c.apiHook = f
}

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, g *global.Data) *DeleteCommand {
	c := DeleteCommand{
		Base: argparser.Base{
			Globals: g,
		},
		apiHook: computeacls.Delete,
	}

	c.CmdClause = parent.Command("delete", "Delete a compute ACL")

	// Required.
	c.CmdClause.Flag("acl-id", "Compute ACL ID").Required().StringVar(&c.id)

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

	err := c.apiHook(context.TODO(), fc, &computeacls.DeleteInput{
		ComputeACLID: &c.id,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
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

	text.Success(out, "Deleted compute ACL (id: %s)", c.id)
	return nil
}
