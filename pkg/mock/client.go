package mock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/fastly/go-fastly/v14/fastly"

	"github.com/fastly/cli/pkg/api"
)

// APIClient takes a mock.API and returns an app.ClientFactory that uses that
// mock, ignoring the token and endpoint. It should only be used for tests.
func APIClient(a API) func(string, string, bool) (api.Interface, error) {
	return func(token, endpoint string, debugMode bool) (api.Interface, error) {
		fmt.Printf("token: %s\n", token)
		fmt.Printf("endpoint: %s\n", endpoint)
		fmt.Printf("debugMode: %t\n", debugMode)
		return a, nil
	}
}

// HTTPClient is used to mock fastly.Client requests.
type HTTPClient struct {
	// Index keeps track of which Responses/Errors index to return.
	Index int
	// Responses tracks different responses to return.
	Responses []*http.Response
	// Errors tracks different errors to return.
	Errors []error
	// SaveRequests toggles recording requests that pass through the
	// client.
	SaveRequests bool
	// Requests stores copies of incoming requests.
	Requests []http.Request
}

// Get mocks a HTTP Client Get request.
func (c *HTTPClient) Get(_ context.Context, p string, _ fastly.RequestOptions) (*http.Response, error) {
	fmt.Printf("p: %#v\n", p)
	// IMPORTANT: Have to increment on defer as index is already 0 by this point.
	// This is opposite to the Do() method which is -1 at the time it's called.
	defer func() { c.Index++ }()
	return c.Responses[c.Index], c.Errors[c.Index]
}

// Do mocks a HTTP Client Do operation.
func (c *HTTPClient) Do(r *http.Request) (*http.Response, error) {
	fmt.Printf("r.URL: %#v\n", r.URL.String())
	fmt.Printf("r: %#v\n", r)
	if c.SaveRequests {
		c.Requests = append(c.Requests, *r.Clone(context.Background()))
	}
	c.Index++
	return c.Responses[c.Index], c.Errors[c.Index]
}

// NewHTTPClient returns a mock HTTP Client that returns stubbed responses and errors.
func NewHTTPClient(res []*http.Response, err []error, saveRequests bool) *HTTPClient {
	if len(res) != len(err) {
		panic("mock.HTTPClient: Responses and Errors length mismatch")
	}
	return &HTTPClient{
		Index:        -1,
		Responses:    res,
		Errors:       err,
		SaveRequests: saveRequests,
	}
}

// NewHTTPClientDefault returns a mock HTTP Client that returns stubbed responses and
// errors, and saves requests.
func NewHTTPClientDefault(res []*http.Response, err []error) *HTTPClient {
	return NewHTTPClient(res, err, true)
}

// NewHTTPClientWithResponses returns a mock HTTP Client that returns stubbed response and
// no errors, and saves requests.
func NewHTTPClientWithResponses(res []*http.Response) *HTTPClient {
	return NewHTTPClientDefault(res, make([]error, len(res)))
}

// NewHTTPClientWithErrors returns a mock HTTP Client that returns errors with no responses,
// and saves requests.
func NewHTTPClientWithErrors(err []error) *HTTPClient {
	return NewHTTPClientDefault(make([]*http.Response, len(err)), err)
}

// HTMLClient returns a mock HTTP Client that returns a stubbed response or
// error.
func HTMLClient(res []*http.Response, err []error) api.HTTPClient {
	return &HTTPClient{
		Index:     -1,
		Responses: res,
		Errors:    err,
	}
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

func NewNetHTTPClientWithMockHTTPClient(httpClient *HTTPClient) *http.Client {
	netHTTPClient := &http.Client{
		Transport: NewRoundTripper(httpClient),
	}
	return netHTTPClient
}
