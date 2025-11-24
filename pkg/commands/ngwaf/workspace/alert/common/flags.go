package common

import "github.com/fastly/cli/pkg/argparser"

// BaseAlertFlags contains flags that are common to all alert commands.
type BaseAlertFlags struct {
	WorkspaceID argparser.OptionalWorkspaceID
}

// AlertIDFlags contains flags for identifying a specific alert (used in update/delete/get).
type AlertIDFlags struct {
	AlertID string
}

// AlertDataFlags contains optional data fields for alerts (used in create/update).
type AlertDataFlags struct {
	Description argparser.OptionalString
}

// DatadogConfigFlags contains Datadog specific configuration flags.
type DatadogConfigFlags struct {
	Key  string
	Site string
}

// JiraCreateConfigFlags contains Jira create specific configuration flags.
type JiraCreateConfigFlags struct {
	Host      string
	IssueType string
	Key       string
	Project   string
	Username  string
}

// JiraUpdateConfigFlags contains Jira update specific configuration flags.
// The `IssueType` field is removed here, as it's not a required field for
// update operations.
type JiraUpdateConfigFlags struct {
	Host     string
	Key      string
	Project  string
	Username string
}

// WebhookConfigFlags contains webhook URL configrations (used by webhook, slack, etc...)
type WebhookConfigFlags struct {
	Webhook string
}
