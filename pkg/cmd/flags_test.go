package cmd

import (
	"sort"
	"strconv"
	"testing"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestParse(t *testing.T) {
	cases := map[string]struct {
		value       string
		omitted     bool // represents flag not provided
		want        int  // represent the service version number
		errExpected bool
	}{
		"latest": {
			value: "latest",
			want:  3,
		},
		"active": {
			value: "active",
			want:  1,
		},
		"editable": {
			value: "editable",
			want:  3,
		},
		"empty": {
			value: "",
			want:  3,
		},
		"omitted": {
			omitted: true,
			want:    3,
		},
		"specific version OK": {
			value: "3",
			want:  3,
		},
		"specific version ERR": {
			value:       "4",
			errExpected: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var sv *OptionalServiceVersion

			if c.omitted {
				sv = &OptionalServiceVersion{}
			} else {
				sv = &OptionalServiceVersion{
					OptionalString: OptionalString{
						Value: c.value,
					},
				}
			}

			api := mock.API{
				ListVersionsFn: listVersions,
			}

			v, err := sv.Parse("123", api)
			if err != nil {
				if c.errExpected {
					return
				} else {
					t.Fatalf("unexpected error: %v", err)
				}
			}
			if err == nil {
				if c.errExpected {
					t.Fatalf("expected error, have %v", v)
				}
			}

			want := c.want
			have := v.Number
			if have != want {
				t.Errorf("wanted %d, have %d", want, have)
			}
		})
	}
}

// listVersions returns a list of service versions in different states.
//
// The first element is active, the second is locked, the third is editable.
func listVersions(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.ServiceID,
			Number:    1,
			Active:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    2,
			Active:    false,
			Locked:    true,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    3,
			Active:    false,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z"),
		},
	}, nil
}

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
			wantError: "no active service version found",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// NOTE: this is a duplicate of the sorting algorithm in
			// cmd/command.go to make the test as realistic as possible
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
			// cmd/command.go to make the test as realistic as possible
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
			wantError:   "specified service version not found: 4",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// NOTE: this is a duplicate of the sorting algorithm in
			// cmd/command.go to make the test as realistic as possible
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
