package mock

import (
	"net/http"

	"github.com/fastly/cli/pkg/api"
)

// ClientFactory takes a mock.API and returns an app.ClientFactory that uses that
// mock, ignoring the token and endpoint. It should only be used for tests.
func ClientFactory(a API) api.ClientFactory {
	return func(token, endpoint string) (api.Interface, error) {
		return a, nil
	}
}

type mockHTTPClient struct {
	res *http.Response
	err error
}

func (c mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.res, c.err
}

// HTMLClient returns a mock HTTP Client that returns a stubbed response or
// error.
func HTMLClient(res *http.Response, err error) api.HTTPClient {
	return mockHTTPClient{
		res: res,
		err: err,
	}
}
