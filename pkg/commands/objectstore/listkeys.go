package objectstore

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

// ListKeysCommand calls the Fastly API to list the keys for a given object store.
type ListKeysCommand struct {
	cmd.Base
	json     bool
	manifest manifest.Data
	Input    fastly.ListObjectStoreKeysInput
}

// NewListKeysCommand returns a usable command registered under the parent.
func NewListKeysCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListKeysCommand {
	var c ListKeysCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Fastly object store keys")

	// required
	c.CmdClause.Flag("id", "ID of object store").Required().StringVar(&c.Input.ID)

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
func (c *ListKeysCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.json {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	o, err := c.Globals.APIClient.ListObjectStoreKeys(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.json {
		data, err := json.Marshal(o)
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

	text.PrintObjectStoreKeys(out, "", o.Data)

	return nil
}
