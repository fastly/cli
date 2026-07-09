package vcl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/bep/debounce"
	"github.com/fsnotify/fsnotify"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/fiddle"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// maxBufferedBody bounds how much of an incoming request body is buffered
// so a request can be replayed after the sandbox re-syncs.
const maxBufferedBody = 64 << 20

// ServeCommand runs VCL as a local HTTP server: a listener on localhost
// forwards every request to the VCL deployed on a Fastly edge sandbox and
// returns the edge's complete response.
type ServeCommand struct {
	argparser.Base

	spec    specFlags
	files   []string
	addr    string
	watch   bool
	timeout time.Duration
}

// NewServeCommand returns a usable command registered under the parent.
func NewServeCommand(parent argparser.Registerer, g *global.Data) *ServeCommand {
	c := ServeCommand{
		Base: argparser.Base{
			Globals:         g,
			SuppressVerbose: true,
		},
	}
	c.CmdClause = parent.Command("serve", "Run VCL as a local HTTP server backed by a real Fastly edge sandbox (no service or API token needed)")
	c.CmdClause.Arg("file", "VCL files holding subroutine bodies, named after their subroutine (recv.vcl, fetch.vcl, ...)").StringsVar(&c.files)
	c.spec.register(c.CmdClause)
	c.CmdClause.Flag("addr", "The local address to listen on").Default("127.0.0.1:7676").StringVar(&c.addr)
	c.CmdClause.Flag("watch", "Redeploy the VCL when its files change").BoolVar(&c.watch)
	c.CmdClause.Flag("timeout", "How long to wait for a deploy to reach the edge").Default("2m").DurationVar(&c.timeout)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ServeCommand) Exec(_ io.Reader, out io.Writer) error {
	vcl, sources, err := c.spec.buildVCL(c.files)
	if err != nil {
		return err
	}
	origins, err := c.spec.buildOrigins()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := &sandboxServer{
		client:  c.spec.client(c.Globals),
		origins: origins,
		timeout: c.timeout,
		out:     out,
		verbose: c.Globals.Verbose(),
	}

	text.Info(out, "Deploying VCL to a test sandbox... (this can take a minute to reach the edge)")
	if err := s.deploy(ctx, vcl, sources); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	text.Success(out, "VCL is live on the edge.")
	printOrigins(out, origins, len(c.spec.origins) == 0)

	listener, err := net.Listen("tcp", c.addr)
	if err != nil {
		return err
	}
	text.Break(out)
	text.Output(out, "Serving VCL on http://%s (Ctrl-C to stop)", listener.Addr())
	text.Break(out)

	if c.watch {
		watchCtx, cancelWatch := context.WithCancel(ctx)
		defer cancelWatch()
		if err := s.watchSources(watchCtx, sources); err != nil {
			return err
		}
	}

	server := &http.Server{
		Handler:           s.handler(),
		ReadHeaderTimeout: 30 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Serve(listener)
	}()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		text.Break(out)
		text.Output(out, "Local server stopped.")
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

// sandboxServer holds the state shared between the proxy handler and the
// deploy/watch logic.
type sandboxServer struct {
	client  *fiddle.Client
	origins []string
	timeout time.Duration
	out     io.Writer
	verbose bool

	fiddleID string
	host     atomic.Pointer[string]

	// deployMu serializes every path that publishes or re-syncs the
	// sandbox: the initial deploy, --watch redeploys, and expired-sandbox
	// resyncs triggered by in-flight requests.
	deployMu sync.Mutex
}

// deploy publishes the VCL and waits until it answers on the edge,
// recording the direct execution host.
// The spec carries one plain GET as a sync probe; its response is
// discarded.
func (s *sandboxServer) deploy(ctx context.Context, vcl map[string]string, sources vclSources) error {
	s.deployMu.Lock()
	defer s.deployMu.Unlock()
	spec := newSpec(s.origins, vcl, fiddle.NewRequest())
	var saved *fiddle.SavedFiddle
	var err error
	if s.fiddleID == "" {
		// Serve owns a dedicated sandbox rather than sharing the one run
		// and check reuse: a long-lived proxy must not have its VCL
		// swapped out under it by an unrelated command.
		saved, err = s.client.Create(ctx, spec)
	} else {
		saved, err = s.client.Update(ctx, s.fiddleID, spec)
	}
	if err != nil {
		return err
	}
	if !saved.Valid {
		printDiagnostics(s.out, flattenDiagnostics(saved.LintStatus, sources))
		return errCompileFailed
	}
	s.fiddleID = saved.ID
	return s.sync(ctx)
}

// sync executes the sandbox once and derives the direct host from the
// result, with a fresh cache bucket.
func (s *sandboxServer) sync(ctx context.Context) error {
	result, err := s.client.Run(ctx, s.fiddleID, freshCacheID(), fiddle.StreamOptions{
		WantFetches: 1,
		MaxWait:     s.timeout,
	})
	if err != nil {
		return err
	}
	execHost, err := fiddle.ParseExecHost(result.ExecHost)
	if err != nil {
		return err
	}
	host := execHost.Hostname(freshCacheID())
	s.host.Store(&host)
	return nil
}

// resync redeploys after the edge reports the sandbox is gone (a long-idle
// serve can be evicted).
// Concurrent callers coalesce on one redeploy.
func (s *sandboxServer) resync(prevHost string) error {
	s.deployMu.Lock()
	defer s.deployMu.Unlock()
	if current := s.host.Load(); current != nil && *current != prevHost {
		return nil // another request already resynced
	}
	text.Info(s.out, "The edge sandbox expired; redeploying...")
	return s.sync(context.Background())
}

// handler returns the local HTTP handler: a reverse proxy onto the sandbox's
// execution host.
func (s *sandboxServer) handler() http.Handler {
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			host := *s.host.Load()
			pr.Out.URL.Scheme = "https"
			pr.Out.URL.Host = host
			pr.Out.Host = host
			pr.Out.Header.Del("Accept-Encoding") // keep bodies inspectable
		},
		Transport:      &resyncTransport{base: http.DefaultTransport, server: s},
		ModifyResponse: s.modifyResponse,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			text.Error(s.out, "%s %s: %s", r.Method, r.URL.Path, err)
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "sandbox proxy error: %s\n", err)
		},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Buffer the body so the request can be replayed if the sandbox
		// needs a redeploy mid-flight.
		if r.Body != nil && r.ContentLength != 0 {
			body, err := io.ReadAll(io.LimitReader(r.Body, maxBufferedBody+1))
			if err != nil {
				http.Error(w, "failed reading request body", http.StatusBadRequest)
				return
			}
			if len(body) > maxBufferedBody {
				http.Error(w, "request body too large for the sandbox proxy", http.StatusRequestEntityTooLarge)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(body)), nil
			}
		}
		proxy.ServeHTTP(w, r)
	})
}

// modifyResponse strips the sandbox's trace headers (surfacing them on the
// terminal instead) and logs the request.
func (s *sandboxServer) modifyResponse(resp *http.Response) error {
	trace := resp.Header.Get("Fastly-Fiddle-Log")
	resp.Header.Del("Fastly-Fiddle-Log")
	resp.Header.Del("Fastly-Fiddle-Starttime")

	details := ""
	if cache := resp.Header.Get("X-Cache"); cache != "" {
		details = " (" + lastCacheHop(cache) + ")"
	}
	req := resp.Request
	text.Output(s.out, "%s %s → %d%s", req.Method, req.URL.Path, resp.StatusCode, details)
	if s.verbose && trace != "" {
		for _, line := range parseFiddleTrace(trace) {
			text.Indent(s.out, 2, "%s", line)
		}
	}
	return nil
}

// parseFiddleTrace decodes the Fastly-Fiddle-Log header: comma-separated
// entries whose last space-separated field is a query-encoded attribute set
// (fnName=recv&return=hash&...).
func parseFiddleTrace(header string) []string {
	var lines []string
	for _, entry := range strings.Split(header, ",") {
		fields := strings.Fields(strings.TrimSpace(entry))
		if len(fields) < 2 {
			continue
		}
		attribs, err := url.ParseQuery(fields[len(fields)-1])
		if err != nil {
			continue
		}
		fn := attribs.Get("fnName")
		if fn == "" {
			continue
		}
		line := "vcl_" + fn
		if ret := attribs.Get("return"); ret != "" {
			line += " → return(" + ret + ")"
		}
		if state := attribs.Get("state"); state != "" {
			line += "  [" + state + "]"
		}
		lines = append(lines, line)
	}
	return lines
}

// resyncTransport retries a request once after redeploying, when the edge
// answers with the sandbox-unknown status.
type resyncTransport struct {
	base   http.RoundTripper
	server *sandboxServer
}

// sandboxGoneStatus is what the execution pool returns for a fiddle it no
// longer has deployed.
const sandboxGoneStatus = 982

func (t *resyncTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Host
	resp, err := t.base.RoundTrip(req)
	if err != nil || resp.StatusCode != sandboxGoneStatus {
		return resp, err
	}
	_ = resp.Body.Close()
	if err := t.server.resync(host); err != nil {
		return nil, fmt.Errorf("the edge sandbox expired and redeploying failed: %w", err)
	}
	retry := req.Clone(req.Context())
	newHost := *t.server.host.Load()
	retry.URL.Host = newHost
	retry.Host = newHost
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		retry.Body = body
	}
	return t.base.RoundTrip(retry)
}

// watchSources redeploys when any of the VCL files change.
func (s *sandboxServer) watchSources(ctx context.Context, sources vclSources) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	// Watch the parent directories rather than the files: atomic-save
	// editors replace the file (rename + create), which silently drops a
	// file-level watch.
	watched := map[string]bool{}
	dirs := map[string]bool{}
	for _, path := range sources {
		abs, err := filepath.Abs(path)
		if err != nil {
			_ = watcher.Close()
			return err
		}
		watched[abs] = true
		dirs[filepath.Dir(abs)] = true
	}
	for dir := range dirs {
		if err := watcher.Add(dir); err != nil {
			_ = watcher.Close()
			return err
		}
	}
	reload := func() {
		fresh, err := readSources(sources)
		if err != nil {
			text.Error(s.out, "reload failed: %s", err)
			return
		}
		text.Info(s.out, "VCL changed; redeploying...")
		started := time.Now()
		if err := s.deploy(ctx, fresh, sources); err != nil {
			if errors.Is(err, errCompileFailed) {
				text.Warning(s.out, "Still serving the previous VCL.")
			} else {
				text.Error(s.out, "redeploy failed: %s", err)
			}
			return
		}
		text.Success(s.out, "VCL reloaded in %s.", time.Since(started).Round(time.Second))
	}
	debounced := debounce.New(500 * time.Millisecond)
	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
					continue
				}
				if abs, err := filepath.Abs(event.Name); err != nil || !watched[abs] {
					continue
				}
				debounced(reload)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				text.Error(s.out, "watch error: %s", err)
			}
		}
	}()
	return nil
}
