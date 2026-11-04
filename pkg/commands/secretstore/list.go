package secretstore

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *ListCommand) AfterApply() error {
	c.Input.Cursor = c.Cursor
	c.Input.Limit = c.Limit
	return nil
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional.
	Cursor string `short:"c" help:"Pagination cursor (Use 'next_cursor' value from list output)"`
	argparser.JSONOutput

	// NOTE: API returns 10 items even when --limit is set to smaller.
	Input fastly.ListSecretStoresInput `kong:"-"`

	Limit int `short:"l" default:"50" help:"Maximum number of items to list"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	data := make([]fastly.SecretStore, 0)

	for {
		o, err := c.Globals.APIClient.ListSecretStores(context.TODO(), &c.Input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if o != nil {
			data = append(data, o.Data...)

			if c.JSONOutput.Enabled || c.Globals.Flags.NonInteractive || c.Globals.Flags.AutoYes {
				if o.Meta.NextCursor != "" {
					c.Input.Cursor = o.Meta.NextCursor
					continue
				}
				break
			}

			text.PrintSecretStoresTbl(out, o.Data)

			if o.Meta.NextCursor != "" {
				text.Break(out)
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					text.Break(out)
					c.Input.Cursor = o.Meta.NextCursor
					continue
				}
			}
		}

		break
	}

	ok, err := c.WriteJSON(out, data)
	if err != nil {
		return err
	}

	// Only print output here if we've not already printed JSON.
	// And only if we're non interactive.
	// Otherwise interactive mode would have displayed each page of data.
	if !ok && (c.Globals.Flags.NonInteractive || c.Globals.Flags.AutoYes) {
		text.PrintSecretStoresTbl(out, data)
	}
	return nil
}
