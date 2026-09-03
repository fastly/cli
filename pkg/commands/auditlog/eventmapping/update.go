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

// UpdateCommand calls the Fastly API to update an audit log event mapping.
//
// Update replaces the entire event mapping, so all required fields must be
// provided; omitted fields are not preserved from the previous version.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	id             string
	name           string
	scopeType      string
	eventTypes     []string
	integrationIDs []string

	// Optional.
	description argparser.OptionalString
	scopeIDs    argparser.OptionalStringSlice
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update an audit log event mapping. This replaces the entire mapping, so all required fields must be provided")

	// Required.
	c.CmdClause.Flag("id", "The unique identifier of the event mapping").Required().StringVar(&c.id)
	c.CmdClause.Flag("name", "A descriptive name for the mapping").Required().StringVar(&c.name)
	c.CmdClause.Flag("scope-type", "The category of Fastly resource the mapping applies to (account, vcl, wasm, ngwaf)").HintOptions(knownScopeTypes...).Required().StringVar(&c.scopeType)
	c.CmdClause.Flag("event-type", "An audit event type that triggers a notification. Set flag multiple times, or provide a comma-separated list, to specify multiple event types").Required().StringsVar(&c.eventTypes, kingpin.Separator(","))
	c.CmdClause.Flag("integration-id", "The ID of an integration that should receive notifications. Set flag multiple times, or provide a comma-separated list, to specify multiple integrations").Required().StringsVar(&c.integrationIDs, kingpin.Separator(","))

	// Optional.
	c.CmdClause.Flag("description", "A description of the mapping").Action(c.description.Set).StringVar(&c.description.Value)
	c.CmdClause.Flag("scope-id", "The ID of a service or workspace to scope the mapping to. Set flag multiple times, or provide a comma-separated list, to specify multiple scope IDs").Action(c.scopeIDs.Set).StringsVar(&c.scopeIDs.Value, kingpin.Separator(","))
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &eventmappings.UpdateInput{
		MappingID:      &c.id,
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

	em, err := eventmappings.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Mapping ID": c.id,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, em); ok {
		return err
	}

	text.Success(out, "Updated event mapping '%s' (id: %s)", em.Name, em.ID)
	return nil
}
