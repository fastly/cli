package errors_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/fastly/cli/pkg/api/undocumented"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v13/fastly"
)

func TestDeduce(t *testing.T) {
	var (
		re1             = errors.RemediationError{Inner: fmt.Errorf("foo")}
		re2             = errors.RemediationError{Inner: fmt.Errorf("bar"), Remediation: "Reticulate your splines."}
		http503         = &fastly.HTTPError{StatusCode: http.StatusInternalServerError}
		http401         = &fastly.HTTPError{StatusCode: http.StatusUnauthorized}
		http403         = &fastly.HTTPError{StatusCode: http.StatusForbidden}
		wrappedNotExist = fmt.Errorf("couldn't do the thing: %w", os.ErrNotExist)
		undoc401        = undocumented.NewError(fmt.Errorf("error response"), http.StatusUnauthorized)
		undoc403        = undocumented.NewError(fmt.Errorf("error response"), http.StatusForbidden)
		undoc500        = undocumented.NewError(fmt.Errorf("error response"), http.StatusInternalServerError)
		wrappedUndoc401 = fmt.Errorf("call failed: %w", undoc401)
	)

	for _, testcase := range []struct {
		name  string
		input error
		want  errors.RemediationError
	}{
		{
			name:  "RemediationError with no remediation",
			input: re1,
			want:  re1,
		},
		{
			name:  "RemediationError with remediation",
			input: re2,
			want:  re2,
		},
		{
			name:  "fastly.HTTPError 503",
			input: http503,
			want:  errors.RemediationError{Inner: errors.SimplifyFastlyError(*http503), Remediation: errors.BugRemediation},
		},
		{
			name:  "fastly.HTTPError 401",
			input: http401,
			want:  errors.RemediationError{Inner: errors.SimplifyFastlyError(*http401), Remediation: errors.AuthRemediation()},
		},
		{
			name:  "fastly.HTTPError 403",
			input: http403,
			want:  errors.RemediationError{Inner: errors.SimplifyFastlyError(*http403), Remediation: errors.ForbiddenRemediation()},
		},
		{
			name:  "undocumented APIError 401",
			input: undoc401,
			want:  errors.RemediationError{Inner: undoc401, Remediation: errors.AuthRemediation()},
		},
		{
			name:  "undocumented APIError 403",
			input: undoc403,
			want:  errors.RemediationError{Inner: undoc403, Remediation: errors.ForbiddenRemediation()},
		},
		{
			name:  "undocumented APIError 500",
			input: undoc500,
			want:  errors.RemediationError{Inner: undoc500, Remediation: errors.BugRemediation},
		},
		{
			name:  "wrapped undocumented APIError 401",
			input: wrappedUndoc401,
			want:  errors.RemediationError{Inner: wrappedUndoc401, Remediation: errors.AuthRemediation()},
		},
		{
			name:  "wrapped os.ErrNotExist",
			input: wrappedNotExist,
			want:  errors.RemediationError{Inner: wrappedNotExist, Remediation: errors.HostRemediation},
		},
		{
			name:  "temporary network error",
			input: isTemporary{fmt.Errorf("baz")},
			want:  errors.RemediationError{Inner: fmt.Errorf("baz"), Remediation: errors.NetworkRemediation},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			have := errors.Deduce(testcase.input)
			testutil.AssertString(t, testcase.want.Error(), have.Error())
			testutil.AssertString(t, testcase.want.Remediation, have.Remediation)
		})
	}
}

type isTemporary struct{ error }

func (isTemporary) Temporary() bool { return true }
