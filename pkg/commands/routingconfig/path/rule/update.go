package rule

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/domainmanagement/v1/routingconfigs/paths/rules"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a rule.
type UpdateCommand struct {
	argparser.Base

	// Required.
	file            string
	pathID          string
	routingConfigID string
	ruleID          string
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a rule")

	// Required.
	c.CmdClause.Flag("file", "Path to a JSON file describing the fields to update, e.g. {\"action\": {\"type\": \"service\", \"value\": \"...\"}, \"conditions\": [...]}. A field omitted from the JSON is left unchanged; set \"conditions\" to an empty array to turn the rule into the default (catch-all) rule for its path").Required().StringVar(&c.file)
	c.CmdClause.Flag("path-id", "The Path Identifier").Required().StringVar(&c.pathID)
	c.CmdClause.Flag("routing-config-id", "The Routing Config Identifier").Required().StringVar(&c.routingConfigID)
	c.CmdClause.Flag("rule-id", "The Rule Identifier").Required().StringVar(&c.ruleID)

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	input := &rules.UpdateInput{
		PathID:          &c.pathID,
		RoutingConfigID: &c.routingConfigID,
		RuleID:          &c.ruleID,
	}

	if err := readRuleFile(c.file, input); err != nil {
		return err
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := rules.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.pathID,
			"Routing Config ID": c.routingConfigID,
			"Rule ID":           c.ruleID,
		})
		return err
	}

	text.Success(out, "Updated rule (rule-id: %s)", d.RuleID)
	return nil
}
