package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
)

// AggregateCommand exposes the Aggregate Stats API.
type AggregateCommand struct {
	argparser.Base

	// Optional.
	From       string `help:"Start time"`
	To         string `help:"End time"`
	By         string `default:"" enum:"minute,hour,day," help:"Aggregation period (minute/hour/day)"`
	Region     string `help:"Filter by region ('stats regions' to list)"`
	FormatFlag string `name:"format" default:"" enum:"json," hidden:"" help:"Output format (json)"`
	JSONFlag   bool   `name:"json" short:"j" help:"Render output as JSON"`
}

// Exec implements the command interface.
func (c *AggregateCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := resolveJSONFormat(&c.FormatFlag, c.JSONFlag, c.Globals); err != nil {
		return err
	}

	input := fastly.GetAggregateInput{}
	if c.By != "" {
		input.By = &c.By
	}
	if c.From != "" {
		input.From = &c.From
	}
	if c.Region != "" {
		input.Region = &c.Region
	}
	if c.To != "" {
		input.To = &c.To
	}

	var envelope statsResponse
	err := c.Globals.APIClient.GetAggregateJSON(context.TODO(), &input, &envelope)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if envelope.Status != statusSuccess {
		return fmt.Errorf("non-success response: %s", envelope.Msg)
	}

	switch c.FormatFlag {
	case "json":
		for _, block := range envelope.Data {
			if err := json.NewEncoder(out).Encode(block); err != nil {
				c.Globals.ErrLog.Add(err)
				return err
			}
		}
	default:
		writeHeader(out, envelope.Meta)
		for _, block := range envelope.Data {
			if err := fmtBlock(out, "aggregate", block); err != nil {
				c.Globals.ErrLog.Add(err)
				return err
			}
		}
	}

	return nil
}
