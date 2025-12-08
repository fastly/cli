package threshold

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/thresholds"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a workspace threshold.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	action      string
	dontNotify  bool
	duration    int
	enabled     bool
	interval    int
	limit       int
	name        string
	signal      string
	workspaceID argparser.OptionalWorkspaceID

	// Optional.
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a workspace threshold").Alias("add")

	// Required.
	c.CmdClause.Flag("action", "The action to take when the threshold is exceeded. Options include 'block' or 'log'.").Required().StringVar(&c.action)
	c.CmdClause.Flag("do-not-notify", "Whether to silence notifications when action is taken.").Required().BoolVar(&c.dontNotify)
	c.CmdClause.Flag("duration", "The duration the action is in place in seconds. Default duration is 86,400 seconds (1 day).").Required().IntVar(&c.duration)
	c.CmdClause.Flag("enabled", "Whether the threshold is active. Options include 'true' or 'false'.").Required().BoolVar(&c.enabled)
	c.CmdClause.Flag("interval", "The threshold interval in seconds. The default interval is 3600 seconds (1 hour).").Required().IntVar(&c.interval)
	c.CmdClause.Flag("limit", "The threshold limit. Input must be between 1 and 10000. Default limit is 10.").Required().IntVar(&c.limit)
	c.CmdClause.Flag("name", "User submitted display name of a signal threshold. Input must be between 3 and 50 characters").Required().StringVar(&c.name)
	c.CmdClause.Flag("signal", "The name of the signal this threshold is acting on").Required().StringVar(&c.signal)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.workspaceID.Value,
		Action:      c.workspaceID.Set,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.workspaceID.Parse(); err != nil {
		return err
	}
	input := &thresholds.CreateInput{
		Action:      &c.action,
		Duration:    &c.duration,
		Enabled:     &c.enabled,
		Interval:    &c.interval,
		Limit:       &c.limit,
		Name:        &c.name,
		DontNotify:  &c.dontNotify,
		Signal:      &c.signal,
		WorkspaceID: &c.workspaceID.Value,
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

	text.Success(out, "Created threshold '%s' for workspace '%s'", data.ThresholdID, c.workspaceID.Value)
	return nil
}
