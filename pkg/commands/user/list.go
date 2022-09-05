package user

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all users from a specified customer id")
	c.Globals = globals
	c.manifest = data
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagCustomerIDName,
		Description: cmd.FlagCustomerIDDesc,
		Dst:         &c.customerID.Value,
		Action:      c.customerID.Set,
	})
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	customerID cmd.OptionalCustomerID
	json       bool
	manifest   manifest.Data
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return fsterr.ErrNoToken
	}
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if err := c.customerID.Parse(); err != nil {
		return err
	}

	input := c.constructInput()

	rs, err := c.Globals.APIClient.ListCustomerUsers(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Customer ID": c.customerID.Value,
		})
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, rs)
	} else {
		err = c.printSummary(out, rs)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListCustomerUsersInput {
	var input fastly.ListCustomerUsersInput

	input.CustomerID = c.customerID.Value

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, us []*fastly.User) {
	for _, u := range us {
		fmt.Fprintf(out, "\nID: %s\n", u.ID)
		fmt.Fprintf(out, "Login: %s\n", u.Login)
		fmt.Fprintf(out, "Name: %s\n", u.Name)
		fmt.Fprintf(out, "Role: %s\n", u.Role)
		fmt.Fprintf(out, "Customer ID: %s\n", u.CustomerID)
		fmt.Fprintf(out, "Email Hash: %s\n", u.EmailHash)
		fmt.Fprintf(out, "Limit Services: %t\n", u.LimitServices)
		fmt.Fprintf(out, "Locked: %t\n", u.Locked)
		fmt.Fprintf(out, "Require New Password: %t\n", u.RequireNewPassword)
		fmt.Fprintf(out, "Two Factor Auth Enabled: %t\n", u.TwoFactorAuthEnabled)
		fmt.Fprintf(out, "Two Factor Setup Required: %t\n\n", u.TwoFactorSetupRequired)

		if u.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", u.CreatedAt)
		}
		if u.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", u.UpdatedAt)
		}
		if u.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", u.DeletedAt)
		}
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, us []*fastly.User) error {
	if c.json {
		data, err := json.Marshal(us)
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
	t := text.NewTable(out)
	t.AddHeader("LOGIN", "NAME", "ROLE", "LOCKED", "ID")
	for _, u := range us {
		t.AddLine(u.Login, u.Name, u.Role, u.Locked, u.ID)
	}
	t.Print()
	return nil
}
