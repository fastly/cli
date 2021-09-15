package user

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all users from a specified customer id")
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause.Flag("customer-id", "c").Required().StringVar(&c.customerID)
	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	customerID string
	manifest   manifest.Data
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
		return errors.ErrNoToken
	}

	input := c.constructInput()

	rs, err := c.Globals.Client.ListCustomerUsers(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Customer ID": c.customerID,
		})
		return err
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, rs)
	} else {
		c.printSummary(out, rs)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListCustomerUsersInput {
	var input fastly.ListCustomerUsersInput

	input.CustomerID = c.customerID

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, rs []*fastly.User) {
	for _, r := range rs {
		fmt.Fprintf(out, "\nID: %s\n", r.ID)
		fmt.Fprintf(out, "Login: %s\n", r.Login)
		fmt.Fprintf(out, "Name: %s\n", r.Name)
		fmt.Fprintf(out, "Role: %s\n", r.Role)
		fmt.Fprintf(out, "Customer ID: %s\n", r.CustomerID)
		fmt.Fprintf(out, "Email Hash: %s\n", r.EmailHash)
		fmt.Fprintf(out, "Limit Services: %t\n", r.LimitServices)
		fmt.Fprintf(out, "Locked: %t\n", r.Locked)
		fmt.Fprintf(out, "Require New Password: %t\n", r.RequireNewPassword)
		fmt.Fprintf(out, "Two Factor Auth Enabled: %t\n", r.TwoFactorAuthEnabled)
		fmt.Fprintf(out, "Two Factor Setup Required: %t\n\n", r.TwoFactorSetupRequired)

		if r.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
		}
		if r.UpdatedAt != nil {
			fmt.Fprintf(out, "Updated at: %s\n", r.UpdatedAt)
		}
		if r.DeletedAt != nil {
			fmt.Fprintf(out, "Deleted at: %s\n", r.DeletedAt)
		}
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.User) {
	t := text.NewTable(out)
	t.AddHeader("LOGIN", "NAME", "ROLE", "LOCKED")
	for _, r := range rs {
		t.AddLine(r.Login, r.Name, r.Role, r.Locked)
	}
	t.Print()
}
