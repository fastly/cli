package dashboard

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dashboard/printer"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required flags
	DashboardID string `name:"id" required:"" help:"ID of the Dashboard to describe"`

	// Optional flags
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	dashboard, err := c.Globals.APIClient.GetObservabilityCustomDashboard(context.TODO(), input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, dashboard); ok {
		return err
	}

	printer.PrintDashboard(out, 0, dashboard)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetObservabilityCustomDashboardInput {
	var input fastly.GetObservabilityCustomDashboardInput

	input.ID = &c.DashboardID

	return &input
}
