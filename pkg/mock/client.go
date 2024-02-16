package mock

import (
	"net/http"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/api"
)

// APIClient takes a mock.API and returns an app.ClientFactory that uses that
// mock, ignoring the token and endpoint. It should only be used for tests.
func APIClient(a API) func(string, string, bool) (api.Interface, error) {
	return func(token, endpoint string, debugMode bool) (api.Interface, error) {
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
}

func (c HTTPClient) Get(p string, _ *fastly.RequestOptions) (*http.Response, error) {
	// IMPORTANT: Have to increment on defer as index is already 0 by this point.
	// This is opposite to the Do() method which is -1 at the time it's called.
	defer func() { c.Index++ }()
	return c.Responses[c.Index], c.Errors[c.Index]
}

func (c HTTPClient) Do(r *http.Request) (*http.Response, error) {
	c.Index++
	return c.Responses[c.Index], c.Errors[c.Index]
}

// HTMLClient returns a mock HTTP Client that returns a stubbed response or
// error.
func HTMLClient(res []*http.Response, err []error) api.HTTPClient {
	return HTTPClient{
		Index:     -1,
		Responses: res,
		Errors:    err,
	}
}
