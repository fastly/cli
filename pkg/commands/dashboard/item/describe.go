package item

import (
	"context"
	"fmt"
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

	// required
	DashboardID string `required:"" help:"ID of the Dashboard containing the item"`
	ItemID      string `required:"" help:"ID of the Item to be described"`

	// Optional flags
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	d, err := c.Globals.APIClient.GetObservabilityCustomDashboard(context.TODO(), input)
	if err != nil {
		return err
	}

	di, err := getItemFromDashboard(context.TODO(), d, c.ItemID)
	if err != nil {
		return err
	}

	if c.JSONOutput.Enabled {
		_, err := c.WriteJSON(out, di)
		if err != nil {
			return err
		}
	} else {
		printer.PrintItem(out, 0, di)
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetObservabilityCustomDashboardInput {
	return &fastly.GetObservabilityCustomDashboardInput{ID: &c.DashboardID}
}

func getItemFromDashboard(_ context.Context, d *fastly.ObservabilityCustomDashboard, itemID string) (*fastly.DashboardItem, error) {
	for _, di := range d.Items {
		if di.ID == itemID {
			return &di, nil
		}
	}
	return nil, fmt.Errorf("could not find item with ID (%s) in Dashboard (%s)", itemID, d.ID)
}
