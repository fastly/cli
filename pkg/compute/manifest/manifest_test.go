package manifest

import (
	"path/filepath"
	"testing"

	errs "github.com/fastly/cli/pkg/errors"
)

func TestManifest(t *testing.T) {
	prefix := filepath.Join("../", "testdata", "init")

	tests := map[string]struct {
		manifest      string
		valid         bool
		expectedError error
	}{
		"valid: semver": {
			manifest: "fastly-valid.toml",
			valid:    true,
		},
		"valid: integer": {
			manifest: "fastly-valid-integer.toml",
			valid:    true,
		},
		"invalid: missing manifest_version": {
			manifest:      "fastly-invalid.toml",
			valid:         false,
			expectedError: errs.ErrMissingManifestVersion,
		},
		"invalid: manifest_version Atoi error": {
			manifest:      "fastly-invalid-version.toml",
			valid:         false,
			expectedError: errs.ErrUnrecognisedManifestVersion,
		},
		"unrecognised: manifest_version set to unknown version": {
			manifest:      "fastly-unrecognised.toml",
			valid:         false,
			expectedError: errs.ErrUnrecognisedManifestVersion,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var m File

			path, err := filepath.Abs(filepath.Join(prefix, tc.manifest))
			if err != nil {
				t.Fatal(err)
			}

			err = m.Read(path)
			if tc.valid {
				// if we expect the manifest to be valid and we get an error, then
				// that's unexpected behaviour.
				if err != nil {
					t.Fatal(err)
				}
			} else {
				// otherwise if we expect the manifest to be invalid/unrecognised then
				// the error should match our expectations.
				if err != tc.expectedError {
					t.Fatalf("incorrect error type: %s", err)
				}
			}
		})
	}
}
