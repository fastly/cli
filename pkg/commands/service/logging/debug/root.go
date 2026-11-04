package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/text"
)

// batch wraps errors for sending to the output loop.
type batch struct {
	Errors []fastly.LoggingEndpointError
}

// Command is the command for streaming logging endpoint errors.
type Command struct {
	argparser.Base

	ServiceIDFlag   *string                         `name:"service-id" short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
	ServiceName     argparser.OptionalServiceNameID `help:"The name of the service"`
	serviceID       string
	From            uint64 `help:"From time, in Unix seconds"`
	To              uint64 `help:"To time, in Unix seconds"`
	Filter          string `help:"Filter errors by logging endpoint name"`
	PrintTimestamps bool   `name:"timestamps" help:"Print full timestamps instead of compact time"`
	JSONOutput      bool   `name:"json" help:"Output error stream as JSON"`
	batchCh         chan batch
	dieCh           chan struct{}
	doneCh          chan struct{}
}

// CommandName is the string to be used to invoke this command.
const CommandName = "debug"

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *Command) AfterApply() error {
	if c.ServiceIDFlag != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceIDFlag
	}
	return nil
}

// Exec implements the command interface.
func (c *Command) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.serviceID = serviceID

	c.dieCh = make(chan struct{})
	c.batchCh = make(chan batch)
	c.doneCh = make(chan struct{})

	text.Info(out, "Streaming logging endpoint errors for service %s\n\n", c.serviceID)

	failure := make(chan error)
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start the output loop.
	go c.outputLoop(out)

	// Start streaming the errors.
	go func() {
		failure <- c.stream(out)
	}()

	select {
	case asyncErr := <-failure:
		close(c.dieCh)
		return asyncErr
	case <-c.doneCh:
		return nil
	case <-sigs:
		close(c.dieCh)
	}

	return nil
}

// stream fetches error data from the API and sends it to the output loop.
func (c *Command) stream(out io.Writer) error {
	var curWindow *uint64
	if c.From != 0 {
		curWindow = &c.From
	}
	var toWindow *uint64
	if c.To != 0 {
		toWindow = &c.To
	}

	// Prepare filter slice
	var filter []string
	if c.Filter != "" {
		filter = []string{c.Filter}
	}

	ctx := context.Background()

	for {
		// Check if we've passed the "to" requirement.
		if toWindow != nil && curWindow != nil && *curWindow > *toWindow {
			text.Info(out, "Reached window: %v which is newer than the requested 'to': %v", *curWindow, *toWindow)
			close(c.doneCh)
			break
		}

		// Use go-fastly to fetch logging endpoint errors
		resp, err := c.Globals.APIClient.GetLoggingEndpointErrors(ctx, &fastly.LoggingEndpointErrorsInput{
			ServiceID: c.serviceID,
			From:      curWindow,
			To:        toWindow,
			Filter:    filter,
		})
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("unable to fetch logging endpoint errors: %w", err)
		}

		// Send errors to the output loop
		if len(resp.Errors) > 0 {
			c.batchCh <- batch{Errors: resp.Errors}
		}

		// Check for next link to continue streaming
		if resp.NextFrom != "" {
			// Parse the next link value (it's already the from parameter value)
			nextFrom, err := strconv.ParseUint(resp.NextFrom, 10, 64)
			if err != nil {
				c.Globals.ErrLog.AddWithContext(err, map[string]any{
					"NextFrom": resp.NextFrom,
				})
				text.Error(out, "error parsing next from")
				continue
			}
			curWindow = &nextFrom
		} else {
			// No next link, we're done
			close(c.doneCh)
			break
		}
	}
	return nil
}

// outputLoop processes the errors out of band from the request/response loop.
func (c *Command) outputLoop(out io.Writer) {
	for {
		select {
		case <-c.dieCh:
			return
		case batch := <-c.batchCh:
			c.printErrors(out, batch.Errors)
		}
	}
}

// printErrors prints error entries.
func (c *Command) printErrors(out io.Writer, errors []fastly.LoggingEndpointError) {
	if len(errors) == 0 {
		return
	}

	if c.JSONOutput {
		// Output as JSON array
		encoder := json.NewEncoder(out)
		for _, e := range errors {
			if err := encoder.Encode(e); err != nil {
				c.Globals.ErrLog.Add(err)
			}
		}
	} else {
		// Find the longest endpoint name in this batch for dynamic width
		maxEndpointLen := 0
		for _, e := range errors {
			if len(e.Endpoint) > maxEndpointLen {
				maxEndpointLen = len(e.Endpoint)
			}
		}

		// Human-readable format - match log-tail style
		for _, e := range errors {
			// Format timestamp
			// #nosec G115 -- Timestamp is in microseconds, multiplication by 1000 for nanoseconds is safe for reasonable time values
			timestamp := time.Unix(0, int64(e.Timestamp)*1000) // Convert microseconds to nanoseconds
			var timeStr string
			if c.PrintTimestamps {
				// Full timestamp with --timestamps flag
				timeStr = timestamp.UTC().Format(time.RFC3339)
			} else {
				// Compact time by default (HH:MM:SS)
				timeStr = timestamp.UTC().Format("15:04:05")
			}

			// Extract clean error message from details JSON if present
			errorSummary := e.Message
			if e.Details != "" {
				var detailsJSON map[string]interface{}
				if err := json.Unmarshal([]byte(e.Details), &detailsJSON); err == nil {
					// Try to extract a cleaner error message
					if errorMsg, ok := detailsJSON["error"].(string); ok {
						// Simplify common error patterns
						errorMsg = strings.TrimPrefix(errorMsg, "non-temporary request err: ")
						errorMsg = strings.TrimPrefix(errorMsg, "temporary request err: ")
						errorSummary = errorMsg
					}
				}
			}

			// Format: time | endpoint | message
			fmt.Fprintf(out, "%s | %-*s | %s\n", timeStr, maxEndpointLen, e.Endpoint, errorSummary)
		}
	}

	// Flush output immediately
	if f, ok := out.(*os.File); ok {
		_ = f.Sync()
	}
}
