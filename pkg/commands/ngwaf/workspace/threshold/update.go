package threshold

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/ngwaf/v1/workspaces/thresholds"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a workspace threshold.
type UpdateCommand struct {
	argparser.Base

	// Required.
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`
	ThresholdID string                        `required:"" help:"Threshold ID"`

	// Optional.
	Action      argparser.OptionalString `help:"The action to take when the threshold is exceeded. [block, log]"`
	DontNotify  argparser.OptionalString `name:"do-not-notify" help:"Whether to silence notifications when action is taken. [true, false]"`
	Duration    argparser.OptionalInt    `help:"The duration the action is in place in seconds. Default duration is 86,400 seconds (1 day)."`
	EnabledFlag argparser.OptionalString `name:"enabled" help:"Whether the threshold is active. [true, false]"`
	Interval    argparser.OptionalInt    `help:"The threshold interval in seconds. The default interval is 3600 seconds (1 hour)."`
	Limit       argparser.OptionalInt    `help:"The threshold limit. Input must be between 1 and 10000. Default limit is 10."`
	NameFlag    argparser.OptionalString `name:"name" help:"User submitted display name of a signal threshold. Input must be between 3 and 50 characters"`
	Signal      argparser.OptionalString `help:"The name of the signal this threshold is acting on"`

	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	input := &thresholds.UpdateInput{
		ThresholdID: &c.ThresholdID,
		WorkspaceID: &c.WorkspaceID.Value,
	}
	if c.Action.WasSet {
		input.Action = &c.Action.Value
	}
	if c.DontNotify.WasSet {
		dontNotify, err := argparser.ConvertBoolFromStringFlag(c.DontNotify.Value, "do-not-notify")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		input.DontNotify = dontNotify
	}
	if c.Duration.WasSet {
		input.Duration = &c.Duration.Value
	}
	if c.EnabledFlag.WasSet {
		enabled, err := argparser.ConvertBoolFromStringFlag(c.EnabledFlag.Value, "enabled")
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		input.Enabled = enabled
	}
	if c.Interval.WasSet {
		input.Interval = &c.Interval.Value
	}
	if c.Limit.WasSet {
		input.Limit = &c.Limit.Value
	}
	if c.NameFlag.WasSet {
		input.Name = &c.NameFlag.Value
	}
	if c.Signal.WasSet {
		input.Signal = &c.Signal.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := thresholds.Update(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Updated threshold '%s' for workspace '%s'", data.ThresholdID, c.WorkspaceID.Value)
	return nil
}
