package mock

import (
	"net/http"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/api"
)

// APIClient takes a mock.API and returns an app.ClientFactory that uses that
// mock, ignoring the token and endpoint. It should only be used for tests.
func APIClient(a API) func(string, string, bool) (api.Interface, error) {
	return func(token, endpoint string, debugMode bool) (api.Interface, error) {
		return a, nil
	}
}

// MockHTTPClient is used to mock fastly.Client requests.
type MockHTTPClient struct {
	// Index keeps track of which Responses/Errors index to return.
	Index int
	// Responses tracks different responses to return.
	Responses []*http.Response
	// Errors tracks different errors to return.
	Errors []error
}

func (c MockHTTPClient) Get(p string, ro *fastly.RequestOptions) (*http.Response, error) {
	c.Index++
	return c.Responses[c.Index], c.Errors[c.Index]
}

func (c MockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	c.Index++
	return c.Responses[c.Index], c.Errors[c.Index]
}

// HTMLClient returns a mock HTTP Client that returns a stubbed response or
// error.
func HTMLClient(res []*http.Response, err []error) api.HTTPClient {
	return MockHTTPClient{
		Index:     -1,
		Responses: res,
		Errors:    err,
	}
}
