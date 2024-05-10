package testutil

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

type httpResult struct {
	Response *http.Response
	Err      error
}

func (h *httpResult) isZero() bool {
	return h.Response == nil && h.Err == nil
}

// RecordingHTTPClient records all requests made through it and
// returns pre-registered responses or errors in order.
type RecordingHTTPClient struct {
	requests     []http.Request
	counter      int
	mu           sync.Mutex
	results      []httpResult
	singleResult httpResult
}

// NewRecordingHTTPClient returns an initialized RecordingHTTPClient.
func NewRecordingHTTPClient() *RecordingHTTPClient {
	return &RecordingHTTPClient{}
}

// Do implements the minimal surface area of net/http.Client.
func (c *RecordingHTTPClient) Do(r *http.Request) (*http.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requests = append(c.requests, *r)

	if !c.singleResult.isZero() {
		return c.singleResult.Response, c.singleResult.Err
	}

	curr := c.counter
	resp := c.results[curr]
	c.counter++
	return resp.Response, resp.Err
}

// SingleResponse registers a response or error to use for all requests.
// It takes precedence over any queued responses.
func (c *RecordingHTTPClient) SingleResponse(resp *http.Response, err error) {
	c.singleResult = httpResult{
		Response: resp,
		Err:      err,
	}
}

// QueueResponse sets the given response or error to be returned for
// requests. Responses are returned for each request in the order they
// are queued.
func (c *RecordingHTTPClient) QueueResponse(resp *http.Response, err error) {
	c.results = append(c.results, httpResult{
		Response: resp,
		Err:      err,
	})
}

// GetRequests returns a slice of all the requests it has recorded.
func (c *RecordingHTTPClient) GetRequests() []http.Request {
	return c.requests
}

// GetRequest retrieves a specific request from the saved requests. It
// is zero-indexed.
func (c *RecordingHTTPClient) GetRequest(i int) (r http.Request, ok bool) {
	if i < len(c.requests) {
		return c.requests[i], true
	}
	return r, ok
}

// NewHTTPResponse fills in the boilerplate needed to create a minimal
// *http.Response.
func NewHTTPResponse(statusCode int, headers map[string]string, body io.ReadCloser) *http.Response {
	if body == nil {
		body = io.NopCloser(bytes.NewReader(nil))
	}
	h := http.Header{}
	for header, value := range headers {
		h.Add(header, value)
	}
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Body:       body,
		Header:     h,
	}
}
