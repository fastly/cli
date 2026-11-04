package tsigkey

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/dns/v1/tsigkeys"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a TSIG key.
type UpdateCommand struct {
	argparser.Base

	// Required.
	TSIGKeyID string `required:"" help:"The TSIG key ID to update."`

	// Optional.
	NameFlag    argparser.OptionalString `name:"name" help:"The name of the TSIG key."`
	Algorithm   argparser.OptionalString `hints:"hmac-sha224,hmac-sha256,hmac-sha384,hmac-sha512" help:"The algorithm of the TSIG key. Valid values are: hmac-sha224, hmac-sha256, hmac-sha384, hmac-sha512."`
	Secret      argparser.OptionalString `help:"The Base64 encoded secret key."`
	Description argparser.OptionalString `help:"A freeform descriptive note."`

	argparser.JSONOutput
}

// AfterApply checks the values kong cannot check by itself.
func (c *UpdateCommand) AfterApply(ctx *kong.Context) error {
	if (c.Algorithm.Value != "" || kongcompat.Given(ctx, "algorithm")) && !slices.Contains([]string{"hmac-sha224", "hmac-sha256", "hmac-sha384", "hmac-sha512"}, c.Algorithm.Value) {
		return fmt.Errorf("enum value must be one of hmac-sha224,hmac-sha256,hmac-sha384,hmac-sha512, got '%s'", c.Algorithm.Value)
	}
	return nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &tsigkeys.UpdateInput{
		TSIGKeyID: &c.TSIGKeyID,
	}
	if c.NameFlag.WasSet {
		if strings.Contains(c.NameFlag.Value, " ") {
			return fmt.Errorf("invalid --name value %q: TSIG key names cannot contain spaces", c.NameFlag.Value)
		}
		if len(c.NameFlag.Value) > 255 {
			return fmt.Errorf("invalid --name value %q: TSIG key names cannot exceed 255 characters", c.NameFlag.Value)
		}
		input.Name = &c.NameFlag.Value
	}
	if c.Algorithm.WasSet {
		input.Algorithm = &c.Algorithm.Value
	}
	if c.Secret.WasSet {
		input.Secret = &c.Secret.Value
	}
	if c.Description.WasSet {
		if strings.TrimSpace(c.Description.Value) == "" {
			input.Description = fastly.NullValue[string]()
		} else {
			input.Description = fastly.NewNullable(c.Description.Value)
		}
	}

	k, err := tsigkeys.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TSIG Key ID": c.TSIGKeyID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, k); ok {
		return err
	}

	text.Success(out, "Updated TSIG key '%s' (tsig-key-id: %s)", *k.Name, *k.ID)
	return nil
}
