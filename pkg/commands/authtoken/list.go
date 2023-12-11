package authtoken

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List API tokens")

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

	var (
		err error
		o   []*fastly.Token
	)

	if err = c.customerID.Parse(); err == nil {
		if !c.customerID.WasSet && !c.Globals.Flags.Quiet {
			text.Info(out, "Listing customer tokens for the FASTLY_CUSTOMER_ID environment variable\n\n")
		}
		input := c.constructInput()
		o, err = c.Globals.APIClient.ListCustomerTokens(input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	} else {
		o, err = c.Globals.APIClient.ListTokens()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
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
func (c *ListCommand) constructInput() *fastly.ListCustomerTokensInput {
	var input fastly.ListCustomerTokensInput

	input.CustomerID = c.customerID.Value

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
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.Token) error {
	t := text.NewTable(out)
	t.AddHeader("NAME", "TOKEN ID", "USER ID", "SCOPE", "SERVICES")
	for _, r := range rs {
		t.AddLine(r.Name, r.ID, r.UserID, r.Scope, strings.Join(r.Services, ", "))
	}
	t.Print()
	return nil
}
