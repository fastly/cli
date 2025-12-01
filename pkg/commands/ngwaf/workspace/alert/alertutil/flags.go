package alertutil

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

// JiraOptConfigFlags contains optional Jira specific configuration flags.
type JiraOptConfigFlags struct {
	IssueType string
}

// JiraConfigFlags contains Jira specific configuration flags.
type JiraConfigFlags struct {
	Host     string
	Key      string
	Project  string
	Username string
}

// AddressConfigFlags contains Address configurations used by mailing lists.
type AddressConfigFlags struct {
	Address string
}

// KeyConfigFlags contains the Key configuration (used by opsgenie, pagerduty, etc...).
type KeyConfigFlags struct {
	Key string
}

// WebhookConfigFlags contains the Webhook configuration (used by webhook, slack, etc...).
type WebhookConfigFlags struct {
	Webhook string
}
