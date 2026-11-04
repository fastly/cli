package eventmapping

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/notifications/v1/eventmappings/scopetypes"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListScopeTypesCommand calls the Fastly API to list the scope types
// supported when creating an event mapping.
type ListScopeTypesCommand struct {
	argparser.Base

	// Optional.
	Sort argparser.OptionalString `hints:"scope_type,-scope_type" help:"The order in which to return results, alphabetically by scope type"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *ListScopeTypesCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &scopetypes.ListInput{}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	types, err := scopetypes.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, types); ok {
		return err
	}

	var data []scopetypes.ScopeType
	if types != nil {
		data = types.Data
	}
	text.PrintScopeTypesTbl(out, data)
	return nil
}
