package user

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all users from a specified customer id")
	c.Globals = g
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagCustomerIDName,
		Description: argparser.FlagCustomerIDDesc,
		Dst:         &c.customerID.Value,
		Action:      c.customerID.Set,
	})
	c.RegisterFlagBool(c.JSONFlag()) // --json
	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	customerID argparser.OptionalCustomerID
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	if err := c.customerID.Parse(); err != nil {
		return err
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListCustomerUsers(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Customer ID": c.customerID.Value,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, o)
	} else {
		err = c.printSummary(out, o)
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
		fmt.Fprintf(out, "\nID: %s\n", fastly.ToValue(u.ID))
		fmt.Fprintf(out, "Login: %s\n", fastly.ToValue(u.Login))
		fmt.Fprintf(out, "Name: %s\n", fastly.ToValue(u.Name))
		fmt.Fprintf(out, "Role: %s\n", fastly.ToValue(u.Role))
		fmt.Fprintf(out, "Customer ID: %s\n", fastly.ToValue(u.CustomerID))
		fmt.Fprintf(out, "Email Hash: %s\n", fastly.ToValue(u.EmailHash))
		fmt.Fprintf(out, "Limit Services: %t\n", fastly.ToValue(u.LimitServices))
		fmt.Fprintf(out, "Locked: %t\n", fastly.ToValue(u.Locked))
		fmt.Fprintf(out, "Require New Password: %t\n", fastly.ToValue(u.RequireNewPassword))
		fmt.Fprintf(out, "Two Factor Auth Enabled: %t\n", fastly.ToValue(u.TwoFactorAuthEnabled))
		fmt.Fprintf(out, "Two Factor Setup Required: %t\n\n", fastly.ToValue(u.TwoFactorSetupRequired))

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
	t := text.NewTable(out)
	t.AddHeader("LOGIN", "NAME", "ROLE", "LOCKED", "ID")
	for _, u := range us {
		t.AddLine(
			fastly.ToValue(u.Login),
			fastly.ToValue(u.Name),
			fastly.ToValue(u.Role),
			fastly.ToValue(u.Locked),
			fastly.ToValue(u.ID),
		)
	}
	t.Print()
	return nil
}
