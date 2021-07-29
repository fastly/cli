package errors

import "fmt"

// ErrNoToken means no --token has been provided.
var ErrNoToken = RemediationError{Inner: fmt.Errorf("no token provided"), Remediation: AuthRemediation}

// ErrNoServiceID means no --service-id or service_id package manifest value has
// been provided.
var ErrNoServiceID = RemediationError{Inner: fmt.Errorf("error reading service: no service ID found"), Remediation: ServiceIDRemediation}

// ErrMissingManifestVersion means an invalid manifest (fastly.toml) has been used.
var ErrMissingManifestVersion = RemediationError{Inner: fmt.Errorf("no manifest_version found in the fastly.toml"), Remediation: BugRemediation}

// ErrUnrecognisedManifestVersion means an invalid manifest (fastly.toml)
// version has been specified.
var ErrUnrecognisedManifestVersion = RemediationError{Inner: fmt.Errorf("unrecognised manifest_version found in the fastly.toml"), Remediation: BugRemediation}

// ErrInvalidManifestVersion means the manifest_version is defined as a toml
// section.
var ErrInvalidManifestVersion = RemediationError{Inner: fmt.Errorf("failed to parse fastly.toml when checking if manifest_version was valid"), Remediation: "Delete `[manifest_version]` from the fastly.toml if present"}

// ErrSignalInterrupt means a SIGINT was received.
var ErrSignalInterrupt = fmt.Errorf("a SIGINT was received")

// ErrSignalKilled means a SIGTERM was received.
var ErrSignalKilled = fmt.Errorf("a SIGTERM was received")

// ErrNoID means no --id value has been provided.
var ErrNoID = RemediationError{Inner: fmt.Errorf("no ID found"), Remediation: IDRemediation}
