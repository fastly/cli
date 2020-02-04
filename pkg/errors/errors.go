package errors

import "fmt"

// ErrNoToken means no --token has been provided.
var ErrNoToken = RemediationError{Inner: fmt.Errorf("no token provided"), Remediation: AuthRemediation}

// ErrNoServiceID means no --service-id or service_id package manifest value has
// been provided.
var ErrNoServiceID = RemediationError{Inner: fmt.Errorf("error reading service: no service ID found"), Remediation: ServiceIDRemediation}
