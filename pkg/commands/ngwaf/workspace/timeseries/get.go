package timeseries

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v19/fastly"
	gots "github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/workspaces/timeseries"
)

// GetCommand calls the Fastly API to get workspace-level time series metrics.
type GetCommand struct {
	argparser.Base

	// Required.
	From        string                        `required:"" help:"The start of a date-time range, expressed in RFC 3339 format"`
	Metrics     string                        `required:"" help:"Comma-separated list of metrics to be included in the timeseries. Metrics can be XSS, SQLI, HTTP404, requests_total, requests_attack, requests_total_blocked, or any custom metric"`
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	To          argparser.OptionalString `help:"The end of a date-time range, expressed in RFC 3339 format"`
	Granularity argparser.OptionalInt    `help:"Level of detail of the sample size in seconds (Default value is 86400)"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := gots.GetInput{
		Start:       &c.From,
		Metrics:     &c.Metrics,
		WorkspaceID: &c.WorkspaceID.Value,
	}

	if c.To.WasSet {
		input.End = &c.To.Value
	}
	if c.Granularity.WasSet {
		input.Granularity = &c.Granularity.Value
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
