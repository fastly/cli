package global_test

import (
	"testing"

	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/threadsafe"
)

func TestToken(t *testing.T) {
	tests := []struct {
		name       string
		data       *global.Data
		wantToken  string
		wantSource lookup.Source
	}{
		{
			name: "token flag matches stored auth token name",
			data: &global.Data{
				Flags: global.Flags{Token: "myname"},
				Config: config.File{
					Auth: config.Auth{
						Default: "myname",
						Tokens: config.AuthTokens{
							"myname": &config.AuthToken{
								Type:  config.AuthTokenTypeStatic,
								Token: "stored-token-value",
							},
						},
					},
				},
			},
			wantToken:  "stored-token-value",
			wantSource: lookup.SourceAuth,
		},
		{
			name: "token flag raw value when no stored name matches",
			data: &global.Data{
				Flags: global.Flags{Token: "raw-api-token"},
				Config: config.File{
					Auth: config.Auth{
						Default: "user",
						Tokens: config.AuthTokens{
							"user": &config.AuthToken{
								Type:  config.AuthTokenTypeStatic,
								Token: "other-token",
							},
						},
					},
				},
			},
			wantToken:  "raw-api-token",
			wantSource: lookup.SourceFlag,
		},
		{
			name: "manifest profile selects stored auth token",
			data: &global.Data{
				Manifest: &manifest.Data{
					File: manifest.File{Profile: "proj"},
				},
				Config: config.File{
					Auth: config.Auth{
						Default: "default-user",
						Tokens: config.AuthTokens{
							"default-user": &config.AuthToken{
								Type:  config.AuthTokenTypeStatic,
								Token: "default-token",
							},
							"proj": &config.AuthToken{
								Type:  config.AuthTokenTypeStatic,
								Token: "project-token",
							},
						},
					},
				},
			},
			wantToken:  "project-token",
			wantSource: lookup.SourceAuth,
		},
		{
			name: "manifest profile falls through when no matching auth token",
			data: &global.Data{
				Manifest: &manifest.Data{
					File: manifest.File{Profile: "missing"},
				},
				Config: config.File{
					Auth: config.Auth{
						Default: "user",
						Tokens: config.AuthTokens{
							"user": &config.AuthToken{
								Type:  config.AuthTokenTypeStatic,
								Token: "default-token",
							},
						},
					},
				},
			},
			wantToken:  "default-token",
			wantSource: lookup.SourceAuth,
		},
		{
			name: "env var takes precedence over manifest profile",
			data: &global.Data{
				Env: config.Environment{APIToken: "env-token"},
				Manifest: &manifest.Data{
					File: manifest.File{Profile: "proj"},
				},
				Config: config.File{
					Auth: config.Auth{
						Default: "user",
						Tokens: config.AuthTokens{
							"proj": &config.AuthToken{
								Type:  config.AuthTokenTypeStatic,
								Token: "project-token",
							},
						},
					},
				},
			},
			wantToken:  "env-token",
			wantSource: lookup.SourceEnvironment,
		},
		{
			name: "no token sources returns undefined",
			data: &global.Data{
				Config: config.File{},
			},
			wantToken:  "",
			wantSource: lookup.SourceUndefined,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, gotSource := tt.data.Token()
			if gotToken != tt.wantToken {
				t.Errorf("Token() token = %q, want %q", gotToken, tt.wantToken)
			}
			if gotSource != tt.wantSource {
				t.Errorf("Token() source = %v, want %v", gotSource, tt.wantSource)
			}
		})
	}
}

func TestAuthTokenName(t *testing.T) {
	tests := []struct {
		name     string
		data     *global.Data
		wantName string
	}{
		{
			name: "token flag matches stored name",
			data: &global.Data{
				Flags: global.Flags{Token: "myname"},
				Config: config.File{
					Auth: config.Auth{
						Tokens: config.AuthTokens{
							"myname": &config.AuthToken{Token: "t"},
						},
					},
				},
			},
			wantName: "myname",
		},
		{
			name: "token flag raw value returns empty",
			data: &global.Data{
				Flags: global.Flags{Token: "raw-value"},
				Config: config.File{
					Auth: config.Auth{
						Tokens: config.AuthTokens{
							"user": &config.AuthToken{Token: "t"},
						},
					},
				},
			},
			wantName: "",
		},
		{
			name: "manifest profile returns profile name",
			data: &global.Data{
				Manifest: &manifest.Data{
					File: manifest.File{Profile: "proj"},
				},
				Config: config.File{
					Auth: config.Auth{
						Default: "default-user",
						Tokens: config.AuthTokens{
							"default-user": &config.AuthToken{Token: "t"},
							"proj":         &config.AuthToken{Token: "t2"},
						},
					},
				},
			},
			wantName: "proj",
		},
		{
			name: "manifest profile missing falls through to default",
			data: &global.Data{
				Manifest: &manifest.Data{
					File: manifest.File{Profile: "missing"},
				},
				Config: config.File{
					Auth: config.Auth{
						Default: "user",
						Tokens: config.AuthTokens{
							"user": &config.AuthToken{Token: "t"},
						},
					},
				},
			},
			wantName: "user",
		},
		{
			name: "default auth token name",
			data: &global.Data{
				Config: config.File{
					Auth: config.Auth{
						Default: "user",
						Tokens: config.AuthTokens{
							"user": &config.AuthToken{Token: "t"},
						},
					},
				},
			},
			wantName: "user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.data.AuthTokenName()
			if got != tt.wantName {
				t.Errorf("AuthTokenName() = %q, want %q", got, tt.wantName)
			}
		})
	}
}

func TestTokenManifestProfileMissingNoSideEffect(t *testing.T) {
	var buf threadsafe.Buffer
	d := &global.Data{
		Manifest: &manifest.Data{
			File: manifest.File{Profile: "missing"},
		},
		Config: config.File{
			Auth: config.Auth{
				Default: "user",
				Tokens: config.AuthTokens{
					"user": &config.AuthToken{
						Type:  config.AuthTokenTypeStatic,
						Token: "default-token",
					},
				},
			},
		},
		Output: &buf,
	}

	token, source := d.Token()
	if token != "default-token" {
		t.Errorf("Token() = %q, want %q", token, "default-token")
	}
	if source != lookup.SourceAuth {
		t.Errorf("Token() source = %v, want %v", source, lookup.SourceAuth)
	}
	if buf.String() != "" {
		t.Errorf("Token() should not write to output, got: %q", buf.String())
	}
}
