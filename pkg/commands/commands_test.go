package commands_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/commands"
	"github.com/fastly/cli/pkg/testutil"
)

// authCommandPrefixes lists the command names (and prefixes for subcommands)
// that should be excluded when FASTLY_DISABLE_AUTH_COMMAND is set.
var authCommandPrefixes = []string{"auth", "auth-token", "sso", "profile", "whoami"}

// isAuthRelated reports whether a command name belongs to an auth-related
// command group.
func isAuthRelated(name string) bool {
	for _, prefix := range authCommandPrefixes {
		if name == prefix || strings.HasPrefix(name, prefix+" ") {
			return true
		}
	}
	return false
}

func TestDefineDisableAuthCommand(t *testing.T) {
	newApp := func(stdout *bytes.Buffer) *kingpin.Application {
		app := kingpin.New("fastly", "test")
		app.Writers(stdout, io.Discard)
		app.Terminate(nil)
		return app
	}

	t.Run("auth-related commands present by default", func(t *testing.T) {
		var stdout bytes.Buffer
		data := testutil.MockGlobalData([]string{"fastly"}, &stdout)
		cmds := commands.Define(newApp(&stdout), data)

		for _, want := range authCommandPrefixes {
			found := false
			for _, cmd := range cmds {
				if cmd.Name() == want || strings.HasPrefix(cmd.Name(), want+" ") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected %q command to be present when FASTLY_DISABLE_AUTH_COMMAND is not set", want)
			}
		}
	})

	t.Run("auth-related commands excluded when FASTLY_DISABLE_AUTH_COMMAND is set", func(t *testing.T) {
		t.Setenv("FASTLY_DISABLE_AUTH_COMMAND", "1")

		var stdout bytes.Buffer
		data := testutil.MockGlobalData([]string{"fastly"}, &stdout)
		cmds := commands.Define(newApp(&stdout), data)

		for _, cmd := range cmds {
			if isAuthRelated(cmd.Name()) {
				t.Errorf("expected no auth-related commands, but found %q", cmd.Name())
			}
		}

		// Non-auth commands still exist.
		found := false
		for _, cmd := range cmds {
			if cmd.Name() == "compute" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected compute command to still be present")
		}
	})
}
