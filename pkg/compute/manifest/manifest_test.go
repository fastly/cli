package manifest

import (
	"errors"
	"io/ioutil"
	"os"
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

	// NOTE: the fixture file "fastly-invalid-missing-version.toml" will be
	// overwritten by the test as the internal logic is supposed to add back into
	// the manifest a manifest_version field if one isn't found.
	//
	// To ensure future test runs complete successfully we do an initial read of
	// the data and then write it back out when the tests have completed.
	path, err := filepath.Abs(filepath.Join(prefix, "fastly-invalid-missing-version.toml"))
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = ioutil.WriteFile(path, b, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}()

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var m File
			m.SetOutput(os.Stdout)

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
