package logtail

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tomnomnom/linkheader"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/debug"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base

	Input       fastly.CreateManagedLoggingInput
	batchCh     chan Batch // send batches to output loop
	cfg         cfg
	dieCh       chan struct{} // channel to end output/printing
	doneCh      chan struct{} // channel to signal we've reached the end of the run
	hClient     *http.Client  // TODO: this will go away when GET is in go-fastly
	serviceName argparser.OptionalServiceNameID
	token       string // TODO: this will go away when GET is in go-fastly
}

// CommandName is the string to be used to invoke this command.
const CommandName = "log-tail"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(CommandName, "Tail Compute logs")
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
	c.CmdClause.Flag("from", "From time, in Unix seconds").Int64Var(&c.cfg.from)
	c.CmdClause.Flag("to", "To time, in Unix seconds").Int64Var(&c.cfg.to)
	c.CmdClause.Flag("sort-buffer", "Duration of sort buffer for received logs").Default("1s").DurationVar(&c.cfg.sortBuffer)
	c.CmdClause.Flag("search-padding", "Time beyond from/to to consider in searches").Default("2s").DurationVar(&c.cfg.searchPadding)
	c.CmdClause.Flag("stream", "Output: stdout, stderr, both (default)").StringVar(&c.cfg.stream)
	c.CmdClause.Flag("timestamps", "Print timestamps with logs").BoolVar(&c.cfg.printTimestamps)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	c.Input.ServiceID = serviceID

	c.Input.Kind = fastly.ManagedLoggingInstanceOutput
	endpoint, _ := c.Globals.APIEndpoint()
	c.cfg.path = fmt.Sprintf("%s/service/%s/log_stream/managed/instance_output", endpoint, c.Input.ServiceID)

	c.dieCh = make(chan struct{})
	c.batchCh = make(chan Batch)
	c.doneCh = make(chan struct{})

	c.hClient = http.DefaultClient
	c.token, _ = c.Globals.Token()

	// Adjust the from/to times if they are
	// defined. We adjust the times based on searchPadding.
	c.adjustTimes()

	// Enable managed logging if not already enabled.
	if err := c.enableManagedLogging(out); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	failure := make(chan error)
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start the output loop.
	go c.outputLoop(out)

	// Start tailing the logs.
	go func() {
		failure <- c.tail(out)
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

// Tail starts the virtual tail process. Tail fetches data from the eventbuffer
// API. It hands off the requested logs to the outputloop for the actual
// printing.
func (c *RootCommand) tail(out io.Writer) error {
	// Start this with --from and --to if set.
	curWindow := c.cfg.from
	toWindow := c.cfg.to

	// Start the loop with an initial address to query.
	path, err := makeNewPath(c.cfg.path, curWindow, "")
	if err != nil {
		return err
	}

	// lastBatchID keeps the last successfully read Batch.ID in case we need
	// re-request on failure.
	var lastBatchID string

	for {
		// Check to see if we already passed the "to" requirement.
		if toWindow != 0 && curWindow > toWindow {
			text.Info(out, "Reached window: %v which is newer than the requested 'to': %v", curWindow, toWindow)
			// We are done, but we still want printing to finish.
			close(c.doneCh)
			break
		}

		req, err := http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				http.MethodGet: path,
			})
			return fmt.Errorf("unable to create new request: %w", err)
		}
		req.Header.Add("Fastly-Key", c.token)

		resp, err := c.doReq(req)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("unable to execute request: %w", err)
		}

		// Check that our request was successful. If the server is
		// having trouble, retry after waiting for some time.
		if resp.StatusCode != http.StatusOK {
			// If the response was a 404, the from time was
			// not valid, give them an error stating this and exit.
			if resp.StatusCode == http.StatusNotFound &&
				c.cfg.from != 0 {
				return fmt.Errorf("specified 'from' time %d not found, either too far in the past or future", c.cfg.from)
			}

			// In an effort to clean up the output, do not print on
			// 503's.
			if resp.StatusCode != http.StatusServiceUnavailable {
				text.Warning(out, "non-200 resp %d", resp.StatusCode)
			}

			// Reuse the connection for the retry, or cleanup in the
			// case of Exit.
			_, _ = io.Copy(io.Discard, resp.Body)
			err := resp.Body.Close()
			if err != nil {
				c.Globals.ErrLog.Add(err)
			}

			// Try the response again after a 1 second wait.
			if resp.StatusCode/100 == 5 && resp.StatusCode != 501 ||
				resp.StatusCode == http.StatusTooManyRequests {
				time.Sleep(1 * time.Second)
				continue
			}

			// Failing at this point is unrecoverable.
			return fmt.Errorf("unrecoverable error, response code: %d", resp.StatusCode)
		}

		// Read and parse response, send batches to the output loop.
		scanner := bufio.NewScanner(resp.Body)

		// Use a 10MB buffer for the bufio scanner, as we don't know
		// how big some of the responses will be.
		const tmb = 10 << 20
		buf := make([]byte, tmb)
		scanner.Buffer(buf, tmb)

		for scanner.Scan() {
			// Scan one line at a time, and get only one batch
			// at a time.
			b := scanner.Bytes()
			batch, err := parseResponseData(b)
			if err != nil {
				c.Globals.ErrLog.Add(err)
				// We can't parse the response, attempt to
				// re-request from the last window & batch.
				text.Warning(out, "unable to parse response body: %v", err)
				path, err = makeNewPath(path, curWindow, lastBatchID)
				if err != nil {
					return err
				}
				continue
			}

			// If we got a batch back, there will be an ID.
			if batch.ID != "" {
				// Record last batchID in case
				// anything fails along the way, we
				// can re-request.
				lastBatchID = batch.ID
				// Send batch down batchCh to the output loop.
				c.batchCh <- batch
			}
		}
		err = resp.Body.Close()
		if err != nil {
			c.Globals.ErrLog.Add(err)
		}

		if err := scanner.Err(); err != nil {
			c.Globals.ErrLog.Add(err)
			// ErrUnexpectedEOFs need to be retried, but they
			// produce a lot of noise for the user, so don't log.
			if err != io.ErrUnexpectedEOF {
				text.Warning(out, "error scanning response body: %v", err)
			}

			// Something happened in the scanner, re-request the
			// current batchID.
			path, err = makeNewPath(path, curWindow, lastBatchID)
			if err != nil {
				return err
			}
			continue
		}

		// Get our next time window to request.
		_, next := getLinks(resp.Header)
		curWindow, err = getTimeFromLink(next)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Next link": next,
			})
			text.Error(out, "error generating window from next link")
		}

		// We do NOT want to specify a batchID, as this
		// request was successful.
		lastBatchID = ""
		path, err = makeNewPath(path, curWindow, lastBatchID)
		if err != nil {
			return err
		}
	}
	return nil
}

// adjustTimes adjusts the passed in from and to flags based on the
// specified padding.
func (c *RootCommand) adjustTimes() {
	if c.cfg.from != 0 {
		// Adjust from based on search padding, we want to
		// look back further.
		c.cfg.from -= int64(c.cfg.searchPadding.Seconds())
	}

	if c.cfg.to != 0 {
		// Adjust to based on search padding, we want look forward more.
		c.cfg.to += int64(c.cfg.searchPadding.Seconds())
	}
}

// enableManagedLogging enables managed logging in our API.
func (c *RootCommand) enableManagedLogging(out io.Writer) error {
	_, err := c.Globals.APIClient.CreateManagedLogging(&c.Input)
	if err != nil && err != fastly.ErrManagedLoggingEnabled {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Info(out, "Managed logging enabled on service %s\n\n", c.Input.ServiceID)
	return nil
}

// outputLoop processes the logs out of band from the request/response loop.
func (c *RootCommand) outputLoop(out io.Writer) {
	type (
		bufferedLog struct {
			reqID string
			seq   int
		}

		receive struct {
			when    time.Time
			highSeq int
		}

		logrecv struct {
			logs     []Log
			receives []receive
		}
	)

	// Channel for timers to notify they are done buffering.
	tdCh := make(chan bufferedLog)

	// Single map to keep all buffered logs by RequestID as
	// well recording when logs were received.
	logmap := make(map[string]logrecv)

	for {
		select {
		case <-c.dieCh:
			return
		case batch := <-c.batchCh: // Got new batch.
			// Range through batch logs, for each
			// RequestID we create a timer based on the
			// highest SequenceNum we got in this batch
			// for that RequestID.  If a timer already
			// exists for the RequestID, we append the new
			// time.Now() and high SequenceNum.  At most
			// there should be one timer per RequestID.
			for reqid, logs := range splitByReqID(batch.Logs) {
				// Required for use in AfterFunc below.
				req := reqid

				// Record highest SequenceNum in this new batch
				// for this RequestID
				highSeq := highSequence(logs)

				// Whether we have the RequestID or not, we
				// append and sort the logs slice.
				reqLogs := logmap[req]
				reqLogs.logs = append(reqLogs.logs, logs...)
				// Sort the current batch of logs by their sequence number.
				sort.Slice(reqLogs.logs,
					func(i, j int) bool {
						return reqLogs.logs[i].SequenceNum < reqLogs.logs[j].SequenceNum
					})

				// Check to see if we already have a timer running or if the current
				// high sequence is higher than the one with the timer.
				// The timer will always be running on the head of the slice.
				// In either case append to the receives slice.
				recv := reqLogs.receives
				if len(recv) == 0 || recv[0].highSeq < highSeq {
					// NOTE: gocritic will warn about appendAssign but we ignore it.
					// Because if we try to address it the code fails to work at runtime.
					//nolint:gocritic
					reqLogs.receives = append(recv, receive{
						when:    time.Now(),
						highSeq: highSeq,
					})
				}

				// In only the empty case, start a new timer
				// since this is the head of the slice.
				if len(recv) == 0 {
					time.AfterFunc(c.cfg.sortBuffer, func() {
						tdCh <- bufferedLog{
							reqID: req,
							seq:   highSeq,
						}
					})
				}

				// Set the new log and receive info back to the
				// logmap for this RequestID.
				logmap[req] = reqLogs
			}

		case bufdLogs := <-tdCh: // A timer expired for a particular request.
			reqID, seq := bufdLogs.reqID, bufdLogs.seq

			// Get the logs for this RequestID and
			// find the index of the sequence in our current logs.
			reqLogs := logmap[reqID]
			idx := findIdxBySeq(reqLogs.logs, seq)

			// Split off the source of this timer, leave
			// remaining logs to be printed later.
			toPrint, remainingLogs := reqLogs.logs[:idx], reqLogs.logs[idx:]
			reqLogs.logs = remainingLogs
			c.printLogs(out, toPrint)

			// Special case if we just printed the entire set of
			// logs, we remove the keys from the maps and finish.
			if len(remainingLogs) == 0 {
				delete(logmap, reqID)
				break
			}

			// Drop the front of the batchReqReceives map and start
			// another timer for any remaining recorded sequences.
			recv := reqLogs.receives[1:]
			reqLogs.receives = recv

			// If anything is left...
			if len(recv) > 0 {
				// We create a new timer, we subtract
				// off time already served from the
				// user defined sortBuffer.
				time.AfterFunc(c.cfg.sortBuffer-time.Since(recv[0].when), func() {
					tdCh <- bufferedLog{
						reqID: reqID,
						seq:   recv[0].highSeq,
					}
				})
			}

			// Set the new log and receive info back to the
			// logmap for this RequestID.
			logmap[reqID] = reqLogs
		}
	}
}

// printLogs is a simple printer for Log slices, only printing requested
// streams.
func (c *RootCommand) printLogs(out io.Writer, logs []Log) {
	if len(logs) > 0 {
		filtered := filterStream(c.cfg.stream, logs)

		for _, l := range filtered {
			if c.cfg.printTimestamps {
				fmt.Fprint(out, l.RequestStartFromRaw().UTC().Format(time.RFC3339))
				fmt.Fprint(out, " | ")
			}
			fmt.Fprintln(out, l.String())
		}
	}
}

// doReq runs the http.Request, returning a http.Response or error.
func (c *RootCommand) doReq(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-c.dieCh:
			cancel()
		}
	}()

	if c.Globals.Flags.Debug {
		debug.DumpHTTPRequest(req)
	}
	resp, err := c.hClient.Do(req)
	if c.Globals.Flags.Debug {
		debug.DumpHTTPResponse(resp)
	}
	return resp, err
}

type (
	// cfg holds the configuration parameters passed in through
	// command line arguments.
	cfg struct {
		// path is the full path to fetch
		path string

		// from is how far in the past to start showing logs.
		from int64

		// to is when to get logs until.
		to int64

		// printTimestamps is whether to print timestamps with logs.
		printTimestamps bool

		// sortBuffer is how long to buffer logs from when the cli
		// receives them to when the cli prints them. It will sort
		// by RequestID for that buffer period.
		sortBuffer time.Duration
		// searchPadding is how much of a window on either side of
		// from and to to use for searching for the beginning or
		// through the end timestamps.
		searchPadding time.Duration
		// stream specifies which of stdout or stderr or both the
		// customer wants to consume.
		// Undefined == both stderr and stdout.
		stream string
	}

	// Log defines the message envelope that the Compute platform wraps the
	// user messages in.
	Log struct {
		// SequenceNum is the message sequence number used to reorder
		// messages.
		SequenceNum int `json:"sequence_number"`
		// RequestTime is the time in microseconds when the request
		// was received.
		RequestStart int64 `json:"request_start_us"`
		// Stream is the Compute stream, either stdout or stderr.
		Stream string `json:"stream"`
		// RequestID is a UUID representing individual requests to the
		// particular Wasm service.
		RequestID string `json:"id"`
		// Message is the actual message body the user wants printed.
		Message string `json:"message"`
	}

	// Batch encompasses a batch ID and the logs for this batch.
	Batch struct {
		ID   string `json:"batch_id"`
		Logs []Log  `json:"logs"`
	}
)

// RequestStartFromRaw return a time.Time object representing the
// RequestStart data.
func (l *Log) RequestStartFromRaw() time.Time {
	// RequestTime comes as unix time in microseconds. Convert to
	// nanoseconds, then parse with stdlib.
	nano := l.RequestStart * 1000
	return time.Unix(0, nano)
}

// String is used to print a log for the tail output.
func (l *Log) String() string {
	// Trim the RequestID for nicer output, it might be a long UUID.
	return fmt.Sprintf("%6s | %8.8s | %s",
		l.Stream,
		l.RequestID,
		l.Message)
}

// makeNewPath generates a new request path based on current
// path, window, and batchID.
func makeNewPath(path string, window int64, batchID string) (string, error) {
	basePath, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("error generating request URL: %w", err)
	}

	// Unset anything in the query parameters that might already exist.
	basePath.RawQuery = ""

	q := basePath.Query()
	if window != 0 {
		q.Set("from", strconv.FormatInt(window, 10))
	}

	if batchID != "" {
		q.Set("batch_id", batchID)
	}

	basePath.RawQuery = q.Encode()
	return basePath.String(), nil
}

// splitByReqID splits slices of logs based on RequestID.
func splitByReqID(in []Log) map[string][]Log {
	out := make(map[string][]Log)
	for _, l := range in {
		out[l.RequestID] = append(out[l.RequestID], l)
	}
	return out
}

// parseResponseData returns the batch from a response.
func parseResponseData(data []byte) (Batch, error) {
	var batch Batch
	reader := bytes.NewReader(data)
	d := json.NewDecoder(reader)

	if err := d.Decode(&batch); err != nil && err != io.EOF {
		return batch, err
	}

	return batch, nil
}

// filterStream returns only logs that are requested by the stream flag.
func filterStream(stream string, logs []Log) []Log {
	// If unset, do not filter out any logs.
	if stream == "" {
		return logs
	}

	var out []Log
	for _, l := range logs {
		// If the stream matches what they wanted, keep it.
		if stream == l.Stream {
			out = append(out, l)
		}
	}
	return out
}

// getTimeFromLink splits a link header format, returning
// the time.
func getTimeFromLink(link string) (int64, error) {
	s := strings.SplitN(link, "=", 2)[1]
	return strconv.ParseInt(s, 10, 64)
}

// getLinks returns the prev and next links from a header.
func getLinks(head http.Header) (prev, next string) {
	links := linkheader.ParseMultiple(head["Link"])
	for _, link := range links {
		switch link.Rel {
		case "prev":
			prev = link.URL
		case "next":
			next = link.URL
		}
	}
	return prev, next
}

// findIdxBySeq returns the slice index after the
// SequenceNum we are searching for.
func findIdxBySeq(logs []Log, seq int) int {
	for i, v := range logs {
		if v.SequenceNum > seq {
			return i
		}
	}
	return len(logs)
}

// highSequence returns the highest SequenceNum
// in a slice of logs.
func highSequence(logs []Log) int {
	var maximum int
	for _, l := range logs {
		if l.SequenceNum > maximum {
			maximum = l.SequenceNum
		}
	}
	return maximum
}
