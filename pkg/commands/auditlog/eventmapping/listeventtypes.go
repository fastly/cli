package eventmapping

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/notifications/v1/eventmappings/eventtypes"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListEventTypesCommand calls the Fastly API to list the audit event types
// that can be used when creating an event mapping.
type ListEventTypesCommand struct {
	argparser.Base

	// Optional.
	ScopeType argparser.OptionalString `hints:"account,vcl,wasm,ngwaf" help:"Filters results to event types compatible with the given scope type"`
	Sort      argparser.OptionalString `hints:"event_type,-event_type" help:"The order in which to return results, alphabetically by event type"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *ListEventTypesCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &eventtypes.ListInput{}
	if c.ScopeType.WasSet {
		input.ScopeType = &c.ScopeType.Value
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	types, err := eventtypes.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, types); ok {
		return err
	}

	var data []eventtypes.EventType
	if types != nil {
		data = types.Data
	}
	text.PrintEventTypesTbl(out, data)
	return nil
}
