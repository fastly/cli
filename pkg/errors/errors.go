package errors

import (
	"errors"
	"fmt"
)

// ErrSignalInterrupt means a SIGINT was received.
var ErrSignalInterrupt = fmt.Errorf("a SIGINT was received")

// ErrSignalKilled means a SIGTERM was received.
var ErrSignalKilled = fmt.Errorf("a SIGTERM was received")

// ErrViceroyRestart means the viceroy binary needs to be restarted due to a
// file modification noticed while running `compute serve --watch`.
var ErrViceroyRestart = fmt.Errorf("a RESTART was initiated")

// ErrIncompatibleServeFlags means no --skip-build can't be used with --watch
// because it defeats the purpose of --watch which is designed to restart
// Viceroy whenever changes are detected (those changes would not be seen if we
// allowed --skip-build with --watch).
var ErrIncompatibleServeFlags = RemediationError{
	Inner:       fmt.Errorf("--skip-build shouldn't be used with --watch"),
	Remediation: ComputeServeRemediation,
}

// ErrNoToken means no --token has been provided.
var ErrNoToken = RemediationError{
	Inner:       fmt.Errorf("no token provided"),
	Remediation: AuthRemediation,
}

// ErrNoServiceID means no --service-id or service_id fastly.toml value has
// been provided.
var ErrNoServiceID = RemediationError{
	Inner:       fmt.Errorf("error reading service: no service ID found"),
	Remediation: ServiceIDRemediation,
}

// ErrNoCustomerID means no --customer-id or FASTLY_CUSTOMER_ID environment
// variable found.
var ErrNoCustomerID = RemediationError{
	Inner:       fmt.Errorf("error reading customer ID: no customer ID found"),
	Remediation: CustomerIDRemediation,
}

// ErrMissingManifestVersion means an invalid manifest (fastly.toml) has been used.
var ErrMissingManifestVersion = RemediationError{
	Inner:       fmt.Errorf("no manifest_version found in the fastly.toml"),
	Remediation: BugRemediation,
}

// ErrUnrecognisedManifestVersion means an invalid manifest (fastly.toml)
// version has been specified.
var ErrUnrecognisedManifestVersion = RemediationError{
	Inner:       fmt.Errorf("unrecognised manifest_version found in the fastly.toml"),
	Remediation: UnrecognisedManifestVersionRemediation,
}

// ErrIncompatibleManifestVersion means the manifest_version defined is no
// longer compatible with the current CLI version.
var ErrIncompatibleManifestVersion = RemediationError{
	Inner:       fmt.Errorf("the fastly.toml contains an incompatible manifest_version number"),
	Remediation: "Update the `manifest_version` in the fastly.toml and refer to https://developer.fastly.com/reference/compute/fastly-toml/ for changes to the manifest structure",
}

// ErrNoID means no --id value has been provided.
var ErrNoID = RemediationError{
	Inner:       fmt.Errorf("no ID found"),
	Remediation: IDRemediation,
}

// ErrReadingManifest means there was a problem reading the fastly.toml.
var ErrReadingManifest = RemediationError{
	Inner:       fmt.Errorf("error reading fastly.toml"),
	Remediation: "Ensure the Fastly CLI is being run within a directory containing a fastly.toml file. " + ComputeInitRemediation,
}

// ErrParsingManifest means there was a problem unmarshalling the fastly.toml.
var ErrParsingManifest = RemediationError{
	Inner:       fmt.Errorf("error parsing fastly.toml"),
	Remediation: ComputeInitRemediation,
}

// ErrStopWalk is used to indicate to filepath.WalkDir that it should stop
// walking the directory tree.
var ErrStopWalk = errors.New("stop directory walking")

// ErrInvalidArchive means the package archive didn't contain a recognised
// directory structure.
var ErrInvalidArchive = RemediationError{
	Inner:       fmt.Errorf("invalid package archive structure"),
	Remediation: "Ensure the archive contains all required package files (such as a 'fastly.toml' manifest, and a 'src' folder etc).",
}

// ErrBuildStopped means the user stopped the build because they were unhappy
// with the custom build defined in the fastly.toml manifest file.
var ErrBuildStopped = RemediationError{
	Inner:       fmt.Errorf("build process stopped by user"),
	Remediation: "Check the [scripts.build] in the fastly.toml manifest is safe to execute or skip this prompt using either `--auto-yes` or `--non-interactive`.",
}

// ErrInvalidVerboseJSONCombo means the user provided both a --verbose and
// --json flag which are mutually exclusive behaviours.
var ErrInvalidVerboseJSONCombo = RemediationError{
	Inner:       fmt.Errorf("invalid flag combination, --verbose and --json"),
	Remediation: "Use either --verbose or --json, not both.",
}

// ErrInvalidDeleteAllJSONKeyCombo means the user provided both a --all and
// --json flag which are mutually exclusive behaviours.
var ErrInvalidDeleteAllJSONKeyCombo = RemediationError{
	Inner:       fmt.Errorf("invalid flag combination, --all and --json"),
	Remediation: "Use either --all or --json, not both.",
}

// ErrInvalidDeleteAllKeyCombo means the user provided both a --all and --key
// flag which are mutually exclusive behaviours.
var ErrInvalidDeleteAllKeyCombo = RemediationError{
	Inner:       fmt.Errorf("invalid flag combination, --all and --key"),
	Remediation: "Use either --all or --key, not both.",
}

// ErrMissingDeleteAllKeyCombo means the user omitted both the --all and --key
// flags and we need at least one of them.
var ErrMissingDeleteAllKeyCombo = RemediationError{
	Inner:       fmt.Errorf("invalid command, neither --all or --key provided"),
	Remediation: "Provide at least one of: --all or --key, not both.",
}

// ErrNoSTDINData indicates the --stdin flag was specified but no data was piped
// into stdin.
var ErrNoSTDINData = RemediationError{
	Inner:       fmt.Errorf("unable to read from STDIN"),
	Remediation: "Provide data to STDIN, or use --file to read from a file",
}

// ErrInvalidKVCombo means the user omitted either the key or value flag.
var ErrInvalidKVCombo = RemediationError{
	Inner:       fmt.Errorf("--key-name and --value are required"),
	Remediation: "Please add both flags or alternatively use either --stdin or --file.",
}

// ErrInvalidStdinFileDirCombo means the user provided more than one of --stdin,
// --file or --dir flags, which are mutally exclusive behaviours.
var ErrInvalidStdinFileDirCombo = RemediationError{
	Inner:       fmt.Errorf("invalid flag combination"),
	Remediation: "Use only one of --stdin, --file or --dir.",
}
