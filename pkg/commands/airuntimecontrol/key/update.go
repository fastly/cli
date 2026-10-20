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

// UpdateCommand calls the Fastly API to update a virtual key.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	keyID string

	// Optional.
	name              argparser.OptionalString
	model             argparser.OptionalString
	provider          argparser.OptionalString
	expiresAt         time.Time
	securityEnabled   argparser.OptionalString
	securityAction    argparser.OptionalString
	rpmLimit          argparser.OptionalInt
	tpmLimit          argparser.OptionalInt
	spendLimitEnabled argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a virtual key")

	// Required.
	c.CmdClause.Flag("key-id", "Alphanumeric string identifying the virtual key").Required().StringVar(&c.keyID)

	// Optional.
	c.CmdClause.Flag("name", "An updated human-readable name for the virtual key").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("model", "An updated AI model identifier").Action(c.model.Set).StringVar(&c.model.Value)
	c.CmdClause.Flag("provider", "An updated AI model provider").Action(c.provider.Set).StringVar(&c.provider.Value)
	c.CmdClause.Flag("expires-at", "An updated expiration timestamp (RFC 3339 format)").HintOptions("2026-07-28T19:24:50+00:00").TimeVar(time.RFC3339, &c.expiresAt)
	c.CmdClause.Flag("security-enabled", "Whether AI Firewall security is enabled for requests made with this key [true, false]").Action(c.securityEnabled.Set).StringVar(&c.securityEnabled.Value)
	c.CmdClause.Flag("security-action", "An updated action AI Firewall takes when it detects a violation. One of: log, block, replace").HintOptions(securityActions...).Action(c.securityAction.Set).EnumVar(&c.securityAction.Value, securityActions...)
	c.CmdClause.Flag("rpm-limit", "An updated requests-per-minute limit for this key. Set to 0 to remove the limit").Action(c.rpmLimit.Set).IntVar(&c.rpmLimit.Value)
	c.CmdClause.Flag("tpm-limit", "An updated tokens-per-minute limit for this key. Set to 0 to remove the limit").Action(c.tpmLimit.Set).IntVar(&c.tpmLimit.Value)
	c.CmdClause.Flag("spend-limit-enabled", "Whether a spend limit is enabled for this key [true, false]").Action(c.spendLimitEnabled.Set).StringVar(&c.spendLimitEnabled.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &key.UpdateInput{
		KeyID: &c.keyID,
	}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.model.WasSet {
		input.Model = &c.model.Value
	}
	if c.provider.WasSet {
		input.Provider = &c.provider.Value
	}
	if !c.expiresAt.IsZero() {
		input.ExpiresAt = &c.expiresAt
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
		// Allow the limit to be reset.
		if c.rpmLimit.Value == 0 {
			input.RPMLimit = fastly.NullValue[int]()
		} else {
			input.RPMLimit = fastly.NewNullable(c.rpmLimit.Value)
		}
	}
	if c.tpmLimit.WasSet {
		// Allow the limit to be reset.
		if c.tpmLimit.Value == 0 {
			input.TPMLimit = fastly.NullValue[int]()
		} else {
			input.TPMLimit = fastly.NewNullable(c.tpmLimit.Value)
		}
	}
	if c.spendLimitEnabled.WasSet {
		b, err := argparser.ConvertBoolFromStringFlag(c.spendLimitEnabled.Value, "spend-limit-enabled")
		if err != nil {
			return err
		}
		input.SpendLimitEnabled = b
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
