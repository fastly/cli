package authtoken

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v7/fastly"
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

	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base

	json     bool
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	_, s := c.Globals.Token()
	if s == lookup.SourceUndefined {
		return fsterr.ErrNoToken
	}
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	r, err := c.Globals.APIClient.GetTokenSelf()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	return c.print(out, r)
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.Token) error {
	if c.json {
		data, err := json.Marshal(r)
		if err != nil {
			return err
		}
		_, err = out.Write(data)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return fmt.Errorf("error: unable to write data to stdout: %w", err)
		}
		return nil
	}

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
