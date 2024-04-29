package testutil

import (
	"bytes"
	"io"
	"net/http"
	"sync"
)

type HTTPResult struct {
	Response *http.Response
	Err      error
}

func (h *HTTPResult) IsZero() bool {
	return h.Response == nil && h.Err == nil
}

type RecordingHTTPClient struct {
	requests     []http.Request
	counter      int
	mu           sync.Mutex
	results      []HTTPResult
	singleResult HTTPResult
}

func NewRecordingHTTPClient() *RecordingHTTPClient {
	return &RecordingHTTPClient{}
}

func (c *RecordingHTTPClient) Do(r *http.Request) (*http.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.requests = append(c.requests, *r)

	if !c.singleResult.IsZero() {
		return c.singleResult.Response, c.singleResult.Err
	}

	curr := c.counter
	resp := c.results[curr]
	c.counter++
	return resp.Response, resp.Err
}

func (c *RecordingHTTPClient) SingleResponse(resp *http.Response, err error) {
	c.singleResult = HTTPResult{
		Response: resp,
		Err:      err,
	}
}

func (c *RecordingHTTPClient) QueueResponse(resp *http.Response, err error) {
	c.results = append(c.results, HTTPResult{
		Response: resp,
		Err:      err,
	})
}

func (c *RecordingHTTPClient) GetRequests() []http.Request {
	return c.requests
}

func (c *RecordingHTTPClient) GetRequest(i int) (r http.Request, ok bool) {
	if i < len(c.requests) {
		return c.requests[i], true
	}
	return r, ok
}

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
