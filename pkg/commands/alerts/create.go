package alerts

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/go-fastly/v10/fastly"
)

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("create", "Create Alerts")

	// Required.
	c.CmdClause.Flag("description", "Additional text that is included in the alert notification.").Required().StringVar(&c.description)
	c.CmdClause.Flag("metric", "Metric name to alert on for a specific source.").Required().StringVar(&c.metric)
	c.CmdClause.Flag("name", "Name of the alert definition.").Required().StringVar(&c.name)
	c.CmdClause.Flag("period", "Period of time to evaluate whether the conditions have been met. The data is polled every minute.").Required().HintOptions(evaluationPeriod...).EnumVar(&c.period, evaluationPeriod...)
	c.CmdClause.Flag("source", "Source where the metric comes from.").Required().StringVar(&c.source)
	c.CmdClause.Flag("threshold", "Threshold used to alert.").Required().Float64Var(&c.threshold)
	c.CmdClause.Flag("type", "Type of strategy to use to evaluate.").Required().HintOptions(evaluationType...).EnumVar(&c.eType, evaluationType...)

	// Optional.
	c.CmdClause.Flag("dimensions", "Dimensions filters depending on the source type.").Action(c.dimensions.Set).StringsVar(&c.dimensions.Value)
	c.CmdClause.Flag("ignoreBelow", "IgnoreBelow is the threshold for the denominator value used in evaluations that calculate a rate or ratio. Usually used to filter out noise.").Action(c.ignoreBelow.Set).Float64Var(&c.ignoreBelow.Value)
	c.CmdClause.Flag("integrations", "Integrations are a list of integrations used to notify when alert fires.").Action(c.integrations.Set).StringsVar(&c.integrations.Value)
	c.RegisterFlagBool(c.JSONFlag())                                                                                                   // --json
	c.CmdClause.Flag(argparser.FlagServiceIDName, "ServiceID of the definition").Action(c.serviceID.Set).StringVar(&c.serviceID.Value) // --service-id

	return &c
}

// CreateCommand calls the Fastly API to list appropriate resources.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	description string
	eType       string
	metric      string
	name        string
	period      string
	source      string
	threshold   float64

	dimensions   argparser.OptionalStringSlice
	ignoreBelow  argparser.OptionalFloat64
	integrations argparser.OptionalStringSlice
	serviceID    argparser.OptionalString
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()
	definition, err := c.Globals.APIClient.CreateAlertDefinition(input)
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
func (c *CreateCommand) constructInput() *fastly.CreateAlertDefinitionInput {
	input := fastly.CreateAlertDefinitionInput{
		Description: &c.description,
		EvaluationStrategy: map[string]any{
			"type":      c.eType,
			"period":    c.period,
			"threshold": c.threshold,
		},
		Metric: &c.metric,
		Name:   &c.name,
		Source: &c.source,
	}

	if c.ignoreBelow.WasSet {
		input.EvaluationStrategy["ignore_below"] = c.ignoreBelow.Value
	}

	if c.serviceID.WasSet {
		input.ServiceID = &c.serviceID.Value
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
