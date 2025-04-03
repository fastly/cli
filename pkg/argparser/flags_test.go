package argparser_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
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
			wantVersion: 4,
		},
		"active": {
			flagValue:   "active",
			wantVersion: 1,
		},
		// NOTE: Default behaviour for an empty flag value (or no flag at all) is to
		// get the active version, and if no active version return the latest.
		"empty": {
			flagValue:   "",
			wantVersion: 1,
		},
		"omitted": {
			flagOmitted: true,
			wantVersion: 1,
		},
		"specific version OK": {
			flagValue:   "2",
			wantVersion: 2,
		},
		"specific version ERR": {
			flagValue:   "5",
			errExpected: true, // there is no version 5
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sv := &argparser.OptionalServiceVersion{}

			if !c.flagOmitted {
				sv.OptionalString = argparser.OptionalString{
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
			have := fastly.ToValue(v.Number)
			if have != want {
				t.Errorf("wanted %d, have %d", want, have)
			}
		})
	}
}

// listVersions returns a list of service versions in different states.
//
// The first element is active, the second is locked, the third is
// editable, the fourth is staged.
func listVersions(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(1),
			Active:    fastly.ToPointer(true),
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(2),
			Locked:    fastly.ToPointer(true),
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(3),
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(4),
			Staging:   fastly.ToPointer(true),
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-04T01:00:00Z"),
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
				{Number: fastly.ToPointer(1), Active: fastly.ToPointer(false), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: fastly.ToPointer(2), Active: fastly.ToPointer(true), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 2,
		},
		{
			name: "draft",
			inputVersions: []*fastly.Version{
				{Number: fastly.ToPointer(1), Active: fastly.ToPointer(true), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: fastly.ToPointer(2), Active: fastly.ToPointer(false), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "locked",
			inputVersions: []*fastly.Version{
				{Number: fastly.ToPointer(1), Active: fastly.ToPointer(true), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: fastly.ToPointer(2), Active: fastly.ToPointer(false), Locked: fastly.ToPointer(true), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "no active",
			inputVersions: []*fastly.Version{
				{Number: fastly.ToPointer(1), Active: fastly.ToPointer(false), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: fastly.ToPointer(2), Active: fastly.ToPointer(false), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
				{Number: fastly.ToPointer(3), Active: fastly.ToPointer(false), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z")},
			},
			wantError: "no active service version found",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// NOTE: this is a duplicate of the sorting algorithm in
			// cmd/command.go to make the test as realistic as possible
			sort.Slice(testcase.inputVersions, func(i, j int) bool {
				return fastly.ToValue(testcase.inputVersions[i].Number) > fastly.ToValue(testcase.inputVersions[j].Number)
			})

			v, err := argparser.GetActiveVersion(testcase.inputVersions)
			if err != nil {
				if testcase.wantError != "" {
					testutil.AssertString(t, testcase.wantError, err.Error())
				} else {
					t.Errorf("unexpected error returned: %v", err)
				}
			} else if fastly.ToValue(v.Number) != testcase.wantVersion {
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
				{Number: fastly.ToPointer(1), Active: fastly.ToPointer(false), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: fastly.ToPointer(2), Active: fastly.ToPointer(true), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z")},
			},
			wantVersion: 1,
		},
		{
			name: "no version available",
			inputVersions: []*fastly.Version{
				{Number: fastly.ToPointer(1), Active: fastly.ToPointer(false), Locked: fastly.ToPointer(true), UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z")},
				{Number: fastly.ToPointer(2), Active: fastly.ToPointer(false), Locked: fastly.ToPointer(true), UpdatedAt: testutil.MustParseTimeRFC3339("2000-02-02T01:00:00Z")},
				{Number: fastly.ToPointer(3), Active: fastly.ToPointer(true), UpdatedAt: testutil.MustParseTimeRFC3339("2000-03-03T01:00:00Z")},
			},
			wantVersion: 4,
			wantError:   "specified service version not found: 4",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			// NOTE: this is a duplicate of the sorting algorithm in
			// cmd/command.go to make the test as realistic as possible
			sort.Slice(testcase.inputVersions, func(i, j int) bool {
				return fastly.ToValue(testcase.inputVersions[i].Number) > fastly.ToValue(testcase.inputVersions[j].Number)
			})

			v, err := argparser.GetSpecifiedVersion(testcase.inputVersions, strconv.Itoa(testcase.wantVersion))
			if err != nil {
				if testcase.wantError != "" {
					testutil.AssertString(t, testcase.wantError, err.Error())
				} else {
					t.Errorf("unexpected error returned: %v", err)
				}
			} else if fastly.ToValue(v.Number) != testcase.wantVersion {
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
				Number: fastly.ToPointer(1),
			},
			wantVersion:    1,
			expectEditable: true,
		},
		"version is locked": {
			version: &fastly.Version{
				Number: fastly.ToPointer(1),
				Locked: fastly.ToPointer(true),
			},
			wantVersion: 2,
		},
		"version is active": {
			version: &fastly.Version{
				Number: fastly.ToPointer(1),
				Active: fastly.ToPointer(true),
			},
			wantVersion: 2,
		},
		"version is locked but flag omitted": {
			version: &fastly.Version{
				Number: fastly.ToPointer(1),
				Locked: fastly.ToPointer(true),
			},
			flagOmitted: true,
			errExpected: true,
		},
		"version is active but flag omitted": {
			version: &fastly.Version{
				Number: fastly.ToPointer(1),
				Active: fastly.ToPointer(true),
			},
			flagOmitted: true,
			errExpected: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var (
				acv *argparser.OptionalAutoClone
				bs  []byte
			)
			buf := bytes.NewBuffer(bs)

			if c.flagOmitted {
				acv = &argparser.OptionalAutoClone{}
			} else {
				acv = &argparser.OptionalAutoClone{
					OptionalBool: argparser.OptionalBool{
						Value: true,
					},
				}
			}

			verboseMode := true
			v, err := acv.Parse(c.version, "123", verboseMode, buf, mock.API{
				CloneVersionFn: cloneVersionResult(fastly.ToValue(c.version.Number) + 1),
			})
			if err != nil {
				if c.errExpected && errMatches(fastly.ToValue(c.version.Number), err) {
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
			have := fastly.ToValue(v.Number)
			if have != want {
				t.Errorf("wanted %d, have %d", want, have)
			}

			if !c.expectEditable {
				want := fmt.Sprintf("Service version %d is not editable, so it was automatically cloned because --autoclone is enabled. Now operating on version %d.", fastly.ToValue(c.version.Number), fastly.ToValue(v.Number))
				have := strings.Trim(strings.ReplaceAll(buf.String(), "\n", " "), " ")
				if !strings.Contains(have, want) {
					t.Errorf("wanted %s, have %s", want, have)
				}
			}
		})
	}
}

func TestServiceID(t *testing.T) {
	cases := map[string]struct {
		ServiceName   argparser.OptionalServiceNameID
		Data          manifest.Data
		API           mock.API
		WantServiceID string
		WantError     string
		WantSource    manifest.Source
		WantFlag      string
	}{
		"service-id flag": {
			Data: manifest.Data{
				Flag: manifest.Flag{ServiceID: "456"},
			},
			WantServiceID: "456",
			WantSource:    manifest.SourceFlag,
			WantFlag:      argparser.FlagServiceIDName,
		},
		"service ID in manifest": {
			Data: manifest.Data{
				File: manifest.File{ServiceID: "456"},
			},
			WantServiceID: "456",
			WantSource:    manifest.SourceFile,
		},
		"service-name flag with service-id flag": {
			ServiceName: argparser.OptionalServiceNameID{argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bar"}},
			Data: manifest.Data{
				Flag: manifest.Flag{ServiceID: "123"},
			},
			WantError: "cannot specify both service-id and service-name",
		},
		"service-name flag with service-id in file": {
			ServiceName: argparser.OptionalServiceNameID{argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bar"}},
			Data: manifest.Data{
				File: manifest.File{ServiceID: "123"},
			},
			API: mock.API{
				GetServicesFn: func(i *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service] {
					return fastly.NewPaginator[fastly.Service](&mock.HTTPClient{
						Errors: []error{nil},
						Responses: []*http.Response{
							{
								Body: io.NopCloser(strings.NewReader(`[{"id": "456", "name": "bar"}]`)),
							},
						},
					}, fastly.ListOpts{}, "/example")
				},
			},
			WantServiceID: "456",
			WantSource:    manifest.SourceFlag,
			WantFlag:      argparser.FlagServiceName,
		},
		"unknown service-name flag value": {
			ServiceName: argparser.OptionalServiceNameID{argparser.OptionalString{Optional: argparser.Optional{WasSet: true}, Value: "bar"}},
			Data:        manifest.Data{},
			API: mock.API{
				GetServicesFn: func(i *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service] {
					return fastly.NewPaginator[fastly.Service](&mock.HTTPClient{
						Errors: []error{nil},
						Responses: []*http.Response{
							{
								Body: io.NopCloser(strings.NewReader(`[{"id": "456", "name": "beepboop"}]`)),
							},
						},
					}, fastly.ListOpts{}, "/example")
				},
			},
			WantError: "error matching service name with available services",
		},
		"no information provided": {
			Data:      manifest.Data{},
			WantError: "error reading service: no service ID found",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			serviceID, source, flag, err := argparser.ServiceID(c.ServiceName, c.Data, c.API, nil)
			testutil.AssertErrorContains(t, err, c.WantError)
			if err == nil {
				testutil.AssertString(t, serviceID, c.WantServiceID)
				testutil.AssertStringContains(t, flag, c.WantFlag)
				testutil.AssertEqual(t, source, c.WantSource)
			}
		})
	}
}

// cloneVersionResult returns a function which returns a specific cloned version.
func cloneVersionResult(version int) func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return func(i *fastly.CloneVersionInput) (*fastly.Version, error) {
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(version),
		}, nil
	}
}

// errMatches validates that the error message is what we expect when given a
// service version that is either locked or active, while also not providing
// the --autoclone flag.
func errMatches(version int, err error) bool {
	return err.Error() == fmt.Sprintf("service version %d is not editable", version)
}
