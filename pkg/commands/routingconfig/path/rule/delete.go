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

// DeleteCommand calls the Fastly API to delete a rule.
type DeleteCommand struct {
	argparser.Base

	// Required.
	PathID          string `required:"" help:"The Path Identifier"`
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
	RuleID          string `required:"" help:"The Rule Identifier"`
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &rules.DeleteInput{
		PathID:          &c.PathID,
		RoutingConfigID: &c.RoutingConfigID,
		RuleID:          &c.RuleID,
	}

	err := rules.Delete(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.PathID,
			"Routing Config ID": c.RoutingConfigID,
			"Rule ID":           c.RuleID,
		})
		return err
	}

	text.Success(out, "Deleted rule (rule-id: %s)", c.RuleID)
	return nil
}
