package errors

import "fmt"

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

// ErrNoServiceID means no --service-id or service_id package manifest value has
// been provided.
var ErrNoServiceID = RemediationError{
	Inner:       fmt.Errorf("error reading service: no service ID found"),
	Remediation: ServiceIDRemediation,
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
	Remediation: CLIUpdateRemediation,
}

// ErrInvalidManifestVersion means the manifest_version is defined as a toml
// section.
var ErrInvalidManifestVersion = RemediationError{
	Inner:       fmt.Errorf("failed to parse fastly.toml when checking if manifest_version was valid"),
	Remediation: "Delete `[manifest_version]` from the fastly.toml if present",
}

// ErrIncompatibleManifestVersion means the manifest_version defined is no
// longer compatible with the current CLI version.
var ErrIncompatibleManifestVersion = RemediationError{
	Inner:       fmt.Errorf("the fastly.toml contains an incompatible manifest_version number"),
	Remediation: "Update the `manifest_version` in the fastly.toml and refer to https://github.com/fastly/cli/releases/tag/v0.39.3 for changes to the manifest structure",
}

// ErrNoID means no --id value has been provided.
var ErrNoID = RemediationError{
	Inner:       fmt.Errorf("no ID found"),
	Remediation: IDRemediation,
}

// ErrReadingManifest means there was a problem reading the package manifest.
var ErrReadingManifest = RemediationError{
	Inner:       fmt.Errorf("error reading package manifest"),
	Remediation: ComputeInitRemediation,
}

// ErrParsingManifest means there was a problem unmarshalling the package manifest.
var ErrParsingManifest = RemediationError{
	Inner:       fmt.Errorf("error parsing package manifest"),
	Remediation: ComputeInitRemediation,
}
