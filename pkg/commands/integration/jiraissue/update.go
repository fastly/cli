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
	ID         string
	BaseURL    string
	Username   string
	APIToken   string
	ProjectKey string
	IssueType  string

	// Optional.
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
	c.RegisterFlag(argparser.IntegrationIDFlag(&c.ID))
	c.CmdClause.Flag("base-url", "The base URL of the Jira instance").Required().StringVar(&c.BaseURL)
	c.CmdClause.Flag("username", "The Jira username (email address) used to authenticate").Required().StringVar(&c.Username)
	c.CmdClause.Flag("api-token", "The Jira API token").Required().StringVar(&c.APIToken)
	c.CmdClause.Flag("project-key", "The key of the Jira project where issues will be created").Required().StringVar(&c.ProjectKey)
	c.CmdClause.Flag("issue-type", "The type of Jira issue to create").Required().StringVar(&c.IssueType)

	// Optional.
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

	config := fastly.JiraIssueConfig{
		BaseURL:    c.BaseURL,
		Username:   c.Username,
		Token:      c.APIToken,
		ProjectKey: c.ProjectKey,
		IssueType:  c.IssueType,
	}

	input := &fastly.UpdateIntegrationInput{
		ID:     c.ID,
		Type:   fastly.ToPointer(fastly.IntegrationTypeJiraIssue),
		Config: config.ToMap(),
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
