package tsigkey

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/dns/v1/tsigkeys"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a TSIG key.
type CreateCommand struct {
	argparser.Base

	// Required.
	NameFlag  string `name:"name" required:"" help:"The name of the TSIG key."`
	Algorithm string `required:"" enum:"hmac-sha224,hmac-sha256,hmac-sha384,hmac-sha512" hints:"hmac-sha224,hmac-sha256,hmac-sha384,hmac-sha512" help:"The algorithm of the TSIG key. Valid values are: hmac-sha224, hmac-sha256, hmac-sha384, hmac-sha512."`
	Secret    string `required:"" help:"The Base64 encoded secret key."`

	// Optional.
	Description argparser.OptionalString `help:"A freeform descriptive note."`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if strings.Contains(c.NameFlag, " ") {
		return fmt.Errorf("invalid --name value %q: TSIG key names cannot contain spaces", c.NameFlag)
	}
	if len(c.NameFlag) > 255 {
		return fmt.Errorf("invalid --name value %q: TSIG key names cannot exceed 255 characters", c.NameFlag)
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &tsigkeys.CreateInput{
		Name:      &c.NameFlag,
		Algorithm: &c.Algorithm,
		Secret:    &c.Secret,
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	k, err := tsigkeys.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Name": c.NameFlag,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, k); ok {
		return err
	}

	text.Success(out, "Created TSIG key '%s' (tsig-key-id: %s)", *k.Name, *k.ID)
	return nil
}
