package timeseries

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v16/fastly"
	gots "github.com/fastly/go-fastly/v16/fastly/ngwaf/v1/workspaces/timeseries"
)

// GetCommand calls the Fastly API to get workspace-level time series metrics.
type GetCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	from        string
	metrics     string
	workspaceID argparser.OptionalWorkspaceID

	// Optional.
	to          argparser.OptionalString
	granularity argparser.OptionalInt
}

// NewGetCommand returns a usable command registered under the parent.
func NewGetCommand(parent argparser.Registerer, g *global.Data) *GetCommand {
	c := GetCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("get", "Get workspace-level time series metrics")

	// Required.
	c.CmdClause.Flag("from", "The start of a date-time range, expressed in RFC 3339 format").Required().StringVar(&c.from)
	c.CmdClause.Flag("metrics", "Comma-separated list of metrics to be included in the timeseries. Metrics can be XSS, SQLI, HTTP404, requests_total, requests_attack, requests_total_blocked, or any custom metric").Required().StringVar(&c.metrics)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
	})

	// Optional.
	c.CmdClause.Flag("to", "The end of a date-time range, expressed in RFC 3339 format").Action(c.to.Set).StringVar(&c.to.Value)
	c.CmdClause.Flag("granularity", "Level of detail of the sample size in seconds (Default value is 86400)").Action(c.granularity.Set).IntVar(&c.granularity.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if err := c.workspaceID.Parse(); err != nil {
		return err
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := gots.GetInput{
		Start:       &c.from,
		Metrics:     &c.metrics,
		WorkspaceID: &c.workspaceID.Value,
	}

	if c.to.WasSet {
		input.End = &c.to.Value
	}
	if c.granularity.WasSet {
		input.Granularity = &c.granularity.Value
	}

	result, err := gots.Get(context.TODO(), fc, &input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, result); ok {
		return err
	}

	text.PrintWorkspaceTimeseries(out, result)
	return nil
}
