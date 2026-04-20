package errors

import (
	"fmt"
	"io"
	"strings"

	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/text"
)

// RemediationError wraps a normal error with a suggested remediation.
type RemediationError struct {
	// Prefix is a custom message displayed without modification.
	Prefix string
	// Inner is the root error.
	Inner error
	// Remediation provides more context and helpful references.
	Remediation string
}

// Unwrap returns the inner error.
func (re RemediationError) Unwrap() error {
	return re.Inner
}

// Error prints the inner error string without any remediation suggestion.
func (re RemediationError) Error() string {
	if re.Inner == nil {
		return ""
	}
	return re.Inner.Error()
}

// Print the error to the io.Writer for human consumption. If a prefix is
// provided, it will be written without modification. The inner error is always
// printed via text.Output with an "Error: " prefix and a "." suffix. If a
// remediation is provided, it's printed via text.Output.
func (re RemediationError) Print(w io.Writer) {
	if re.Prefix != "" {
		fmt.Fprintf(w, "%s\n\n", strings.TrimRight(re.Prefix, "\r\n"))
	}
	if re.Inner != nil {
		text.Error(w, "%s.\n\n", re.Inner.Error()) // single "\n" ensured by text.Error
	}
	if re.Remediation != "" {
		fmt.Fprintf(w, "%s\n", strings.TrimRight(re.Remediation, "\r\n"))
	}
}

// FormatTemplate represents a generic error message prefix.
var FormatTemplate = "To fix this error, run the following command:\n\n\t$ %s"

// AuthRemediation suggests checking the provided --token.
func AuthRemediation() string {
	var parts []string
	if env.AuthCommandDisabled() {
		parts = []string{
			"This error is likely caused by a missing, incorrect, or expired Fastly API token.",
			fmt.Sprintf("Token precedence: %s > fastly.toml profile > default auth token.", env.APIToken),
			fmt.Sprintf("Supply a token via %s.", env.APIToken),
		}
	} else {
		parts = []string{
			"This error is likely caused by a missing, incorrect, or expired Fastly API token.",
			fmt.Sprintf("Token precedence: --token (raw or stored name) > %s > fastly.toml profile > default auth token.", env.APIToken),
			fmt.Sprintf("Run `fastly auth login` to authenticate, or supply a token via --token or %s.", env.APIToken),
		}
	}
	parts = append(parts, "Learn more: fastly.help/cli/cli-auth")
	return strings.Join(parts, " ")
}

// ForbiddenRemediation suggests the token may lack required permissions.
func ForbiddenRemediation() string {
	parts := []string{
		"This error may indicate insufficient token permissions, an incorrect account context,",
		"or restricted access to the requested resource.",
		"Check that your token has the required scope for this operation.",
	}
	if env.AuthCommandDisabled() {
		parts = append(parts, fmt.Sprintf("Verify your token has the required scope via %s or the Fastly dashboard.", env.APIToken))
	} else {
		parts = append(parts, "You can re-authenticate with `fastly auth login` or check your current identity with `fastly whoami`.")
	}
	parts = append(parts, "Learn more: fastly.help/cli/cli-auth")
	return strings.Join(parts, " ")
}

// NetworkRemediation suggests, somewhat unhelpfully, to try again later.
var NetworkRemediation = strings.Join([]string{
	"This error may be caused by transient network issues.",
	"Please verify your network connection and DNS configuration, and try again.",
}, " ")

// HostRemediation suggests there might be an issue with the local host.
var HostRemediation = strings.Join([]string{
	"This error may be caused by a problem with your host environment, for example",
	"too-restrictive file permissions, files that already exist, or a full disk.",
}, " ")

// BugRemediation suggests filing a bug on the GitHub repo. It's good to include
// as the final suggested remediation in many errors.
var BugRemediation = strings.Join([]string{
	"If you believe this error is the result of a bug, please file an issue:",
	"https://github.com/fastly/cli/issues/new?labels=bug&template=bug_report.md",
}, " ")

// ConfigRemediation informs the user that an error with loading the config
// isn't a breaking error and the CLI can still be used.
var ConfigRemediation = strings.Join([]string{
	"There is a fallback version of the configuration provided with the CLI install",
	"(run `fastly config` to view the config) which enables the CLI to continue to be usable even though the config couldn't be updated.",
}, " ")

// ServiceIDRemediation suggests provide a service ID via --service-id flag or
// fastly.toml.
var ServiceIDRemediation = strings.Join([]string{
	"Please provide one via the --service-id or --service-name flag, or by setting the FASTLY_SERVICE_ID environment variable, or within your fastly.toml",
}, " ")

// CustomerIDRemediation suggests provide a customer ID via --customer-id flag
// or via environment variable.
var CustomerIDRemediation = strings.Join([]string{
	"Please provide one via the --customer-id flag, or by setting the FASTLY_CUSTOMER_ID environment variable",
}, " ")

// WorkspaceIDRemediation suggests provide a customer ID via --workspace-id flag
// or via environment variable.
var WorkspaceIDRemediation = strings.Join([]string{
	"Please provide one via the --workspace-id flag, or by setting the FASTLY_WORKSPACE_ID environment variable",
}, " ")

// ExistingDirRemediation suggests moving to another directory and retrying.
var ExistingDirRemediation = strings.Join([]string{
	"Please create a new directory and initialize a new project using:",
	"`fastly compute init`.",
}, " ")

// AutoCloneRemediation suggests provide an --autoclone flag.
var AutoCloneRemediation = strings.Join([]string{
	"Repeat the command with the --autoclone flag to allow the version to be cloned",
}, " ")

// IDRemediation suggests an ID via --id flag should be provided.
var IDRemediation = strings.Join([]string{
	"Please provide one via the --id flag",
}, " ")

// PackageSizeRemediation suggests checking the resources documentation for the
// current package size limit.
var PackageSizeRemediation = strings.Join([]string{
	"Please check our Compute resource limits:",
	"https://www.fastly.com/documentation/guides/compute#limitations-and-constraints",
}, " ")

// UnrecognisedManifestVersionRemediation suggests steps to resolve an issue
// where the project contains a manifest_version that is larger than what the
// current CLI version supports.
var UnrecognisedManifestVersionRemediation = strings.Join([]string{
	"Please try updating the installed CLI version using: `fastly update`.",
	"See also https://www.fastly.com/documentation/reference/compute/fastly-toml to check your fastly.toml manifest is up-to-date with the latest data model.",
	BugRemediation,
}, " ")

// ComputeInitRemediation suggests re-running `compute init` to resolve
// manifest issue.
var ComputeInitRemediation = strings.Join([]string{
	"Run `fastly compute init` to ensure a correctly configured manifest.",
	"See more at https://www.fastly.com/documentation/reference/compute/fastly-toml",
}, " ")

// ComputeServeRemediation suggests re-running `compute serve` with one of the
// incompatible flags removed.
var ComputeServeRemediation = strings.Join([]string{
	"The --watch flag enables hot reloading of your project to support a faster feedback loop during local development, and subsequently conflicts with the --skip-build flag which avoids rebuilding your project altogether.",
	"Remove one of the flags based on the outcome you require.",
}, " ")

// ComputeBuildRemediation suggests configuring a `[scripts.build]` setting in
// the fastly.toml manifest.
var ComputeBuildRemediation = strings.Join([]string{
	"Add a [scripts] section with `build = \"%s\"`.",
	"See more at https://www.fastly.com/documentation/reference/compute/fastly-toml",
}, " ")

// ProfileRemediation suggests running auth commands.
func ProfileRemediation() string {
	if env.AuthCommandDisabled() {
		return fmt.Sprintf("Supply a token via the %s environment variable.", env.APIToken)
	}
	return "Run `fastly auth login` to authenticate, or `fastly auth list` to view stored tokens."
}

// InvalidStaticConfigRemediation indicates an unexpected error occurred when
// deserialising the CLI's internal configuration.
var InvalidStaticConfigRemediation = strings.Join([]string{
	"The Fastly CLI attempted to parse an internal configuration file but failed.",
	"Run `fastly update` to upgrade your current CLI version.",
	"If this does not resolve the issue, then please file an issue:",
	"https://github.com/fastly/cli/issues/new?labels=bug&template=bug_report.md",
}, " ")

// TokenExpirationRemediation indicates that a stored OIDC token has expired.
func TokenExpirationRemediation() string {
	if env.AuthCommandDisabled() {
		return fmt.Sprintf("Supply a fresh token via the %s environment variable.", env.APIToken)
	}
	return "Run 'fastly auth login --sso --token <name>' to refresh the token."
}

// TokenExpirationRemediationForType returns remediation text appropriate for
// the given token type ("static", "sso", or "" for unknown/default).
//
// NOTE: tokenType values must match config.AuthTokenTypeStatic / config.AuthTokenTypeSSO.
// We cannot import pkg/config here (pkg/errors is a foundational package), so the
// string literals are used directly. Callers should pass config.AuthToken.Type.
func TokenExpirationRemediationForType(tokenType string) string {
	if env.AuthCommandDisabled() {
		return fmt.Sprintf("Supply a fresh token via the %s environment variable.", env.APIToken)
	}
	if tokenType == "static" {
		return "Generate a new token from the Fastly dashboard or run 'fastly auth add'."
	}
	return "Run 'fastly auth login --sso --token <name>' to refresh the token."
}

// NonInteractiveAuthRemediation tells the user how to supply a token when
// interactive prompts are suppressed.
func NonInteractiveAuthRemediation() string {
	parts := []string{"Interactive authentication is not available in this mode."}
	if env.AuthCommandDisabled() {
		parts = append(parts, fmt.Sprintf("Supply a token via the %s environment variable.", env.APIToken))
	} else {
		parts = append(parts, fmt.Sprintf("Supply a token via --token or the %s environment variable.", env.APIToken))
	}
	parts = append(parts, "Learn more: fastly.help/cli/cli-auth")
	return strings.Join(parts, " ")
}
