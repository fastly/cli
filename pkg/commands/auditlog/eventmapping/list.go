package eventmapping

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/notifications/v1/eventmappings"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// knownMappingStatuses lists the supported mapping status values, offered as
// shell-completion hints for --mapping-status.
var knownMappingStatuses = []string{
	eventmappings.MappingStatusActive,
	eventmappings.MappingStatusInactive,
}

// knownSortValues lists the supported sort values, offered as
// shell-completion hints for --sort.
var knownSortValues = []string{"created_at", "-created_at"}

// ListCommand calls the Fastly API to list audit log event mappings.
type ListCommand struct {
	argparser.Base

	// Optional.
	IntegrationID argparser.OptionalString `help:"Filters results to mappings that reference the given integration ID"`
	MappingStatus argparser.OptionalString `hints:"active,inactive" help:"Filters results by mapping status"`
	NameFlag      argparser.OptionalString `name:"name" help:"Filters results to mappings whose name contains the given string (case-insensitive)"`
	ScopeID       argparser.OptionalString `help:"Filters results to mappings that apply to the given service or workspace ID"`
	ScopeType     argparser.OptionalString `hints:"account,vcl,wasm,ngwaf" help:"Filters results to the given scope type"`
	Sort          argparser.OptionalString `hints:"created_at,-created_at" help:"The order in which to return results by creation date"`

	argparser.JSONOutput
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

	input := &eventmappings.ListInput{}
	if c.IntegrationID.WasSet {
		input.IntegrationID = &c.IntegrationID.Value
	}
	if c.MappingStatus.WasSet {
		input.MappingStatus = &c.MappingStatus.Value
	}
	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.ScopeID.WasSet {
		input.ScopeID = &c.ScopeID.Value
	}
	if c.ScopeType.WasSet {
		input.ScopeType = &c.ScopeType.Value
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	ems, err := eventmappings.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, ems); ok {
		return err
	}

	text.PrintEventMappingsTbl(out, ems)
	return nil
}
