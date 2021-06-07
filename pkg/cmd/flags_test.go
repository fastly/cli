package cmd

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/go-fastly/v3/fastly"
)

func TestOptionalServiceVersionParse(t *testing.T) {
	cases := map[string]struct {
		flagValue   string
		flagOmitted bool
		wantVersion int
		errExpected bool
	}{
		"latest": {
			flagValue:   "latest",
			wantVersion: 3,
		},
		"active": {
			flagValue:   "active",
			wantVersion: 1,
		},
		"empty": {
			flagValue:   "",
			wantVersion: 3,
		},
		"omitted": {
			flagOmitted: true,
			wantVersion: 3,
		},
		"specific version OK": {
			flagValue:   "2",
			wantVersion: 2,
		},
		"specific version ERR": {
			flagValue:   "4",
			errExpected: true, // there is no version 4
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sv := &OptionalServiceVersion{}

			if !c.flagOmitted {
				sv.OptionalString = OptionalString{
					Value: c.flagValue,
				}
			}

			v, err := sv.Parse("123", mock.API{
				ListVersionsFn: listVersions,
			})
			if err != nil {
				if c.errExpected {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil {
				if c.errExpected {
					t.Fatalf("expected error, have %v", v)
				}
			}

			want := c.wantVersion
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

			v, err := getActiveVersion(testcase.inputVersions)
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

func TestOptionalAutoCloneParse(t *testing.T) {
	cases := map[string]struct {
		version        *fastly.Version
		flagOmitted    bool
		wantVersion    int
		errExpected    bool
		expectEditable bool
	}{
		"version is editable": {
			version: &fastly.Version{
				Number: 1,
			},
			wantVersion:    1,
			expectEditable: true,
		},
		"version is locked": {
			version: &fastly.Version{
				Number: 1,
				Locked: true,
			},
			wantVersion: 2,
		},
		"version is active": {
			version: &fastly.Version{
				Number: 1,
				Active: true,
			},
			wantVersion: 2,
		},
		"version is locked but flag omitted": {
			version: &fastly.Version{
				Number: 1,
				Locked: true,
			},
			flagOmitted: true,
			errExpected: true,
		},
		"version is active but flag omitted": {
			version: &fastly.Version{
				Number: 1,
				Active: true,
			},
			flagOmitted: true,
			errExpected: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var (
				acv *OptionalAutoClone
				bs  []byte
			)
			buf := bytes.NewBuffer(bs)

			if c.flagOmitted {
				acv = &OptionalAutoClone{}
			} else {
				acv = &OptionalAutoClone{
					OptionalBool: OptionalBool{
						Value: true,
					},
				}
			}

			verboseMode := true
			v, err := acv.Parse(c.version, "123", verboseMode, buf, mock.API{
				CloneVersionFn: cloneVersionResult(c.version.Number + 1),
			})
			if err != nil {
				if c.errExpected && errMatches(c.version.Number, err) {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil {
				if c.errExpected {
					t.Fatalf("expected error, have %v", v)
				}
			}

			want := c.wantVersion
			have := v.Number
			if have != want {
				t.Errorf("wanted %d, have %d", want, have)
			}

			if !c.expectEditable {
				want := fmt.Sprintf("Service version %d is not editable, so it was automatically cloned because --autoclone is enabled. Now operating on version %d.", c.version.Number, v.Number)
				have := strings.Trim(strings.ReplaceAll(buf.String(), "\n", " "), " ")
				if !strings.Contains(have, want) {
					t.Errorf("wanted %s, have %s", want, have)
				}
			}
		})
	}
}

// cloneVersionResult returns a function which returns a specific cloned version.
func cloneVersionResult(version int) func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
		return &fastly.Version{
			ServiceID: i.ServiceID,
			Number:    version,
		}, nil
	}
}

// errMatches validates that the error message is what we expect when given a
// service version that is either locked or active, while also not providing
// the --autoclone flag.
func errMatches(version int, err error) bool {
	return err.Error() == fmt.Sprintf("service version %d is not editable", version)
}
