package cmd

import (
	"sort"
	"strconv"
	"testing"

	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
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
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "draft",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "locked",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "no active",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: 3, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z")},
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
					testutil.AssertString(t, testcase.wantError, err.Error())
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
				{Number: 1, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "no editable",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, Locked: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-02-02T01:00:00Z")},
				{Number: 3, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-03-03T01:00:00Z")},
			},
			wantError: "error retrieving an editable service version: no editable version found",
		},
		{
			name: "editable should be ahead of last active",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
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
					testutil.AssertString(t, testcase.wantError, err.Error())
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
				{Number: 1, Active: false, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "no version available",
			inputVersions: []*fastly.Version{
				{Number: 1, Active: false, Locked: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: 2, Active: false, Locked: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-02-02T01:00:00Z")},
				{Number: 3, Active: true, UpdatedAt: testutil.MustParseTimeRFC3339("2000-03-03T01:00:00Z")},
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
					testutil.AssertString(t, testcase.wantError, err.Error())
				} else {
					t.Errorf("unexpected error returned: %v", err)
				}
			} else if v.Number != testcase.wantVersion {
				t.Errorf("wanted version %d, got %d", testcase.wantVersion, v.Number)
			}
		})
	}
}
