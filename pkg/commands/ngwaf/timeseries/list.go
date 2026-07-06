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
	ts "github.com/fastly/go-fastly/v16/fastly/ngwaf/v1/timeseries"
)

// ListCommand calls the Fastly API to list an account-level time series metrics.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	from    string
	metrics string

	// Optional.
	dimensions  argparser.OptionalString
	granularity argparser.OptionalInt
	to          argparser.OptionalString
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("list", "List account-level time series metrics")

	// Required.
	c.CmdClause.Flag("from", "The start of a date-time range, expressed in RFC 3339 format").Required().StringVar(&c.from)
	c.CmdClause.Flag("metrics", "Comma-separated list of metrics to be included in the timeseries. Metrics can be XSS, SQLI, HTTP404, requests_total, requests_attack, requests_total_blocked, or any custom metric").Required().StringVar(&c.metrics)

	// Optional.
	c.CmdClause.Flag("dimensions", "Comma separated list of grouping dimensions to be included in the timeseries. Allowed values are workspaces and time. (Default value is time)").Action(c.dimensions.Set).StringVar(&c.dimensions.Value)
	c.CmdClause.Flag("granularity", "Level of detail of the sample size in seconds. (Default value is 86400)").Action(c.granularity.Set).IntVar(&c.granularity.Value)
	c.CmdClause.Flag("to", "The end of a date-time range, expressed in RFC 3339 format").Action(c.to.Set).StringVar(&c.to.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := ts.ListInput{
		From:    &c.from,
		Metrics: &c.metrics,
	}

	if c.dimensions.WasSet {
		input.Dimensions = &c.dimensions.Value
	}
	if c.granularity.WasSet {
		input.Granularity = &c.granularity.Value
	}
	if c.to.WasSet {
		input.To = &c.to.Value
	}

	result, err := ts.List(context.TODO(), fc, &input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, result); ok {
		return err
	}

	text.PrintTimeseries(out, result)
	return nil
}
