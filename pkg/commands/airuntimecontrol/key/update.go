package key

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/airuntimecontrol/v1/key"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a virtual key.
type UpdateCommand struct {
	argparser.Base

	// Required.
	KeyID string `required:"" help:"Alphanumeric string identifying the virtual key"`

	// Optional.
	NameFlag  argparser.OptionalString `name:"name" help:"An updated human-readable name for the virtual key"`
	Model     argparser.OptionalString `help:"An updated AI model identifier"`
	Provider  argparser.OptionalString `help:"An updated AI model provider"`
	ExpiresAt time.Time                `hints:"2026-07-28T19:24:50+00:00" format:"2006-01-02T15:04:05Z07:00" help:"An updated expiration timestamp (RFC 3339 format)"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &key.UpdateInput{
		KeyID: &c.KeyID,
	}
	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Model.WasSet {
		input.Model = &c.Model.Value
	}
	if c.Provider.WasSet {
		input.Provider = &c.Provider.Value
	}
	if !c.ExpiresAt.IsZero() {
		input.ExpiresAt = &c.ExpiresAt
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := key.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintVirtualKey(out, data)
	return nil
}
