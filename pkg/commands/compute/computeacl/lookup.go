package computeacl

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/go-fastly/v19/fastly/computeacls"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// LookupCommand calls the Fastly API to lookup a compute ACL entry.
type LookupCommand struct {
	argparser.Base

	// Required.
	ID string `name:"acl-id" required:"" help:"Compute ACL ID"`
	IP string `required:"" help:"Valid IPv4 or IPv6 address"`

	// Optional.
	argparser.JSONOutput
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

	entry, err := computeacls.Lookup(context.TODO(), fc, &computeacls.LookupInput{
		ComputeACLID: &c.ID,
		ComputeACLIP: &c.IP,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, entry); ok {
		return err
	}

	// Status 204 - No Content
	if entry == nil {
		text.Info(out, "Compute ACL (%s) has no entry with IP (%s)", c.ID, c.IP)
		return nil
	}

	text.PrintComputeACLEntry(out, "", entry)
	return nil
}
