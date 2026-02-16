package config

import (
	"testing"
	"time"
)

func TestMigrateProfilesToAuth_EmptyProfiles(t *testing.T) {
	t.Parallel()

	f := &File{}
	migrated := f.MigrateProfilesToAuth()

	if migrated {
		t.Fatal("expected no migration when profiles are empty")
	}
	if f.Auth.Tokens != nil {
		t.Fatalf("expected Auth.Tokens to remain nil, got %v", f.Auth.Tokens)
	}
	if f.Auth.Default != "" {
		t.Fatalf("expected Auth.Default to be empty, got %q", f.Auth.Default)
	}
}

func TestMigrateProfilesToAuth_StaticToken(t *testing.T) {
	t.Parallel()

	f := &File{
		Profiles: Profiles{
			"work": {
				Token:      "tok_abc123",
				Email:      "dev@example.com",
				CustomerID: "cust_42",
			},
		},
	}

	migrated := f.MigrateProfilesToAuth()
	if !migrated {
		t.Fatal("expected migration to occur")
	}

	tok := f.Auth.Tokens["work"]
	if tok == nil {
		t.Fatal("expected auth token 'work' to exist after migration")
	}
	if tok.Type != AuthTokenTypeStatic {
		t.Fatalf("expected type %q, got %q", AuthTokenTypeStatic, tok.Type)
	}
	if tok.Token != "tok_abc123" {
		t.Fatalf("expected token %q, got %q", "tok_abc123", tok.Token)
	}
	if tok.Email != "dev@example.com" {
		t.Fatalf("expected email %q, got %q", "dev@example.com", tok.Email)
	}
	if tok.AccountID != "cust_42" {
		t.Fatalf("expected account_id %q, got %q", "cust_42", tok.AccountID)
	}
	if tok.AccessToken != "" {
		t.Fatalf("expected empty access_token for static type, got %q", tok.AccessToken)
	}
	if tok.RefreshToken != "" {
		t.Fatalf("expected empty refresh_token for static type, got %q", tok.RefreshToken)
	}
}

func TestMigrateProfilesToAuth_SSOToken(t *testing.T) {
	t.Parallel()

	// Use timestamps far in the future so NeedsReauth stays false.
	now := time.Now()
	created := now.Unix()
	ttlSeconds := 86400 // 24 hours

	f := &File{
		Profiles: Profiles{
			"sso-user": {
				Token:               "tok_sso",
				Email:               "sso@example.com",
				CustomerID:          "cust_99",
				CustomerName:        "Acme Corp",
				AccessToken:         "access_xyz",
				AccessTokenCreated:  created,
				AccessTokenTTL:      ttlSeconds,
				RefreshToken:        "refresh_xyz",
				RefreshTokenCreated: created,
				RefreshTokenTTL:     ttlSeconds * 30, // 30 days
			},
		},
	}

	migrated := f.MigrateProfilesToAuth()
	if !migrated {
		t.Fatal("expected migration to occur")
	}

	tok := f.Auth.Tokens["sso-user"]
	if tok == nil {
		t.Fatal("expected auth token 'sso-user' to exist after migration")
	}
	if tok.Type != AuthTokenTypeSSO {
		t.Fatalf("expected type %q, got %q", AuthTokenTypeSSO, tok.Type)
	}
	if tok.Token != "tok_sso" {
		t.Fatalf("expected token %q, got %q", "tok_sso", tok.Token)
	}
	if tok.AccessToken != "access_xyz" {
		t.Fatalf("expected access_token %q, got %q", "access_xyz", tok.AccessToken)
	}
	if tok.RefreshToken != "refresh_xyz" {
		t.Fatalf("expected refresh_token %q, got %q", "refresh_xyz", tok.RefreshToken)
	}
	if tok.Label != "Acme Corp (sso@example.com)" {
		t.Fatalf("expected label %q, got %q", "Acme Corp (sso@example.com)", tok.Label)
	}

	// Verify expiry timestamps were computed.
	expectedAccessExpiry := time.Unix(created, 0).Add(time.Duration(ttlSeconds) * time.Second).Format(time.RFC3339)
	if tok.AccessExpiresAt != expectedAccessExpiry {
		t.Fatalf("expected access_expires_at %q, got %q", expectedAccessExpiry, tok.AccessExpiresAt)
	}
	expectedRefreshExpiry := time.Unix(created, 0).Add(time.Duration(ttlSeconds*30) * time.Second).Format(time.RFC3339)
	if tok.RefreshExpiresAt != expectedRefreshExpiry {
		t.Fatalf("expected refresh_expires_at %q, got %q", expectedRefreshExpiry, tok.RefreshExpiresAt)
	}
	if tok.NeedsReauth {
		t.Fatal("expected NeedsReauth to be false for a token with a future refresh expiry")
	}
}

func TestMigrateProfilesToAuth_SSOToken_NeedsReauth(t *testing.T) {
	t.Parallel()

	// Use timestamps in the past so the refresh token is expired.
	pastCreated := time.Now().Add(-48 * time.Hour).Unix()
	ttlSeconds := 3600 // 1 hour -- well in the past

	f := &File{
		Profiles: Profiles{
			"expired-sso": {
				Token:               "tok_old",
				Email:               "old@example.com",
				AccessToken:         "access_old",
				AccessTokenCreated:  pastCreated,
				AccessTokenTTL:      ttlSeconds,
				RefreshToken:        "refresh_old",
				RefreshTokenCreated: pastCreated,
				RefreshTokenTTL:     ttlSeconds,
			},
		},
	}

	f.MigrateProfilesToAuth()

	tok := f.Auth.Tokens["expired-sso"]
	if tok == nil {
		t.Fatal("expected auth token to exist")
	}
	if tok.Type != AuthTokenTypeSSO {
		t.Fatalf("expected type %q, got %q", AuthTokenTypeSSO, tok.Type)
	}
	if !tok.NeedsReauth {
		t.Fatal("expected NeedsReauth to be true for an expired refresh token")
	}
}

func TestMigrateProfilesToAuth_DefaultPreserved(t *testing.T) {
	t.Parallel()

	f := &File{
		Profiles: Profiles{
			"alpha": {
				Token: "tok_alpha",
				Email: "alpha@example.com",
			},
			"beta": {
				Token:   "tok_beta",
				Email:   "beta@example.com",
				Default: true,
			},
			"gamma": {
				Token: "tok_gamma",
				Email: "gamma@example.com",
			},
		},
	}

	migrated := f.MigrateProfilesToAuth()
	if !migrated {
		t.Fatal("expected migration to occur")
	}

	if f.Auth.Default != "beta" {
		t.Fatalf("expected default to be %q, got %q", "beta", f.Auth.Default)
	}

	// All three should be migrated.
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if f.Auth.Tokens[name] == nil {
			t.Fatalf("expected auth token %q to exist", name)
		}
	}
}

func TestMigrateProfilesToAuth_NoDefaultPicksOne(t *testing.T) {
	t.Parallel()

	f := &File{
		Profiles: Profiles{
			"only": {
				Token: "tok_only",
				Email: "only@example.com",
				// Default is false
			},
		},
	}

	f.MigrateProfilesToAuth()

	// When no profile has Default=true, migration should still pick a default.
	if f.Auth.Default == "" {
		t.Fatal("expected a default to be assigned when none was explicitly set")
	}
	if f.Auth.Tokens[f.Auth.Default] == nil {
		t.Fatalf("expected the assigned default %q to exist in tokens", f.Auth.Default)
	}
}

func TestMigrateProfilesToAuth_AlreadyMigrated(t *testing.T) {
	t.Parallel()

	existing := &AuthToken{
		Type:  AuthTokenTypeStatic,
		Token: "original_token",
		Email: "original@example.com",
	}

	f := &File{
		Auth: Auth{
			Default: "existing",
			Tokens: AuthTokens{
				"existing": existing,
			},
		},
		Profiles: Profiles{
			"existing": {
				Token: "profile_token_different",
				Email: "different@example.com",
			},
		},
	}

	migrated := f.MigrateProfilesToAuth()

	// The profile name matches an existing auth token, so it should be skipped.
	// Since no new tokens were added, migrated should be false.
	if migrated {
		t.Fatal("expected no migration when all profiles already have corresponding auth tokens")
	}

	// The original auth token should be untouched.
	tok := f.Auth.Tokens["existing"]
	if tok.Token != "original_token" {
		t.Fatalf("expected original token %q to be preserved, got %q", "original_token", tok.Token)
	}
	if tok.Email != "original@example.com" {
		t.Fatalf("expected original email to be preserved, got %q", tok.Email)
	}
}

func TestMigrateProfilesToAuth_Idempotent(t *testing.T) {
	t.Parallel()

	f := &File{
		Profiles: Profiles{
			"user1": {
				Token:   "tok_1",
				Email:   "user1@example.com",
				Default: true,
			},
			"user2": {
				Token: "tok_2",
				Email: "user2@example.com",
			},
		},
	}

	// First migration.
	first := f.MigrateProfilesToAuth()
	if !first {
		t.Fatal("expected first migration to return true")
	}

	if len(f.Auth.Tokens) != 2 {
		t.Fatalf("expected 2 auth tokens after first migration, got %d", len(f.Auth.Tokens))
	}

	// Capture state after first migration.
	tok1Token := f.Auth.Tokens["user1"].Token
	tok2Token := f.Auth.Tokens["user2"].Token
	defaultName := f.Auth.Default

	// Second migration: should be a no-op.
	second := f.MigrateProfilesToAuth()
	if second {
		t.Fatal("expected second migration to return false (no new tokens added)")
	}

	if len(f.Auth.Tokens) != 2 {
		t.Fatalf("expected 2 auth tokens after second migration, got %d", len(f.Auth.Tokens))
	}
	if f.Auth.Tokens["user1"].Token != tok1Token {
		t.Fatal("user1 token was modified on second migration")
	}
	if f.Auth.Tokens["user2"].Token != tok2Token {
		t.Fatal("user2 token was modified on second migration")
	}
	if f.Auth.Default != defaultName {
		t.Fatalf("default changed from %q to %q on second migration", defaultName, f.Auth.Default)
	}
}

func TestAuthToken_CRUD(t *testing.T) {
	t.Parallel()

	f := &File{}

	// Initially there are no tokens.
	if f.AuthInitialized() {
		t.Fatal("expected AuthInitialized to return false on empty File")
	}
	if tok := f.GetAuthToken("anything"); tok != nil {
		t.Fatalf("expected nil from GetAuthToken on empty File, got %v", tok)
	}

	// Set a token.
	token1 := &AuthToken{
		Type:  AuthTokenTypeStatic,
		Token: "tok_crud_1",
		Email: "crud1@example.com",
	}
	f.SetAuthToken("first", token1)

	if !f.AuthInitialized() {
		t.Fatal("expected AuthInitialized to return true after SetAuthToken")
	}

	got := f.GetAuthToken("first")
	if got == nil {
		t.Fatal("expected to get auth token 'first'")
	}
	if got.Token != "tok_crud_1" {
		t.Fatalf("expected token %q, got %q", "tok_crud_1", got.Token)
	}

	// Set another token.
	token2 := &AuthToken{
		Type:  AuthTokenTypeSSO,
		Token: "tok_crud_2",
		Email: "crud2@example.com",
	}
	f.SetAuthToken("second", token2)

	if len(f.Auth.Tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(f.Auth.Tokens))
	}

	// Overwrite an existing token.
	token1Updated := &AuthToken{
		Type:  AuthTokenTypeStatic,
		Token: "tok_crud_1_updated",
		Email: "crud1_updated@example.com",
	}
	f.SetAuthToken("first", token1Updated)

	got = f.GetAuthToken("first")
	if got.Token != "tok_crud_1_updated" {
		t.Fatalf("expected updated token %q, got %q", "tok_crud_1_updated", got.Token)
	}
	if len(f.Auth.Tokens) != 2 {
		t.Fatalf("expected 2 tokens after overwrite, got %d", len(f.Auth.Tokens))
	}

	// Set a default.
	err := f.SetDefaultAuthToken("second")
	if err != nil {
		t.Fatalf("unexpected error setting default: %v", err)
	}
	if f.Auth.Default != "second" {
		t.Fatalf("expected default %q, got %q", "second", f.Auth.Default)
	}

	// Try to set default to a non-existent token.
	err = f.SetDefaultAuthToken("nonexistent")
	if err == nil {
		t.Fatal("expected error when setting default to non-existent token")
	}

	// Delete a non-default token.
	deleted := f.DeleteAuthToken("first")
	if !deleted {
		t.Fatal("expected DeleteAuthToken to return true")
	}
	if f.GetAuthToken("first") != nil {
		t.Fatal("expected 'first' to be deleted")
	}
	if len(f.Auth.Tokens) != 1 {
		t.Fatalf("expected 1 token after deletion, got %d", len(f.Auth.Tokens))
	}
	// Default should still be "second".
	if f.Auth.Default != "second" {
		t.Fatalf("expected default to remain %q, got %q", "second", f.Auth.Default)
	}

	// Delete a non-existent token.
	deleted = f.DeleteAuthToken("nonexistent")
	if deleted {
		t.Fatal("expected DeleteAuthToken to return false for non-existent token")
	}

	// Delete from nil tokens map.
	emptyFile := &File{}
	deleted = emptyFile.DeleteAuthToken("anything")
	if deleted {
		t.Fatal("expected DeleteAuthToken to return false on nil tokens map")
	}
}

func TestDeleteAuthToken_ReassignsDefault(t *testing.T) {
	t.Parallel()

	f := &File{}

	f.SetAuthToken("primary", &AuthToken{
		Type:  AuthTokenTypeStatic,
		Token: "tok_primary",
	})
	f.SetAuthToken("secondary", &AuthToken{
		Type:  AuthTokenTypeStatic,
		Token: "tok_secondary",
	})

	err := f.SetDefaultAuthToken("primary")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Delete the default token.
	deleted := f.DeleteAuthToken("primary")
	if !deleted {
		t.Fatal("expected DeleteAuthToken to return true")
	}

	// The default should be reassigned to the remaining token.
	if f.Auth.Default == "" {
		t.Fatal("expected default to be reassigned after deleting the default token")
	}
	if f.Auth.Default == "primary" {
		t.Fatal("expected default to no longer be the deleted token")
	}
	if f.Auth.Tokens[f.Auth.Default] == nil {
		t.Fatalf("expected reassigned default %q to reference an existing token", f.Auth.Default)
	}
}

func TestDeleteAuthToken_LastToken(t *testing.T) {
	t.Parallel()

	f := &File{}

	f.SetAuthToken("only", &AuthToken{
		Type:  AuthTokenTypeStatic,
		Token: "tok_only",
	})
	err := f.SetDefaultAuthToken("only")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	deleted := f.DeleteAuthToken("only")
	if !deleted {
		t.Fatal("expected DeleteAuthToken to return true")
	}

	// With no tokens remaining, default should be empty.
	if f.Auth.Default != "" {
		t.Fatalf("expected default to be empty after deleting the only token, got %q", f.Auth.Default)
	}
	if len(f.Auth.Tokens) != 0 {
		t.Fatalf("expected 0 tokens, got %d", len(f.Auth.Tokens))
	}
}

func TestGetDefaultAuthToken(t *testing.T) {
	t.Parallel()

	// No default set, no tokens.
	f := &File{}
	name, tok := f.GetDefaultAuthToken()
	if name != "" || tok != nil {
		t.Fatalf("expected empty name and nil token, got name=%q tok=%v", name, tok)
	}

	// Tokens exist but no default is set.
	f.SetAuthToken("orphan", &AuthToken{
		Type:  AuthTokenTypeStatic,
		Token: "tok_orphan",
	})
	name, tok = f.GetDefaultAuthToken()
	if name != "" || tok != nil {
		t.Fatalf("expected empty name and nil token when no default is set, got name=%q tok=%v", name, tok)
	}

	// Set a default.
	err := f.SetDefaultAuthToken("orphan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	name, tok = f.GetDefaultAuthToken()
	if name != "orphan" {
		t.Fatalf("expected default name %q, got %q", "orphan", name)
	}
	if tok == nil {
		t.Fatal("expected non-nil token for default")
	}
	if tok.Token != "tok_orphan" {
		t.Fatalf("expected token value %q, got %q", "tok_orphan", tok.Token)
	}

	// Default points to a token that has been removed out-of-band
	// (e.g. by direct map manipulation).
	f.Auth.Default = "ghost"
	name, tok = f.GetDefaultAuthToken()
	if name != "" || tok != nil {
		t.Fatalf("expected empty name and nil token when default references non-existent token, got name=%q tok=%v", name, tok)
	}
}

func TestSetDefaultAuthToken_NoTokens(t *testing.T) {
	t.Parallel()

	f := &File{}
	err := f.SetDefaultAuthToken("anything")
	if err == nil {
		t.Fatal("expected error when setting default with nil tokens map")
	}
}

func TestProfileToAuthToken_StaticMinimal(t *testing.T) {
	t.Parallel()

	p := &Profile{
		Token: "minimal_tok",
		Email: "min@example.com",
	}

	tok := profileToAuthToken(p)

	if tok.Type != AuthTokenTypeStatic {
		t.Fatalf("expected type %q, got %q", AuthTokenTypeStatic, tok.Type)
	}
	if tok.Token != "minimal_tok" {
		t.Fatalf("expected token %q, got %q", "minimal_tok", tok.Token)
	}
	if tok.Email != "min@example.com" {
		t.Fatalf("expected email %q, got %q", "min@example.com", tok.Email)
	}
	if tok.Label != "" {
		t.Fatalf("expected empty label when CustomerName is empty, got %q", tok.Label)
	}
	if tok.AccessToken != "" || tok.RefreshToken != "" {
		t.Fatal("expected empty SSO fields for static token")
	}
	if tok.AccessExpiresAt != "" || tok.RefreshExpiresAt != "" {
		t.Fatal("expected empty expiry fields for static token")
	}
}

func TestProfileToAuthToken_SSOWithLabel(t *testing.T) {
	t.Parallel()

	now := time.Now()
	created := now.Unix()

	p := &Profile{
		Token:               "sso_tok",
		Email:               "sso@example.com",
		CustomerID:          "cust_sso",
		CustomerName:        "SSO Corp",
		AccessToken:         "at_123",
		AccessTokenCreated:  created,
		AccessTokenTTL:      3600,
		RefreshToken:        "rt_456",
		RefreshTokenCreated: created,
		RefreshTokenTTL:     86400,
	}

	tok := profileToAuthToken(p)

	if tok.Type != AuthTokenTypeSSO {
		t.Fatalf("expected type %q, got %q", AuthTokenTypeSSO, tok.Type)
	}
	if tok.Label != "SSO Corp (sso@example.com)" {
		t.Fatalf("expected label %q, got %q", "SSO Corp (sso@example.com)", tok.Label)
	}
	if tok.AccountID != "cust_sso" {
		t.Fatalf("expected account_id %q, got %q", "cust_sso", tok.AccountID)
	}
	if tok.AccessToken != "at_123" {
		t.Fatalf("expected access_token %q, got %q", "at_123", tok.AccessToken)
	}
	if tok.RefreshToken != "rt_456" {
		t.Fatalf("expected refresh_token %q, got %q", "rt_456", tok.RefreshToken)
	}
	if tok.AccessExpiresAt == "" {
		t.Fatal("expected access_expires_at to be set")
	}
	if tok.RefreshExpiresAt == "" {
		t.Fatal("expected refresh_expires_at to be set")
	}
}

func TestProfileToAuthToken_SSOPartialFields(t *testing.T) {
	t.Parallel()

	// Only AccessToken is set, no timestamps -- should still be classified as SSO
	// but without computed expiry.
	p := &Profile{
		Token:       "partial_tok",
		Email:       "partial@example.com",
		AccessToken: "at_partial",
	}

	tok := profileToAuthToken(p)

	if tok.Type != AuthTokenTypeSSO {
		t.Fatalf("expected type %q, got %q", AuthTokenTypeSSO, tok.Type)
	}
	if tok.AccessExpiresAt != "" {
		t.Fatalf("expected empty access_expires_at when created/ttl are zero, got %q", tok.AccessExpiresAt)
	}
	if tok.RefreshExpiresAt != "" {
		t.Fatalf("expected empty refresh_expires_at when created/ttl are zero, got %q", tok.RefreshExpiresAt)
	}
	if tok.NeedsReauth {
		t.Fatal("expected NeedsReauth to be false when RefreshExpiresAt is empty")
	}
}

func TestMigrateProfilesToAuth_MixedPartial(t *testing.T) {
	t.Parallel()

	existing := &AuthToken{
		Type:  AuthTokenTypeStatic,
		Token: "existing_tok",
	}

	f := &File{
		Auth: Auth{
			Default: "existing",
			Tokens: AuthTokens{
				"existing": existing,
			},
		},
		Profiles: Profiles{
			"existing": {
				Token: "should_be_skipped",
				Email: "skipped@example.com",
			},
			"new-profile": {
				Token:   "new_tok",
				Email:   "new@example.com",
				Default: true,
			},
		},
	}

	migrated := f.MigrateProfilesToAuth()
	if !migrated {
		t.Fatal("expected migration for the new profile")
	}

	// The existing token should be untouched.
	if f.Auth.Tokens["existing"].Token != "existing_tok" {
		t.Fatal("existing token was overwritten during partial migration")
	}

	// The new profile should be migrated.
	newTok := f.Auth.Tokens["new-profile"]
	if newTok == nil {
		t.Fatal("expected 'new-profile' to be migrated")
	}
	if newTok.Token != "new_tok" {
		t.Fatalf("expected token %q, got %q", "new_tok", newTok.Token)
	}

	// Default should remain "existing" because it was already set, even though
	// new-profile has Default=true in the profile. The migration only sets
	// Auth.Default when Auth.Default is empty.
	if f.Auth.Default != "existing" {
		t.Fatalf("expected default to remain %q, got %q", "existing", f.Auth.Default)
	}

	if len(f.Auth.Tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(f.Auth.Tokens))
	}
}
