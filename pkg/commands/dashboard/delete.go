package dashboard

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, globals *global.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Delete a custom dashboard").Alias("remove")
	c.Globals = globals

	// Required flags
	c.CmdClause.Flag("id", "Alphanumeric string identifying a Dashboard").Required().StringVar(&c.dashboardID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	dashboardID string
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	err := c.Globals.APIClient.DeleteObservabilityCustomDashboard(input)
	if err != nil {
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"dashboard_id"`
			Deleted bool   `json:"deleted"`
		}{
			c.dashboardID,
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

	input.ID = &c.dashboardID

	return &input
}
