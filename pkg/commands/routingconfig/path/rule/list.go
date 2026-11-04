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

// ListCommand calls the Fastly API to list rules within a path.
type ListCommand struct {
	argparser.Base

	// Required.
	PathID          string `required:"" help:"The Path Identifier"`
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`

	argparser.JSONOutput

	// Optional.
	Sort argparser.OptionalString `help:"The order in which to list the results"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &rules.ListInput{
		PathID:          &c.PathID,
		RoutingConfigID: &c.RoutingConfigID,
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	cl, err := rules.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path ID":           c.PathID,
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, cl); ok {
		return err
	}

	if c.Globals.Verbose() {
		printVerbose(out, cl)
	} else {
		printSummary(out, cl)
	}
	return nil
}
