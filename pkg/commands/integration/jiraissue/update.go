package jiraissue

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update a Jira Issue notification integration.
type UpdateCommand struct {
	argparser.Base

	// Required.
	ID string `arg:"" required:"" help:"Integration ID"`

	// Optional.
	BaseURL         argparser.OptionalString `help:"The base URL of the Jira instance"`
	Username        argparser.OptionalString `help:"The Jira username (email address) used to authenticate"`
	APIToken        argparser.OptionalString `help:"The Jira API token"`
	ProjectKey      argparser.OptionalString `help:"The key of the Jira project where issues will be created"`
	IssueType       argparser.OptionalString `help:"The type of Jira issue to create"`
	IntegrationName argparser.OptionalString `name:"name" short:"n" help:"The name of the integration"`
	Description     argparser.OptionalString `help:"A description of the integration"`

	argparser.JSONOutput
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
		Type: new(fastly.IntegrationTypeJiraIssue),
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
