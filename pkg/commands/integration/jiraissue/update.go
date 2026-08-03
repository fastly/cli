package jiraissue

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a Jira Issue notification integration.
type UpdateCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	ID string

	// Optional.
	BaseURL         argparser.OptionalString
	Username        argparser.OptionalString
	APIToken        argparser.OptionalString
	ProjectKey      argparser.OptionalString
	IssueType       argparser.OptionalString
	IntegrationName argparser.OptionalString
	Description     argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Jira Issue notification integration")

	// Required.
	c.CmdClause.Arg("id", "Integration ID").Required().StringVar(&c.ID)

	// Optional.
	c.CmdClause.Flag("base-url", "The base URL of the Jira instance").Action(c.BaseURL.Set).StringVar(&c.BaseURL.Value)
	c.CmdClause.Flag("username", "The Jira username (email address) used to authenticate").Action(c.Username.Set).StringVar(&c.Username.Value)
	c.CmdClause.Flag("api-token", "The Jira API token").Action(c.APIToken.Set).StringVar(&c.APIToken.Value)
	c.CmdClause.Flag("project-key", "The key of the Jira project where issues will be created").Action(c.ProjectKey.Set).StringVar(&c.ProjectKey.Value)
	c.CmdClause.Flag("issue-type", "The type of Jira issue to create").Action(c.IssueType.Set).StringVar(&c.IssueType.Value)
	c.CmdClause.Flag("name", "The name of the integration").Short('n').Action(c.IntegrationName.Set).StringVar(&c.IntegrationName.Value)
	c.CmdClause.Flag("description", "A description of the integration").Action(c.Description.Set).StringVar(&c.Description.Value)
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	config := map[string]string{}
	if c.BaseURL.WasSet {
		config["baseurl"] = c.BaseURL.Value
	}
	if c.Username.WasSet {
		config["username"] = c.Username.Value
	}
	if c.APIToken.WasSet {
		config["token"] = c.APIToken.Value
	}
	if c.ProjectKey.WasSet {
		config["projectkey"] = c.ProjectKey.Value
	}
	if c.IssueType.WasSet {
		config["issuetype"] = c.IssueType.Value
	}

	input := &fastly.UpdateIntegrationInput{
		ID:   c.ID,
		Type: fastly.ToPointer(fastly.IntegrationTypeJiraIssue),
	}
	if len(config) > 0 {
		input.Config = config
	}
	if c.IntegrationName.WasSet {
		input.Name = &c.IntegrationName.Value
	}
	if c.Description.WasSet {
		input.Description = &c.Description.Value
	}

	if err := c.Globals.APIClient.UpdateIntegration(context.TODO(), input); err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if c.JSONOutput.Enabled {
		o := struct {
			ID      string `json:"id"`
			Updated bool   `json:"updated"`
		}{
			c.ID,
			true,
		}
		_, err := c.WriteJSON(out, o)
		return err
	}

	text.Success(out, "Updated Jira Issue integration (id: %s)", c.ID)
	return nil
}
