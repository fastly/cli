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
	Prefix      string
	Inner       error
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
		text.Error(w, "%s.", re.Inner.Error()) // single "\n" ensured by text.Error
	}
	if re.Inner != nil && re.Remediation != "" {
		fmt.Fprintln(w) // additional "\n" to allow breathing room
	}
	if re.Remediation != "" {
		fmt.Fprintf(w, "%s\n", strings.TrimRight(re.Remediation, "\r\n"))
	}
}

// AuthRemediation suggests checking the provided --token.
var AuthRemediation = fmt.Sprintf(strings.Join([]string{
	"This error may be caused by a missing, incorrect, or expired Fastly API token.",
	"Check that you're supplying a valid token, either via --token,",
	"through the environment variable %s, or through the config file via `fastly configure`.",
	"Verify that the token is still valid via `fastly whoami`.",
}, " "), env.Token)

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

// ServiceIDRemediation suggests provide a service ID via --service-id flag or
// package manifest.
var ServiceIDRemediation = strings.Join([]string{
	"Please provide one via the --service-id flag, or by setting the FASTLY_SERVICE_ID environment variable, or within your package manifest",
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
	"Please check our Compute@Edge resource limits:",
	"https://developer.fastly.com/learning/compute/#limitations-and-constraints",
}, " ")

// CLIUpdateRemediation suggests updating the installed CLI version.
var CLIUpdateRemediation = strings.Join([]string{
	"Please try updating the installed CLI version using:",
	"`fastly update`.",
	BugRemediation,
}, " ")

// ComputeInitRemediation suggests re-running `compute init` to resolve
// manifest issue.
var ComputeInitRemediation = strings.Join([]string{
	"Run `fastly compute init` to ensure a correctly configured manifest.",
	"See more at https://developer.fastly.com/reference/fastly-toml/",
}, " ")
