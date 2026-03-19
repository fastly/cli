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

// APILookupFunc defines the type of 'mock' function which can be provided
// by tests to to replace the function from go-fastly. The signature
// must exactly match the corresponding function in go-fastly.
type APILookupFunc func(context.Context, *fastly.Client, *computeacls.LookupInput) (*computeacls.ComputeACLEntry, error)

// LookupCommand calls the Fastly API to lookup a compute ACL entry.
type LookupCommand struct {
	argparser.Base
	argparser.JSONOutput

	apiHook APILookupFunc

	// Required.
	id string
	ip string
}

// SetHook allows a test to supply a 'mock' function to replace the
// function from go-fastly, and satisfies the
// argparser.HookableCommand interface.
func (c *LookupCommand) SetHook(f APILookupFunc) {
	c.apiHook = f
}

// NewLookupCommand returns a usable command registered under the parent.
func NewLookupCommand(parent argparser.Registerer, g *global.Data) *LookupCommand {
	c := LookupCommand{
		Base: argparser.Base{
			Globals: g,
		},
		apiHook: computeacls.Lookup,
	}

	c.CmdClause = parent.Command("lookup", "Find a matching ACL entry for an IP address")

	// Required.
	c.CmdClause.Flag("acl-id", "Compute ACL ID").Required().StringVar(&c.id)
	c.CmdClause.Flag("ip", "Valid IPv4 or IPv6 address").Required().StringVar(&c.ip)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *LookupCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	entry, err := c.apiHook(context.TODO(), fc, &computeacls.LookupInput{
		ComputeACLID: &c.id,
		ComputeACLIP: &c.ip,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, entry); ok {
		return err
	}

	// no match found
	if entry == nil {
		text.Info(out, "Compute ACL (%s) has no entry with IP (%s)", c.id, c.ip)
		return nil
	}

	text.PrintComputeACLEntry(out, "", entry)
	return nil
}
