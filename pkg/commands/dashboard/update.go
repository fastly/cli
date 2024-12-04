package dashboard

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, globals *global.Data) *UpdateCommand {
	var c UpdateCommand
	c.CmdClause = parent.Command("update", "Update a custom dashboard")
	c.Globals = globals

	// Required flags
	c.CmdClause.Flag("id", "Alphanumeric string identifying a Dashboard").Required().StringVar(&c.dashboardID)

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag())                                                                                                  // --json
	c.CmdClause.Flag("name", "A human-readable name for the dashboard").Short('n').Action(c.name.Set).StringVar(&c.name.Value)        // --name
	c.CmdClause.Flag("description", "A short description of the dashboard").Action(c.description.Set).StringVar(&c.description.Value) // --description

	return &c
}

// UpdateCommand calls the Fastly API to update an appropriate resource.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	dashboardID string
	name        argparser.OptionalString
	description argparser.OptionalString
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	dashboard, err := c.Globals.APIClient.UpdateObservabilityCustomDashboard(input)
	if err != nil {
		return err
	}

	text.Success(out, `Updated Custom Dashboard "%s" (id: %s)`, dashboard.Name, dashboard.ID)
	dashboards := []*fastly.ObservabilityCustomDashboard{dashboard}
	if c.Globals.Verbose() {
		printVerbose(out, dashboards)
	} else {
		printSummary(out, dashboards)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateObservabilityCustomDashboardInput {
	var input fastly.UpdateObservabilityCustomDashboardInput

	input.ID = &c.dashboardID

	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.description.WasSet {
		input.Description = &c.description.Value
	}

	return &input
}
