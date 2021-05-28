package common

import (
	"testing"
	"time"

	"github.com/fastly/go-fastly/v3/fastly"
)

func TestGetLatestNonActiveVersion(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		inputVersions []*fastly.Version
		wantVersion   int
	}{
		{
			name: "ignore active",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "draft",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "locked not favoured",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, Locked: true, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 3, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-03T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "no locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 3, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-03T01:00:00Z")},
			},
			wantVersion: 3,
		},
		{
			name: "not sorted",
			inputVersions: []*fastly.Version{
				{Number: 3, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-03T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 4, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-04T01:00:00Z")},
				{Number: 1, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
			},
			wantVersion: 1,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			v, err := getLatestNonActiveVersion(testcase.inputVersions)
			assertNoError(t, err)
			if v.Number != testcase.wantVersion {
				t.Errorf("wanted version %d, got %d", testcase.wantVersion, v.Number)
			}
		})
	}
}

// NOTE: The common package can't import the testutil package, as that imports
// the common package already. To avoid a cyclic import error we copy the
// implementation of the following functions from the testutil package.

// mustParseTimeRFC3339 is a small helper to initialize time constants.
func mustParseTimeRFC3339(s string) *time.Time {
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &tm
}

// assertNoError fatals a test if the error is not nil.
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
