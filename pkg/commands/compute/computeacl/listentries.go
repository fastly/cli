package computeacl

import (
	"errors"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"
	"github.com/fastly/go-fastly/v10/fastly/computeacls"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListEntriesCommand calls the Fastly API to list all entries of a compute ACLs.
type ListEntriesCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	id string

	// Optional.
	cursor string
	limit  int
}

// NewListEntriesCommand returns a usable command registered under the parent.
func NewListEntriesCommand(parent argparser.Registerer, g *global.Data) *ListEntriesCommand {
	c := ListEntriesCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("list-entries", "List all entries of a compute ACL")

	// Required.
	c.CmdClause.Flag("acl-id", "Compute ACL ID").Required().StringVar(&c.id)

	// Optional.
	c.RegisterFlag(argparser.CursorFlag(&c.cursor))
	c.RegisterFlagInt(argparser.LimitFlag(&c.limit))
	c.RegisterFlagBool(c.JSONFlag())

	return &c
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
		o, err := computeacls.ListEntries(fc, &computeacls.ListEntriesInput{
			ComputeACLID: &c.id,
			Cursor:       &c.cursor,
			Limit:        &c.limit,
		})
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if o != nil {
			entries = append(entries, o.Entries...)

			if loadAllPages {
				if next := o.Meta.NextCursor; next != "" {
					c.cursor = next
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
					c.cursor = next
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
