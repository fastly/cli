package testutil

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/commands/whoami"
)

// Err represents a generic error.
var Err = errors.New("test error")

// ListVersions returns a list of service versions in different states.
//
// The first element is active, the second is locked, the third is editable.
//
// NOTE: consult the entire test suite before adding any new entries to the
// returned type as the tests currently use testutil.CloneVersionResult() as a
// way of making the test output and expectations as accurate as possible.
func ListVersions(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(1),
			Active:    fastly.ToPointer(true),
			UpdatedAt: MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(2),
			Active:    fastly.ToPointer(false),
			Locked:    fastly.ToPointer(true),
			UpdatedAt: MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(3),
			Active:    fastly.ToPointer(false),
			UpdatedAt: MustParseTimeRFC3339("2000-01-03T01:00:00Z"),
		},
	}, nil
}

// ListVersionsError returns a generic error message when attempting to list
// service versions.
func ListVersionsError(_ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return nil, Err
}

// CloneVersionResult returns a function which returns a specific cloned version.
func CloneVersionResult(version int) func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(version),
		}, nil
	}
}

// CloneVersionError returns a generic error message when attempting to clone a
// service version.
func CloneVersionError(_ *fastly.CloneVersionInput) (*fastly.Version, error) {
	return nil, Err
}

// WhoamiVerifyClient is used by `whoami` and `sso` tests.
type WhoamiVerifyClient whoami.VerifyResponse

// Do executes the HTTP request.
func (c WhoamiVerifyClient) Do(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	_ = json.NewEncoder(rec).Encode(whoami.VerifyResponse(c))
	return rec.Result(), nil
}

// WhoamiBasicResponse is used by `whoami` and `sso` tests.
var WhoamiBasicResponse = whoami.VerifyResponse{
	Customer: whoami.Customer{
		ID:   "abc",
		Name: "Computer Company",
	},
	User: whoami.User{
		ID:    "123",
		Name:  "Alice Programmer",
		Login: "alice@example.com",
	},
	Services: map[string]string{
		"1xxaa": "First service",
		"2baba": "Second service",
	},
	Token: whoami.Token{
		ID:        "abcdefg",
		Name:      "Token name",
		CreatedAt: "2019-01-01T12:00:00Z",
		// no ExpiresAt
		Scope: "global",
	},
}
