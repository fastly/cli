package stats

import (
	"encoding/json"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// RealtimeCommand exposes the Realtime Metrics API.
type RealtimeCommand struct {
	cmd.Base
	manifest manifest.Data

	formatFlag string
}

// NewRealtimeCommand is the "stats realtime" subcommand.
func NewRealtimeCommand(parent cmd.Registerer, globals *config.Data) *RealtimeCommand {
	var c RealtimeCommand
	c.Globals = globals

	c.CmdClause = parent.Command("realtime", "View realtime stats for a Fastly service")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)

	c.CmdClause.Flag("format", "Output format (json)").EnumVar(&c.formatFlag, "json")

	return &c
}

// Exec implements the command interface.
func (c *RealtimeCommand) Exec(in io.Reader, out io.Writer) error {
	service, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}

	switch c.formatFlag {
	case "json":
		if err := loopJSON(c.Globals.RTSClient, service, out); err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}

	default:
		if err := loopText(c.Globals.RTSClient, service, out); err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	}

	return nil
}

func loopJSON(client api.RealtimeStatsInterface, service string, out io.Writer) error {
	var timestamp uint64
	for {
		var envelope struct {
			Timestamp uint64            `json:"timestamp"`
			Data      []json.RawMessage `json:"data"`
		}

		err := client.GetRealtimeStatsJSON(&fastly.GetRealtimeStatsInput{
			ServiceID: service,
			Timestamp: timestamp,
		}, &envelope)
		if err != nil {
			text.Error(out, "fetching stats: %w", err)
			continue
		}
		timestamp = envelope.Timestamp

		for _, block := range envelope.Data {
			out.Write(block)
			text.Break(out)
		}
	}
}

func loopText(client api.RealtimeStatsInterface, service string, out io.Writer) error {
	var timestamp uint64
	for {
		var envelope realtimeResponse

		err := client.GetRealtimeStatsJSON(&fastly.GetRealtimeStatsInput{
			ServiceID: service,
			Timestamp: timestamp,
		}, &envelope)
		if err != nil {
			text.Error(out, "fetching stats: %w", err)
			continue
		}
		timestamp = envelope.Timestamp

		for _, block := range envelope.Data {
			agg := block.Aggregated

			// FIXME: These are heavy-handed compatibility
			// fixes for stats vs realtime, so we can use
			// fmtBlock for both.
			agg["start_time"] = block.Recorded
			delete(agg, "miss_histogram")

			if err := fmtBlock(out, service, agg); err != nil {
				text.Error(out, "formatting stats: %w", err)
				continue
			}
		}
	}
}
