package authtoken

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List API tokens")
	c.Globals = globals
	c.manifest = data
	c.CmdClause.Flag("customer-id", "Alphanumeric string identifying the customer").StringVar(&c.customerID)
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

	var (
		err error
		rs  []*fastly.Token
	)

	if c.customerID != "" {
		input := c.constructInput()

		rs, err = c.Globals.Client.ListCustomerTokens(input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	} else {
		rs, err = c.Globals.Client.ListTokens()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	}

	if c.Globals.Verbose() {
		c.printVerbose(out, rs)
	} else {
		c.printSummary(out, rs)
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListCustomerTokensInput {
	var input fastly.ListCustomerTokensInput

	input.CustomerID = c.customerID

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func (c *ListCommand) printVerbose(out io.Writer, rs []*fastly.Token) {
	for _, r := range rs {
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
	}
	fmt.Fprintf(out, "\n")
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.Token) {
	t := text.NewTable(out)
	t.AddHeader("NAME", "TOKEN ID", "USER ID", "SCOPE", "SERVICES")
	for _, r := range rs {
		t.AddLine(r.Name, r.ID, r.UserID, r.Scope, strings.Join(r.Services, ", "))
	}
	t.Print()
}
