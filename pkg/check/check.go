package check

import "time"

// Stale validates if the given time is older than 24hrs.
func Stale(lastVersionCheck string) bool {
	if t, _ := time.Parse(time.RFC3339, lastVersionCheck); !t.Before(time.Now().Add(-24 * time.Hour)) {
		return false
	}
	return true
}
