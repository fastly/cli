package config

// Auth represents the new auth configuration section.
// It stores named tokens and tracks which one is the default.
type Auth struct {
	Default string     `toml:"default" json:"default"`
	Tokens  AuthTokens `toml:"tokens" json:"tokens"`
}

// AuthTokens is a map of token name to token entry.
type AuthTokens map[string]*AuthToken

// AuthToken represents a single stored credential.
type AuthToken struct {
	Type      string `toml:"type" json:"type"`
	Token     string `toml:"token" json:"token"`
	Label     string `toml:"label,omitempty" json:"label,omitempty"`
	AccountID string `toml:"account_id,omitempty" json:"account_id,omitempty"`
	Email     string `toml:"email,omitempty" json:"email,omitempty"`

	// API token metadata (populated from /tokens/self when available).

	APITokenName      string `toml:"api_token_name,omitempty" json:"api_token_name,omitempty"`
	APITokenScope     string `toml:"api_token_scope,omitempty" json:"api_token_scope,omitempty"`
	APITokenExpiresAt string `toml:"api_token_expires_at,omitempty" json:"api_token_expires_at,omitempty"`
	APITokenID        string `toml:"api_token_id,omitempty" json:"api_token_id,omitempty"`

	// SSO-specific fields (only populated when Type == "sso").

	RefreshToken     string `toml:"refresh_token,omitempty" json:"refresh_token,omitempty"`
	AccessExpiresAt  string `toml:"access_expires_at,omitempty" json:"access_expires_at,omitempty"`
	RefreshExpiresAt string `toml:"refresh_expires_at,omitempty" json:"refresh_expires_at,omitempty"`
	AccessToken      string `toml:"access_token,omitempty" json:"access_token,omitempty"`
	NeedsReauth      bool   `toml:"needs_reauth,omitempty" json:"needs_reauth,omitempty"`
}

const AuthTokenTypeStatic = "static"

const AuthTokenTypeSSO = "sso"
