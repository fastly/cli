package stats

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/fastly"
)

// RealtimeCommand exposes the Realtime Metrics API.
type RealtimeCommand struct {
	common.Base
	manifest manifest.Data

	formatFlag string
}

func (c *RealtimeCommand) checkArgs() error {
	switch c.formatFlag {
	case "json", "":
		// OK
	default:
		return fmt.Errorf("unsupported value for 'format': %q", c.formatFlag)
	}

	return nil
}

// NewRealtimeCommand is the "stats realtime" subcommand.
func NewRealtimeCommand(parent common.Registerer, globals *config.Data) *RealtimeCommand {
	var c RealtimeCommand
	c.Globals = globals

	c.CmdClause = parent.Command("realtime", "Query realtime stats")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Required().StringVar(&c.manifest.Flag.ServiceID)

	c.CmdClause.Flag("format", "Output format (json)").StringVar(&c.formatFlag)

	return &c
}

// Exec implements the command interface.
func (c *RealtimeCommand) Exec(in io.Reader, out io.Writer) error {
	if err := c.checkArgs(); err != nil {
		return fmt.Errorf("realtime: %w", err)
	}

	service, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}

	switch c.formatFlag {
	case "json":
		if err := loopJSON(c.Globals.RTSClient, service, out); err != nil {
			return err
		}

	default:
		if err := loopText(c.Globals.RTSClient, service, out); err != nil {
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
			Service:   service,
			Timestamp: timestamp,
		}, &envelope)
		if err != nil {
			fmt.Fprintf(out, "ERROR: fetching stats: %s", err)
			continue
		}
		timestamp = envelope.Timestamp

		for _, block := range envelope.Data {
			out.Write(block)
			out.Write([]byte{'\n'})
		}
	}

	return nil
}

func loopText(client api.RealtimeStatsInterface, service string, out io.Writer) error {
	var timestamp uint64
	for {
		var envelope realtimeResponse

		err := client.GetRealtimeStatsJSON(&fastly.GetRealtimeStatsInput{
			Service:   service,
			Timestamp: timestamp,
		}, &envelope)
		if err != nil {
			fmt.Fprintf(out, "ERROR: fetching stats: %s", err)
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
				fmt.Fprintf(out, "ERROR: formatting stats: %s", err)
				continue
			}
		}
	}

	return nil
}
