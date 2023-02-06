package objectstorekeys

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// GetCommand calls the Fastly API to fetch the value of a key from an object store.
type GetCommand struct {
	cmd.Base
	json     bool
	manifest manifest.Data
	Input    fastly.GetObjectStoreKeyInput
}

// NewGetCommand returns a usable command registered under the parent.
func NewGetCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *GetCommand {
	c := GetCommand{
		Base: cmd.Base{
			Globals: globals,
		},
		manifest: data,
	}
	c.CmdClause = parent.Command("get", "Get the value associated with a key")
	c.CmdClause.Flag("store-id", "Store ID").Short('s').Required().StringVar(&c.Input.ID)
	c.CmdClause.Flag("key-name", "Key name").Short('k').Required().StringVar(&c.Input.Key)

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
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
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

	if c.Globals.Flag.Verbose {
		text.PrintObjectStoreKeyValue(out, "", c.Input.Key, value)
		return nil
	}

	text.Output(out, value)
	return nil
}
