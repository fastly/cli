package argparser_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v15/fastly"

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
		"empty with WasSet": {
			flagValue:   "",
			errExpected: true,
		},
		"omitted": {
			flagOmitted: true,
			wantVersion: 1, // Returns active version when flag not provided (falls back to latest if no active)
		},
		"specific version": {
			flagValue:   "2",
			wantVersion: 2,
		},
		"specific version not found": {
			flagValue:   "999",
			errExpected: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			sv := &argparser.OptionalServiceVersion{}

			if !c.flagOmitted {
				sv.OptionalString = argparser.OptionalString{
					Optional: argparser.Optional{WasSet: true},
					Value:    c.flagValue,
				}
			}

			v, err := sv.Parse("123", mock.API{
				GetVersionFn:        testutil.GetVersion,
				ListVersionsFn:      listVersions,
				GetServiceDetailsFn: getServiceDetails,
			})
			if err != nil {
				if c.errExpected {
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}
			if c.errExpected {
				t.Fatalf("expected error, have %v", v)
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
// Versions are returned in descending order by version number (highest first),
// matching the real Fastly API behavior.
// Version 4 (staged), Version 3 (editable), Version 2 (locked), Version 1 (active).
func listVersions(_ context.Context, i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(4),
			Staging:   fastly.ToPointer(true),
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-04T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(3),
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-03T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(2),
			Locked:    fastly.ToPointer(true),
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(1),
			Active:    fastly.ToPointer(true),
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
	}, nil
}

// getServiceDetails returns service details with active and latest version info.
func getServiceDetails(_ context.Context, i *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
	result := &fastly.ServiceDetail{
		ServiceID: fastly.ToPointer(i.ServiceID),
	}

	// Check if specific version is requested
	if i.Version != nil {
		result.Version = &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    i.Version,
			UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		}
		return result, nil
	}

	// Check filters
	for _, filter := range i.Filters {
		if filter.Key == "versions.active" && filter.Value {
			result.ActiveVersion = &fastly.Version{
				ServiceID: fastly.ToPointer(i.ServiceID),
				Number:    fastly.ToPointer(1),
				Active:    fastly.ToPointer(true),
				UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
			}
			return result, nil
		}
	}

	// Default: return both active and latest
	result.ActiveVersion = &fastly.Version{
		ServiceID: fastly.ToPointer(i.ServiceID),
		Number:    fastly.ToPointer(1),
		Active:    fastly.ToPointer(true),
		UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
	}
	result.Version = &fastly.Version{
		ServiceID: fastly.ToPointer(i.ServiceID),
		Number:    fastly.ToPointer(4),
		Staging:   fastly.ToPointer(true),
		UpdatedAt: testutil.MustParseTimeRFC3339("2000-01-04T01:00:00Z"),
	}
	return result, nil
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
				Active: fastly.ToPointer(false),
				Locked: fastly.ToPointer(false),
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
		"version state unknown with autoclone": {
			version: &fastly.Version{
				Number: fastly.ToPointer(1),
				Active: nil,
				Locked: nil,
			},
			wantVersion: 2,
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
			if c.errExpected {
				t.Fatalf("expected error, have %v", v)
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
		EnvVars       map[string]string
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
			EnvVars:       map[string]string{"FASTLY_SERVICE_ID": ""},
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
				GetServicesFn: func(ctx context.Context, _ *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service] {
					return fastly.NewPaginator[fastly.Service](ctx, &mock.HTTPClient{
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
				GetServicesFn: func(ctx context.Context, _ *fastly.GetServicesInput) *fastly.ListPaginator[fastly.Service] {
					return fastly.NewPaginator[fastly.Service](ctx, &mock.HTTPClient{
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
			EnvVars:   map[string]string{"FASTLY_SERVICE_ID": ""},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// Set environment variables for this test case
			for k, v := range c.EnvVars {
				t.Setenv(k, v)
			}
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

func TestContent(t *testing.T) {
	const expectedContent = "This is a test"
	const expectedPath = "fixtures/content_test.txt"
	for _, testcase := range []struct {
		name    string
		content string
	}{
		{
			name:    "regular string",
			content: expectedContent,
		},
		{
			name:    "path",
			content: expectedPath,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			content := argparser.Content(testcase.content)
			if content != expectedContent {
				t.Errorf("for test %s, wanted content %s, got %s", testcase.name, expectedContent, content)
			}
		})
	}
}

// cloneVersionResult returns a function which returns a specific cloned version.
func cloneVersionResult(version int) func(_ context.Context, i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return func(_ context.Context, i *fastly.CloneVersionInput) (*fastly.Version, error) {
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
