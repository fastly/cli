package config

import (
	"fmt"
	"time"
)

// MigrateProfilesToAuth migrates existing profile entries into the new [auth]
// config section. It is safe to call multiple times; it only migrates profiles
// that do not already have a corresponding auth token entry.
//
// Returns true if any migration was performed.
func (f *File) MigrateProfilesToAuth() bool {
	if len(f.Profiles) == 0 {
		return false
	}

	if f.Auth.Tokens == nil {
		f.Auth.Tokens = make(AuthTokens)
	}

	migrated := false
	for name, p := range f.Profiles {
		if _, exists := f.Auth.Tokens[name]; exists {
			continue
		}

		entry := profileToAuthToken(p)
		f.Auth.Tokens[name] = entry

		if p.Default && f.Auth.Default == "" {
			f.Auth.Default = name
		}

		migrated = true
	}

	// If no default was set but we migrated something, pick the first one.
	if f.Auth.Default == "" && len(f.Auth.Tokens) > 0 {
		for name := range f.Auth.Tokens {
			f.Auth.Default = name
			break
		}
	}

	return migrated
}

// profileToAuthToken converts a legacy Profile to the new AuthToken format.
func profileToAuthToken(p *Profile) *AuthToken {
	entry := &AuthToken{
		Token:     p.Token,
		Email:     p.Email,
		AccountID: p.CustomerID,
	}

	// Determine if this is an SSO-backed profile.
	if p.AccessToken != "" || p.RefreshToken != "" || p.AccessTokenCreated != 0 || p.RefreshTokenCreated != 0 {
		entry.Type = AuthTokenTypeSSO
		entry.AccessToken = p.AccessToken
		entry.RefreshToken = p.RefreshToken

		if p.AccessTokenCreated != 0 && p.AccessTokenTTL != 0 {
			expiresAt := time.Unix(p.AccessTokenCreated, 0).Add(time.Duration(p.AccessTokenTTL) * time.Second)
			entry.AccessExpiresAt = expiresAt.Format(time.RFC3339)
		}
		if p.RefreshTokenCreated != 0 && p.RefreshTokenTTL != 0 {
			expiresAt := time.Unix(p.RefreshTokenCreated, 0).Add(time.Duration(p.RefreshTokenTTL) * time.Second)
			entry.RefreshExpiresAt = expiresAt.Format(time.RFC3339)
		}

		// Check if the SSO session is expired and cannot be auto-refreshed.
		if entry.RefreshExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, entry.RefreshExpiresAt)
			if err == nil && time.Now().After(t) {
				entry.NeedsReauth = true
			}
		}
	} else {
		entry.Type = AuthTokenTypeStatic
	}

	if p.CustomerName != "" {
		entry.Label = fmt.Sprintf("%s (%s)", p.CustomerName, p.Email)
	}

	return entry
}

func (f *File) AuthInitialized() bool {
	return len(f.Auth.Tokens) > 0
}

func (f *File) GetAuthToken(name string) *AuthToken {
	return f.Auth.Tokens[name]
}

func (f *File) GetDefaultAuthToken() (string, *AuthToken) {
	if f.Auth.Default == "" {
		return "", nil
	}
	if t := f.Auth.Tokens[f.Auth.Default]; t != nil {
		return f.Auth.Default, t
	}
	return "", nil
}

func (f *File) SetAuthToken(name string, token *AuthToken) {
	if f.Auth.Tokens == nil {
		f.Auth.Tokens = make(AuthTokens)
	}
	f.Auth.Tokens[name] = token
}

func (f *File) DeleteAuthToken(name string) bool {
	if f.Auth.Tokens == nil {
		return false
	}
	if _, ok := f.Auth.Tokens[name]; !ok {
		return false
	}
	delete(f.Auth.Tokens, name)
	if f.Auth.Default == name {
		f.Auth.Default = ""
		// Set a new default if possible.
		for n := range f.Auth.Tokens {
			f.Auth.Default = n
			break
		}
	}
	return true
}

func (f *File) SetDefaultAuthToken(name string) error {
	if f.Auth.Tokens == nil {
		return fmt.Errorf("no auth tokens configured")
	}
	if _, ok := f.Auth.Tokens[name]; !ok {
		return fmt.Errorf("token %q not found", name)
	}
	f.Auth.Default = name
	return nil
}
