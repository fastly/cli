package tsigkey

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v15/fastly"
	"github.com/fastly/go-fastly/v15/fastly/dns/v1/tsigkeys"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a TSIG key.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	tsigKeyID string

	// Optional.
	name        argparser.OptionalString
	algorithm   argparser.OptionalString
	secret      argparser.OptionalString
	description argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a TSIG key")

	// Required.
	c.CmdClause.Flag("tsig-key-id", "The TSIG key ID to update.").Required().StringVar(&c.tsigKeyID)

	// Optional.
	c.CmdClause.Flag("name", "The name of the TSIG key.").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("algorithm", "The algorithm of the TSIG key. Valid values are: hmac-sha224, hmac-sha256, hmac-sha384, hmac-sha512.").Action(c.algorithm.Set).HintOptions(AlgorithmOptions...).EnumVar(&c.algorithm.Value, AlgorithmOptions...)
	c.CmdClause.Flag("secret", "The Base64 encoded secret key.").Action(c.secret.Set).StringVar(&c.secret.Value)
	c.CmdClause.Flag("description", "A freeform descriptive note.").Action(c.description.Set).StringVar(&c.description.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
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
		TSIGKeyID: &c.tsigKeyID,
	}
	if c.name.WasSet {
		if strings.Contains(c.name.Value, " ") {
			return fmt.Errorf("invalid --name value %q: TSIG key names cannot contain spaces", c.name.Value)
		}
		if len(c.name.Value) > 255 {
			return fmt.Errorf("invalid --name value %q: TSIG key names cannot exceed 255 characters", c.name.Value)
		}
		input.Name = &c.name.Value
	}
	if c.algorithm.WasSet {
		input.Algorithm = &c.algorithm.Value
	}
	if c.secret.WasSet {
		input.Secret = &c.secret.Value
	}
	if c.description.WasSet {
		if strings.TrimSpace(c.description.Value) == "" {
			input.Description = fastly.NullValue[string]()
		} else {
			input.Description = fastly.NewNullable(c.description.Value)
		}
	}

	k, err := tsigkeys.Update(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"TSIG Key ID": c.tsigKeyID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, k); ok {
		return err
	}

	text.Success(out, "Updated TSIG key '%s' (tsig-key-id: %s)", *k.Name, *k.ID)
	return nil
}
