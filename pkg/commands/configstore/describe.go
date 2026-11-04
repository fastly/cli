package configstore

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *DescribeCommand) AfterApply() error {
	c.input.StoreID = c.StoreID
	return nil
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	argparser.Base

	// Required.
	StoreID string `short:"s" required:"" help:"Store ID"`

	// Optional.
	argparser.JSONOutput
	input    fastly.GetConfigStoreInput
	Metadata bool `short:"m" help:"Include config store metadata"`
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	cs, err := c.Globals.APIClient.GetConfigStore(context.TODO(), &c.input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	var csm *fastly.ConfigStoreMetadata
	if c.Metadata {
		csm, err = c.Globals.APIClient.GetConfigStoreMetadata(context.TODO(), &fastly.GetConfigStoreMetadataInput{
			StoreID: c.input.StoreID,
		})
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	}

	if c.JSONOutput.Enabled {
		// Create an ad-hoc structure for JSON representation of the config store
		// and its metadata.
		data := struct {
			*fastly.ConfigStore
			Metadata *fastly.ConfigStoreMetadata `json:"metadata,omitempty"`
		}{
			ConfigStore: cs,
			Metadata:    csm,
		}

		if ok, err := c.WriteJSON(out, data); ok {
			return err
		}
	}

	text.PrintConfigStore(out, cs, csm)

	return nil
}
