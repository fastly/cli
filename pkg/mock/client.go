package mock

import (
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
)

// APIClient takes a mock.API and returns an app.ClientFactory that uses that
// mock, ignoring the token and endpoint. It should only be used for tests.
func APIClient(a API) app.APIClientFactory {
	return func(token, endpoint string) (api.Interface, error) {
		return a, nil
	}
}
