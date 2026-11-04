package privatekey

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	argparser.Base

	// Optional.
	FilterInUse string `default:"" enum:"false," hints:"false" help:"Limit the returned keys to those without any matching TLS certificates"`
	argparser.JSONOutput

	PageNumber int `name:"page" help:"Page number of data set to fetch"`
	PageSize   int `name:"per-page" help:"Number of records per page"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	o, err := c.Globals.APIClient.ListPrivateKeys(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter In Use": c.FilterInUse,
			"Page Number":   c.PageNumber,
			"Page Size":     c.PageSize,
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

	if c.FilterInUse != "" {
		input.FilterInUse = c.FilterInUse
	}
	if c.PageNumber > 0 {
		input.PageNumber = c.PageNumber
	}
	if c.PageSize > 0 {
		input.PageSize = c.PageSize
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
