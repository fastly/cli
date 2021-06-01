package common

import (
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/google/go-cmp/cmp"
)

func TestGetLatestActiveVersion(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		inputVersions []*fastly.Version
		wantVersion   int
		wantError     string
	}{
		{
			name: "active",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "draft",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "no active",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 3, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-03T01:00:00Z")},
			},
			wantError: "error locating latest active service version",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// NOTE: this is a duplicate of the sorting algorithm in
			// common/command.go to make the test as realistic as possible
			sort.Slice(testcase.inputVersions, func(i, j int) bool {
				return testcase.inputVersions[i].Number > testcase.inputVersions[j].Number
			})

			v, err := getLatestActiveVersion(testcase.inputVersions)
			if err != nil {
				if testcase.wantError != "" {
					assertString(t, testcase.wantError, err.Error())
				} else {
					t.Errorf("unexpected error returned: %v", err)
				}
			} else if v.Number != testcase.wantVersion {
				t.Errorf("wanted version %d, got %d", testcase.wantVersion, v.Number)
			}
		})
	}
}

func TestGetLatestEditableVersion(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		inputVersions []*fastly.Version
		wantVersion   int
		wantError     string
	}{
		{
			name: "success",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "no editable",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, Locked: true, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: mustParseTimeRFC3339("2000-02-02T01:00:00Z")},
				{Number: 3, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-03-03T01:00:00Z")},
			},
			wantError: "error retrieving an editable service version: no editable version found",
		},
		{
			name: "editable should be ahead of last active",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantError: "error retrieving an editable service version: no editable version found",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// NOTE: this is a duplicate of the sorting algorithm in
			// common/command.go to make the test as realistic as possible
			sort.Slice(testcase.inputVersions, func(i, j int) bool {
				return testcase.inputVersions[i].Number > testcase.inputVersions[j].Number
			})

			v, err := getLatestEditableVersion(testcase.inputVersions)
			if err != nil {
				if testcase.wantError != "" {
					assertString(t, testcase.wantError, err.Error())
				} else {
					t.Errorf("unexpected error returned: %v", err)
				}
			} else if v.Number != testcase.wantVersion {
				t.Errorf("wanted version %d, got %d", testcase.wantVersion, v.Number)
			}
		})
	}
}

func TestGetSpecifiedVersion(t *testing.T) {
	for _, testcase := range []struct {
		name          string
		inputVersions []*fastly.Version
		wantVersion   int
		wantError     string
	}{
		{
			name: "success",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "no version available",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, Locked: true, UpdatedAt: mustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: mustParseTimeRFC3339("2000-02-02T01:00:00Z")},
				{Number: 3, Active: true, UpdatedAt: mustParseTimeRFC3339("2000-03-03T01:00:00Z")},
			},
			wantVersion: 4,
			wantError:   "error getting specified service version: 4",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// NOTE: this is a duplicate of the sorting algorithm in
			// common/command.go to make the test as realistic as possible
			sort.Slice(testcase.inputVersions, func(i, j int) bool {
				return testcase.inputVersions[i].Number > testcase.inputVersions[j].Number
			})

			v, err := getSpecifiedVersion(testcase.inputVersions, strconv.Itoa(testcase.wantVersion))
			if err != nil {
				if testcase.wantError != "" {
					assertString(t, testcase.wantError, err.Error())
				} else {
					t.Errorf("unexpected error returned: %v", err)
				}
			} else if v.Number != testcase.wantVersion {
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

// assertString fatals a test if the parameters aren't equal.
func assertString(t *testing.T, want, have string) {
	t.Helper()
	if want != have {
		t.Fatal(cmp.Diff(want, have))
	}
}
