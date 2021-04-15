package config

import (
	"testing"
	"time"
)

func TestBinaryUpdatedScenarios(t *testing.T) {
	tests := map[string]struct {
		dur    string
		expect bool
	}{
		"binary updated outside config ttl": {
			dur:    "10m",
			expect: false,
		},
		"binary updated within config ttl": {
			dur:    "1m",
			expect: true,
		},
	}

	var file File

	file.CLI.TTL = "5m"

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d, err := time.ParseDuration("-" + tc.dur)
			if err != nil {
				t.Fatal(err)
			}

			file.CLI.BinaryUpdated = time.Now().Add(d).Format(time.RFC3339)

			if file.ShouldFetch() != tc.expect {
				t.Fatal("expected `false` (the binary wasn't updated within config TTL), got `true` (it was updated within config TTL)")
			}
		})
	}
}
