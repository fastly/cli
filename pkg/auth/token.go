package auth

import (
	"net/http"

	"github.com/fastly/cli/pkg/api"
)

// ValidateToken retrieves the given token data and returns.
//
// NOTE: We don't use the SDK GetTokenSelf method because it returns data and
// not the response, meaning we can't inspect the status code returned.
func ValidateToken(token string, apiClient api.Interface) (status int, err error) {
	if resp, err := apiClient.Get("/tokens/self", nil); err != nil {
		return http.StatusInternalServerError, err
	} else {
		return resp.StatusCode, nil
	}
}
