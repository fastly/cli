package key

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// GetFailoverPolicyCommand calls the Fastly API to get a virtual key's failover
// policy.
type GetFailoverPolicyCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	keyID string
}

// NewGetFailoverPolicyCommand returns a usable command registered under the parent.
func NewGetFailoverPolicyCommand(parent argparser.Registerer, g *global.Data) *GetFailoverPolicyCommand {
	c := GetFailoverPolicyCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("get-failover-policy", "Get a virtual key's failover policy")

	// Required.
	c.CmdClause.Flag("key-id", "Alphanumeric string identifying the virtual key").Required().StringVar(&c.keyID)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetFailoverPolicyCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := key.GetFailoverPolicy(context.TODO(), fc, &key.GetFailoverPolicyInput{
		KeyID: &c.keyID,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintFailoverPolicy(out, data)
	return nil
}
