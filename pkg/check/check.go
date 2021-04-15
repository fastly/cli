package check

import (
	"time"
)

// Stale validates if the given time is older than the given duration.
//
// EXAMPLE:
// dur is a string like "24h", "10m" or "5s".
func Stale(lastVersionCheck string, dur string) bool {
	d, err := time.ParseDuration("-" + dur)
	if err != nil {
		// If there is no duration provided, then we should presume the loading of
		// remote configuration failed and that we should retry that operation.
		return true
	}

	if t, _ := time.Parse(time.RFC3339, lastVersionCheck); !t.Before(time.Now().Add(d)) {
		return false
	}

	return true
}
