package privatekey

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Show a TLS private key").Alias("get")
	c.Globals = g
	c.manifest = m

	// required
	c.CmdClause.Flag("id", "Alphanumeric string identifying a private Key").Required().StringVar(&c.id)

	// optional
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

	id       string
	json     bool
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := c.constructInput()

	r, err := c.Globals.APIClient.GetPrivateKey(input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TLS Certificate ID": c.id,
		})
		return err
	}

	return c.print(out, r)
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() *fastly.GetPrivateKeyInput {
	var input fastly.GetPrivateKeyInput

	input.ID = c.id

	return &input
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, r *fastly.PrivateKey) error {
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
	fmt.Fprintf(out, "Key Length: %d\n", r.KeyLength)
	fmt.Fprintf(out, "Key Type: %s\n", r.KeyType)
	fmt.Fprintf(out, "Public Key SHA1: %s\n", r.PublicKeySHA1)

	if r.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", r.CreatedAt)
	}

	fmt.Fprintf(out, "Replace: %t\n", r.Replace)

	return nil
}
