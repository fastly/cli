package auth

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
)

type RootCommand struct {
	argparser.Base

	Login  LoginCommand  `cmd:"" help:"Authenticate and store a default token (paste token or use --sso)"`
	Add    AddCommand    `cmd:"" help:"Store a named token"`
	Delete DeleteCommand `cmd:"" help:"Delete a stored token"`
	List   ListCommand   `cmd:"" help:"List stored tokens and show the default"`
	Show   ShowCommand   `cmd:"" help:"Show details for a stored token"`
	Use    UseCommand    `cmd:"" help:"Set the default stored token for CLI commands"`
	Revoke RevokeCommand `cmd:"" help:"Revoke a token via the API and remove it from local config"`
	Token  TokenCommand  `cmd:"" help:"Output the active API token (for use in shell substitutions)"`
}

const CommandName = "auth"

func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
