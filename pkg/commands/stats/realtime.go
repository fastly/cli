package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// RealtimeCommand exposes the Realtime Metrics API.
type RealtimeCommand struct {
	argparser.Base

	// Optional.
	ServiceID   *string                         `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName argparser.OptionalServiceNameID `help:"The name of the service"`
	FormatFlag  string                          `name:"format" default:"" enum:"json," hidden:"" help:"Output format (json)"`
	JSONFlag    bool                            `name:"json" short:"j" help:"Render output as JSON"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *RealtimeCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec implements the command interface.
func (c *RealtimeCommand) Exec(_ io.Reader, out io.Writer) error {
	if err := resolveJSONFormat(&c.FormatFlag, c.JSONFlag, c.Globals); err != nil {
		return err
	}

	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	switch c.FormatFlag {
	case "json":
		err := loopJSON(c.Globals.RTSClient, serviceID, out)
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err

	default:
		err := loopText(c.Globals.RTSClient, serviceID, out)
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID": serviceID,
		})
		return err
	}
}

func loopJSON(client api.RealtimeStatsInterface, service string, out io.Writer) error {
	var timestamp uint64
	for {
		var envelope struct {
			Timestamp uint64            `json:"timestamp"`
			Data      []json.RawMessage `json:"data"`
		}

		err := client.GetRealtimeStatsJSON(context.TODO(), &fastly.GetRealtimeStatsInput{
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

		err := client.GetRealtimeStatsJSON(context.TODO(), &fastly.GetRealtimeStatsInput{
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
