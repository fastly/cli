package path

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/paths"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// ListCommand calls the Fastly API to list paths within a routing config.
type ListCommand struct {
	argparser.Base

	// Required.
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`

	argparser.JSONOutput

	// Optional.
	Match argparser.OptionalString `help:"Filters results using the given path pattern matching strategy"`
	Path  argparser.OptionalString `help:"Filters results by path pattern"`
	Sort  argparser.OptionalString `help:"The order in which to list the results"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &paths.ListInput{
		RoutingConfigID: &c.RoutingConfigID,
	}
	if c.Match.WasSet {
		input.Match = &c.Match.Value
	}
	if c.Path.WasSet {
		input.Path = &c.Path.Value
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	cl, err := paths.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
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
