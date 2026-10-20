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

// GetVertexSetupCommand calls the Fastly API to get the setup information
// needed to create a gcp-iam (Vertex AI) provider connection.
type GetVertexSetupCommand struct {
	argparser.Base
	argparser.JSONOutput
}

// NewGetVertexSetupCommand returns a usable command registered under the parent.
func NewGetVertexSetupCommand(parent argparser.Registerer, g *global.Data) *GetVertexSetupCommand {
	c := GetVertexSetupCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("get-vertex-setup", "Get Vertex AI setup information")

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetVertexSetupCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := providerconnection.GetVertexSetup(context.TODO(), fc)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintVertexSetup(out, data)
	return nil
}
