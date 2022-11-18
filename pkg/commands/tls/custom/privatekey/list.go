package privatekey

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.CmdClause = parent.Command("list", "List all TLS private keys")
	c.Globals = globals
	c.manifest = data

	// optional
	c.CmdClause.Flag("filter-in-use", "Limit the returned keys to those without any matching TLS certificates").HintOptions("false").EnumVar(&c.filterInUse, "false")
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})
	c.CmdClause.Flag("page", "Page number of data set to fetch").IntVar(&c.pageNumber)
	c.CmdClause.Flag("per-page", "Number of records per page").IntVar(&c.pageSize)

	return &c
}

// ListCommand calls the Fastly API to list appropriate resources.
type ListCommand struct {
	cmd.Base

	filterInUse string
	json        bool
	manifest    manifest.Data
	pageNumber  int
	pageSize    int
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	rs, err := c.Globals.APIClient.ListPrivateKeys(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Filter In Use": c.filterInUse,
			"Page Number":   c.pageNumber,
			"Page Size":     c.pageSize,
		})
		return err
	}

	if c.Globals.Verbose() {
		printVerbose(out, rs)
	} else {
		err = c.printSummary(out, rs)
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
	t.AddHeader("ID", "NAME", "KEY LENGTH", "KEY TYPE", "PUBLIC KEY SHA1", "REPLACE")
	for _, r := range rs {
		t.AddLine(r.ID, r.Name, r.KeyLength, r.KeyType, r.PublicKeySHA1, r.Replace)
	}
	t.Print()
	return nil
}
