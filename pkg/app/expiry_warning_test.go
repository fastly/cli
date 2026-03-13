package app

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
)

// expiringTokenData returns a global.Data configured with a stored token that
// expires soon. Callers can override Flags and commandName to test suppression.
func expiringTokenData(out *bytes.Buffer) *global.Data {
	soon := time.Now().Add(3 * 24 * time.Hour).Format(time.RFC3339)
	return &global.Data{
		Output:   out,
		ErrLog:   fsterr.Log,
		Manifest: &manifest.Data{},
		Config: config.File{
			Auth: config.Auth{
				Default: "mytoken",
				Tokens: config.AuthTokens{
					"mytoken": &config.AuthToken{
						Type:              config.AuthTokenTypeStatic,
						Token:             "tok_abc123",
						APITokenExpiresAt: soon,
					},
				},
			},
		},
	}
}

func TestCheckTokenExpirationWarning(t *testing.T) {
	soon := time.Now().Add(3 * 24 * time.Hour).Format(time.RFC3339)
	farFuture := time.Now().Add(60 * 24 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name        string
		commandName string
		data        func(out *bytes.Buffer) *global.Data
		wantWarn    bool
		wantSubstr  string
	}{
		{
			name:        "SourceAuth expiring soon shows warning",
			commandName: "service list",
			data: func(out *bytes.Buffer) *global.Data {
				return &global.Data{
					Output: out,
					ErrLog: fsterr.Log,
					Config: config.File{
						Auth: config.Auth{
							Default: "mytoken",
							Tokens: config.AuthTokens{
								"mytoken": &config.AuthToken{
									Type:              config.AuthTokenTypeStatic,
									Token:             "tok_abc123",
									APITokenExpiresAt: soon,
								},
							},
						},
					},
				}
			},
			wantWarn:   true,
			wantSubstr: "expires in",
		},
		{
			name:        "SourceAuth not expiring soon no warning",
			commandName: "service list",
			data: func(out *bytes.Buffer) *global.Data {
				return &global.Data{
					Output: out,
					ErrLog: fsterr.Log,
					Config: config.File{
						Auth: config.Auth{
							Default: "mytoken",
							Tokens: config.AuthTokens{
								"mytoken": &config.AuthToken{
									Type:              config.AuthTokenTypeStatic,
									Token:             "tok_abc123",
									APITokenExpiresAt: farFuture,
								},
							},
						},
					},
				}
			},
			wantWarn: false,
		},
		{
			name:        "SourceEnvironment JWT-like token no warning",
			commandName: "service list",
			data: func(out *bytes.Buffer) *global.Data {
				return &global.Data{
					Output: out,
					ErrLog: fsterr.Log,
					Env:    config.Environment{APIToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.fake"},
					Config: config.File{
						Auth: config.Auth{
							Default: "mytoken",
							Tokens: config.AuthTokens{
								"mytoken": &config.AuthToken{
									Type:              config.AuthTokenTypeStatic,
									Token:             "tok_abc123",
									APITokenExpiresAt: soon,
								},
							},
						},
					},
				}
			},
			wantWarn: false,
		},
		{
			name:        "SourceFlag raw token no warning",
			commandName: "service list",
			data: func(out *bytes.Buffer) *global.Data {
				return &global.Data{
					Output: out,
					ErrLog: fsterr.Log,
					Flags:  global.Flags{Token: "some-raw-token"},
					Config: config.File{
						Auth: config.Auth{
							Default: "mytoken",
							Tokens: config.AuthTokens{
								"mytoken": &config.AuthToken{
									Type:              config.AuthTokenTypeStatic,
									Token:             "tok_abc123",
									APITokenExpiresAt: soon,
								},
							},
						},
					},
				}
			},
			wantWarn: false,
		},
		{
			name:        "stale default name nil token no panic",
			commandName: "service list",
			data: func(out *bytes.Buffer) *global.Data {
				return &global.Data{
					Output: out,
					ErrLog: fsterr.Log,
					Config: config.File{
						Auth: config.Auth{
							Default: "deleted-token",
							Tokens:  config.AuthTokens{},
						},
					},
				}
			},
			wantWarn: false,
		},
		{
			name:        "malformed expiry logs error no visible warning",
			commandName: "service list",
			data: func(out *bytes.Buffer) *global.Data {
				return &global.Data{
					Output: out,
					ErrLog: fsterr.Log,
					Config: config.File{
						Auth: config.Auth{
							Default: "mytoken",
							Tokens: config.AuthTokens{
								"mytoken": &config.AuthToken{
									Type:              config.AuthTokenTypeStatic,
									Token:             "tok_abc123",
									APITokenExpiresAt: "not-a-date",
								},
							},
						},
					},
				}
			},
			wantWarn: false,
		},
		{
			name:        "nil ErrLog does not panic",
			commandName: "service list",
			data: func(out *bytes.Buffer) *global.Data {
				return &global.Data{
					Output: out,
					Config: config.File{
						Auth: config.Auth{
							Default: "mytoken",
							Tokens: config.AuthTokens{
								"mytoken": &config.AuthToken{
									Type:              config.AuthTokenTypeStatic,
									Token:             "tok_abc123",
									APITokenExpiresAt: "not-a-date",
								},
							},
						},
					},
				}
			},
			wantWarn: false,
		},
		{
			name:        "SSO token expiring soon shows remediation",
			commandName: "service list",
			data: func(out *bytes.Buffer) *global.Data {
				return &global.Data{
					Output: out,
					ErrLog: fsterr.Log,
					Config: config.File{
						Auth: config.Auth{
							Default: "sso-tok",
							Tokens: config.AuthTokens{
								"sso-tok": &config.AuthToken{
									Type:             config.AuthTokenTypeSSO,
									Token:            "tok_sso",
									RefreshExpiresAt: soon,
								},
							},
						},
					},
				}
			},
			wantWarn:   true,
			wantSubstr: "fastly auth login --sso",
		},
	}

	// Ensure FASTLY_DISABLE_AUTH_COMMAND is not set.
	originalEnv := os.Getenv("FASTLY_DISABLE_AUTH_COMMAND")
	os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "")
	defer os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", originalEnv)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			data := tt.data(&buf)

			// Ensure Manifest is initialized to avoid nil panics in Token().
			if data.Manifest == nil {
				data.Manifest = &manifest.Data{}
			}

			checkTokenExpirationWarning(data, tt.commandName)

			output := buf.String()
			if tt.wantWarn && output == "" {
				t.Error("expected warning output but got none")
			}
			if !tt.wantWarn && output != "" {
				t.Errorf("expected no warning but got: %s", output)
			}
			if tt.wantSubstr != "" && !strings.Contains(output, tt.wantSubstr) {
				t.Errorf("expected output to contain %q, got: %s", tt.wantSubstr, output)
			}
		})
	}
}

func TestCheckTokenExpirationWarningDisabledAuth(t *testing.T) {
	originalEnv := os.Getenv("FASTLY_DISABLE_AUTH_COMMAND")
	os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")
	defer os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", originalEnv)

	var buf bytes.Buffer
	data := expiringTokenData(&buf)

	checkTokenExpirationWarning(data, "service list")

	output := buf.String()
	if !strings.Contains(output, "FASTLY_API_TOKEN") {
		t.Errorf("expected FASTLY_API_TOKEN remediation in disabled-auth mode, got: %s", output)
	}
	if strings.Contains(output, "fastly auth") {
		t.Errorf("should not mention fastly auth in disabled mode, got: %s", output)
	}
}

// TestCheckTokenExpirationWarningSuppression tests that the warning is
// suppressed for all auth-related commands, --quiet, and --json (which sets
// Quiet=true). The auth-related set matches FASTLY_DISABLE_AUTH_COMMAND
// (pkg/env/env.go): auth, auth-token, sso, profile, whoami.
func TestCheckTokenExpirationWarningSuppression(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		flags       global.Flags
	}{
		// auth family.
		{
			name:        "suppressed for auth list",
			commandName: "auth list",
		},
		{
			name:        "suppressed for auth show",
			commandName: "auth show",
		},
		{
			name:        "suppressed for auth login",
			commandName: "auth login",
		},
		{
			name:        "suppressed for bare auth",
			commandName: "auth",
		},
		// Other auth-related families.
		{
			name:        "suppressed for sso",
			commandName: "sso",
		},
		{
			name:        "suppressed for auth-token create",
			commandName: "auth-token create",
		},
		{
			name:        "suppressed for bare auth-token",
			commandName: "auth-token",
		},
		{
			name:        "suppressed for profile switch",
			commandName: "profile switch",
		},
		{
			name:        "suppressed for bare profile",
			commandName: "profile",
		},
		{
			name:        "suppressed for whoami",
			commandName: "whoami",
		},
		// Flag-based suppression.
		{
			name:        "suppressed with --quiet flag",
			commandName: "service list",
			flags:       global.Flags{Quiet: true},
		},
		{
			name:        "suppressed with --json flag (sets Quiet)",
			commandName: "service list",
			flags:       global.Flags{Quiet: true}, // --json sets Quiet=true in Exec
		},
	}

	originalEnv := os.Getenv("FASTLY_DISABLE_AUTH_COMMAND")
	os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "")
	defer os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", originalEnv)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			data := expiringTokenData(&buf)
			data.Flags = tt.flags

			checkTokenExpirationWarning(data, tt.commandName)

			output := buf.String()
			if output != "" {
				t.Errorf("expected no output for %q with flags %+v, got: %s", tt.commandName, tt.flags, output)
			}
		})
	}
}

// TestCheckTokenExpirationWarningNotSuppressedForNonAuth ensures that commands
// starting with "auth" as a prefix of another word (e.g. "authtoken") are not
// incorrectly suppressed.
func TestCheckTokenExpirationWarningNotSuppressedForNonAuth(t *testing.T) {
	originalEnv := os.Getenv("FASTLY_DISABLE_AUTH_COMMAND")
	os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "")
	defer os.Setenv("FASTLY_DISABLE_AUTH_COMMAND", originalEnv)

	var buf bytes.Buffer
	data := expiringTokenData(&buf)

	// "authtoken" is not "auth" or "auth <sub>", so warning should fire.
	checkTokenExpirationWarning(data, "authtoken list")

	output := buf.String()
	if output == "" {
		t.Error("expected warning for non-auth command 'authtoken list' but got none")
	}
}
