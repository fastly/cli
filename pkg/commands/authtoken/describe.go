package authtoken

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Get the current API token").Alias("get")

	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.GetTokenSelf()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	return c.print(out, o)
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.Token) error {
	fmt.Fprintf(out, "\nID: %s\n", fastly.ToValue(r.ID))
	fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(r.Name))
	fmt.Fprintf(out, "User ID: %s\n", fastly.ToValue(r.UserID))
	fmt.Fprintf(out, "Services: %s\n", strings.Join(r.Services, ", "))
	fmt.Fprintf(out, "Scope: %s\n", fastly.ToValue(r.Scope))
	fmt.Fprintf(out, "IP: %s\n\n", fastly.ToValue(r.IP))

	if r.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
	}
	if r.LastUsedAt != nil {
		fmt.Fprintf(out, "Last used at: %s\n", r.LastUsedAt)
	}
	if r.ExpiresAt != nil {
		fmt.Fprintf(out, "Expires at: %s\n", r.ExpiresAt)
	}
	return nil
}
