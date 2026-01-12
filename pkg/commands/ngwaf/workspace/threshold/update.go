package threshold

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/thresholds"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/flagconversion"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a workspace threshold.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	thresholdID string
	workspaceID argparser.OptionalWorkspaceID

	// Optional.
	action     argparser.OptionalString
	dontNotify argparser.OptionalString
	duration   argparser.OptionalInt
	enabled    argparser.OptionalString
	interval   argparser.OptionalInt
	limit      argparser.OptionalInt
	name       argparser.OptionalString
	signal     argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a workspace threshold")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
	})
	c.CmdClause.Flag("threshold-id", "Threshold ID").Required().StringVar(&c.thresholdID)

	// Optional.
	c.CmdClause.Flag("action", "The action to take when the threshold is exceeded. [block, log]").Action(c.action.Set).StringVar(&c.action.Value)
	c.CmdClause.Flag("do-not-notify", "Whether to silence notifications when action is taken. [true, false]").Action(c.dontNotify.Set).StringVar(&c.dontNotify.Value)
	c.CmdClause.Flag("duration", "The duration the action is in place in seconds. Default duration is 86,400 seconds (1 day).").Action(c.duration.Set).IntVar(&c.duration.Value)
	c.CmdClause.Flag("enabled", "Whether the threshold is active. [true, false]").Action(c.enabled.Set).StringVar(&c.enabled.Value)
	c.CmdClause.Flag("interval", "The threshold interval in seconds. The default interval is 3600 seconds (1 hour).").Action(c.interval.Set).IntVar(&c.interval.Value)
	c.CmdClause.Flag("limit", "The threshold limit. Input must be between 1 and 10000. Default limit is 10.").Action(c.limit.Set).IntVar(&c.limit.Value)
	c.CmdClause.Flag("name", "User submitted display name of a signal threshold. Input must be between 3 and 50 characters").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("signal", "The name of the signal this threshold is acting on").Action(c.signal.Set).StringVar(&c.signal.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.workspaceID.Parse(); err != nil {
		return err
	}
	input := &thresholds.UpdateInput{
		ThresholdID: &c.thresholdID,
		WorkspaceID: &c.workspaceID.Value,
	}
	if c.action.WasSet {
		input.Action = &c.action.Value
	}
	if c.dontNotify.WasSet {
		dontNotify, err := flagconversion.ConvertBoolFromStringFlag(c.dontNotify.Value)
		if err != nil {
			err := errors.New("'do-not-notify' flag must be one of the following [true, false]")
			c.Globals.ErrLog.Add(err)
			return err
		}
		input.DontNotify = dontNotify
	}
	if c.duration.WasSet {
		input.Duration = &c.duration.Value
	}
	if c.enabled.WasSet {
		enabled, err := flagconversion.ConvertBoolFromStringFlag(c.enabled.Value)
		if err != nil {
			err := errors.New("'enabled' flag must be one of the following [true, false]")
			c.Globals.ErrLog.Add(err)
			return err
		}
		input.Enabled = enabled
	}
	if c.interval.WasSet {
		input.Interval = &c.interval.Value
	}
	if c.limit.WasSet {
		input.Limit = &c.limit.Value
	}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.signal.WasSet {
		input.Signal = &c.signal.Value
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

	text.Success(out, "Updated threshold '%s' for workspace '%s'", data.ThresholdID, c.workspaceID.Value)
	return nil
}
