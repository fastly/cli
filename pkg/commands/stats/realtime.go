package stats

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RealtimeCommand exposes the Realtime Metrics API.
type RealtimeCommand struct {
	argparser.Base

	formatFlag  string
	serviceName argparser.OptionalServiceNameID
}

// NewRealtimeCommand is the "stats realtime" subcommand.
func NewRealtimeCommand(parent argparser.Registerer, g *global.Data) *RealtimeCommand {
	var c RealtimeCommand
	c.Globals = g

	c.CmdClause = parent.Command("realtime", "View realtime stats for a Fastly service")
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

	c.CmdClause.Flag("format", "Output format (json)").EnumVar(&c.formatFlag, "json")

	return &c
}

// Exec implements the command interface.
func (c *RealtimeCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	switch c.formatFlag {
	case "json":
		if err := loopJSON(c.Globals.RTSClient, serviceID, out); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
			return err
		}

	default:
		if err := loopText(c.Globals.RTSClient, serviceID, out); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Service ID": serviceID,
			})
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

		for _, data := range envelope.Data {
			_, err = out.Write(data)
			if err != nil {
				return fmt.Errorf("error: unable to write data to stdout: %w", err)
			}
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
