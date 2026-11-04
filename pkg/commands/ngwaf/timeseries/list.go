package timeseries

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"
	ts "github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/timeseries"
)

// ListCommand calls the Fastly API to list an account-level time series metrics.
type ListCommand struct {
	argparser.Base

	// Required.
	From    string `required:"" help:"The start of a date-time range, expressed in RFC 3339 format"`
	Metrics string `required:"" help:"Comma-separated list of metrics to be included in the timeseries. Metrics can be XSS, SQLI, HTTP404, requests_total, requests_attack, requests_total_blocked, or any custom metric"`

	// Optional.
	Dimensions  argparser.OptionalString `help:"Comma separated list of grouping dimensions to be included in the timeseries. Allowed values are workspaces and time. (Default value is time)"`
	Granularity argparser.OptionalInt    `help:"Level of detail of the sample size in seconds. (Default value is 86400)"`
	To          argparser.OptionalString `help:"The end of a date-time range, expressed in RFC 3339 format"`

	argparser.JSONOutput
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
		From:    &c.From,
		Metrics: &c.Metrics,
	}

	if c.Dimensions.WasSet {
		input.Dimensions = &c.Dimensions.Value
	}
	if c.Granularity.WasSet {
		input.Granularity = &c.Granularity.Value
	}
	if c.To.WasSet {
		input.To = &c.To.Value
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
