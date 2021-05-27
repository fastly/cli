package testutil

import (
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListVersionsOk returns a list of service versions in different states.
//
// The first element is active, the second is locked, the third is editable.
func ListVersionsOk(i *fastly.ListVersionsInput) ([]*fastly.Version, error) {
	return []*fastly.Version{
		{
			ServiceID: i.ServiceID,
			Number:    1,
			Active:    true,
			UpdatedAt: MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    2,
			Active:    false,
			Locked:    true,
			UpdatedAt: MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
		{
			ServiceID: i.ServiceID,
			Number:    3,
			Active:    false,
			UpdatedAt: MustParseTimeRFC3339("2000-01-02T01:00:00Z"),
		},
	}, nil
}

// GetActiveVersionOK returns an active service version (Number: 1).
func GetActiveVersionOK(i *fastly.GetVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		ServiceID: i.ServiceID,
		Number:    1,
		Active:    true,
		UpdatedAt: MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
	}, nil
}

// GetInactiveVersionOK returns an inactive service version (Number: 1).
func GetInactiveVersionOK(i *fastly.GetVersionInput) (*fastly.Version, error) {
	return &fastly.Version{
		ServiceID: i.ServiceID,
		Number:    1,
		Active:    false,
		UpdatedAt: MustParseTimeRFC3339("2000-01-01T01:00:00Z"),
	}, nil
}

// CloneVersionOK returns an incremented service version.
func CloneVersionOK(i *fastly.CloneVersionInput) (*fastly.Version, error) {
	return &fastly.Version{ServiceID: i.ServiceID, Number: i.ServiceVersion + 1}, nil
}
