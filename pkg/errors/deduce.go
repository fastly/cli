package errors

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fastly/go-fastly/v3/fastly"
)

// Deduce attempts to deduce a RemediationError from a plain error. If the error
// is already a RemediationError it is returned directly. Certain deep error
// types, like a Fastly SDK HTTPError, are detected and converted in appropriate
// cases to e.g. AuthRemediation. If no specific remediation can be suggested, a
// remediation to file a bug is used.
func Deduce(err error) RemediationError {
	var re RemediationError
	if errors.As(err, &re) {
		return re // assume the useful suggestion is already baked-in
	}

	var httpError *fastly.HTTPError
	if errors.As(err, &httpError) {
		var remediation string
		switch httpError.StatusCode {
		case http.StatusUnauthorized:
			remediation = AuthRemediation
		}
		return RemediationError{Inner: SimplifyFastlyError(*httpError), Remediation: remediation}
	}

	if errors.Is(err, os.ErrNotExist) {
		return RemediationError{Inner: err, Remediation: HostRemediation}
	}

	if t, ok := err.(interface{ Temporary() bool }); ok && t.Temporary() {
		return RemediationError{Inner: err, Remediation: NetworkRemediation}
	}

	return RemediationError{Inner: err, Remediation: BugRemediation}
}

// SimplifyFastlyError reduces the potentially complex and multi-line Error
// rendering of a fastly.HTTPError to something more palatable for a CLI.
func SimplifyFastlyError(httpError fastly.HTTPError) error {
	switch len(httpError.Errors) {
	case 1:
		s := fmt.Sprintf(
			"the Fastly API returned %d %s: %s",
			httpError.StatusCode,
			http.StatusText(httpError.StatusCode),
			strings.TrimSpace(httpError.Errors[0].Title),
		)
		if detail := httpError.Errors[0].Detail; detail != "" {
			s += fmt.Sprintf(" (%s)", detail)
		}
		return fmt.Errorf(s)

	default:
		return fmt.Errorf(
			"the Fastly API returned %d %s",
			httpError.StatusCode,
			http.StatusText(httpError.StatusCode),
		)
	}
}
