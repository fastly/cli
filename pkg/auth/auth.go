package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/cap/jwt"
	"github.com/hashicorp/cap/oidc"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/api/undocumented"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// Remediation is a generic remediation message for an error authorizing.
const Remediation = "Please re-run the command. If the problem persists, please file an issue: https://github.com/fastly/cli/issues/new?labels=bug&template=bug_report.md"

// ClientID is the auth provider's Client ID.
const ClientID = "fastly-cli"

// RedirectURL is the endpoint the auth provider will pass an authorization code to.
const RedirectURL = "http://localhost:8080/callback"

// OIDCMetadata is OpenID Connect's metadata discovery mechanism.
// https://swagger.io/docs/specification/authentication/openid-connect-discovery/
const OIDCMetadata = "https://accounts.fastly.com/realms/fastly/.well-known/openid-configuration"

// WellKnownEndpoints represents the OpenID Connect metadata.
type WellKnownEndpoints struct {
	// Auth is the authorization_endpoint.
	Auth string `json:"authorization_endpoint"`
	// Certs is the jwks_uri.
	Certs string `json:"jwks_uri"`
	// Token is the token_endpoint.
	Token string `json:"token_endpoint"`
}

// Starter defines the behaviour for the authentication server.
type Starter interface {
	// GetResult returns the results channel
	GetResult() chan AuthorizationResult
	// SetAccountEndpoint sets the account endpoint.
	SetAccountEndpoint(endpoint string)
	// SetEndpoint sets the API endpoint.
	SetAPIEndpoint(endpoint string)
	// SetVerifier sets the code verifier.
	SetVerifier(verifier *oidc.S256Verifier)
	// Start starts a local server for handling authentication processing.
	Start() error
}

// Server is a local server responsible for authentication processing.
type Server struct {
	// APIEndpoint is the API endpoint.
	APIEndpoint string
	// AccountEndpoint is the accounts endpoint.
	AccountEndpoint string
	// DebugMode indicates to the CLI it can display debug information.
	DebugMode string
	// HTTPClient is a HTTP client used to call the API to exchange the access token for a session token.
	HTTPClient api.HTTPClient
	// Result is a channel that reports the result of authorization.
	Result chan AuthorizationResult
	// Router is an HTTP request multiplexer.
	Router *http.ServeMux
	// Verifier represents an OAuth PKCE code verifier that uses the S256 challenge method.
	Verifier *oidc.S256Verifier
	// WellKnownEndpoints is the .well-known metadata.
	WellKnownEndpoints WellKnownEndpoints
}

// GetResult returns the result channel.
func (s Server) GetResult() chan AuthorizationResult {
	return s.Result
}

// SetAccountEndpoint sets the account endpoint.
func (s *Server) SetAccountEndpoint(endpoint string) {
	s.AccountEndpoint = endpoint
}

// SetAPIEndpoint sets the API endpoint.
func (s *Server) SetAPIEndpoint(endpoint string) {
	s.APIEndpoint = endpoint
}

// SetVerifier sets the code verifier endpoint.
func (s *Server) SetVerifier(verifier *oidc.S256Verifier) {
	s.Verifier = verifier
}

// Start starts a local server for handling authentication processing.
func (s *Server) Start() error {
	server := &http.Server{
		Addr:         ":8080",
		Handler:      s.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to start local server: %w", err),
			Remediation: Remediation,
		}
	}
	return nil
}

// HandleCallback processes the callback from the authentication service.
func (s *Server) HandleCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authorizationCode := r.URL.Query().Get("code")
		if authorizationCode == "" {
			fmt.Fprint(w, "ERROR: no authorization code returned\n")
			s.Result <- AuthorizationResult{
				Err: fmt.Errorf("no authorization code returned"),
			}
			return
		}

		// Exchange the authorization code and the code verifier for a JWT.
		// NOTE: I use the identifier `j` to avoid overlap with the `jwt` package.
		codeVerifier := s.Verifier.Verifier()
		j, err := GetJWT(s.AccountEndpoint, codeVerifier, authorizationCode)
		if err != nil || j.AccessToken == "" || j.IDToken == "" {
			fmt.Fprint(w, "ERROR: failed to exchange code for JWT\n")
			s.Result <- AuthorizationResult{
				Err: fmt.Errorf("failed to exchange code for JWT"),
			}
			return
		}

		email, at, err := ValidateAndRetrieveAPIToken(s.AccountEndpoint, s.APIEndpoint, j.AccessToken, s.DebugMode, s.HTTPClient)
		if err != nil {
			s.Result <- AuthorizationResult{
				Err: err,
			}
			return
		}

		fmt.Fprint(w, "Authenticated successfully. Please close this page and return to the Fastly CLI in your terminal.")
		s.Result <- AuthorizationResult{
			Email:        email,
			Jwt:          j,
			SessionToken: at.AccessToken,
		}
	}
}

// ValidateAndRetrieveAPIToken verifies the signature and the claims and
// exchanges the access token for an API token.
//
// NOTE: This function exists as it's called by this package + app.Run().
func ValidateAndRetrieveAPIToken(accountEndpoint, apiEndpoint, accessToken, debugMode string, httpClient api.HTTPClient) (string, *APIToken, error) {
	claims, err := VerifyJWTSignature(accountEndpoint, accessToken)
	if err != nil {
		return "", nil, err
	}

	azp, ok := claims["azp"]
	if !ok {
		return "", nil, errors.New("failed to extract azp from JWT claims")
	}
	if azp != ClientID {
		if !ok {
			return "", nil, fmt.Errorf("failed to match expected azp: %s", azp)
		}
	}

	aud, ok := claims["aud"]
	if !ok {
		return "", nil, errors.New("failed to extract aud from JWT claims")
	}

	if aud != apiEndpoint {
		if !ok {
			return "", nil, fmt.Errorf("failed to match expected aud: %s", apiEndpoint)
		}
	}

	email, ok := claims["email"]
	if !ok {
		return "", nil, errors.New("failed to extract email from JWT claims")
	}

	// Exchange the access token for a Fastly API token.
	at, err := ExchangeAccessToken(accessToken, apiEndpoint, httpClient, debugMode)
	if err != nil {
		return "", nil, fmt.Errorf("failed to exchange access token for an API token: %w", err)
	}

	e, ok := email.(string)
	if !ok {
		return "", nil, fmt.Errorf("failed to type assert 'email' (%#v) to a string", email)
	}
	return e, at, nil
}

// APIToken is returned from the /login-enhanced endpoint.
type APIToken struct {
	// AccessToken is used to access the Fastly API.
	AccessToken string `json:"access_token"`
	// CustomerID is the customer ID.
	CustomerID string `json:"customer_id"`
	// ExpiresAt is when the access token will expire.
	ExpiresAt string `json:"expires_at"`
	// ID is a unique ID.
	ID string `json:"id"`
	// Name is a description of the token.
	Name string `json:"name"`
	// UserID is the user's ID.
	UserID string `json:"user_id"`
}

// AuthorizationResult represents the result of the authorization process.
type AuthorizationResult struct {
	// Email address extracted from JWT claims.
	Email string
	// Err is any error received during authentication.
	Err error
	// Jwt is the JWT token returned by the authorization server.
	Jwt JWT
	// SessionToken is a temporary API token.
	SessionToken string
}

// GenVerifier creates a code verifier.
func GenVerifier() (*oidc.S256Verifier, error) {
	return oidc.NewCodeVerifier()
}

// GenURL constructs the required authorization_endpoint path.
func GenURL(accountEndpoint, apiEndpoint string, verifier *oidc.S256Verifier) (string, error) {
	challenge, err := oidc.CreateCodeChallenge(verifier)
	if err != nil {
		return "", err
	}

	authorizationURL := fmt.Sprintf(
		"%s/realms/fastly/protocol/openid-connect/auth?audience=%s"+
			"&scope=openid"+
			"&response_type=code&client_id=%s"+
			"&code_challenge=%s"+
			"&code_challenge_method=S256&redirect_uri=%s",
		accountEndpoint, apiEndpoint, ClientID, challenge, RedirectURL)

	return authorizationURL, nil
}

// GetJWT constructs and calls the token_endpoint path, returning a JWT
// containing the access and refresh tokens and associated TTLs.
func GetJWT(accountEndpoint, codeVerifier, authorizationCode string) (JWT, error) {
	path := "/realms/fastly/protocol/openid-connect/token"

	payload := fmt.Sprintf(
		"grant_type=authorization_code&client_id=%s&code_verifier=%s&code=%s&redirect_uri=%s",
		ClientID,
		codeVerifier,
		authorizationCode,
		"http://localhost:8080/callback", // NOTE: not redirected to, just a security check.
	)

	req, err := http.NewRequest("POST", accountEndpoint+path, strings.NewReader(payload))
	if err != nil {
		return JWT{}, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return JWT{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return JWT{}, fmt.Errorf("failed to exchange code for jwt (status: %s)", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return JWT{}, err
	}

	var j JWT
	err = json.Unmarshal(body, &j)
	if err != nil {
		return JWT{}, err
	}

	return j, nil
}

// JWT is the API response for a Token request.
//
// Access Token typically has a TTL of 5mins.
// Refresh Token typically has a TTL of 30mins.
type JWT struct {
	// AccessToken can be exchanged for a Fastly API token.
	AccessToken string `json:"access_token"`
	// ExpiresIn indicates the lifetime (in seconds) of the access token.
	ExpiresIn int `json:"expires_in"`
	// IDToken contains user information that must be decoded and extracted.
	IDToken string `json:"id_token"`
	// RefreshExpiresIn indicates the lifetime (in seconds) of the refresh token.
	RefreshExpiresIn int `json:"refresh_expires_in"`
	// RefreshToken contains a token used to refresh the issued access token.
	RefreshToken string `json:"refresh_token"`
	// TokenType indicates which HTTP authentication scheme is used (e.g. Bearer).
	TokenType string `json:"token_type"`
}

// VerifyJWTSignature calls the jwks_uri endpoint and extracts its claims.
func VerifyJWTSignature(accountEndpoint, token string) (claims map[string]any, err error) {
	ctx := context.Background()
	path := "/realms/fastly/protocol/openid-connect/certs"

	// NOTE: The last argument is optional and is for validating the JWKs endpoint
	// (which we don't need to do, so we pass an empty string)
	keySet, err := jwt.NewJSONWebKeySet(ctx, accountEndpoint+path, "")
	if err != nil {
		return claims, fmt.Errorf("failed to verify signature of access token: %w", err)
	}

	claims, err = keySet.VerifySignature(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signature of access token: %w", err)
	}

	return claims, nil
}

// RefreshAccessToken constructs and calls the token_endpoint with the
// refresh token so we can refresh and return the access token.
func RefreshAccessToken(accountEndpoint, refreshToken string) (JWT, error) {
	path := "/realms/fastly/protocol/openid-connect/token"

	payload := fmt.Sprintf(
		"grant_type=refresh_token&client_id=%s&refresh_token=%s",
		ClientID,
		refreshToken,
	)

	req, err := http.NewRequest("POST", accountEndpoint+path, strings.NewReader(payload))
	if err != nil {
		return JWT{}, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return JWT{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return JWT{}, err
	}

	if res.StatusCode != http.StatusOK {
		return JWT{}, fmt.Errorf("failed to refresh the access token (status: %s)", res.Status)
	}

	var j JWT
	err = json.Unmarshal(body, &j)
	if err != nil {
		return JWT{}, err
	}

	return j, nil
}

// ExchangeAccessToken exchanges `accessToken` for a Fastly API token.
func ExchangeAccessToken(accessToken, apiEndpoint string, httpClient api.HTTPClient, debugMode string) (*APIToken, error) {
	debug, _ := strconv.ParseBool(debugMode)
	resp, err := undocumented.Call(undocumented.CallOptions{
		APIEndpoint: apiEndpoint,
		HTTPClient:  httpClient,
		HTTPHeaders: []undocumented.HTTPHeader{
			{
				Key:   "Authorization",
				Value: fmt.Sprintf("Bearer %s", accessToken),
			},
		},
		Method: http.MethodPost,
		Path:   "/login-enhanced",
		Debug:  debug,
	})
	if err != nil {
		if apiErr, ok := err.(undocumented.APIError); ok {
			if apiErr.StatusCode != http.StatusConflict {
				err = fmt.Errorf("%w: %d %s", err, apiErr.StatusCode, http.StatusText(apiErr.StatusCode))
			}
		}
		return nil, err
	}

	at := &APIToken{}
	err = json.Unmarshal(resp, at)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json containing API token: %w", err)
	}

	return at, nil
}

// TokenExpired indicates if the specified TTL has past.
func TokenExpired(ttl int, timestamp int64) bool {
	d := time.Duration(ttl) * time.Second
	ttlAgo := time.Now().Add(-d).Unix()
	return timestamp < ttlAgo
}
