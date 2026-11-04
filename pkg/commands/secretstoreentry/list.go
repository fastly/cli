package secretstoreentry

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
	c.Input.StoreID = c.StoreID
	c.Input.Cursor = c.Cursor
	c.Input.Limit = c.Limit
	return nil
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	Input fastly.ListSecretsInput `kong:"-"`

	// Required.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	Cursor string `short:"c" help:"Pagination cursor (Use 'next_cursor' value from list output)"`
	argparser.JSONOutput
	Limit int `short:"l" default:"50" help:"Maximum number of items to list"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	for {
		o, err := c.Globals.APIClient.ListSecrets(context.TODO(), &c.Input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

		if ok, err := c.WriteJSON(out, o); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		text.PrintSecretsTbl(out, o)

		if o != nil && o.Meta.NextCursor != "" {
			// Check if 'out' is interactive before prompting.
			if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					c.Input.Cursor = o.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}
