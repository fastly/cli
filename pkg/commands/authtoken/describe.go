package authtoken

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("describe", "Get the current API token").Alias("get")

	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest manifest.Data
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
	fmt.Fprintf(out, "\nID: %s\n", r.ID)
	fmt.Fprintf(out, "Name: %s\n", r.Name)
	fmt.Fprintf(out, "User ID: %s\n", r.UserID)
	fmt.Fprintf(out, "Services: %s\n", strings.Join(r.Services, ", "))
	fmt.Fprintf(out, "Scope: %s\n", r.Scope)
	fmt.Fprintf(out, "IP: %s\n\n", r.IP)

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
