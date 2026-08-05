package eventmapping

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
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
	argparser.JSONOutput

	// Optional.
	integrationID argparser.OptionalString
	mappingStatus argparser.OptionalString
	name          argparser.OptionalString
	scopeID       argparser.OptionalString
	scopeType     argparser.OptionalString
	sort          argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List audit log event mappings")

	// Optional.
	c.CmdClause.Flag("integration-id", "Filters results to mappings that reference the given integration ID").Action(c.integrationID.Set).StringVar(&c.integrationID.Value)
	c.CmdClause.Flag("mapping-status", "Filters results by mapping status").HintOptions(knownMappingStatuses...).Action(c.mappingStatus.Set).StringVar(&c.mappingStatus.Value)
	c.CmdClause.Flag("name", "Filters results to mappings whose name contains the given string (case-insensitive)").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("scope-id", "Filters results to mappings that apply to the given service or workspace ID").Action(c.scopeID.Set).StringVar(&c.scopeID.Value)
	c.CmdClause.Flag("scope-type", "Filters results to the given scope type").HintOptions(knownScopeTypes...).Action(c.scopeType.Set).StringVar(&c.scopeType.Value)
	c.CmdClause.Flag("sort", "The order in which to return results by creation date").HintOptions(knownSortValues...).Action(c.sort.Set).StringVar(&c.sort.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
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
	if c.integrationID.WasSet {
		input.IntegrationID = &c.integrationID.Value
	}
	if c.mappingStatus.WasSet {
		input.MappingStatus = &c.mappingStatus.Value
	}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.scopeID.WasSet {
		input.ScopeID = &c.scopeID.Value
	}
	if c.scopeType.WasSet {
		input.ScopeType = &c.scopeType.Value
	}
	if c.sort.WasSet {
		input.Sort = &c.sort.Value
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
