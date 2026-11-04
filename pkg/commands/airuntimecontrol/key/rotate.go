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

// RotateCommand calls the Fastly API to rotate a virtual key.
type RotateCommand struct {
	argparser.Base

	// Required.
	KeyID     string    `required:"" help:"Alphanumeric string identifying the virtual key"`
	ExpiresAt time.Time `required:"" hints:"2026-07-28T19:24:50+00:00" format:"2006-01-02T15:04:05Z07:00" help:"Expiration timestamp for the rotated key (RFC 3339 format)"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *RotateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := key.Rotate(context.TODO(), fc, &key.RotateInput{
		KeyID:     &c.KeyID,
		ExpiresAt: &c.ExpiresAt,
	})
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
