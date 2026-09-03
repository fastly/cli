package eventmapping

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings/eventtypes"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListEventTypesCommand calls the Fastly API to list the audit event types
// that can be used when creating an event mapping.
type ListEventTypesCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Optional.
	scopeType argparser.OptionalString
	sort      argparser.OptionalString
}

// NewListEventTypesCommand returns a usable command registered under the parent.
func NewListEventTypesCommand(parent argparser.Registerer, g *global.Data) *ListEventTypesCommand {
	c := ListEventTypesCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list-event-types", "List the audit event types supported when creating an event mapping")

	// Optional.
	c.CmdClause.Flag("scope-type", "Filters results to event types compatible with the given scope type").HintOptions(knownScopeTypes...).Action(c.scopeType.Set).StringVar(&c.scopeType.Value)
	c.CmdClause.Flag("sort", "The order in which to return results, alphabetically by event type").HintOptions("event_type", "-event_type").Action(c.sort.Set).StringVar(&c.sort.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
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
	if c.scopeType.WasSet {
		input.ScopeType = &c.scopeType.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
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
