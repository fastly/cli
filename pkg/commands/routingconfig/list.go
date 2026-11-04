package routingconfig

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// ListCommand calls the Fastly API to list routing configs.
type ListCommand struct {
	argparser.Base

	// Optional.
	argparser.JSONOutput

	Sort  argparser.OptionalString      `help:"The order in which to list the results"`
	State argparser.OptionalStringSlice `sep:"none" help:"Filter results by lifecycle state. Set flag multiple times to filter by multiple states"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &routingconfigs.ListInput{}

	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}
	if c.State.WasSet {
		input.State = c.State.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	cl, err := routingconfigs.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Sort":  c.Sort.Value,
			"State": c.State.Value,
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
