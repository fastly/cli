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

// CreateCommand calls the Fastly API to create a workspace threshold.
type CreateCommand struct {
	argparser.Base

	// Required.
	Action      string                        `required:"" help:"The action to take when the threshold is exceeded. [block, log]"`
	DontNotify  argparser.OptionalString      `name:"do-not-notify" required:"" help:"Whether to silence notifications when action is taken. [true, false]"`
	Duration    int                           `required:"" help:"The duration the action is in place in seconds. Default duration is 86,400 seconds (1 day)."`
	EnabledFlag argparser.OptionalString      `name:"enabled" required:"" help:"Whether the threshold is active. [true, false]"`
	Interval    int                           `required:"" help:"The threshold interval in seconds. The default interval is 3600 seconds (1 hour)."`
	Limit       int                           `required:"" help:"The threshold limit. Input must be between 1 and 10000. Default limit is 10."`
	NameFlag    string                        `name:"name" required:"" help:"User submitted display name of a signal threshold. Input must be between 3 and 50 characters"`
	Signal      string                        `required:"" help:"The name of the signal this threshold is acting on"`
	WorkspaceID argparser.OptionalWorkspaceID `help:"Alphanumeric string identifying the NGWAF Workspace (falls back to FASTLY_WORKSPACE_ID)"`

	// Optional.
	argparser.JSONOutput
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}

	enabled, err := argparser.ConvertBoolFromStringFlag(c.EnabledFlag.Value, "enabled")
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	dontNotify, err := argparser.ConvertBoolFromStringFlag(c.DontNotify.Value, "do-not-notify")
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	input := &thresholds.CreateInput{
		Action:      &c.Action,
		Duration:    &c.Duration,
		Enabled:     enabled,
		Interval:    &c.Interval,
		Limit:       &c.Limit,
		Name:        &c.NameFlag,
		DontNotify:  dontNotify,
		Signal:      &c.Signal,
		WorkspaceID: &c.WorkspaceID.Value,
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := thresholds.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created threshold '%s' for workspace '%s'", data.ThresholdID, c.WorkspaceID.Value)
	return nil
}
