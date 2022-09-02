package privatekey

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v6/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Show a TLS private key").Alias("get")
	c.Globals = globals
	c.manifest = data

	// Required flags
	c.CmdClause.Flag("id", "Alphanumeric string identifying a private Key").Required().StringVar(&c.id)

	// Optional Flags
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
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
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

	err = c.print(out, r)
	if err != nil {
		return err
	}
	return nil
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
		out.Write(data)
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
