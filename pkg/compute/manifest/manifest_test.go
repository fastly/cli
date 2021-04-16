package manifest

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/env"
	errs "github.com/fastly/cli/pkg/errors"
	toml "github.com/pelletier/go-toml"
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

	// NOTE: the fixture file "fastly-missing-spec-url.toml" will be
	// overwritten by the test as the internal logic is supposed to add into the
	// manifest a reference to the fastly.toml specification.
	//
	// To ensure future test runs complete successfully we do an initial read of
	// the data and then write it back out when the tests have completed.

	fpath := "fastly-missing-spec-url.toml"

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

func TestDataServiceID(t *testing.T) {
	sid := os.Getenv(env.ServiceID)
	defer func(sid string) {
		os.Setenv(env.ServiceID, sid)
	}(sid)

	err := os.Setenv(env.ServiceID, "001")
	if err != nil {
		t.Fatal(err)
	}

	// SourceFlag
	d := Data{
		Flag: Flag{ServiceID: "123"},
		File: File{ServiceID: "456"},
	}
	_, src := d.ServiceID()
	if src != SourceFlag {
		t.Fatal("expected SourceFlag")
	}

	// SourceEnv
	d.Flag = Flag{}
	_, src = d.ServiceID()
	if src != SourceEnv {
		t.Fatal("expected SourceEnv")
	}

	// SourceFile
	err = os.Setenv(env.ServiceID, "")
	if err != nil {
		t.Fatal(err)
	}
	_, src = d.ServiceID()
	if src != SourceFile {
		t.Fatal("expected SourceFile")
	}
}

// This test validates that manually added changes, such as the toml
// syntax for Viceroy local testing, are not accidentally deleted after
// decoding and encoding flows.
func TestManifestPersistsLocalServerSection(t *testing.T) {
	fpath := filepath.Join("../", "testdata", "init", "fastly-viceroy-update.toml")

	b, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	defer func(fpath string, b []byte) {
		err := os.WriteFile(fpath, b, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}(fpath, b)

	original, err := toml.LoadFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	ot := original.Get("local_server")
	if ot == nil {
		t.Fatal("expected [local_server] block to exist in fastly.toml but is missing")
	}

	osid := original.Get("service_id")
	if osid != nil {
		t.Fatal("did not expect service_id key to exist in fastly.toml but is present")
	}

	var m File

	err = m.Read(fpath)
	if err != nil {
		t.Fatal(err)
	}

	m.ServiceID = "a change occurred to the data structure"

	err = m.Write(fpath)
	if err != nil {
		t.Fatal(err)
	}

	latest, err := toml.LoadFile(fpath)
	if err != nil {
		t.Fatal(err)
	}

	lsid := latest.Get("service_id")
	if lsid == nil {
		t.Fatal("expected service_id key to exist in fastly.toml but is missing")
	}

	lt := latest.Get("local_server")
	if lt == nil {
		t.Fatal("expected [local_server] block to exist in fastly.toml but is missing")
	}

	if lt.(*toml.Tree).String() != ot.(*toml.Tree).String() {
		t.Fatal("testing section between original and updated fastly.toml do not match")
	}
}
