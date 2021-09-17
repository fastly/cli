package authtoken

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Get the current API token").Alias("get")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base

	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	r, err := c.Globals.Client.GetTokenSelf()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.print(out, r)
	return nil
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.Token) {
	fmt.Fprintf(out, "\nID: %s\n", r.ID)
	fmt.Fprintf(out, "Name: %s\n", r.Name)
	fmt.Fprintf(out, "User ID: %s\n", r.UserID)
	fmt.Fprintf(out, "Services: %s\n", strings.Join(r.Services, ", "))
	fmt.Fprintf(out, "Access Token: %s\n", r.AccessToken)
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
}
