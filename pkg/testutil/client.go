package testutil

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// MockRoundTripper implements [http.RoundTripper] for mocking HTTP responses.
type MockRoundTripper struct {
	Response *http.Response
	Err      error
}

// RoundTrip executes a single HTTP transaction, returning a Response for the
// provided Request.
func (m *MockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}

// MultiResponseRoundTripper implements [http.RoundTripper] for mocking multiple
// sequential HTTP responses. This is useful when the code under test makes
// multiple HTTP calls (e.g., GET then PATCH).
//
// When we perform a get and update in go-fastly operations (such as for alerts),
// we need to be able to parse multiple responses back from the API.
type MultiResponseRoundTripper struct {
	Responses []*http.Response
	index     int
}

// RoundTrip executes a single HTTP transaction, returning the next Response
// in sequence.
func (m *MultiResponseRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.index >= len(m.Responses) {
		return m.Responses[len(m.Responses)-1], nil
	}
	resp := m.Responses[m.index]
	m.index++
	return resp, nil
}

// RequestRecorder implements [http.RoundTripper] by replying with a fixed
// response and recording the last request it received, so tests can assert
// on what the command under test actually sent to the API.
type RequestRecorder struct {
	// Status is the HTTP status code of the response.
	Status int
	// Body is the response body.
	Body []byte

	// Method, Path, Query, and RequestBody are populated from the last
	// request received.
	Method      string
	Path        string
	Query       url.Values
	RequestBody []byte
}

// RoundTrip records the request and returns the configured response.
func (r *RequestRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	r.Method = req.Method
	r.Path = req.URL.Path
	// The scenario runner configures its client with a scheme-less endpoint
	// ("api.example.com"), which leaves the host at the front of the path.
	if req.URL.Host == "" && !strings.HasPrefix(r.Path, "/") {
		if i := strings.Index(r.Path, "/"); i >= 0 {
			r.Path = r.Path[i:]
		}
	}
	r.Query = req.URL.Query()
	r.RequestBody = nil
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		r.RequestBody = b
	}
	return &http.Response{
		StatusCode: r.Status,
		Status:     http.StatusText(r.Status),
		Body:       io.NopCloser(bytes.NewReader(r.Body)),
	}, nil
}

// Client returns an [http.Client] that uses the recorder as its transport.
func (r *RequestRecorder) Client() *http.Client {
	return &http.Client{Transport: r}
}
