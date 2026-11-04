package dashboard

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base

	// Required flags
	DashboardID string `name:"id" required:"" help:"ID of the Dashboard to update"`

	// Optional flags
	argparser.JSONOutput

	NameFlag    argparser.OptionalString `name:"name" short:"n" help:"A human-readable name for the dashboard"`
	Description argparser.OptionalString `help:"A short description of the dashboard"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	dashboard, err := c.Globals.APIClient.UpdateObservabilityCustomDashboard(context.TODO(), input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, dashboard); ok {
		return err
	}

	text.Success(out, `Updated Custom Dashboard "%s" (id: %s)`, dashboard.Name, dashboard.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateObservabilityCustomDashboardInput {
	var input fastly.UpdateObservabilityCustomDashboardInput

	input.ID = &c.DashboardID

	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	return &input
}
