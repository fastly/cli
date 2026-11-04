package session

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/airuntimecontrol/v1/session"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list session logs.
type ListCommand struct {
	argparser.Base

	// Optional.
	Key      argparser.OptionalString `help:"Filter by a specific virtual key ID"`
	Provider argparser.OptionalString `help:"Filter by provider connection ID"`
	Model    argparser.OptionalString `help:"Filter by AI model identifier"`
	From     time.Time                `hints:"2026-07-28T19:24:50+00:00" format:"2006-01-02T15:04:05Z07:00" help:"Start of the time range (RFC 3339 format)"`
	To       time.Time                `hints:"2026-07-28T19:24:50+00:00" format:"2006-01-02T15:04:05Z07:00" help:"End of the time range (RFC 3339 format). Defaults to now"`
	argparser.JSONOutput
	Limit int                      `short:"l" default:"50" help:"Maximum number of items to list"`
	Sort  argparser.OptionalString `default:"created_at" placeholder:"\"created_at\"" help:"Sort field. Prefix with '-' for descending order (e.g. \"-created_at\")"`
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

	input := &session.ListInput{
		Limit: &c.Limit,
	}
	if c.Key.WasSet {
		input.Key = &c.Key.Value
	}
	if c.Provider.WasSet {
		input.Provider = &c.Provider.Value
	}
	if c.Model.WasSet {
		input.Model = &c.Model.Value
	}
	if !c.From.IsZero() {
		input.From = &c.From
	}
	if !c.To.IsZero() {
		input.To = &c.To
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	// session.List() automatically paginates through all pages internally.
	data, err := session.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintSessionsTbl(out, data)
	return nil
}
