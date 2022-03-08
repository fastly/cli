package auth

import (
	"fmt"
	"net/http"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v6/fastly"
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

// PersistToken stores the access token and user email in the local config.
func PersistToken(token, cfgPath, apiEndpoint string, cfg *config.File, cf api.ClientFactory) error {
	c, err := cf(token, apiEndpoint)
	if err != nil {
		return fmt.Errorf("error constructing Fastly API client: %w", err)
	}

	ts, err := c.GetTokenSelf()
	if err != nil {
		return fmt.Errorf("error validating token: %w", err)
	}
	u, err := c.GetUser(&fastly.GetUserInput{
		ID: ts.UserID,
	})
	if err != nil {
		return fmt.Errorf("error fetching token user: %w", err)
	}

	cfg.User.Email = u.Login
	cfg.User.Token = token
	err = cfg.Write(cfgPath)
	if err != nil {
		return err
	}

	return nil
}
