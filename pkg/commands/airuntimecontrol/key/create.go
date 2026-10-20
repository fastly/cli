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
	provider string

	// Optional.
	expiresAt       time.Time
	model           argparser.OptionalString
	securityEnabled argparser.OptionalString
	securityAction  argparser.OptionalString
	rpmLimit        argparser.OptionalInt
	tpmLimit        argparser.OptionalInt
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
	c.CmdClause.Flag("provider", "AI model provider name (e.g. \"Anthropic\")").Required().StringVar(&c.provider)

	// Optional.
	c.CmdClause.Flag("expires-at", "Expiration timestamp (RFC 3339 format)").HintOptions("2026-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &c.expiresAt)
	c.CmdClause.Flag("model", "AI model identifier (e.g. \"claude-fable-5-1\")").Action(c.model.Set).StringVar(&c.model.Value)
	c.CmdClause.Flag("security-enabled", "Whether AI Firewall security is enabled for requests made with this key [true, false]").Action(c.securityEnabled.Set).StringVar(&c.securityEnabled.Value)
	c.CmdClause.Flag("security-action", "Action AI Firewall takes when it detects a violation. One of: log, block, replace").HintOptions(securityActions...).Action(c.securityAction.Set).EnumVar(&c.securityAction.Value, securityActions...)
	c.CmdClause.Flag("rpm-limit", "Requests-per-minute limit for this key").Action(c.rpmLimit.Set).IntVar(&c.rpmLimit.Value)
	c.CmdClause.Flag("tpm-limit", "Tokens-per-minute limit for this key").Action(c.tpmLimit.Set).IntVar(&c.tpmLimit.Value)
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
		Provider: &c.provider,
	}
	if !c.expiresAt.IsZero() {
		input.ExpiresAt = &c.expiresAt
	}
	if c.model.WasSet {
		input.Model = &c.model.Value
	}
	if c.securityEnabled.WasSet {
		b, err := argparser.ConvertBoolFromStringFlag(c.securityEnabled.Value, "security-enabled")
		if err != nil {
			return err
		}
		input.SecurityEnabled = b
	}
	if c.securityAction.WasSet {
		input.SecurityAction = &c.securityAction.Value
	}
	if c.rpmLimit.WasSet {
		input.RPMLimit = &c.rpmLimit.Value
	}
	if c.tpmLimit.WasSet {
		input.TPMLimit = &c.tpmLimit.Value
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
