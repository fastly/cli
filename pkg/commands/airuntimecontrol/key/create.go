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

// CreateCommand calls the Fastly API to create a virtual key.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	name     string
	model    string
	provider string
	userID   string

	// Optional.
	expiresAt time.Time
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a virtual key").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "Human-readable name for the virtual key").Required().StringVar(&c.name)
	c.CmdClause.Flag("model", "AI model identifier (e.g. \"claude-sonnet-4-20250514\")").Required().StringVar(&c.model)
	c.CmdClause.Flag("provider", "AI model provider name (e.g. \"Anthropic\")").Required().StringVar(&c.provider)
	c.CmdClause.Flag("user-id", "ID of the user creating the key").Required().StringVar(&c.userID)

	// Optional.
	c.CmdClause.Flag("expires-at", "Expiration timestamp (RFC 3339 format)").HintOptions("2026-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &c.expiresAt)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &key.CreateInput{
		Name:     &c.name,
		Model:    &c.model,
		Provider: &c.provider,
		UserID:   &c.userID,
	}
	if !c.expiresAt.IsZero() {
		input.ExpiresAt = &c.expiresAt
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
