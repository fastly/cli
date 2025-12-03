package jira

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v12/fastly"
	"github.com/fastly/go-fastly/v12/fastly/ngwaf/v1/workspaces/alerts/jira"
)

// CreateCommand calls the Fastly API to create Jira alerts.
type CreateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	alert.WorkspaceIDFlags
	ConfigFlags

	// Optional.
	alert.DataFlags
	OptConfigFlags
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a Jira alert").Alias("add")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagNGWAFWorkspaceID,
		Description: argparser.FlagNGWAFWorkspaceIDDesc,
		Dst:         &c.WorkspaceID.Value,
		Action:      c.WorkspaceID.Set,
	})
	c.CmdClause.Flag("host", "Host name of the Jira instance.").Required().StringVar(&c.Host)
	c.CmdClause.Flag("key", "Jira API key.").Required().StringVar(&c.Key)
	c.CmdClause.Flag("project", "Specifies the Jira project where the issue will be created.").Required().StringVar(&c.Project)
	c.CmdClause.Flag("username", "Jira username of the user who created the ticket.").Required().StringVar(&c.Username)

	// Optional.
	c.CmdClause.Flag("description", "An optional description for the alert.").Action(c.Description.Set).StringVar(&c.Description.Value)
	c.CmdClause.Flag("issue-type", "An optional Jira issue type associated with the ticket. (Default Task)").StringVar(&c.IssueType)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	// Call Parse() to ensure that we check if workspaceID
	// is set or to throw the appropriate error.
	if err := c.WorkspaceID.Parse(); err != nil {
		return err
	}
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &jira.CreateInput{
		WorkspaceID: &c.WorkspaceID.Value,
		Config: &jira.CreateConfig{
			Host:     &c.Host,
			Key:      &c.Key,
			Project:  &c.Project,
			Username: &c.Username,
		},
		// Set 'Events' to the only possible value, 'flag'
		Events: alert.GetDefaultEvents(),
	}
	if c.IssueType != "" {
		input.Config.IssueType = &c.IssueType
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := jira.Create(context.TODO(), fc, input)
	if err != nil {
		return err
	}

	if ok, err := c.WriteJSON(out, data); ok {
		return err
	}

	text.Success(out, "Created a '%s' alert '%s' (workspace-id: %s)", data.Type, data.ID, c.WorkspaceID.Value)
	return nil
}
