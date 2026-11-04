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

// CreateCommand calls the Fastly API to create a virtual key.
type CreateCommand struct {
	argparser.Base

	// Required.
	NameFlag string `name:"name" required:"" help:"Human-readable name for the virtual key"`
	Model    string `required:"" help:"AI model identifier (e.g. \"claude-sonnet-4-20250514\")"`
	Provider string `required:"" help:"AI model provider name (e.g. \"Anthropic\")"`

	// Optional.
	ExpiresAt time.Time `hints:"2026-07-28T19:24:50+00:00" format:"2006-01-02T15:04:05Z07:00" help:"Expiration timestamp (RFC 3339 format)"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &key.CreateInput{
		Name:     &c.NameFlag,
		Model:    &c.Model,
		Provider: &c.Provider,
	}
	if !c.ExpiresAt.IsZero() {
		input.ExpiresAt = &c.ExpiresAt
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := key.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintVirtualKeyWithToken(out, data)
	return nil
}
