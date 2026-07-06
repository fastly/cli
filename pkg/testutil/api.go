package testutil

// Service Version Testing Guide
//
// This package provides standard mock functions for testing service version operations.
// The version mocks follow a consistent pattern across all tests:
//
// Version States:
//   - Version 1: ACTIVE (cannot modify without --autoclone)
//   - Version 2: LOCKED (cannot modify without --autoclone)
//   - Version 3: EDITABLE (can modify directly)
//   - Version 4: STAGING (can modify directly)
//
// Test Scenario Guide:
//
// 1. Testing --autoclone behavior:
//      Use: --version 1 or --version 2
//      Why: Tests that the CLI correctly clones active/locked versions
//      Example: "validate --autoclone on locked version"
//
// 2. Testing successful modifications (create/update/delete):
//      Use: --version 3
//      Why: Version is editable, so modification succeeds directly
//      Example: "validate backend creation"
//
// 3. Testing API errors:
//      Use: --version 3
//      Why: Avoids autoclone logic interfering with error testing
//      Example: "validate API error when creating backend"
//
// 4. Testing version activation:
//      Use: --version 3
//      Why: Version 1 is already active, so test would fail validation
//      Example: "validate version activation"
//
// 5. Testing without --version flag:
//      Use: No --version flag (defaults to active or latest)
//      Mock: ListVersionsFn only (GetVersionFn not needed)
//
// 6. Testing keyword versions (--version active/latest):
//      Use: --version active or --version latest
//      Mock: ListVersionsFn only (GetVersionFn not needed)
//
// Required Mock Functions:
//   - GetVersionFn: REQUIRED when using numeric --version flags (--version 1, --version 2, etc.)
//   - ListVersionsFn: REQUIRED when using keyword versions or no --version flag
//   - Both: REQUIRED when using --autoclone (needs GetVersion for initial parse, ListVersions for clone check)
//
// Example Test Patterns:
//
//	// Pattern 1: Testing --autoclone on locked version
//	{
//	    Args: "--service-id 123 --version 2 --name test --autoclone",
//	    API: &mock.API{
//	        GetVersionFn:   testutil.GetVersion,           // Parse --version 2
//	        ListVersionsFn: testutil.ListVersions,         // Check if locked
//	        CloneVersionFn: testutil.CloneVersionResult(4), // Clone to v4
//	        UpdateBackendFn: updateBackendOK,               // Modify v4
//	    },
//	    WantOutput: "Updated backend (service 123 version 4)",
//	}
//
//	// Pattern 2: Testing API error
//	{
//	    Args: "--service-id 123 --version 3 --name test",
//	    API: &mock.API{
//	        GetVersionFn:   testutil.GetVersion,  // Parse --version 3
//	        ListVersionsFn: testutil.ListVersions, // Check if editable
//	        UpdateBackendFn: updateBackendError,   // API fails
//	    },
//	    WantError: "API error",
//	}
//
//	// Pattern 3: Testing with keyword version
//	{
//	    Args: "--service-id 123 --version active --name test",
//	    API: &mock.API{
//	        ListVersionsFn: testutil.ListVersions, // Find active version
//	        GetBackendFn:   getBackendOK,
//	    },
//	    WantOutput: "Backend details",
//	}

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/fastly/go-fastly/v16/fastly"

	authcmd "github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/commands/whoami"
)

// Err represents a generic error.
var Err = errors.New("test error")

// ListVersions returns a list of service versions in different states.
//
// Versions are returned in descending order by version number (highest first),
// matching the real Fastly API behavior.
//
// Version states:
//   - Version 4: Staged (can be modified directly)
//   - Version 3: Editable (can be modified directly)
//   - Version 2: Locked (cannot be modified without --autoclone)
//   - Version 1: Active (cannot be modified without --autoclone)
//
// Usage guide for test scenarios:
//   - Testing --autoclone behavior: Use version 1 or 2 (active/locked)
//   - Testing successful modifications: Use version 3 (editable)
//   - Testing API errors: Use version 3 (so autoclone logic doesn't interfere)
//   - Testing version activation: Use version 3 or 4 (not already active)
//   - Testing staging/unstaging: Use version 4 (staged)
//
// NOTE: consult the entire test suite before adding any new entries to the
// returned type as the tests currently use testutil.CloneVersionResult() as a
// way of making the test output and expectations as accurate as possible.
//
// IMPORTANT: When using numeric --version flags in tests, you must include
// GetVersionFn: testutil.GetVersion in the mock API, as the CLI now calls
// GetVersion for numeric versions.
func ListVersions(_ context.Context, i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(4),
			Staging:   fastly.ToPointer(true),
			UpdatedAt: MustParseTimeRFC3339("2000-01-04T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(3),
			UpdatedAt: MustParseTimeRFC3339("2000-01-03T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(2),
			Locked:    fastly.ToPointer(true),
			UpdatedAt: MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
		{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(1),
			Active:    fastly.ToPointer(true),
			UpdatedAt: MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
	}, nil
}

// ListVersionsError returns a generic error message when attempting to list
// service versions.
func ListVersionsError(_ context.Context, _ *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return nil, Err
}

// GetVersion returns a version matching the requested version number.
//
// This function must be included in mock APIs when tests use numeric --version
// flags (e.g., --version 1, --version 2), as the CLI now calls GetVersion for
// numeric versions before processing them.
//
// Version states returned:
//   - Version 1: Active (Active=true)
//   - Version 2: Locked (Locked=true)
//   - Version 3: Editable (no Active/Locked flags)
//   - Version 4: Staging (Staging=true)
//   - Version 5: Generic editable version (commonly used after cloning version 4)
//   - Version 999: Returns an error (version not found - use this to test error handling)
//   - Other numbers: Returns a generic editable version
//
// This matches the versions returned by ListVersions for versions 1-4.
//
// Example test setup:
//
//	API: &mock.API{
//	    GetVersionFn:   testutil.GetVersion,    // Required for numeric --version flags
//	    ListVersionsFn: testutil.ListVersions,  // Required for keyword versions (active/latest)
//	    CloneVersionFn: testutil.CloneVersionResult(4),
//	}
func GetVersion(_ context.Context, i *fastly.GetVersionInput) (*fastly.Version, error) {
	switch i.ServiceVersion {
	case 1:
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(1),
			Active:    fastly.ToPointer(true),
			UpdatedAt: MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		}, nil
	case 2:
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(2),
			Locked:    fastly.ToPointer(true),
			UpdatedAt: MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		}, nil
	case 3:
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(3),
			UpdatedAt: MustParseTimeRFC3339("2000-01-03T01:00:00Z"),
		}, nil
	case 4:
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(4),
			Staging:   fastly.ToPointer(true),
			UpdatedAt: MustParseTimeRFC3339("2000-01-04T01:00:00Z"),
		}, nil
	case 5:
		// Version 5 is commonly used in tests after cloning version 4
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(5),
			UpdatedAt: MustParseTimeRFC3339("2000-01-05T01:00:00Z"),
		}, nil
	case 999:
		// Return an error for test cases that explicitly want to test version not found
		return nil, Err
	default:
		// Return a generic version for any other number to avoid breaking existing tests
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(i.ServiceVersion),
			UpdatedAt: MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		}, nil
	}
}

// CloneVersionResult returns a function which returns a specific cloned version.
func CloneVersionResult(version int) func(_ context.Context, i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return func(_ context.Context, i *fastly.CloneVersionInput) (*fastly.Version, error) {
		return &fastly.Version{
			ServiceID: fastly.ToPointer(i.ServiceID),
			Number:    fastly.ToPointer(version),
		}, nil
	}
}

// CloneVersionError returns a generic error message when attempting to clone a
// service version.
func CloneVersionError(_ context.Context, _ *fastly.CloneVersionInput) (*fastly.Version, error) {
	return nil, Err
}

// WhoamiVerifyClient is used by `whoami` and auth tests.
type WhoamiVerifyClient whoami.VerifyResponse

// Do executes the HTTP request.
func (c WhoamiVerifyClient) Do(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	_ = json.NewEncoder(rec).Encode(whoami.VerifyResponse(c))
	return rec.Result(), nil
}

// WhoamiBasicResponse is used by `whoami` and auth tests.
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

// CurrentCustomerClient is used by SSO auth tests.
type CurrentCustomerClient authcmd.CurrentCustomerResponse

// Do executes the HTTP request.
func (c CurrentCustomerClient) Do(*http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	_ = json.NewEncoder(rec).Encode(authcmd.CurrentCustomerResponse(c))
	return rec.Result(), nil
}

// CurrentCustomerResponse is used by SSO auth tests.
var CurrentCustomerResponse = authcmd.CurrentCustomerResponse{
	ID:   "abc",
	Name: "Computer Company",
}

// GetServiceDetails returns service details with versions matching the filter.
//
// This function must be included in mock APIs when the CLI calls GetServiceDetails,
// which happens when using the --version active flag.
//
// Filters supported:
//   - versions.active: Returns version 1 (active)
//
// This matches the version states returned by ListVersions and GetVersion.
//
// Example test setup:
//
//	API: &mock.API{
//	    GetServiceDetailsFn: testutil.GetServiceDetails,  // Required for --version active
//	    GetVersionFn:        testutil.GetVersion,         // Required for numeric --version flags
//	    ListVersionsFn:      testutil.ListVersions,       // Required for --version latest or omitted flag
//	}
func GetServiceDetails(_ context.Context, i *fastly.GetServiceDetailsInput) (*fastly.ServiceDetail, error) {
	detail := &fastly.ServiceDetail{
		ServiceID: fastly.ToPointer(i.ServiceID),
		Versions:  []*fastly.Version{},
	}

	// Check filters to determine which version to return
	for _, filter := range i.Filters {
		if filter.Key == "versions.active" && filter.Value {
			version := &fastly.Version{
				ServiceID: fastly.ToPointer(i.ServiceID),
				Number:    fastly.ToPointer(1),
				Active:    fastly.ToPointer(true),
				UpdatedAt: MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
			}
			detail.ActiveVersion = version
			detail.Version = version
			detail.Versions = append(detail.Versions, version)
		}
		if filter.Key == "versions.staged" && filter.Value {
			version := &fastly.Version{
				ServiceID: fastly.ToPointer(i.ServiceID),
				Number:    fastly.ToPointer(4),
				Staging:   fastly.ToPointer(true),
				UpdatedAt: MustParseTimeRFC3339("2000-01-04T01:00:00Z"),
			}
			detail.Version = version
			detail.Versions = append(detail.Versions, version)
		}
	}

	return detail, nil
}
