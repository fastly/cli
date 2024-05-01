package alerts

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update Alerts")
	c.CmdClause.Flag("id", "Alphanumeric string identifying an Alert definition").Required().StringVar(&c.definitionID)
	c.CmdClause.Flag("name", "Name of the alert definition").Required().StringVar(&c.name)
	c.CmdClause.Flag("description", "Description of the alert definition").Required().StringVar(&c.description)
	c.CmdClause.Flag("metric", "Metric to alert on").Required().StringVar(&c.metric)
	c.CmdClause.Flag("source", "Source to get the metric from").Required().StringVar(&c.source)
	c.CmdClause.Flag("type", "Type of evaluation strategy, one of [all_above_threshold, above_threshold, below_threshold, percent_absolute, percent_decrease, percent_increase]").Required().StringVar(&c.eType)
	c.CmdClause.Flag("period", "Period of time to evaluate, one of [2m, 3m, 5m, 15m, 30m]").Required().StringVar(&c.period)
	c.CmdClause.Flag("threshold", "Threshold use to compare during evaluation").Required().Float64Var(&c.threshold)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("ignoreBelow", "IgnoreBelow use to discard data that are below this threshold").Action(c.ignoreBelow.Set).Float64Var(&c.ignoreBelow.Value)
	c.CmdClause.Flag("dimensions", "Dimensions to filter on origins or domains").Action(c.dimensions.Set).StringsVar(&c.dimensions.Value)
	c.CmdClause.Flag("integrations", "Integrations used to notify when an alert is firing or resolving").Action(c.integrations.Set).StringsVar(&c.integrations.Value)

	return &c
}

// UpdateCommand calls the Fastly API to list appropriate resources.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	definitionID string
	name         string
	description  string
	metric       string
	source       string
	eType        string
	period       string
	threshold    float64

	ignoreBelow  argparser.OptionalFloat64
	dimensions   argparser.OptionalStringSlice
	integrations argparser.OptionalStringSlice
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	definition, err := c.Globals.APIClient.UpdateAlertDefinition(input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, definition); ok {
		return err
	}

	definitions := []*fastly.AlertDefinition{definition}
	if c.Globals.Verbose() {
		printVerbose(out, nil, definitions)
	} else {
		printSummary(out, nil, definitions)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) constructInput() *fastly.UpdateAlertDefinitionInput {
	input := fastly.UpdateAlertDefinitionInput{
		ID:          &c.definitionID,
		Description: &c.description,
		EvaluationStrategy: map[string]any{
			"type":      c.eType,
			"period":    c.period,
			"threshold": c.threshold,
		},
		Metric: &c.metric,
		Name:   &c.name,
	}

	if c.ignoreBelow.WasSet {
		input.EvaluationStrategy["ignore_below"] = c.ignoreBelow.Value
	}

	dimensions := map[string][]string{}
	if c.source == "origins" || c.source == "domains" {
		var filter []string
		if c.dimensions.WasSet {
			filter = c.dimensions.Value
		}
		dimensions[c.source] = filter
	}
	input.Dimensions = dimensions

	input.IntegrationIDs = []string{}
	if c.integrations.WasSet {
		input.IntegrationIDs = c.integrations.Value
	}

	return &input
}
