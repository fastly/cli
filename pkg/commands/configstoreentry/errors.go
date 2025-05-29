package configstoreentry

import (
	"errors"
	"fmt"

	fsterr "github.com/fastly/cli/pkg/errors"
)

const (
	maxKeyLen = 256
	// maxValueLen is the maximum length of Config Store entry's value. It's set to 64k,
	// even though customers may have a smaller limit. The API will reject requests if the
	// value is larger than the customer's limit.
	maxValueLen = 2 << 15
)

var errNoSTDINData = fsterr.RemediationError{
	Inner:       errors.New("unable to read from STDIN"),
	Remediation: "Provide data to STDIN, or use --value to specify item value",
}

var errNoValue = fsterr.RemediationError{
	Inner:       errors.New("no value provided"),
	Remediation: "Use --value or --stdin to specify item value",
}

var errMaxKeyLen = fsterr.RemediationError{
	Inner:       errors.New("key max length"),
	Remediation: fmt.Sprintf("Key must be less than or equal to %d bytes", maxKeyLen),
}

var errMaxValueLen = fsterr.RemediationError{
	Inner:       errors.New("value max length"),
	Remediation: fmt.Sprintf("Value must be less than or equal to %d bytes", maxValueLen),
}
