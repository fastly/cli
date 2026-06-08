package auth_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/fastly/kingpin"

	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

func newTokenCommand(g *global.Data) *authcmd.TokenCommand {
	app := kingpin.New("fastly", "test")
	parent := app.Command("auth", "test auth")
	return authcmd.NewTokenCommand(parent, g)
}

func globalDataWithToken(token string) *global.Data {
	return &global.Data{
		Config: config.File{
			Auth: config.Auth{
				Default: "user",
				Tokens: config.AuthTokens{
					"user": &config.AuthToken{
						Type:  config.AuthTokenTypeStatic,
						Token: token,
					},
				},
			},
		},
	}
}

func TestToken_NonTTY_Success(t *testing.T) {
	var buf bytes.Buffer
	cmd := newTokenCommand(globalDataWithToken("test-api-token-value"))
	err := cmd.Exec(nil, &buf)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got := buf.String(); got != "test-api-token-value" {
		t.Errorf("expected token %q, got %q", "test-api-token-value", got)
	}
	if got := buf.Bytes(); got[len(got)-1] == '\n' {
		t.Error("output should not have a trailing newline")
	}
}

func TestToken_NonTTY_NoToken(t *testing.T) {
	var buf bytes.Buffer
	g := &global.Data{
		Config: config.File{},
	}

	cmd := newTokenCommand(g)
	err := cmd.Exec(nil, &buf)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("expected RemediationError, got %T: %v", err, err)
	}
	if re.Inner == nil || re.Inner.Error() != "no API token configured" {
		t.Errorf("unexpected inner error: %v", re.Inner)
	}
}

func TestToken_NonTTY_ProfileFlagUnknown(t *testing.T) {
	var buf bytes.Buffer
	g := globalDataWithToken("test-api-token-value")
	g.Flags.Profile = "bogus"

	cmd := newTokenCommand(g)
	err := cmd.Exec(nil, &buf)
	if err == nil {
		t.Fatal("expected error for unknown --profile")
	}
	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("expected RemediationError, got %T: %v", err, err)
	}
	if re.Inner == nil || !strings.Contains(re.Inner.Error(), `"bogus"`) {
		t.Errorf("unexpected inner error: %v", re.Inner)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no token to be written, got: %q", buf.String())
	}
}

func TestToken_NonTTY_ProfileFlagKnown(t *testing.T) {
	var buf bytes.Buffer
	g := globalDataWithToken("default-token")
	g.Config.Auth.Tokens["alt"] = &config.AuthToken{Type: config.AuthTokenTypeStatic, Token: "alt-token"}
	g.Flags.Profile = "alt"

	cmd := newTokenCommand(g)
	err := cmd.Exec(nil, &buf)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got := buf.String(); got != "alt-token" {
		t.Errorf("expected token %q, got %q", "alt-token", got)
	}
}

func TestToken_NonTTY_TokenFlagBeatsProfileFlag(t *testing.T) {
	var buf bytes.Buffer
	g := globalDataWithToken("default-token")
	g.Flags.Token = "raw-xyz"
	g.Flags.Profile = "bogus"

	cmd := newTokenCommand(g)
	err := cmd.Exec(nil, &buf)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got := buf.String(); got != "raw-xyz" {
		t.Errorf("expected token %q, got %q", "raw-xyz", got)
	}
}
