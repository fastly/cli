package auth

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/text"
)

// TokenCommand prints the active API token to non-terminal stdout.
type TokenCommand struct {
	argparser.Base
}

// NewTokenCommand returns a new command registered under the parent.
func NewTokenCommand(parent argparser.Registerer, g *global.Data) *TokenCommand {
	var c TokenCommand
	c.Globals = g
	c.CmdClause = parent.Command("token", "Output the active API token (for use in shell substitutions)")
	return &c
}

// Exec implements the command interface.
func (c *TokenCommand) Exec(_ io.Reader, out io.Writer) error {
	if text.IsTTY(out) {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("refusing to print token to a terminal"),
			Remediation: "Use this command in a shell substitution or pipe, e.g. $(fastly auth token).",
		}
	}

	token, src := c.Globals.Token()
	if src == lookup.SourceUndefined || token == "" {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("no API token configured"),
			Remediation: fsterr.ProfileRemediation(),
		}
	}

	fmt.Fprint(out, token)
	return nil
}
