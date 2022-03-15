package auth

import (
	"fmt"
	"io"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// ValidateToken retrieves the given token data and validates if it's OK to
// continue using.
func ValidateToken(apiClient api.Interface, stdout io.Writer) (ok bool) {
	t, err := apiClient.GetTokenSelf()
	if err != nil {
		text.Info(stdout, "We were not able to validate your access token.")
		return false
	}

	if t.ExpiresAt == nil {
		text.Info(stdout, "Your current access token has no expiration and will be replaced with a short-lived token.")
		return false
	}

	if (*t.ExpiresAt).Before(time.Now()) {
		text.Info(stdout, "Your access token has expired.")
		return false
	}

	return true
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
