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

// WebhookConfigFlags contains webhook URL configrations (used by webhook, slack, etc...)
type WebhookConfigFlags struct {
	Webhook string
}
