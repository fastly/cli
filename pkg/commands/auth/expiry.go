package auth

import (
	"fmt"
	"math"
	"time"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
)

// ExpirationStatus represents the expiration state of a token.
type ExpirationStatus int

const (
	// StatusNoExpiry means no expiration information is available.
	StatusNoExpiry ExpirationStatus = iota
	// StatusOK means the token is valid and not close to expiring.
	StatusOK
	// StatusExpiringSoon means the token will expire within the warning threshold.
	StatusExpiringSoon
	// StatusExpired means the token has already expired.
	StatusExpired
	// StatusNeedsReauth means the token requires re-authentication (NeedsReauth flag set).
	StatusNeedsReauth
)

const (
	// expiryWarningThreshold is the warning window for API tokens and SSO refresh tokens.
	expiryWarningThreshold = 7 * 24 * time.Hour
	// accessOnlyWarningThreshold is the warning window when only an access token
	// expiry is available (no refresh token). This is shorter because access
	// tokens are short-lived and auto-refresh.
	accessOnlyWarningThreshold = 10 * time.Minute
)

// GetExpirationStatus computes the expiration status for a token.
// It returns the status, the effective expiry time (zero if no expiry), and a
// parse error if timestamp fields are present but malformed.
func GetExpirationStatus(at *config.AuthToken, now time.Time) (ExpirationStatus, time.Time, error) {
	if at == nil {
		return StatusNoExpiry, time.Time{}, nil
	}

	if at.NeedsReauth {
		return StatusNeedsReauth, time.Time{}, nil
	}

	switch at.Type {
	case config.AuthTokenTypeStatic:
		return staticExpirationStatus(at, now)
	case config.AuthTokenTypeSSO:
		return ssoExpirationStatus(at, now)
	default:
		// Unknown type; try static-style check on APITokenExpiresAt.
		return staticExpirationStatus(at, now)
	}
}

func staticExpirationStatus(at *config.AuthToken, now time.Time) (ExpirationStatus, time.Time, error) {
	if at.APITokenExpiresAt == "" {
		return StatusNoExpiry, time.Time{}, nil
	}

	expires, err := time.Parse(time.RFC3339, at.APITokenExpiresAt)
	if err != nil {
		return StatusNoExpiry, time.Time{}, fmt.Errorf("invalid api_token_expires_at %q: %w", at.APITokenExpiresAt, err)
	}

	return classifyExpiry(expires, now, expiryWarningThreshold), expires, nil
}

// ssoExpirationStatus handles expiration for SSO tokens. It prefers
// RefreshExpiresAt (the user-actionable deadline) and falls back to
// AccessExpiresAt when refresh info is missing or malformed.
func ssoExpirationStatus(at *config.AuthToken, now time.Time) (ExpirationStatus, time.Time, error) {
	if at.RefreshExpiresAt != "" {
		expires, err := time.Parse(time.RFC3339, at.RefreshExpiresAt)
		if err == nil {
			return classifyExpiry(expires, now, expiryWarningThreshold), expires, nil
		}
		// Malformed refresh_expires_at; fall through to access token.
	}

	if at.AccessExpiresAt != "" {
		expires, err := time.Parse(time.RFC3339, at.AccessExpiresAt)
		if err == nil {
			return classifyExpiry(expires, now, accessOnlyWarningThreshold), expires, nil
		}
		return StatusNoExpiry, time.Time{}, fmt.Errorf("invalid access_expires_at %q: %w", at.AccessExpiresAt, err)
	}

	// Neither field set.
	if at.RefreshExpiresAt != "" {
		// RefreshExpiresAt was set but malformed, and no AccessExpiresAt to fall back to.
		return StatusNoExpiry, time.Time{}, fmt.Errorf("invalid refresh_expires_at %q", at.RefreshExpiresAt)
	}
	return StatusNoExpiry, time.Time{}, nil
}

func classifyExpiry(expires, now time.Time, threshold time.Duration) ExpirationStatus {
	if now.After(expires) {
		return StatusExpired
	}
	remaining := expires.Sub(now)
	if remaining <= threshold {
		return StatusExpiringSoon
	}
	return StatusOK
}

// ExpirationSummary returns a human-readable string describing the time until
// or since expiry. Returns "" for StatusNoExpiry and StatusNeedsReauth.
func ExpirationSummary(status ExpirationStatus, expires time.Time, now time.Time) string {
	switch status {
	case StatusOK, StatusExpiringSoon:
		return "expires in " + humanDuration(expires.Sub(now))
	case StatusExpired:
		return "expired " + humanDuration(now.Sub(expires)) + " ago"
	case StatusNoExpiry, StatusNeedsReauth:
		return ""
	}
	return ""
}

// ExpirationRemediation returns actionable remediation text for the given token type.
func ExpirationRemediation(tokenType string) string {
	return fsterr.TokenExpirationRemediationForType(tokenType)
}

// humanDuration formats a duration into a short human-readable string like
// "3 days", "2 hours", "45 minutes". Always returns a positive representation.
func humanDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	switch {
	case d < time.Minute:
		s := int(d.Seconds())
		if s <= 1 {
			return "1 second"
		}
		return fmt.Sprintf("%d seconds", s)
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", m)
	case d < 24*time.Hour:
		h := int(math.Round(d.Hours()))
		if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	default:
		days := int(math.Round(d.Hours() / 24))
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}
