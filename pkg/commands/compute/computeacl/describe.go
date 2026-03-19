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

// APIDescribeFunc defines the type of 'mock' function which can be provided
// by tests to to replace the function from go-fastly. The signature
// must exactly match the corresponding function in go-fastly.
type APIDescribeFunc func(context.Context, *fastly.Client, *computeacls.DescribeInput) (*computeacls.ComputeACL, error)

// DescribeCommand calls the Fastly API to describe a compute ACL.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	apiHook APIDescribeFunc

	// Required.
	id string
}

// SetHook allows a test to supply a 'mock' function to replace the
// function from go-fastly, and satisfies the
// argparser.HookableCommand interface.
func (c *DescribeCommand) SetHook(f APIDescribeFunc) {
	c.apiHook = f
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
		apiHook: computeacls.Describe,
	}

	c.CmdClause = parent.Command("describe", "Describe a compute ACL")

	// Required.
	c.CmdClause.Flag("acl-id", "Compute ACL ID").Required().StringVar(&c.id)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

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

	acl, err := c.apiHook(context.TODO(), fc, &computeacls.DescribeInput{
		ComputeACLID: &c.id,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, acl); ok {
		return err
	}

	text.PrintComputeACL(out, "", acl)
	return nil
}
