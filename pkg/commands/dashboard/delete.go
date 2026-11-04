package dashboard

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base

	// Required flags
	DashboardID string `name:"id" required:"" help:"ID of the Dashboard to delete"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	err := c.Globals.APIClient.DeleteObservabilityCustomDashboard(context.TODO(), input)
	if err != nil {
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"dashboard_id"`
			Deleted bool   `json:"deleted"`
		}{
			c.DashboardID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, `Deleted Custom Dashboard %s`, fastly.ToValue(input.ID))
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput() *fastly.DeleteObservabilityCustomDashboardInput {
	var input fastly.DeleteObservabilityCustomDashboardInput

	input.ID = &c.DashboardID

	return &input
}
