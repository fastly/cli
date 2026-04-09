package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/fastly/go-fastly/v14/fastly"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
)

// TokenMetadata holds validated token information from the Fastly API.
type TokenMetadata struct {
	Email     string
	AccountID string

	APITokenName      string
	APITokenScope     string
	APITokenExpiresAt string
	APITokenID        string
}

// FetchTokenMetadata validates a token by calling GetCurrentUser (required)
// and GetTokenSelf (best-effort), returning metadata for storage.
// It constructs its own API client from the provided token string.
func FetchTokenMetadata(g *global.Data, token string) (*TokenMetadata, error) {
	endpoint, _ := g.APIEndpoint()
	apiClient, err := g.APIClientFactory(token, endpoint, g.Flags.Debug)
	if err != nil {
		return nil, fmt.Errorf("error creating API client: %w", err)
	}

	user, err := apiClient.GetCurrentUser(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("token validation failed (could not look up current user): %w", err)
	}

	md := &TokenMetadata{
		Email:     fastly.ToValue(user.Login),
		AccountID: fastly.ToValue(user.CustomerID),
	}

	fetchTokenSelf(g, apiClient, md)

	return md, nil
}

// FetchTokenMetadataLenient is like FetchTokenMetadata but treats
// GetCurrentUser as best-effort. Use this for scoped tokens that may lack
// permission to call /current_user. At least one of GetCurrentUser or
// GetTokenSelf must succeed to confirm the token is valid.
func FetchTokenMetadataLenient(g *global.Data, token string) (*TokenMetadata, error) {
	endpoint, _ := g.APIEndpoint()
	apiClient, err := g.APIClientFactory(token, endpoint, g.Flags.Debug)
	if err != nil {
		return nil, fmt.Errorf("error creating API client: %w", err)
	}

	md := &TokenMetadata{}
	anyOK := false

	user, err := apiClient.GetCurrentUser(context.TODO())
	if err != nil {
		g.ErrLog.Add(fmt.Errorf("GetCurrentUser failed (best-effort): %w", err))
	} else {
		md.Email = fastly.ToValue(user.Login)
		md.AccountID = fastly.ToValue(user.CustomerID)
		anyOK = true
	}

	if fetchTokenSelf(g, apiClient, md) {
		anyOK = true
	}

	if !anyOK {
		return nil, fmt.Errorf("token validation failed: neither /current_user nor /tokens/self responded successfully")
	}

	return md, nil
}

// EnrichWithTokenSelf calls GetTokenSelf to populate API token metadata on
// an existing AuthToken. It constructs its own API client from the token.
// This is best-effort: failures are logged and existing fields are preserved.
func EnrichWithTokenSelf(g *global.Data, at *config.AuthToken) {
	endpoint, _ := g.APIEndpoint()
	apiClient, err := g.APIClientFactory(at.Token, endpoint, g.Flags.Debug)
	if err != nil {
		g.ErrLog.Add(fmt.Errorf("EnrichWithTokenSelf: error creating API client: %w", err))
		return
	}

	var md TokenMetadata
	if fetchTokenSelf(g, apiClient, &md) {
		at.APITokenName = md.APITokenName
		at.APITokenScope = md.APITokenScope
		at.APITokenExpiresAt = md.APITokenExpiresAt
		at.APITokenID = md.APITokenID
	}
}

// BuildAndStoreStaticToken constructs an AuthToken from pre-fetched metadata
// and stores it under the given name. Does NOT write config to disk.
func BuildAndStoreStaticToken(g *global.Data, token, name string, md *TokenMetadata, makeDefault bool) {
	entry := &config.AuthToken{
		Type:              config.AuthTokenTypeStatic,
		Token:             token,
		Email:             md.Email,
		AccountID:         md.AccountID,
		APITokenName:      md.APITokenName,
		APITokenScope:     md.APITokenScope,
		APITokenExpiresAt: md.APITokenExpiresAt,
		APITokenID:        md.APITokenID,
	}

	g.Config.SetAuthToken(name, entry)

	if makeDefault {
		g.Config.Auth.Default = name
	}
}

// StoreStaticToken validates a raw API token, fetches metadata, and stores it
// in the auth config as the default token. Returns the stored name and
// metadata.
func StoreStaticToken(g *global.Data, token string) (name string, md *TokenMetadata, err error) {
	md, err = FetchTokenMetadata(g, token)
	if err != nil {
		return "", nil, err
	}

	name = md.APITokenName
	if name == "" {
		name = "default"
	}

	BuildAndStoreStaticToken(g, token, name, md, true)

	if err := g.Config.Write(g.ConfigPath); err != nil {
		return "", nil, fmt.Errorf("error saving config: %w", err)
	}

	return name, md, nil
}

// fetchTokenSelf calls GetTokenSelf and populates md on success.
// Returns true if the call succeeded and md was populated, false otherwise.
// Failures (including panics from unmocked test doubles) are logged.
func fetchTokenSelf(g *global.Data, apiClient api.Interface, md *TokenMetadata) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			g.ErrLog.Add(fmt.Errorf("GetTokenSelf panicked (best-effort): %v", r))
			ok = false
		}
	}()

	tok, err := apiClient.GetTokenSelf(context.TODO())
	if err != nil {
		g.ErrLog.Add(fmt.Errorf("GetTokenSelf failed (best-effort): %w", err))
		return false
	}

	md.APITokenName = fastly.ToValue(tok.Name)
	if tok.Scope != nil {
		md.APITokenScope = string(*tok.Scope)
	}
	if tok.ExpiresAt != nil {
		md.APITokenExpiresAt = tok.ExpiresAt.Format(time.RFC3339)
	}
	md.APITokenID = fastly.ToValue(tok.TokenID)
	return true
}
