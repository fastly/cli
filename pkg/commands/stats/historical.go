package stats

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

const statusSuccess = "success"

// HistoricalCommand exposes the Historical Stats API.
type HistoricalCommand struct {
	argparser.Base

	by          string
	formatFlag  string
	from        string
	region      string
	serviceName argparser.OptionalServiceNameID
	to          string
}

// NewHistoricalCommand is the "stats historical" subcommand.
func NewHistoricalCommand(parent argparser.Registerer, g *global.Data) *HistoricalCommand {
	var c HistoricalCommand
	c.Globals = g

	c.CmdClause = parent.Command("historical", "View historical stats for a Fastly service")
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})

	c.CmdClause.Flag("from", "From time, accepted formats at https://fastly.dev/reference/api/metrics-stats/historical-stats").StringVar(&c.from)
	c.CmdClause.Flag("to", "To time").StringVar(&c.to)
	c.CmdClause.Flag("by", "Aggregation period (minute/hour/day)").EnumVar(&c.by, "minute", "hour", "day")
	c.CmdClause.Flag("region", "Filter by region ('stats regions' to list)").StringVar(&c.region)

	c.CmdClause.Flag("format", "Output format (json)").EnumVar(&c.formatFlag, "json")

	return &c
}

// Exec implements the command interface.
func (c *HistoricalCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	input := fastly.GetStatsInput{
		Service: fastly.ToPointer(serviceID),
	}
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
	err = c.Globals.APIClient.GetStatsJSON(&input, &envelope)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}

	if envelope.Status != statusSuccess {
		return fmt.Errorf("non-success response: %s", envelope.Msg)
	}

	switch c.formatFlag {
	case "json":
		err := writeBlocksJSON(out, serviceID, envelope.Data)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
		}

	default:
		writeHeader(out, envelope.Meta)
		err := writeBlocks(out, serviceID, envelope.Data)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
		}
	}

	return nil
}

func writeHeader(out io.Writer, meta statsResponseMeta) {
	fmt.Fprintf(out, "From: %s\n", meta.From)
	fmt.Fprintf(out, "To: %s\n", meta.To)
	fmt.Fprintf(out, "By: %s\n", meta.By)
	fmt.Fprintf(out, "Region: %s\n", meta.Region)
	fmt.Fprintf(out, "---\n")
}

func writeBlocks(out io.Writer, service string, blocks []statsResponseData) error {
	for _, block := range blocks {
		if err := fmtBlock(out, service, block); err != nil {
			return err
		}
	}

	return nil
}

func writeBlocksJSON(out io.Writer, _ string, blocks []statsResponseData) error {
	for _, block := range blocks {
		if err := json.NewEncoder(out).Encode(block); err != nil {
			return err
		}
	}

	return nil
}
