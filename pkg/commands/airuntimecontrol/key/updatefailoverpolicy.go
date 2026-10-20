package key

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/kingpin"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/key"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateFailoverPolicyCommand calls the Fastly API to update a virtual key's
// failover policy.
type UpdateFailoverPolicyCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	keyID   string
	targets []string

	// Optional.
	timeoutEnabled        argparser.OptionalString
	rateLimit429Enabled   argparser.OptionalString
	serverError5xxEnabled argparser.OptionalString
}

// NewUpdateFailoverPolicyCommand returns a usable command registered under the parent.
func NewUpdateFailoverPolicyCommand(parent argparser.Registerer, g *global.Data) *UpdateFailoverPolicyCommand {
	c := UpdateFailoverPolicyCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update-failover-policy", "Update a virtual key's failover policy")

	// Required.
	c.CmdClause.Flag("key-id", "Alphanumeric string identifying the virtual key").Required().StringVar(&c.keyID)
	c.CmdClause.Flag("targets", "Comma-separated list of the provider connections to retry against, in order").Required().StringsVar(&c.targets, kingpin.Separator(","))

	// Optional.
	c.CmdClause.Flag("timeout-enabled", "Whether timeouts trigger failover [true, false]").Action(c.timeoutEnabled.Set).StringVar(&c.timeoutEnabled.Value)
	c.CmdClause.Flag("rate-limit-429-enabled", "Whether 429 rate-limit responses trigger failover [true, false]").Action(c.rateLimit429Enabled.Set).StringVar(&c.rateLimit429Enabled.Value)
	c.CmdClause.Flag("server-error-5xx-enabled", "Whether 5xx server errors trigger failover [true, false]").Action(c.serverError5xxEnabled.Set).StringVar(&c.serverError5xxEnabled.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateFailoverPolicyCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &key.UpdateFailoverPolicyInput{
		KeyID: &c.keyID,
	}
	for _, id := range c.targets {
		input.Targets = append(input.Targets, key.FailoverTarget{ProviderConnectionID: id})
	}
	if c.timeoutEnabled.WasSet {
		b, err := argparser.ConvertBoolFromStringFlag(c.timeoutEnabled.Value, "timeout-enabled")
		if err != nil {
			return err
		}
		input.TimeoutEnabled = b
	}
	if c.rateLimit429Enabled.WasSet {
		b, err := argparser.ConvertBoolFromStringFlag(c.rateLimit429Enabled.Value, "rate-limit-429-enabled")
		if err != nil {
			return err
		}
		input.RateLimit429Enabled = b
	}
	if c.serverError5xxEnabled.WasSet {
		b, err := argparser.ConvertBoolFromStringFlag(c.serverError5xxEnabled.Value, "server-error-5xx-enabled")
		if err != nil {
			return err
		}
		input.ServerError5xxEnabled = b
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := key.UpdateFailoverPolicy(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.PrintFailoverPolicy(out, data)
	return nil
}
