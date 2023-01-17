package objectstore

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// GetKeyCommand calls the Fastly API to fetch the value of a key from an object store.
type GetKeyCommand struct {
	cmd.Base
	json     bool
	manifest manifest.Data
	Input    fastly.GetObjectStoreKeyInput
}

// NewGetKeyCommand returns a usable command registered under the parent.
func NewGetKeyCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *GetKeyCommand {
	var c GetKeyCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("get", "Get Fastly object store key")

	// required
	c.CmdClause.Flag("id", "ID of object store").Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("key", "Key to fetch").Short('k').Required().StringVar(&c.Input.Key)

	// optional
	c.RegisterFlagBool(cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &c.json,
		Short:       'j',
	})

	return &c
}

// Exec invokes the application logic for the command.
func (c *GetKeyCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	value, err := c.Globals.APIClient.GetObjectStoreKey(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		text.Output(out, `{"%s": "%s"}`, c.Input.Key, value)
		return nil
	}

	text.PrintObjectStoreKeyValue(out, "", c.Input.Key, value)

	return nil
}
