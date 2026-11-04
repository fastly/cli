package dashboard

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an appropriate resource.
type CreateCommand struct {
	argparser.Base

	// Required flags
	NameFlag string `name:"name" short:"n" required:"" help:"A human-readable name for the dashboard"`

	// Optional flags
	argparser.JSONOutput

	Description argparser.OptionalString `help:"A short description of the dashboard"`
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	dashboard, err := c.Globals.APIClient.CreateObservabilityCustomDashboard(context.TODO(), input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, dashboard); ok {
		return err
	}

	text.Success(out, `Created Custom Dashboard "%s" (id: %s)`, dashboard.Name, dashboard.ID)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) constructInput() *fastly.CreateObservabilityCustomDashboardInput {
	input := fastly.CreateObservabilityCustomDashboardInput{
		Name:  c.NameFlag,
		Items: []fastly.DashboardItem{},
	}

	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	return &input
}
