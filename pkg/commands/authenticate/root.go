package authenticate

// NOTE: See the link for information on the 'Device Authorization Flow' used.
// https://auth0.com/docs/get-started/authentication-and-authorization-flow/device-authorization-flow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base
}

// Auth0DeviceCodeURL is the Auth0 device code URL.
const Auth0DeviceCodeURL = "https://dev-37kjpso9.us.auth0.com"

// Auth0ClientID is the Auth0 Client ID.
const Auth0ClientID = "7TAnqT4DTDhJTyXk9aXcuD48JoHRXK2X"

// Auth0Audience is the unique identifier of the API your app wants to access.
const Auth0Audience = "https://api.secretcdn-stg.net/"

// Auth0GrantType is an extension grant type (MUST be URL encoded).
var Auth0GrantType = url.QueryEscape("urn:ietf:params:oauth:grant-type:device_code")

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, globals *config.Data) *RootCommand {
	var c RootCommand
	c.Globals = globals
	c.CmdClause = parent.Command("authenticate", "Authenticate with Fastly (returns temporary, auto-rotated, API token)")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	deviceCodeResponse, err := getDeviceCode()
	if err != nil {
		return err
	}

	intro := "Please open the following URL and enter your user code: " + deviceCodeResponse.UserCode
	text.Description(out, intro, deviceCodeResponse.VerificationURI)

	var accessTokenResponse chan *AccessTokenResponse

	interval := time.Duration(deviceCodeResponse.Interval) * time.Second
	deviceCodeExpiration := time.Duration(deviceCodeResponse.ExpiresIn) * time.Second

	go pollForAccessToken(
		deviceCodeResponse.DeviceCode,
		interval,
		deviceCodeExpiration,
		accessTokenResponse,
		c.Globals.ErrLog,
	)

	select {
	case atr := <-accessTokenResponse:
		fmt.Printf("%+v\n", atr)
	case <-time.After(deviceCodeExpiration):
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("user code expired"),
			Remediation: "Please re-run the command and complete the authorization flow.",
		}
	}

	return nil
}

// getDeviceCode retrieves a device code from Auth0.
func getDeviceCode() (deviceCodeResponse DeviceCodeResponse, err error) {
	path := "/oauth/device/code"

	// TODO: In the future we may want to restrict the API scope (see 'scope').
	// https://auth0.com/docs/get-started/authentication-and-authorization-flow/call-your-api-using-the-device-authorization-flow#device-code-parameters
	payload := fmt.Sprintf("client_id=%s&audience=%s", Auth0ClientID, url.QueryEscape(Auth0Audience))

	req, err := http.NewRequest("POST", Auth0DeviceCodeURL+path, strings.NewReader(payload))
	if err != nil {
		return deviceCodeResponse, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return deviceCodeResponse, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return deviceCodeResponse, err
	}

	err = json.Unmarshal(body, &deviceCodeResponse)
	if err != nil {
		return deviceCodeResponse, err
	}

	return deviceCodeResponse, nil
}

// DeviceCodeResponse is the API response for an Auth0 Device Code request.
type DeviceCodeResponse struct {
	// DeviceCode is the unique code for the device.
	DeviceCode string `json:"device_code"`
	// ExpiresIn indicates the lifetime (in seconds) of the device_code and user_code.
	ExpiresIn int `json:"expires_in"`
	// Interval indicates the interval (in seconds) at which the app should poll the token URL to request a token.
	Interval int `json:"interval"`
	// UserCode contains the code that should be input at the verification_uri to authorize the device.
	UserCode string `json:"user_code"`
	// VerificationURI contains the URL the user should visit to authorize the device.
	VerificationURI string `json:"verification_uri"`
	// VerificationURIComplete contains the complete URL the user should visit to authorize the device.
	VerificationURIComplete string `json:"verification_uri_complete"`
}

func pollForAccessToken(
	deviceCode string,
	interval time.Duration,
	deviceCodeExpiration time.Duration,
	accessTokenResponse chan *AccessTokenResponse,
	errLog fsterr.LogInterface,
) {
	path := "/oauth/token"
	payload := fmt.Sprintf("grant_type=%s&device_code=%s&client_id=%s", Auth0GrantType, deviceCode, Auth0ClientID)
	ctx := map[string]any{
		"path":    path,
		"payload": payload,
	}

	req, err := http.NewRequest("POST", Auth0DeviceCodeURL+path, strings.NewReader(payload))
	if err != nil {
		errLog.AddWithContext(err, ctx)
		return
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	done := make(chan bool)
	go func() {
		time.Sleep(deviceCodeExpiration)
		done <- true
	}()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// NOTE: We extract the logic into a func to avoid a defer within a loop.
			checkAccessToken(req, errLog, ctx, accessTokenResponse, done)
		}
	}
}

func checkAccessToken(
	req *http.Request,
	errLog fsterr.LogInterface,
	ctx map[string]any,
	accessTokenResponse chan *AccessTokenResponse,
	done chan bool,
) {
	// TODO: Handle all the different error scenarios appropriately.
	// https://auth0.com/docs/get-started/authentication-and-authorization-flow/call-your-api-using-the-device-authorization-flow#token-responses
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errLog.AddWithContext(err, ctx)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		errLog.AddWithContext(err, ctx)
		return
	}

	var atr *AccessTokenResponse
	err = json.Unmarshal(body, atr)
	if err != nil {
		errLog.AddWithContext(err, ctx)
		return
	}

	done <- true
	accessTokenResponse <- atr
}

// AccessTokenResponse is the API response for an Auth0 Access Token request.
type AccessTokenResponse struct {
	// AccessToken can be exchanged for a Fastly API token.
	AccessToken string `json:"access_token"`
	// ExpiresIn indicates the lifetime (in seconds) of the access token.
	ExpiresIn int `json:"expires_in"`
	// IDToken contains user information that must be decoded and extracted.
	IDToken string `json:"id_token"`
	// RefreshToken is used to obtain a new Access Token or ID Token after the previous one has expired.
	RefreshToken string `json:"refresh_token"`
	// TokenType indicates which HTTP authentication scheme is used (e.g. Bearer).
	TokenType string `json:"token_type"`
}
