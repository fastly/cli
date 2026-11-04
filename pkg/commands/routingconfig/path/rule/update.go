package rule

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/paths/rules"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a rule.
type UpdateCommand struct {
	argparser.Base

	// Required.
	File            string `required:"" help:"Path to a JSON file describing the fields to update, e.g. {\"action\": {\"type\": \"service\", \"value\": \"...\"}, \"conditions\": [...]}. A field omitted from the JSON is left unchanged; set \"conditions\" to an empty array to turn the rule into the default (catch-all) rule for its path"`
	PathID          string `required:"" help:"The Path Identifier"`
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
	RuleID          string `required:"" help:"The Rule Identifier"`
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &rules.UpdateInput{
		PathID:          &c.PathID,
		RoutingConfigID: &c.RoutingConfigID,
		RuleID:          &c.RuleID,
	}

	if err := readRuleFile(c.File, input); err != nil {
		return err
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := rules.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.PathID,
			"Routing Config ID": c.RoutingConfigID,
			"Rule ID":           c.RuleID,
		})
		return err
	}

	text.Success(out, "Updated rule (rule-id: %s)", d.RuleID)
	return nil
}
