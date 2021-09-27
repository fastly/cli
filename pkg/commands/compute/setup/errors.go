package setup

import (
	"fmt"

	"github.com/fastly/cli/pkg/errors"
)

const (
	RemediationErrFormat = "Check the fastly.toml configuration for a missing or invalid '%s' field, and consult https://developer.fastly.com/reference/fastly-toml/"
	InnerErrFormat       = "error parsing the [setup.%s] configuration"
)

// RemediationError reduces the boilerplate of serving a remediation error
// whose only difference is the field it applies to.
func RemediationError(field string, remediation string, err error) error {
	return errors.RemediationError{
		Inner:       err,
		Remediation: fmt.Sprintf(remediation, field),
	}
}
