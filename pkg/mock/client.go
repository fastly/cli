package mock

import (
	"github.com/fastly/cli/pkg/api"
)

// APIClient takes a mock.API and returns an app.ClientFactory that uses that
// mock, ignoring the token and endpoint. It should only be used for tests.
func APIClient(a API) func(string, string) (api.Interface, error) {
	return func(token, endpoint string) (api.Interface, error) {
		return a, nil
	}
}
