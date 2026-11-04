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

// ListEntriesCommand calls the Fastly API to list all entries of a compute ACLs.
type ListEntriesCommand struct {
	argparser.Base

	// Required.
	ID string `name:"acl-id" required:"" help:"Compute ACL ID"`

	// Optional.
	Cursor string `short:"c" help:"Pagination cursor (Use 'next_cursor' value from list output)"`
	Limit  int    `short:"l" default:"50" help:"Maximum number of items to list"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *ListEntriesCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	var entries []computeacls.ComputeACLEntry
	loadAllPages := c.JSONOutput.Enabled || c.Globals.Flags.NonInteractive || c.Globals.Flags.AutoYes

	for {
		o, err := computeacls.ListEntries(context.TODO(), fc, &computeacls.ListEntriesInput{
			ComputeACLID: &c.ID,
			Cursor:       &c.Cursor,
			Limit:        &c.Limit,
		})
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if o != nil {
			entries = append(entries, o.Entries...)

			if loadAllPages {
				if next := o.Meta.NextCursor; next != "" {
					c.Cursor = next
					continue
				}
				break
			}

			text.PrintComputeACLEntriesTbl(out, o.Entries)

			if next := o.Meta.NextCursor; next != "" {
				text.Break(out)
				printNextPage, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNextPage {
					c.Cursor = next
					continue
				}
			}
		}

		break
	}

	ok, err := c.WriteJSON(out, entries)
	if err != nil {
		return err
	}

	// Only print output here if we've not already printed JSON.
	// And only if we're non interactive.
	// Otherwise interactive mode would have displayed each page of data.
	if !ok && (c.Globals.Flags.NonInteractive || c.Globals.Flags.AutoYes) {
		text.PrintComputeACLEntriesTbl(out, entries)
	}

	return nil
}
