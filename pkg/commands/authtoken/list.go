package authtoken

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
	c.CmdClause = parent.Command("list", "List API tokens")
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

	var (
		err error
		rs  []*fastly.Token
	)

	if err = c.customerID.Parse(); err == nil {
		if !c.customerID.WasSet {
			text.Info(out, "Listing customer tokens for the FASTLY_CUSTOMER_ID environment variable")
			text.Break(out)
		}

		input := c.constructInput()

		rs, err = c.Globals.APIClient.ListCustomerTokens(input)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	} else {
		rs, err = c.Globals.APIClient.ListTokens()
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
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
	if c.json {
		data, err := json.Marshal(rs)
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
	t.AddHeader("NAME", "TOKEN ID", "USER ID", "SCOPE", "SERVICES")
	for _, r := range rs {
		t.AddLine(r.Name, r.ID, r.UserID, r.Scope, strings.Join(r.Services, ", "))
	}
	t.Print()
	return nil
}
