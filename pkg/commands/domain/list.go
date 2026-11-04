package domain

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/domains"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list domains.
type ListCommand struct {
	argparser.Base

	// Optional.
	Cursor argparser.OptionalString `help:"Cursor value from the next_cursor field of a previous response, used to retrieve the next page"`
	Fqdn   argparser.OptionalString `help:"Filters results by the FQDN using a fuzzy/partial match"`
	argparser.JSONOutput
	Limit     argparser.OptionalInt    `help:"Limit how many results are returned"`
	ServiceID argparser.OptionalString `short:"s" help:"Filter results based on a service_id"`
	Sort      argparser.OptionalString `help:"The order in which to list the results"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &domains.ListInput{}

	if c.ServiceID.WasSet {
		input.ServiceID = &c.ServiceID.Value
	}
	if c.Cursor.WasSet {
		input.Cursor = &c.Cursor.Value
	}
	if c.Fqdn.WasSet {
		input.FQDN = &c.Fqdn.Value
	}
	if c.Limit.WasSet {
		input.Limit = &c.Limit.Value
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	for {
		cl, err := domains.List(context.TODO(), fc, input)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Cursor":     c.Cursor.Value,
				"FQDN":       c.Fqdn.Value,
				"Limit":      c.Limit.Value,
				"Service ID": c.ServiceID.Value,
				"Sort":       c.Sort.Value,
			})
			return err
		}

		if ok, err := c.WriteJSON(out, cl); ok {
			// No pagination prompt w/ JSON output.
			return err
		}

		if c.Globals.Verbose() {
			printVerbose(out, cl.Data)
		} else {
			printSummary(out, cl.Data)
		}

		if cl != nil && cl.Meta.NextCursor != "" {
			// Check if 'out' is interactive before prompting.
			if !c.Globals.Flags.NonInteractive && !c.Globals.Flags.AutoYes && text.IsTTY(out) {
				printNext, err := text.AskYesNo(out, "Print next page [y/N]: ", in)
				if err != nil {
					return err
				}
				if printNext {
					input.Cursor = &cl.Meta.NextCursor
					continue
				}
			}
		}

		return nil
	}
}
