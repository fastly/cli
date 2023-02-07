package secretstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// NewDescribeSecretCommand returns a usable command registered under the parent.
func NewDescribeSecretCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeSecretCommand {
	c := DescribeSecretCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}

	c.CmdClause = parent.Command("describe", "Retrieve a single secret").Alias("get")

	// Required.
	c.RegisterFlag(secretNameFlag(&c.Input.Name)) // --name
	c.RegisterFlag(storeIDFlag(&c.Input.ID))      // --store-id

	// Optional.
	c.RegisterFlagBool(c.jsonFlag()) // --json

	return &c
}

// DescribeSecretCommand calls the Fastly API to describe an appropriate resource.
type DescribeSecretCommand struct {
	cmd.Base
	jsonOutput

	Input    fastly.GetSecretInput
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (cmd *DescribeSecretCommand) Exec(_ io.Reader, out io.Writer) error {
	if cmd.Globals.Verbose() && cmd.jsonOutput.enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := cmd.Globals.APIClient.GetSecret(&cmd.Input)
	if err != nil {
		cmd.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := cmd.WriteJSON(out, o); ok {
		return err
	}

	text.PrintSecret(out, "", o)

	return nil
}
