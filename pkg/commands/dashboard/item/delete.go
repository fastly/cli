package item

import (
	"io"
	"slices"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewDeleteCommand returns a usable command registered under the parent.
func NewDeleteCommand(parent argparser.Registerer, globals *global.Data) *DeleteCommand {
	var c DeleteCommand
	c.CmdClause = parent.Command("delete", "Delete a custom dashboard item").Alias("add")
	c.Globals = globals

	// Required flags
	c.CmdClause.Flag("dashboard-id", "ID of the Dashboard containing the item").Required().StringVar(&c.dashboardID) // --dashboard-id
	c.CmdClause.Flag("item-id", "ID of the Item to be deleted").Required().StringVar(&c.itemID)      // --item-id

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base
	argparser.JSONOutput

	// required
	dashboardID string
	itemID      string
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	d, err := c.Globals.APIClient.GetObservabilityCustomDashboard(&fastly.GetObservabilityCustomDashboardInput{ID: &c.dashboardID})
	if err != nil {
		return err
	}

	success := false
	numItems := len(d.Items)

	if slices.ContainsFunc(d.Items, func(di fastly.DashboardItem) bool {
		return di.ID == c.itemID
	}) {
		input := c.constructInput(d)
		d, err = c.Globals.APIClient.UpdateObservabilityCustomDashboard(input)
		if err != nil {
			return err
		}

		success = true
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID       string                               `json:"item_id"`
			Deleted  bool                                 `json:"deleted"`
			NewState *fastly.ObservabilityCustomDashboard `json:"dashboard_state"`
		}{
			c.itemID,
			success,
			d,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	if success {
		text.Success(out, `Removed %d dashboard item(s) from Custom Dashboard "%s" (dashboardID: %s)`, (numItems - (len(d.Items))), d.Name, d.ID)
	} else {
		text.Warning(out, "dashboard (%s) has no item with ID (%s)", d.ID, c.itemID)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DeleteCommand) constructInput(d *fastly.ObservabilityCustomDashboard) *fastly.UpdateObservabilityCustomDashboardInput {
	input := fastly.UpdateObservabilityCustomDashboardInput{
		ID:          &d.ID,
		Name:        &d.Name,
		Description: &d.Description,
	}

	items := slices.DeleteFunc(d.Items, func(di fastly.DashboardItem) bool {
		return di.ID == c.itemID
	})
	input.Items = &items

	return &input
}
