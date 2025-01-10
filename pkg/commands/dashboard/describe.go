package dashboard

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/dashboard/common"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, globals *global.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Show detailed information about a custom dashboard").Alias("get")
	c.Globals = globals

	// Required flags
	c.CmdClause.Flag("id", "ID of the Dashboard to describe").Required().StringVar(&c.dashboardID)

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag())
	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	dashboardID string
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	dashboard, err := c.Globals.APIClient.GetObservabilityCustomDashboard(input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, dashboard); ok {
		return err
	}

	common.PrintDashboard(out, 0, dashboard)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetObservabilityCustomDashboardInput {
	var input fastly.GetObservabilityCustomDashboardInput

	input.ID = &c.dashboardID

	return &input
}
