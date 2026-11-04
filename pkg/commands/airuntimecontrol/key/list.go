package key

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/airuntimecontrol/v1/key"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list virtual keys.
type ListCommand struct {
	argparser.Base

	// Optional.
	Model          argparser.OptionalString `help:"Filter by AI model identifier"`
	Provider       argparser.OptionalString `help:"Filter by AI model provider"`
	IncludeDeleted argparser.OptionalBool   `help:"Include deleted virtual keys"`
	Search         argparser.OptionalString `help:"Search virtual keys by substring match on key name"`
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

	input := &key.ListInput{
		Limit: &c.Limit,
	}
	if c.Model.WasSet {
		input.Model = &c.Model.Value
	}
	if c.Provider.WasSet {
		input.Provider = &c.Provider.Value
	}
	if c.IncludeDeleted.WasSet {
		input.IncludeDeleted = &c.IncludeDeleted.Value
	}
	if c.Search.WasSet {
		input.Search = &c.Search.Value
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	// key.List() automatically paginates through all pages internally.
	data, err := key.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintVirtualKeysTbl(out, data)
	return nil
}
