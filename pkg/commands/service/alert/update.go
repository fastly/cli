package alerts

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// UpdateCommand calls the Fastly API to list appropriate resources.
type UpdateCommand struct {
	argparser.Base

	// Required.
	Description  string  `required:"" help:"Additional text that is included in the alert notification."`
	DefinitionID string  `name:"id" required:"" help:"A unique identifier for a definition."`
	Metric       string  `required:"" help:"Metric name to alert on for a specific source."`
	NameFlag     string  `name:"name" required:"" help:"Name of the alert definition."`
	Period       string  `required:"" enum:"2m,3m,5m,15m,30m" hints:"2m,3m,5m,15m,30m" help:"Period of time to evaluate whether the conditions have been met. The data is polled every minute."`
	Source       string  `required:"" help:"Source where the metric comes from."`
	Threshold    float64 `required:"" help:"Threshold used to alert."`
	EType        string  `name:"type" required:"" enum:"above_threshold,all_above_threshold,below_threshold,percent_absolute,percent_decrease,percent_increase" hints:"above_threshold,all_above_threshold,below_threshold,percent_absolute,percent_decrease,percent_increase" help:"Type of strategy to use to evaluate."`

	// Optional.
	Dimensions   argparser.OptionalStringSlice `sep:"none" help:"Dimensions filters depending on the source type."`
	IgnoreBelow  argparser.OptionalFloat64     `name:"ignoreBelow" help:"IgnoreBelow is the threshold for the denominator value used in evaluations that calculate a rate or ratio. Usually used to filter out noise."`
	Integrations argparser.OptionalStringSlice `sep:"none" help:"Integrations are a list of integrations used to notify when alert fires."`
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	definition, err := c.Globals.APIClient.UpdateAlertDefinition(context.TODO(), input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, definition); ok {
		return err
	}

	definitions := []*fastly.AlertDefinition{definition}
	if c.Globals.Verbose() {
		printVerbose(out, definitions)
	} else {
		printSummary(out, definitions)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateAlertDefinitionInput {
	input := fastly.UpdateAlertDefinitionInput{
		ID:          &c.DefinitionID,
		Description: &c.Description,
		EvaluationStrategy: map[string]any{
			"type":      c.EType,
			"period":    c.Period,
			"threshold": c.Threshold,
		},
		Metric: &c.Metric,
		Name:   &c.NameFlag,
	}

	if c.IgnoreBelow.WasSet {
		input.EvaluationStrategy["ignore_below"] = c.IgnoreBelow.Value
	}

	dimensions := map[string][]string{}
	if c.Source == "origins" || c.Source == "domains" {
		var filter []string
		if c.Dimensions.WasSet {
			filter = c.Dimensions.Value
		}
		dimensions[c.Source] = filter
	}
	input.Dimensions = dimensions

	input.IntegrationIDs = []string{}
	if c.Integrations.WasSet {
		input.IntegrationIDs = c.Integrations.Value
	}

	return &input
}
