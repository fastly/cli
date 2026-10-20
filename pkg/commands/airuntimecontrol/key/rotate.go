package key

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RotateCommand calls the Fastly API to rotate a virtual key.
type RotateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	keyID string

	// Optional.
	expiresAt time.Time
}

// NewRotateCommand returns a usable command registered under the parent.
func NewRotateCommand(parent argparser.Registerer, g *global.Data) *RotateCommand {
	c := RotateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("rotate", "Rotate a virtual key, generating a new access token and invalidating the previous one")

	// Required.
	c.CmdClause.Flag("key-id", "Alphanumeric string identifying the virtual key").Required().StringVar(&c.keyID)

	// Optional.
	c.CmdClause.Flag("expires-at", "Expiration timestamp for the rotated key (RFC 3339 format). If not set, the key's existing expiration is cleared").HintOptions("2026-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &c.expiresAt)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
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

	input := &key.RotateInput{
		KeyID: &c.keyID,
	}
	if !c.expiresAt.IsZero() {
		input.ExpiresAt = &c.expiresAt
	}

	data, err := key.Rotate(context.TODO(), fc, input)
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
