package item

import (
	"context"
	"io"
	"slices"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete an appropriate resource.
type DeleteCommand struct {
	argparser.Base

	// Required flags

	// required
	DashboardID string `required:"" help:"ID of the Dashboard containing the item"`
	ItemID      string `required:"" help:"ID of the Item to be deleted"`

	// Optional flags
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	d, err := c.Globals.APIClient.GetObservabilityCustomDashboard(context.TODO(), &fastly.GetObservabilityCustomDashboardInput{ID: &c.DashboardID})
	if err != nil {
		return err
	}

	success := false
	numItems := len(d.Items)

	if slices.ContainsFunc(d.Items, func(di fastly.DashboardItem) bool {
		return di.ID == c.ItemID
	}) {
		input := c.constructInput(d)
		d, err = c.Globals.APIClient.UpdateObservabilityCustomDashboard(context.TODO(), input)
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
			c.ItemID,
			success,
			d,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	if success {
		text.Success(out, `Removed %d dashboard item(s) from Custom Dashboard "%s" (dashboardID: %s)`, (numItems - len(d.Items)), d.Name, d.ID)
	} else {
		text.Warning(out, "dashboard (%s) has no item with ID (%s)", d.ID, c.ItemID)
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
		return di.ID == c.ItemID
	})
	input.Items = &items

	return &input
}
