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
