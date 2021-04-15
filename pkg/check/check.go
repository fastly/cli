package check

import (
	"os"
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

// BinaryUpdated checks if the CLI binary was recently updated and if so it
// won't treat the application configuration as cached (even if the TTL hasn't
// yet expired) because the new binary might need to reference new
// configuration fields that the old configuration doesn't have defined.
//
// NOTE: If any of the operations within this function fail, then we'll be
// cautious and consider the binary updated so we can force a config update.
func BinaryUpdated(dur string) bool {
	d, err := time.ParseDuration("-" + dur)
	if err != nil {
		return true
	}

	bin, err := os.Executable()
	if err != nil {
		return true
	}

	fi, err := os.Stat(bin)
	if err != nil {
		return true
	}

	if t := fi.ModTime(); !t.Before(time.Now().Add(d)) {
		return false
	}

	return false
}
