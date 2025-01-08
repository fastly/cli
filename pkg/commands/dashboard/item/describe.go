package item

import (
	"fmt"
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
	c.CmdClause = parent.Command("describe", "Describe a custom dashboard item").Alias("add")
	c.Globals = globals

	// Required flags
	c.CmdClause.Flag("dashboard-id", "ID of the Dashboard containing the item").Required().StringVar(&c.dashboardID) // --dashboard-id
	c.CmdClause.Flag("item-id", "ID of the Item to be described").Required().StringVar(&c.itemID)      // --item-id

	// Optional flags
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	// required
	dashboardID string
	itemID      string
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	d, err := c.Globals.APIClient.GetObservabilityCustomDashboard(input)
	if err != nil {
		return err
	}

	di, err := getItemFromDashboard(d, c.itemID)
	if err != nil {
		return err
	}

	if c.JSONOutput.Enabled {
		c.WriteJSON(out, di)
	} else {
		common.PrintItem(out, 0, di)
	}

	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetObservabilityCustomDashboardInput {
	return &fastly.GetObservabilityCustomDashboardInput{ID: &c.dashboardID}
}

func getItemFromDashboard(d *fastly.ObservabilityCustomDashboard, itemID string) (*fastly.DashboardItem, error) {
	for _, di := range d.Items {
		if di.ID == itemID {
			return &di, nil
		}
	}
	return nil, fmt.Errorf("could not find item with ID (%s) in Dashboard (%s)", itemID, d.ID)
}
