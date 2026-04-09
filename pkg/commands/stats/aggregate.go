package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v14/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// AggregateCommand exposes the Aggregate Stats API.
type AggregateCommand struct {
	argparser.Base

	by         string
	formatFlag string
	from       string
	jsonFlag   bool
	region     string
	to         string
}

// NewAggregateCommand is the "stats aggregate" subcommand.
func NewAggregateCommand(parent argparser.Registerer, g *global.Data) *AggregateCommand {
	var c AggregateCommand
	c.Globals = g

	c.CmdClause = parent.Command("aggregate", "View aggregated stats across all services")

	// Optional.
	c.CmdClause.Flag("from", "Start time").StringVar(&c.from)
	c.CmdClause.Flag("to", "End time").StringVar(&c.to)
	c.CmdClause.Flag("by", "Aggregation period (minute/hour/day)").EnumVar(&c.by, "minute", "hour", "day")
	c.CmdClause.Flag("region", "Filter by region ('stats regions' to list)").StringVar(&c.region)
	c.CmdClause.Flag("format", "Output format (json)").Hidden().EnumVar(&c.formatFlag, "json")
	c.CmdClause.Flag("json", argparser.FlagJSONDesc).Short('j').BoolVar(&c.jsonFlag)

	return &c
}

// Exec implements the command interface.
func (c *AggregateCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := resolveJSONFormat(&c.formatFlag, c.jsonFlag, c.Globals); err != nil {
		return err
	}

	input := fastly.GetAggregateInput{}
	if c.by != "" {
		input.By = &c.by
	}
	if c.from != "" {
		input.From = &c.from
	}
	if c.region != "" {
		input.Region = &c.region
	}
	if c.to != "" {
		input.To = &c.to
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

	switch c.formatFlag {
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
