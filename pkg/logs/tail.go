package logs

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

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/tomnomnom/linkheader"
)

type (
	// TailCommand represents the CLI subcommand for Log Tailing.
	TailCommand struct {
		cmd.Base
		manifest manifest.Data
		Input    fastly.CreateManagedLoggingInput

		cfg cfg

		dieCh   chan struct{} // channel to end output/printing
		batchCh chan Batch    // send batches to output loop
		doneCh  chan struct{} // channel to signal we've reached the end of the run

		hClient *http.Client // TODO: this will go away when GET is in go-fastly
		token   string       // TODO: this will go away when GET is in go-fastly
	}

	// cfg holds the configuration parameters passed in through
	// command line arguments.
	cfg struct {
		// path is the full path to fetch
		path string

		// from is how far in the past to start showing logs.
		from int64

		// to is when to get logs until.
		to int64

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

	// Log defines the message envelope that compute@edge (C@E) wraps the
	// user messages in.
	Log struct {
		// SequenceNum is the message sequence number used to reorder
		// messages.
		SequenceNum int `json:"sequence_number"`
		// RequestTime is the time in microseconds when the request
		// was received.
		RequestStart int64 `json:"request_start_us"`
		// Stream is the C@E stream, either stdout or stderr.
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

// NewTailCommand returns a usable command registered under the parent.
func NewTailCommand(parent cmd.Registerer, globals *config.Data) *TailCommand {
	var c TailCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("tail", "Tail Compute@Edge logs")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("from", "From time, in unix seconds").Int64Var(&c.cfg.from)
	c.CmdClause.Flag("to", "To time, in unix seconds").Int64Var(&c.cfg.to)
	c.CmdClause.Flag("sort-buffer",
		"Sort buffer is how long to buffer logs, attempting to sort them before printing, defaults to 1s (second)").Default("1s").DurationVar(&c.cfg.sortBuffer)
	c.CmdClause.Flag("search-padding",
		"Search padding is how much of a window on either side of From and To to use for searching, defaults to 2s (seconds)").Default("2s").DurationVar(&c.cfg.searchPadding)
	c.CmdClause.Flag("stream", "Stream specifies which of 'stdout' or 'stderr' to output, defaults to undefined (all streams)").StringVar(&c.cfg.stream)

	return &c
}

// Exec invokes the application logic for the command.
func (c *TailCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	c.Input.Kind = fastly.ManagedLoggingInstanceOutput
	c.cfg.path = fmt.Sprintf("%s/service/%s/log_stream/managed/instance_output", config.DefaultEndpoint, c.Input.ServiceID)

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

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// Start the output loop.
	go c.outputLoop(out)

	// Start tailing the logs.
	go c.tail(out)

	<-sigs
	close(c.dieCh)

	return nil
}

//
// Client
//

// Tail starts the virtual tail process. Tail fetches data from the eventbuffer
// API. It hands off the requested logs to the outputloop for the actual
// printing.
func (c *TailCommand) tail(out io.Writer) {
	// Start this with --from and --to if set.
	curWindow := c.cfg.from
	toWindow := c.cfg.to

	// Start the loop with an initial address to query.
	path := makeNewPath(out, c.cfg.path, curWindow, "")

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

		req, err := http.NewRequest("GET", path, nil)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			text.Error(out, "unable to create new request: %v", err)
			os.Exit(1)
		}
		req.Header.Add("Fastly-Key", c.token)

		resp, err := c.doReq(req)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			text.Error(out, "unable to execute request: %v", err)
			os.Exit(1)
		}

		// Check that our request was successful. If the server is
		// having trouble, retry after waiting for some time.
		if resp.StatusCode != http.StatusOK {
			// If the response was a 404, the from time was
			// not valid, give them an error stating this and exit.
			if resp.StatusCode == http.StatusNotFound &&
				c.cfg.from != 0 {
				text.Error(out, "specified 'from' time %d not found, either too far in the past or future", c.cfg.from)
				os.Exit(1)
			}

			// In an effort to clean up the output, do not print on
			// 503's.
			if resp.StatusCode != http.StatusServiceUnavailable {
				text.Warning(out, "non-200 resp %d", resp.StatusCode)
			}

			// Reuse the connection for the retry, or cleanup in the
			// case of Exit.
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			// Try the response again after a 1 second wait.
			if resp.StatusCode/100 == 5 && resp.StatusCode != 501 ||
				resp.StatusCode == 429 {
				time.Sleep(1 * time.Second)
				continue
			}

			// Failing at this point is unrecoverable.
			text.Error(out, "unrecoverable error, response code: %d", resp.StatusCode)
			os.Exit(1)
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
				path = makeNewPath(out, path, curWindow, lastBatchID)
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
		resp.Body.Close()

		if err := scanner.Err(); err != nil {
			c.Globals.ErrLog.Add(err)
			// ErrUnexpectedEOFs need to be retried, but they
			// produce a lot of noise for the user, so don't log.
			if err != io.ErrUnexpectedEOF {
				text.Warning(out, "error scanning response body: %v", err)
			}

			// Something happened in the scanner, re-request the
			// current batchID.
			path = makeNewPath(out, path, curWindow, lastBatchID)
			continue
		}

		// Get our next time window to request.
		_, next := getLinks(resp.Header)
		curWindow, err = getTimeFromLink(next)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			text.Error(out, "error generating window from next link")
		}

		// We do NOT want to specify a batchID, as this
		// request was successful.
		lastBatchID = ""
		path = makeNewPath(out, path, curWindow, lastBatchID)
	}
}

// adjustTimes adjusts the passed in from and to flags based on the
// specified padding.
func (c *TailCommand) adjustTimes() {
	if c.cfg.from != 0 {
		// Adjust from based on search padding, we want to
		// look back further.
		c.cfg.from = c.cfg.from - int64(c.cfg.searchPadding.Seconds())
	}

	if c.cfg.to != 0 {
		// Adjust to based on search padding, we want look forward more.
		c.cfg.to = c.cfg.to + int64(c.cfg.searchPadding.Seconds())
	}
}

// enableManagedLogging enables managed logging in our API.
func (c *TailCommand) enableManagedLogging(out io.Writer) error {
	_, err := c.Globals.Client.CreateManagedLogging(&c.Input)
	if err != nil && err != fastly.ErrManagedLoggingEnabled {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Info(out, "Managed logging enabled on service %s", c.Input.ServiceID)
	return nil
}

// outputLoop processes the logs out of band from the request/response loop.
func (c *TailCommand) outputLoop(out io.Writer) {
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

				// Check to see if we already have a timer
				// running or if the current high sequence is
				// higher than the one with the timer.
				// The timer will always be running on the head
				// of the slice.
				recv := reqLogs.receives

				// In either case append to the receives slice.
				if len(recv) == 0 || recv[0].highSeq < highSeq {
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

		case <-c.doneCh:
			os.Exit(0)
		}
	}
}

// printLogs is a simple printer for Log slices, only printing requested
// streams.
func (c *TailCommand) printLogs(out io.Writer, logs []Log) {
	if len(logs) > 0 {
		filtered := filterStream(c.cfg.stream, logs)

		for _, l := range filtered {
			fmt.Fprintln(out, l.String())
		}
	}
}

// doReq runs the http.Request, returning a http.Response or error.
func (c *TailCommand) doReq(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)
	go func() {
		select {
		case <-ctx.Done():
		case <-c.dieCh:
			cancel()
		}
	}()

	resp, err := c.hClient.Do(req)
	return resp, err
}

//
// Log
//

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

//
// Helpers
//

// makeNewPath generates a new request path based on current
// path, window, and batchID.
func makeNewPath(out io.Writer, path string, window int64, batchID string) string {
	basePath, err := url.Parse(path)
	if err != nil {
		// No reasonable way to carry on from an error at this point
		// and it should never happen, so error & exit.
		text.Error(out, "error generating request URL: %v", err)
		os.Exit(1)
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
	return basePath.String()
}

// splitByReqID splits slices of logs based on RequestID,
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
	return
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
	var max int
	for _, l := range logs {
		if l.SequenceNum > max {
			max = l.SequenceNum
		}
	}
	return max
}
