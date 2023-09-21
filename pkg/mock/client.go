package mock

import (
	"fmt"
	"net/http"

	"github.com/fastly/cli/pkg/api"
)

// APIClient takes a mock.API and returns an app.ClientFactory that uses that
// mock, ignoring the token and endpoint. It should only be used for tests.
func APIClient(a API) func(string, string, bool) (api.Interface, error) {
	return func(token, endpoint string, debugMode bool) (api.Interface, error) {
		return a, nil
	}
}

type mockHTTPClient struct {
	// index keeps track of which response/error to return
	index int
	res   []*http.Response
	err   []error
}

func (c mockHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	c.index++
	fmt.Printf("res: %#v | len res: %d | index: %d\n", c.res, len(c.res), c.index)
	return c.res[c.index], c.err[c.index]
}

// HTMLClient returns a mock HTTP Client that returns a stubbed response or
// error.
func HTMLClient(res []*http.Response, err []error) api.HTTPClient {
	return mockHTTPClient{
		index: -1,
		res:   res,
		err:   err,
	}
}
