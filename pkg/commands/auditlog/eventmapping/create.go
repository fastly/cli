package eventmapping

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/notifications/v1/eventmappings"

	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// knownScopeTypes lists the supported scope type values, offered as
// shell-completion hints for --scope-type.
var knownScopeTypes = []string{
	eventmappings.ScopeTypeAccount,
	eventmappings.ScopeTypeVCL,
	eventmappings.ScopeTypeWasm,
	eventmappings.ScopeTypeNGWAF,
}

// CreateCommand calls the Fastly API to create an audit log event mapping.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	name           string
	scopeType      string
	eventTypes     []string
	integrationIDs []string

	// Optional.
	description argparser.OptionalString
	scopeIDs    argparser.OptionalStringSlice
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create an audit log event mapping").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "A descriptive name for the mapping").Required().StringVar(&c.name)
	c.CmdClause.Flag("scope-type", "The category of Fastly resource the mapping applies to (account, vcl, wasm, ngwaf)").HintOptions(knownScopeTypes...).Required().StringVar(&c.scopeType)
	c.CmdClause.Flag("event-type", "An audit event type that triggers a notification. Set flag multiple times, or provide a comma-separated list, to specify multiple event types").Required().StringsVar(&c.eventTypes, kingpin.Separator(","))
	c.CmdClause.Flag("integration-id", "The ID of an integration that should receive notifications. Set flag multiple times, or provide a comma-separated list, to specify multiple integrations").Required().StringsVar(&c.integrationIDs, kingpin.Separator(","))

	// Optional.
	c.CmdClause.Flag("description", "A description of the mapping").Action(c.description.Set).StringVar(&c.description.Value)
	c.CmdClause.Flag("scope-id", "The ID of a service or workspace to scope the mapping to. Set flag multiple times, or provide a comma-separated list, to specify multiple scope IDs. Omit to apply the mapping to all resources of the given scope type").Action(c.scopeIDs.Set).StringsVar(&c.scopeIDs.Value, kingpin.Separator(","))
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &eventmappings.CreateInput{
		Name:           &c.name,
		ScopeType:      &c.scopeType,
		EventTypes:     c.eventTypes,
		IntegrationIDs: c.integrationIDs,
	}

	if c.description.WasSet {
		input.Description = &c.description.Value
	}
	if c.scopeIDs.WasSet {
		input.ScopeIDs = c.scopeIDs.Value
	}

	em, err := eventmappings.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Name":       c.name,
			"Scope Type": c.scopeType,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, em); ok {
		return err
	}

	text.Success(out, "Created event mapping '%s' (id: %s)", em.Name, em.ID)
	return nil
}
