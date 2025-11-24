package datadog

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/common"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/datadog"
)

// CreateCommand calls the Fastly API to create Datadog alerts.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	common.BaseAlertFlags
	common.DatadogConfigFlags

	// Optional.
	common.AlertDataFlags
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a Datadog alert").Alias("add")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.WorkspaceID.Value,
		Action:      c.WorkspaceID.Set,
	})
	c.CmdClause.Flag("key", "Datadog integration key.").Required().StringVar(&c.Key)
	c.CmdClause.Flag("site", "Datadog site.").Required().StringVar(&c.Site)

	// Optional.
	c.CmdClause.Flag("description", "An optional description for the alert.").Action(c.Description.Set).StringVar(&c.Description.Value)
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
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	input := &datadog.CreateInput{
		WorkspaceID: &c.WorkspaceID.Value,
		Config: &datadog.CreateConfig{
			Key:  &c.Key,
			Site: &c.Site,
		},
		// Set 'Events' to the only possible value, 'flag'
		Events: common.GetDefaultEvents(),
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := datadog.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created a '%s' alert '%s' (workspace-id: %s)", data.Type, data.ID, c.WorkspaceID.Value)
	return nil
}
