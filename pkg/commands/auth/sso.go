package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/fastly/cli/pkg/api/undocumented"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/useragent"
)

// RunSSO executes the SSO authentication flow as a plain function.
// It derives the token name from the current context via identifyTokenName,
// which preserves naming behavior for re-auth flows (refresh/expired tokens).
func RunSSO(in io.Reader, out io.Writer, g *global.Data, forceReAuth bool, skipPrompt bool) error {
	return RunSSOWithTokenName(in, out, g, forceReAuth, skipPrompt, identifyTokenName(g))
}

// RunSSOWithTokenName is like RunSSO but accepts an explicit token name
// instead of deriving it from the current context.
func RunSSOWithTokenName(in io.Reader, out io.Writer, g *global.Data, forceReAuth bool, skipPrompt bool, tokenName string) error {
	if forceReAuth {
		g.AuthServer.SetParam("prompt", "login select_account")
	} else {
		if at := g.Config.GetAuthToken(tokenName); at != nil {
			g.AuthServer.SetParam("prompt", "login")
			if at.Email != "" {
				g.AuthServer.SetParam("login_hint", at.Email)
			}
			if at.AccountID != "" {
				g.AuthServer.SetParam("account_hint", at.AccountID)
			}
		} else {
			g.AuthServer.SetParam("prompt", "login select_account")
		}
	}

	if !skipPrompt && !g.Flags.AutoYes && !g.Flags.NonInteractive {
		msg := fmt.Sprintf("We're going to authenticate the '%s' token", tokenName)
		text.Important(out, "%s. We need to open your browser to authenticate you.", msg)
		text.Break(out)
		cont, err := text.AskYesNo(out, text.BoldYellow("Do you want to continue? [y/N]: "), in)
		text.Break(out)
		if err != nil {
			return err
		}
		if !cont {
			return fsterr.SkipExitError{
				Skip: true,
				Err:  fsterr.ErrDontContinue,
			}
		}
	}

	var serverErr error
	go func() {
		err := g.AuthServer.Start()
		if err != nil {
			serverErr = err
		}
	}()
	if serverErr != nil {
		return serverErr
	}

	text.Info(out, "Starting a local server to handle the authentication flow.")

	authorizationURL, err := g.AuthServer.AuthURL()
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to generate an authorization URL: %w", err),
			Remediation: auth.Remediation,
		}
	}

	text.Break(out)
	text.Description(out, "We're opening the following URL in your default web browser so you may authenticate with Fastly", authorizationURL)

	err = g.Opener(authorizationURL)
	if err != nil {
		return fmt.Errorf("failed to open your default browser: %w", err)
	}

	ar := <-g.AuthServer.GetResult()
	if ar.Err != nil || ar.SessionToken == "" {
		err := ar.Err
		if ar.Err == nil {
			err = errors.New("no session token")
		}
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to authorize: %w", err),
			Remediation: auth.Remediation,
		}
	}

	customerID, customerName, err := processCustomer(g, ar)
	if err != nil {
		return fmt.Errorf("failed to use session token to get customer data: %w", err)
	}

	err = storeAuthToken(g, ar, tokenName, customerID, customerName)
	if err != nil {
		g.ErrLog.Add(err)
		return fmt.Errorf("failed to store auth token: %w", err)
	}

	msg := fmt.Sprintf("Session token '%s' has been stored.", tokenName)
	if !env.AuthCommandDisabled() {
		msg += " Use 'fastly auth list' to view tokens."
	}
	text.Success(out, msg)
	text.Info(out, "Token saved to %s", g.ConfigPath)
	return nil
}

// identifyTokenName determines which auth token name to use for SSO.
func identifyTokenName(g *global.Data) string {
	if g.Flags.Token != "" {
		return g.Flags.Token
	}
	if g.Manifest != nil && g.Manifest.File.Profile != "" {
		return g.Manifest.File.Profile
	}
	if name, _ := g.Config.GetDefaultAuthToken(); name != "" {
		return name
	}
	return "default"
}

// CurrentCustomerResponse models the Fastly API response for the
// /current_customer endpoint.
type CurrentCustomerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func processCustomer(g *global.Data, ar auth.AuthorizationResult) (customerID, customerName string, err error) {
	debugMode, _ := strconv.ParseBool(g.Env.DebugMode)
	apiEndpoint, _ := g.APIEndpoint()
	data, err := undocumented.Call(undocumented.CallOptions{
		APIEndpoint: apiEndpoint,
		HTTPClient:  g.HTTPClient,
		HTTPHeaders: []undocumented.HTTPHeader{
			{
				Key:   "Accept",
				Value: "application/json",
			},
			{
				Key:   "User-Agent",
				Value: useragent.Name,
			},
		},
		Method: http.MethodGet,
		Path:   "/current_customer",
		Token:  ar.SessionToken,
		Debug:  debugMode,
	})
	if err != nil {
		g.ErrLog.Add(err)
		return "", "", fmt.Errorf("error executing current_customer API request: %w", err)
	}

	var response CurrentCustomerResponse
	if err := json.Unmarshal(data, &response); err != nil {
		g.ErrLog.Add(err)
		return "", "", fmt.Errorf("error decoding current_customer API response: %w", err)
	}

	return response.ID, response.Name, nil
}

func storeAuthToken(g *global.Data, ar auth.AuthorizationResult, tokenName, customerID, customerName string) error {
	now := time.Now()
	label := customerName
	if ar.Email != "" {
		label = fmt.Sprintf("%s (%s)", customerName, ar.Email)
	}

	at := &config.AuthToken{
		Type:             config.AuthTokenTypeSSO,
		Token:            ar.SessionToken,
		Label:            label,
		AccountID:        customerID,
		Email:            ar.Email,
		AccessToken:      ar.Jwt.AccessToken,
		RefreshToken:     ar.Jwt.RefreshToken,
		AccessExpiresAt:  now.Add(time.Duration(ar.Jwt.ExpiresIn) * time.Second).Format(time.RFC3339),
		RefreshExpiresAt: now.Add(time.Duration(ar.Jwt.RefreshExpiresIn) * time.Second).Format(time.RFC3339),
	}

	EnrichWithTokenSelf(g, at)

	g.Config.SetAuthToken(tokenName, at)

	if g.Config.Auth.Default == "" {
		g.Config.Auth.Default = tokenName
	}

	if err := g.Config.Write(g.ConfigPath); err != nil {
		return fmt.Errorf("failed to update config file: %w", err)
	}
	return nil
}
