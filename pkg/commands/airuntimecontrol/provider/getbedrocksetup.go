package provider

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/providerconnection"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// GetBedrockSetupCommand calls the Fastly API to get the setup information
// needed to create an aws-iam (Bedrock) provider connection.
type GetBedrockSetupCommand struct {
	argparser.Base
	argparser.JSONOutput
}

// NewGetBedrockSetupCommand returns a usable command registered under the parent.
func NewGetBedrockSetupCommand(parent argparser.Registerer, g *global.Data) *GetBedrockSetupCommand {
	c := GetBedrockSetupCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("get-bedrock-setup", "Get Bedrock setup information")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetBedrockSetupCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := providerconnection.GetBedrockSetup(context.TODO(), fc)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintBedrockSetup(out, data)
	return nil
}
