package text

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/datadog"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/jira"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/mailinglist"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/microsoftteams"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/opsgenie"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/pagerduty"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/slack"
	"github.com/fastly/go-fastly/v13/fastly/ngwaf/v1/workspaces/alerts/webhook"
)

// PrintAlert displays a single alert.
// Accepts any alert type (datadog, slack, webhook, etc.) via any.
func PrintAlert(out io.Writer, alert any) {
	var id, alertType, description, createdAt, createdBy string
	var config any

	// Extract common fields based on type
	switch a := alert.(type) {
	case *datadog.Alert:
		id = a.ID
		alertType = a.Type
		description = a.Description
		createdAt = a.CreatedAt
		createdBy = a.CreatedBy
		config = a.Config
	case *jira.Alert:
		id = a.ID
		alertType = a.Type
		description = a.Description
		createdAt = a.CreatedAt
		createdBy = a.CreatedBy
		config = a.Config
	case *mailinglist.Alert:
		id = a.ID
		alertType = a.Type
		description = a.Description
		createdAt = a.CreatedAt
		createdBy = a.CreatedBy
		config = a.Config
	case *microsoftteams.Alert:
		id = a.ID
		alertType = a.Type
		description = a.Description
		createdAt = a.CreatedAt
		createdBy = a.CreatedBy
		config = a.Config
	case *opsgenie.Alert:
		id = a.ID
		alertType = a.Type
		description = a.Description
		createdAt = a.CreatedAt
		createdBy = a.CreatedBy
		config = a.Config
	case *pagerduty.Alert:
		id = a.ID
		alertType = a.Type
		description = a.Description
		createdAt = a.CreatedAt
		createdBy = a.CreatedBy
		config = a.Config
	case *slack.Alert:
		id = a.ID
		alertType = a.Type
		description = a.Description
		createdAt = a.CreatedAt
		createdBy = a.CreatedBy
		config = a.Config
	case *webhook.Alert:
		id = a.ID
		alertType = a.Type
		description = a.Description
		createdAt = a.CreatedAt
		createdBy = a.CreatedBy
		config = a.Config
	default:
		fmt.Fprintf(out, "Unknown alert type\n")
		return
	}

	fmt.Fprintf(out, "ID: %s\n", id)
	fmt.Fprintf(out, "Type: %s\n", alertType)
	fmt.Fprintf(out, "Description: %s\n", description)
	fmt.Fprintf(out, "Created At: %s\n", createdAt)
	fmt.Fprintf(out, "Created By: %s\n", createdBy)
	printAlertConfig(out, alertType, config)
}

// printAlertConfig prints alert configuration based on type.
func printAlertConfig(out io.Writer, alertType string, config any) {
	fmt.Fprint(out, "Config:\n")
	switch alertType {
	case "datadog":
		printDatadogConfig(out, config)
	case "jira":
		printJiraConfig(out, config)
	case "mailinglist":
		printMailingListConfig(out, config)
	case "microsoftteams", "slack", "webhook":
		printWebhookConfig(out, config)
	case "opsgenie", "pagerduty":
		printKeyConfig(out, config)
	default:
		fmt.Fprintf(out, "  (unknown type: %s)\n", alertType)
	}
}

// printDatadogConfig prints Datadog-specific configuration.
func printDatadogConfig(out io.Writer, config any) {
	if cfg, ok := config.(datadog.ResponseConfig); ok {
		if cfg.Key != nil {
			fmt.Fprintf(out, "  Key: <redacted>\n")
		}
		if cfg.Site != nil {
			fmt.Fprintf(out, "  Site: %s\n", *cfg.Site)
		}
	}
}

// printWebhookConfig prints webhook-based configuration (slack, webhook, microsoftteams).
func printWebhookConfig(out io.Writer, config any) {
	var hasWebhook bool

	switch cfg := config.(type) {
	case slack.ResponseConfig:
		hasWebhook = cfg.Webhook != nil
	case webhook.ResponseConfig:
		hasWebhook = cfg.Webhook != nil
	case microsoftteams.ResponseConfig:
		hasWebhook = cfg.Webhook != nil
	}

	if hasWebhook {
		fmt.Fprintf(out, "  Webhook: <redacted>\n")
	}
}

// printJiraConfig prints Jira-specific configuration.
func printJiraConfig(out io.Writer, config any) {
	if cfg, ok := config.(jira.ResponseConfig); ok {
		if cfg.Host != nil {
			fmt.Fprintf(out, "  Host: %s\n", *cfg.Host)
		}
		if cfg.Username != nil {
			fmt.Fprintf(out, "  Username: %s\n", *cfg.Username)
		}
		if cfg.Project != nil {
			fmt.Fprintf(out, "  Project: %s\n", *cfg.Project)
		}
		if cfg.IssueType != nil {
			fmt.Fprintf(out, "  Issue Type: %s\n", *cfg.IssueType)
		}
		if cfg.Key != nil {
			fmt.Fprintf(out, "  Key: <redacted>\n")
		}
	}
}

// printMailingListConfig prints Mailing List-specific configuration.
func printMailingListConfig(out io.Writer, config any) {
	if cfg, ok := config.(mailinglist.ResponseConfig); ok {
		if cfg.Address != nil {
			fmt.Fprintf(out, "  Address: %s\n", *cfg.Address)
		}
	}
}

// printKeyConfig prints key-based configuration (opsgenie, pagerduty).
func printKeyConfig(out io.Writer, config any) {
	var hasKey bool

	switch cfg := config.(type) {
	case opsgenie.ResponseConfig:
		hasKey = cfg.Key != nil
	case pagerduty.ResponseConfig:
		hasKey = cfg.Key != nil
	}

	if hasKey {
		fmt.Fprintf(out, "  Key: <redacted>\n")
	}
}

// getConfigSummary returns a string summary of the alert config with sensitive fields redacted.
func getConfigSummary(alertType string, config any) string {
	switch alertType {
	case "datadog":
		if cfg, ok := config.(datadog.ResponseConfig); ok {
			parts := []string{}
			if cfg.Site != nil {
				parts = append(parts, fmt.Sprintf("Site: %s", *cfg.Site))
			}
			if cfg.Key != nil {
				parts = append(parts, "Key: <redacted>")
			}
			return strings.Join(parts, ", ")
		}
	case "jira":
		if cfg, ok := config.(jira.ResponseConfig); ok {
			parts := []string{}
			if cfg.Host != nil {
				parts = append(parts, fmt.Sprintf("Host: %s", *cfg.Host))
			}
			if cfg.IssueType != nil {
				parts = append(parts, fmt.Sprintf("Issue Type: %s", *cfg.IssueType))
			}
			if cfg.Key != nil {
				parts = append(parts, "Key: <redacted>")
			}
			if cfg.Project != nil {
				parts = append(parts, fmt.Sprintf("Project: %s", *cfg.Project))
			}
			if cfg.Username != nil {
				parts = append(parts, fmt.Sprintf("Username: %s", *cfg.Username))
			}
			return strings.Join(parts, ", ")
		}
	case "mailinglist":
		if cfg, ok := config.(mailinglist.ResponseConfig); ok {
			if cfg.Address != nil {
				return fmt.Sprintf("Address: %s", *cfg.Address)
			}
		}
	case "microsoftteams", "slack", "webhook":
		return "Webhook: <redacted>"
	case "opsgenie", "pagerduty":
		return "Key: <redacted>"
	}
	return ""
}

// PrintAlertTbl prints a table of alerts.
func PrintAlertTbl(out io.Writer, alerts any) {
	tbl := NewTable(out)
	tbl.AddHeader("ID", "Type", "Description", "Created At", "Created By", "Config")

	// Handle different alert type slices
	switch a := alerts.(type) {
	case []datadog.Alert:
		for _, alert := range a {
			configSummary := getConfigSummary(alert.Type, alert.Config)
			tbl.AddLine(alert.ID, alert.Type, alert.Description, alert.CreatedAt, alert.CreatedBy, configSummary)
		}
	case []slack.Alert:
		for _, alert := range a {
			configSummary := getConfigSummary(alert.Type, alert.Config)
			tbl.AddLine(alert.ID, alert.Type, alert.Description, alert.CreatedAt, alert.CreatedBy, configSummary)
		}
	case []webhook.Alert:
		for _, alert := range a {
			configSummary := getConfigSummary(alert.Type, alert.Config)
			tbl.AddLine(alert.ID, alert.Type, alert.Description, alert.CreatedAt, alert.CreatedBy, configSummary)
		}
	case []jira.Alert:
		for _, alert := range a {
			configSummary := getConfigSummary(alert.Type, alert.Config)
			tbl.AddLine(alert.ID, alert.Type, alert.Description, alert.CreatedAt, alert.CreatedBy, configSummary)
		}
	case []mailinglist.Alert:
		for _, alert := range a {
			configSummary := getConfigSummary(alert.Type, alert.Config)
			tbl.AddLine(alert.ID, alert.Type, alert.Description, alert.CreatedAt, alert.CreatedBy, configSummary)
		}
	case []microsoftteams.Alert:
		for _, alert := range a {
			configSummary := getConfigSummary(alert.Type, alert.Config)
			tbl.AddLine(alert.ID, alert.Type, alert.Description, alert.CreatedAt, alert.CreatedBy, configSummary)
		}
	case []opsgenie.Alert:
		for _, alert := range a {
			configSummary := getConfigSummary(alert.Type, alert.Config)
			tbl.AddLine(alert.ID, alert.Type, alert.Description, alert.CreatedAt, alert.CreatedBy, configSummary)
		}
	case []pagerduty.Alert:
		for _, alert := range a {
			configSummary := getConfigSummary(alert.Type, alert.Config)
			tbl.AddLine(alert.ID, alert.Type, alert.Description, alert.CreatedAt, alert.CreatedBy, configSummary)
		}
	}

	tbl.Print()
}
