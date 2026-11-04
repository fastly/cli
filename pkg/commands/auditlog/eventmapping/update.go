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

// UpdateCommand calls the Fastly API to update an audit log event mapping.
//
// Update replaces the entire event mapping, so all required fields must be
// provided; omitted fields are not preserved from the previous version.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ID             string   `required:"" help:"The unique identifier of the event mapping"`
	NameFlag       string   `name:"name" required:"" help:"A descriptive name for the mapping"`
	ScopeType      string   `required:"" hints:"account,vcl,wasm,ngwaf" help:"The category of Fastly resource the mapping applies to (account, vcl, wasm, ngwaf)"`
	EventTypes     []string `name:"event-type" required:"" sep:"," help:"An audit event type that triggers a notification. Set flag multiple times, or provide a comma-separated list, to specify multiple event types"`
	IntegrationIDs []string `name:"integration-id" required:"" sep:"," help:"The ID of an integration that should receive notifications. Set flag multiple times, or provide a comma-separated list, to specify multiple integrations"`

	// Optional.
	Description argparser.OptionalString      `help:"A description of the mapping"`
	ScopeIDs    argparser.OptionalStringSlice `name:"scope-id" sep:"," help:"The ID of a service or workspace to scope the mapping to. Set flag multiple times, or provide a comma-separated list, to specify multiple scope IDs"`

	argparser.JSONOutput
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
		MappingID:      &c.ID,
		Name:           &c.NameFlag,
		ScopeType:      &c.ScopeType,
		EventTypes:     c.EventTypes,
		IntegrationIDs: c.IntegrationIDs,
	}

	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}
	if c.ScopeIDs.WasSet {
		input.ScopeIDs = c.ScopeIDs.Value
	}

	em, err := eventmappings.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Mapping ID": c.ID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, em); ok {
		return err
	}

	text.Success(out, "Updated event mapping '%s' (id: %s)", em.Name, em.ID)
	return nil
}
