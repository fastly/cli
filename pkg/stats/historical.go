package stats

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

const statusSuccess = "success"

// HistoricalCommand exposes the Historical Stats API.
type HistoricalCommand struct {
	cmd.Base
	manifest manifest.Data

	Input      fastly.GetStatsInput
	formatFlag string
}

// NewHistoricalCommand is the "stats historical" subcommand.
func NewHistoricalCommand(parent cmd.Registerer, globals *config.Data) *HistoricalCommand {
	var c HistoricalCommand
	c.Globals = globals

	c.CmdClause = parent.Command("historical", "View historical stats for a Fastly service")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	c.CmdClause.Flag("from", "From time, accepted formats at https://fastly.dev/reference/api/metrics-stats/historical-stats").StringVar(&c.Input.From)
	c.CmdClause.Flag("to", "To time").StringVar(&c.Input.To)
	c.CmdClause.Flag("by", "Aggregation period (minute/hour/day)").EnumVar(&c.Input.By, "minute", "hour", "day")
	c.CmdClause.Flag("region", "Filter by region ('stats regions' to list)").StringVar(&c.Input.Region)

	c.CmdClause.Flag("format", "Output format (json)").EnumVar(&c.formatFlag, "json")

	return &c
}

// Exec implements the command interface.
func (c *HistoricalCommand) Exec(in io.Reader, out io.Writer) error {
	service, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = service

	var envelope statsResponse
	err := c.Globals.Client.GetStatsJSON(&c.Input, &envelope)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if envelope.Status != statusSuccess {
		return fmt.Errorf("non-success response: %s", envelope.Msg)
	}

	switch c.formatFlag {
	case "json":
		err := writeBlocksJSON(out, service, envelope.Data)
		if err != nil {
			c.Globals.ErrLog.Add(err)
		}

	default:
		writeHeader(out, envelope.Meta)
		err := writeBlocks(out, service, envelope.Data)
		if err != nil {
			c.Globals.ErrLog.Add(err)
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

func writeBlocksJSON(out io.Writer, service string, blocks []statsResponseData) error {
	for _, block := range blocks {
		if err := json.NewEncoder(out).Encode(block); err != nil {
			return err
		}
	}

	return nil
}
