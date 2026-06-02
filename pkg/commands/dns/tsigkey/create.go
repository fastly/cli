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

// CreateCommand calls the Fastly API to create a TSIG key.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	name      string
	algorithm string
	secret    string

	// Optional.
	description argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a TSIG key").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The name of the TSIG key.").Required().StringVar(&c.name)
	c.CmdClause.Flag("algorithm", "The algorithm of the TSIG key. Valid values are: hmac-sha224, hmac-sha256, hmac-sha384, hmac-sha512.").Required().HintOptions(AlgorithmOptions...).EnumVar(&c.algorithm, AlgorithmOptions...)
	c.CmdClause.Flag("secret", "The Base64 encoded secret key.").Required().StringVar(&c.secret)

	// Optional.
	c.CmdClause.Flag("description", "A freeform descriptive note.").Action(c.description.Set).StringVar(&c.description.Value)
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	if strings.Contains(c.name, " ") {
		return fmt.Errorf("invalid --name value %q: TSIG key names cannot contain spaces", c.name)
	}
	if len(c.name) > 255 {
		return fmt.Errorf("invalid --name value %q: TSIG key names cannot exceed 255 characters", c.name)
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	input := &tsigkeys.CreateInput{
		Name:      &c.name,
		Algorithm: &c.algorithm,
		Secret:    &c.secret,
	}
	if c.description.WasSet {
		input.Description = &c.description.Value
	}

	k, err := tsigkeys.Create(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Name": c.name,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, k); ok {
		return err
	}

	text.Success(out, "Created TSIG key '%s' (tsig-key-id: %s)", *k.Name, *k.ID)
	return nil
}
