package privatekey

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all TLS private keys")
	c.Globals = g

	// Optional.
	c.CmdClause.Flag("filter-in-use", "Limit the returned keys to those without any matching TLS certificates").HintOptions("false").EnumVar(&c.filterInUse, "false")
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.pageNumber)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.pageSize)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	filterInUse string
	pageNumber  int
	pageSize    int
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListPrivateKeys(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter In Use": c.filterInUse,
			"Page Number":   c.pageNumber,
			"Page Size":     c.pageSize,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if c.Globals.Verbose() {
		printVerbose(out, o)
	} else {
		err = c.printSummary(out, o)
		if err != nil {
			return err
		}
	}
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *ListCommand) constructInput() *fastly.ListPrivateKeysInput {
	var input fastly.ListPrivateKeysInput

	if c.filterInUse != "" {
		input.FilterInUse = c.filterInUse
	}
	if c.pageNumber > 0 {
		input.PageNumber = c.pageNumber
	}
	if c.pageSize > 0 {
		input.PageSize = c.pageSize
	}

	return &input
}

// printVerbose displays the information returned from the API in a verbose
// format.
func printVerbose(out io.Writer, rs []*fastly.PrivateKey) {
	for _, r := range rs {
		fmt.Fprintf(out, "\nID: %s\n", r.ID)
		fmt.Fprintf(out, "Name: %s\n", r.Name)
		fmt.Fprintf(out, "Key Length: %d\n", r.KeyLength)
		fmt.Fprintf(out, "Key Type: %s\n", r.KeyType)
		fmt.Fprintf(out, "Public Key SHA1: %s\n", r.PublicKeySHA1)

		if r.CreatedAt != nil {
			fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
		}

		fmt.Fprintf(out, "Replace: %t\n", r.Replace)
		fmt.Fprintf(out, "\n")
	}
}

// printSummary displays the information returned from the API in a summarised
// format.
func (c *ListCommand) printSummary(out io.Writer, rs []*fastly.PrivateKey) error {
	t := text.NewTable(out)
	t.AddHeader("ID", "NAME", "KEY LENGTH", "KEY TYPE", "PUBLIC KEY SHA1", "REPLACE")
	for _, r := range rs {
		t.AddLine(r.ID, r.Name, r.KeyLength, r.KeyType, r.PublicKeySHA1, r.Replace)
	}
	t.Print()
	return nil
}
