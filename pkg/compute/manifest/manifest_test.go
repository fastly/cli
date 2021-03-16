package manifest

import (
	"errors"
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
			manifest: "fastly-valid-semver.toml",
			valid:    true,
		},
		"valid: integer": {
			manifest: "fastly-valid-integer.toml",
			valid:    true,
		},
		"invalid: missing manifest_version causes default to be set": {
			manifest: "fastly-invalid-missing-version.toml",
			valid:    true,
		},
		"invalid: manifest_version Atoi error": {
			manifest:      "fastly-invalid-unrecognised.toml",
			valid:         false,
			expectedError: errs.ErrUnrecognisedManifestVersion,
		},
		"unrecognised: manifest_version exceeded limit": {
			manifest:      "fastly-invalid-version-exceeded.toml",
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
				if !errors.As(err, &tc.expectedError) {
					t.Fatalf("incorrect error type: %T, expected: %T", err, tc.expectedError)
				}
			}
		})
	}
}
