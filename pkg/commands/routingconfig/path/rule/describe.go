package rule

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/paths/rules"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// DescribeCommand calls the Fastly API to describe a rule.
type DescribeCommand struct {
	argparser.Base

	// Required.
	PathID          string `required:"" help:"The Path Identifier"`
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`
	RuleID          string `required:"" help:"The Rule Identifier"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &rules.GetInput{
		PathID:          &c.PathID,
		RoutingConfigID: &c.RoutingConfigID,
		RuleID:          &c.RuleID,
	}

	d, err := rules.Get(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.PathID,
			"Routing Config ID": c.RoutingConfigID,
			"Rule ID":           c.RuleID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, d); ok {
		return err
	}

	if d != nil {
		cl := []rules.Data{*d}
		if c.Globals.Verbose() {
			printVerbose(out, cl)
		} else {
			printSummary(out, cl)
		}
	}
	return nil
}
