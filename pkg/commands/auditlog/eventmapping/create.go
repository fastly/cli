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

	// Required.
	NameFlag       string   `name:"name" required:"" help:"A descriptive name for the mapping"`
	ScopeType      string   `required:"" hints:"account,vcl,wasm,ngwaf" help:"The category of Fastly resource the mapping applies to (account, vcl, wasm, ngwaf)"`
	EventTypes     []string `name:"event-type" required:"" sep:"," help:"An audit event type that triggers a notification. Set flag multiple times, or provide a comma-separated list, to specify multiple event types"`
	IntegrationIDs []string `name:"integration-id" required:"" sep:"," help:"The ID of an integration that should receive notifications. Set flag multiple times, or provide a comma-separated list, to specify multiple integrations"`

	// Optional.
	Description argparser.OptionalString      `help:"A description of the mapping"`
	ScopeIDs    argparser.OptionalStringSlice `name:"scope-id" sep:"," help:"The ID of a service or workspace to scope the mapping to. Set flag multiple times, or provide a comma-separated list, to specify multiple scope IDs. Omit to apply the mapping to all resources of the given scope type"`

	argparser.JSONOutput
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

	em, err := eventmappings.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Name":       c.NameFlag,
			"Scope Type": c.ScopeType,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, em); ok {
		return err
	}

	text.Success(out, "Created event mapping '%s' (id: %s)", em.Name, em.ID)
	return nil
}
