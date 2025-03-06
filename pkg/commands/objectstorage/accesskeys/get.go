package accesskeys

import (
	"errors"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/objectstorage/accesskeys"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand calls the Fastly API to get an access key.
type GetCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	accessKeyId string
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewGetCommand(parent argparser.Registerer, g *global.Data) *GetCommand {
	c := GetCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("get", "Get an access key")

	// Required.
	c.CmdClause.Flag("ak-id", "Access key ID").Required().StringVar(&c.accessKeyId)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	accessKey, err := accesskeys.Get(fc, &accesskeys.GetInput{
		AccessKeyID: &c.accessKeyId,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, accessKey); ok {
		return err
	}

	text.PrintAccessKey(out, accessKey)
	return nil
}
