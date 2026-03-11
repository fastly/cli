package auth_test

import (
	"os"
	"testing"
	"time"

	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/config"
)

func TestGetExpirationStatus(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		token      *config.AuthToken
		wantStatus authcmd.ExpirationStatus
		wantErr    bool
	}{
		// Nil token.
		{
			name:       "nil token",
			token:      nil,
			wantStatus: authcmd.StatusNoExpiry,
		},

		// NeedsReauth precedence.
		{
			name: "needs reauth takes precedence over valid expiry",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				NeedsReauth:      true,
				RefreshExpiresAt: now.Add(30 * 24 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusNeedsReauth,
		},
		{
			name: "needs reauth takes precedence over expired",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				NeedsReauth:      true,
				RefreshExpiresAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusNeedsReauth,
		},

		// Static tokens.
		{
			name: "static no expiry",
			token: &config.AuthToken{
				Type: config.AuthTokenTypeStatic,
			},
			wantStatus: authcmd.StatusNoExpiry,
		},
		{
			name: "static future expiry OK",
			token: &config.AuthToken{
				Type:              config.AuthTokenTypeStatic,
				APITokenExpiresAt: now.Add(30 * 24 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusOK,
		},
		{
			name: "static expiring soon (within 7 days)",
			token: &config.AuthToken{
				Type:              config.AuthTokenTypeStatic,
				APITokenExpiresAt: now.Add(3 * 24 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpiringSoon,
		},
		{
			name: "static expiring soon (exactly 7 days)",
			token: &config.AuthToken{
				Type:              config.AuthTokenTypeStatic,
				APITokenExpiresAt: now.Add(7 * 24 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpiringSoon,
		},
		{
			name: "static expired",
			token: &config.AuthToken{
				Type:              config.AuthTokenTypeStatic,
				APITokenExpiresAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpired,
		},
		{
			name: "static malformed expiry",
			token: &config.AuthToken{
				Type:              config.AuthTokenTypeStatic,
				APITokenExpiresAt: "not-a-date",
			},
			wantStatus: authcmd.StatusNoExpiry,
			wantErr:    true,
		},

		// SSO tokens: RefreshExpiresAt primary.
		{
			name: "sso refresh OK",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				RefreshExpiresAt: now.Add(30 * 24 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusOK,
		},
		{
			name: "sso refresh expiring soon",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				RefreshExpiresAt: now.Add(3 * 24 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpiringSoon,
		},
		{
			name: "sso refresh expired",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				RefreshExpiresAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpired,
		},

		// SSO tokens: AccessExpiresAt fallback.
		{
			name: "sso no refresh, access OK (beyond threshold)",
			token: &config.AuthToken{
				Type:            config.AuthTokenTypeSSO,
				AccessExpiresAt: now.Add(2 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusOK,
		},
		{
			name: "sso no refresh, access expiring soon (within 10m)",
			token: &config.AuthToken{
				Type:            config.AuthTokenTypeSSO,
				AccessExpiresAt: now.Add(5 * time.Minute).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpiringSoon,
		},
		{
			name: "sso no refresh, access expired",
			token: &config.AuthToken{
				Type:            config.AuthTokenTypeSSO,
				AccessExpiresAt: now.Add(-10 * time.Minute).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpired,
		},

		// SSO tokens: malformed timestamps.
		{
			name: "sso malformed refresh, valid access fallback",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				RefreshExpiresAt: "garbage",
				AccessExpiresAt:  now.Add(2 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusOK,
		},
		{
			name: "sso malformed refresh, no access",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				RefreshExpiresAt: "garbage",
			},
			wantStatus: authcmd.StatusNoExpiry,
			wantErr:    true,
		},
		{
			name: "sso no refresh, malformed access",
			token: &config.AuthToken{
				Type:            config.AuthTokenTypeSSO,
				AccessExpiresAt: "garbage",
			},
			wantStatus: authcmd.StatusNoExpiry,
			wantErr:    true,
		},
		{
			name: "sso both malformed",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				RefreshExpiresAt: "bad1",
				AccessExpiresAt:  "bad2",
			},
			wantStatus: authcmd.StatusNoExpiry,
			wantErr:    true,
		},
		{
			name: "sso no expiry fields at all",
			token: &config.AuthToken{
				Type: config.AuthTokenTypeSSO,
			},
			wantStatus: authcmd.StatusNoExpiry,
		},

		// Unknown type falls through to static-style check.
		{
			name: "unknown type with expiry",
			token: &config.AuthToken{
				Type:              "unknown",
				APITokenExpiresAt: now.Add(3 * 24 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpiringSoon,
		},

		// Consistency with checkAndRefreshAuthSSOToken: when both access and
		// refresh are expired, the refresh function returns reauth=true.
		// ExpirationStatus must not return StatusOK.
		{
			name: "consistency: both expired yields StatusExpired not StatusOK",
			token: &config.AuthToken{
				Type:             config.AuthTokenTypeSSO,
				AccessExpiresAt:  now.Add(-2 * time.Hour).Format(time.RFC3339),
				RefreshExpiresAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpired,
		},
		{
			name: "consistency: access expired, no refresh yields StatusExpired",
			token: &config.AuthToken{
				Type:            config.AuthTokenTypeSSO,
				AccessExpiresAt: now.Add(-2 * time.Hour).Format(time.RFC3339),
			},
			wantStatus: authcmd.StatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, _, err := authcmd.GetExpirationStatus(tt.token, now)
			if status != tt.wantStatus {
				t.Errorf("GetExpirationStatus() status = %v, want %v", status, tt.wantStatus)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetExpirationStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpirationSummary(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		status  authcmd.ExpirationStatus
		expires time.Time
		want    string
	}{
		{
			name:    "expires in 3 days",
			status:  authcmd.StatusExpiringSoon,
			expires: now.Add(3 * 24 * time.Hour),
			want:    "expires in 3 days",
		},
		{
			name:    "expires in 2 hours",
			status:  authcmd.StatusExpiringSoon,
			expires: now.Add(2 * time.Hour),
			want:    "expires in 2 hours",
		},
		{
			name:    "expired 2 hours ago",
			status:  authcmd.StatusExpired,
			expires: now.Add(-2 * time.Hour),
			want:    "expired 2 hours ago",
		},
		{
			name:    "expired 1 day ago",
			status:  authcmd.StatusExpired,
			expires: now.Add(-24 * time.Hour),
			want:    "expired 1 day ago",
		},
		{
			name:    "OK returns summary too",
			status:  authcmd.StatusOK,
			expires: now.Add(30 * 24 * time.Hour),
			want:    "expires in 30 days",
		},
		{
			name:   "no expiry returns empty",
			status: authcmd.StatusNoExpiry,
			want:   "",
		},
		{
			name:   "needs reauth returns empty",
			status: authcmd.StatusNeedsReauth,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := authcmd.ExpirationSummary(tt.status, tt.expires, now)
			if got != tt.want {
				t.Errorf("ExpirationSummary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExpirationRemediation(t *testing.T) {
	// Ensure we test in a clean env state.
	originalEnv := os.Getenv("FASTLY_DISABLE_AUTH_COMMAND")
	defer os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", originalEnv)

	tests := []struct {
		name       string
		tokenType  string
		disableEnv string
		wantSubstr string
	}{
		{
			name:       "sso with auth enabled",
			tokenType:  "sso",
			wantSubstr: "fastly auth login --sso",
		},
		{
			name:       "static with auth enabled",
			tokenType:  "static",
			wantSubstr: "fastly auth add",
		},
		{
			name:       "unknown type with auth enabled defaults to sso",
			tokenType:  "",
			wantSubstr: "fastly auth login --sso",
		},
		{
			name:       "sso with auth disabled",
			tokenType:  "sso",
			disableEnv: "1",
			wantSubstr: "FASTLY_API_TOKEN",
		},
		{
			name:       "static with auth disabled",
			tokenType:  "static",
			disableEnv: "1",
			wantSubstr: "FASTLY_API_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", tt.disableEnv)
			got := authcmd.ExpirationRemediation(tt.tokenType)
			if got == "" {
				t.Fatal("ExpirationRemediation() returned empty string")
			}
			found := false
			if len(tt.wantSubstr) > 0 {
				for i := 0; i <= len(got)-len(tt.wantSubstr); i++ {
					if got[i:i+len(tt.wantSubstr)] == tt.wantSubstr {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("ExpirationRemediation() = %q, want substring %q", got, tt.wantSubstr)
			}
		})
	}
}
