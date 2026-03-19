package computeacl

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/go-fastly/v13/fastly/computeacls"
)

// APIListFunc defines the type of 'mock' function which can be provided
// by tests to to replace the function from go-fastly. The signature
// must exactly match the corresponding function in go-fastly.
type APIListFunc func(context.Context, *fastly.Client) (*computeacls.ComputeACLs, error)

// ListCommand calls the Fastly API to list all compute ACLs.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	apiHook APIListFunc
}

// SetHook allows a test to supply a 'mock' function to replace the
// function from go-fastly, and satisfies the
// argparser.HookableCommand interface.
func (c *ListCommand) SetHook(f APIListFunc) {
	c.apiHook = f
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
		apiHook: computeacls.ListACLs,
	}

	c.CmdClause = parent.Command("list-acls", "List all compute ACLs")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	acls, err := c.apiHook(context.TODO(), fc)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, acls); ok {
		return err
	}

	text.PrintComputeACLsTbl(out, acls.Data)
	return nil
}
