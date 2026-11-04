package kvstoreentry

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// GetCommand calls the Fastly API to fetch the value of a key from a kv store.
type GetCommand struct {
	argparser.Base

	Input fastly.GetKVStoreItemInput `kong:"-"`

	// Required.
	Key     string `short:"k" required:"" help:"Key name"`
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	Generation string `name:"if-generation-match" help:"Compares if the provided generation marker matches that of the object"`
	argparser.JSONOutput
}

// BeforeApply sets the command up before kong applies the command line.
func (c *GetCommand) BeforeApply() error {
	// This argument suppresses the 'Fastly API' output from the global verbose command.
	c.SuppressVerbose = true
	return nil
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *GetCommand) AfterApply() error {
	c.Input.Key = c.Key
	c.Input.StoreID = c.StoreID
	return nil
}

// Exec invokes the application logic for the command.
func (c *GetCommand) Exec(_ io.Reader, out io.Writer) error {
	// As the 'describe' command provides the object attributes,
	// we won't be supporting a --verbose flag here.
	if c.Globals.Flags.Verbose {
		return fmt.Errorf("the 'get' command does not support the --verbose flag")
	}

	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	// Validate generation value before making API call
	var inputGeneration uint64
	if c.Generation != "" {
		var err error
		inputGeneration, err = strconv.ParseUint(c.Generation, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid generation value: %s", c.Generation)
		}
	}

	result, err := c.Globals.APIClient.GetKVStoreItem(context.TODO(), &c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	// Check if the generation marker matches the API result
	if c.Generation != "" {
		if inputGeneration != result.Generation {
			return fmt.Errorf("generation value does not match: expected %d, got %d", result.Generation, inputGeneration)
		}
	}

	// Ensure we close the value reader.
	if result.Value != nil {
		defer result.Value.Close()
	}

	// Read the value from ReadCloser.
	var value string
	if result.Value != nil {
		valueBytes, err := io.ReadAll(result.Value)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		value = string(valueBytes)
	}

	if c.JSONOutput.Enabled {
		// We are encoding the value of the item here to ensure safe
		// output for binary content along with other outputs.
		encodedValue := base64.StdEncoding.EncodeToString([]byte(value))
		text.Output(out, `{"%s": "%s"}`, c.Input.Key, encodedValue)
		return nil
	}

	// IMPORTANT: Don't use `text` package as binary data can be messed up.
	fmt.Fprint(out, value)
	return nil
}
