package testutil

import (
	"time"
)

// MustParseTimeRFC3339 is a small helper to initialize time constants.
func MustParseTimeRFC3339(s string) *time.Time {
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return &tm
}
