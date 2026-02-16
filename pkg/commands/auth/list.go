package auth

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand lists stored tokens.
type ListCommand struct {
	argparser.Base
}

func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.Globals = g
	c.CmdClause = parent.Command("list", "List stored tokens and show the default")
	return &c
}

func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	tokens := c.Globals.Config.Auth.Tokens
	if len(tokens) == 0 {
		text.Output(out, "No tokens stored. Run `fastly auth login` to add one.\n")
		return nil
	}

	for name, entry := range tokens {
		marker := "  "
		if name == c.Globals.Config.Auth.Default {
			marker = "* "
		}

		info := entry.Type
		if entry.Email != "" {
			info = entry.Email
		}

		reauthStr := ""
		if entry.NeedsReauth {
			reauthStr = " (needs re-authentication)"
		}

		text.Output(out, "%s%s (%s)%s\n", marker, name, info, reauthStr)
	}

	return nil
}
