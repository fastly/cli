package manifest

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	errs "github.com/fastly/cli/pkg/errors"
)

func TestManifest(t *testing.T) {
	prefix := filepath.Join("../", "testdata", "init")

	tests := map[string]struct {
		manifest      string
		valid         bool
		expectedError error
		checkRef      bool
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
			checkRef: true,
		},
		"invalid: manifest_version as a section causes default to be set": {
			manifest: "fastly-invalid-section-version.toml",
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

	// NOTE: the fixture files "fastly-invalid-missing-version.toml" and
	// "fastly-invalid-section-version.toml" will be overwritten by the test as
	// the internal logic is supposed to add back into the manifest a
	// manifest_version field if one isn't found (or is invalid).
	//
	// To ensure future test runs complete successfully we do an initial read of
	// the data and then write it back out when the tests have completed.

	for _, fpath := range []string{
		"fastly-invalid-missing-version.toml",
		"fastly-invalid-section-version.toml",
	} {
		path, err := filepath.Abs(filepath.Join(prefix, fpath))
		if err != nil {
			t.Fatal(err)
		}

		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		defer func(path string, b []byte) {
			err := os.WriteFile(path, b, 0644)
			if err != nil {
				t.Fatal(err)
			}
		}(path, b)
	}

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

				// NOTE: we only assign checkRef to one of the invalid test cases, even
				// though in practice the reference link will be added anytime a write
				// operation is executed on the manifest file, because as far as the
				// test suite is concerned it's constrained by the use of `once.Do()`
				// at the package level of the manifest package.
				//
				// This is validating the write operation called when the
				// manifest_version is invalid. We have a separate test function for
				// validating the function that prepends the reference.
				if tc.checkRef {
					b, err := os.ReadFile(path)
					if err != nil {
						t.Fatal(err)
					}

					content := string(b)

					if !strings.Contains(content, SpecIntro) || !strings.Contains(content, SpecURL) {
						t.Fatal("missing fastly.toml specification reference link")
					}
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

func TestManifestPrepend(t *testing.T) {
	prefix := filepath.Join("../", "testdata", "init")

	// NOTE: the fixture file "fastly-invalid-missing-version.toml" will be
	// overwritten by the test as the internal logic is supposed to add into the
	// manifest a reference to the fastly.toml specification.
	//
	// To ensure future test runs complete successfully we do an initial read of
	// the data and then write it back out when the tests have completed.

	fpath := "fastly-invalid-missing-version.toml"

	path, err := filepath.Abs(filepath.Join(prefix, fpath))
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	defer func(path string, b []byte) {
		err := os.WriteFile(path, b, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}(path, b)

	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = prependSpecRefToManifest(f)
	if err != nil {
		t.Fatal(err)
	}

	bs, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	content := string(bs)

	if !strings.Contains(content, SpecIntro) || !strings.Contains(content, SpecURL) {
		t.Fatal("missing fastly.toml specification reference link")
	}

	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}
