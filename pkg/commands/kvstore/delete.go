package kvstore

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/kvstoreentry"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// DeleteCommand calls the Fastly API to delete a kv store.
type DeleteCommand struct {
	argparser.Base

	// Required.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	DeleteAll bool `name:"all" short:"a" help:"Delete all entries within the store"`
	PoolSize  int  `name:"concurrency" short:"r" default:"100" help:"The thread pool size (ignored when set without the --all flag)"`
	argparser.JSONOutput

	MaxErrors int                       `short:"m" default:"100" help:"The number of errors to accept before stopping (ignored when set without the --all flag)"`
	Input     fastly.DeleteKVStoreInput `kong:"-"`
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DeleteCommand) AfterApply() error {
	c.Input.StoreID = c.StoreID
	return nil
}

// Exec invokes the application logic for the command.
func (c *DeleteCommand) Exec(in io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if c.DeleteAll {
		if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
			text.Warning(out, "This will delete ALL entries from your store!\n\n")
			cont, err := text.AskYesNo(out, "Are you sure you want to continue? [y/N]: ", in)
			if err != nil {
				return err
			}
			if !cont {
				return nil
			}
			text.Break(out)
		}
		dc := kvstoreentry.DeleteCommand{
			Base: argparser.Base{
				Globals: c.Globals,
			},
			DeleteAll: c.DeleteAll,
			MaxErrors: c.MaxErrors,
			PoolSize:  c.PoolSize,
			StoreID:   c.Input.StoreID,
		}
		if err := dc.DeleteMultipleKeys(out, ""); err != nil {
			return err
		}
		text.Break(out)
	}

	err := c.Globals.APIClient.DeleteKVStore(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("failed to delete KV store: %w", err)
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}{
			c.Input.StoreID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Deleted KV Store '%s'", c.Input.StoreID)
	return nil
}
