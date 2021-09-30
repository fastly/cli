package manifest_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/env"
	errs "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/testutil"
	toml "github.com/pelletier/go-toml"
)

func TestManifest(t *testing.T) {
	prefix := filepath.Join("../", "testdata", "init")

	tests := map[string]struct {
		manifest             string
		valid                bool
		expectedError        error
		wantRemediationError string
	}{
		"valid: semver": {
			manifest: "fastly-valid-semver.toml",
			valid:    true,
		},
		"valid: integer": {
			manifest: "fastly-valid-integer.toml",
			valid:    true,
		},
		"invalid: missing manifest_version": {
			manifest:      "fastly-invalid-missing-version.toml",
			valid:         false,
			expectedError: errs.ErrMissingManifestVersion,
		},
		"invalid: manifest_version as a section": {
			manifest:      "fastly-invalid-section-version.toml",
			valid:         false,
			expectedError: errs.ErrMissingManifestVersion,
		},
		"invalid: manifest_version Atoi error": {
			manifest:             "fastly-invalid-unrecognised.toml",
			valid:                false,
			expectedError:        errs.ErrUnrecognisedManifestVersion,
			wantRemediationError: errs.ErrUnrecognisedManifestVersion.Remediation,
		},
		"unrecognised: manifest_version exceeded limit": {
			manifest:             "fastly-invalid-version-exceeded.toml",
			valid:                false,
			expectedError:        errs.ErrIncompatibleManifestVersion,
			wantRemediationError: errs.ErrIncompatibleManifestVersion.Remediation,
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
			var m manifest.File
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
				// Ensure the remediation error is as expected.
				testutil.AssertRemediationErrorContains(t, err, tc.wantRemediationError)
			}
		})
	}
}

func TestManifestPrepend(t *testing.T) {
	var (
		manifestBody []byte
		manifestPath string
	)

	// NOTE: the fixture file "fastly-missing-spec-url.toml" will be
	// overwritten by the test as the internal logic is supposed to add into the
	// manifest a reference to the fastly.toml specification.
	//
	// To ensure future test runs complete successfully we do an initial read of
	// the data and then write it back out when the tests have completed.
	{
		path, err := filepath.Abs(filepath.Join("../", "testdata", "init", "fastly-missing-spec-url.toml"))
		if err != nil {
			t.Fatal(err)
		}

		manifestBody, err = os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		defer func(path string, b []byte) {
			err := os.WriteFile(path, b, 0644)
			if err != nil {
				t.Fatal(err)
			}
		}(path, manifestBody)
	}

	// Create temp environment to run test code within.
	{
		wd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		rootdir := testutil.NewEnv(testutil.EnvOpts{
			T: t,
			Write: []testutil.FileIO{
				{Src: string(manifestBody), Dst: "fastly.toml"},
			},
		})
		manifestPath = filepath.Join(rootdir, "fastly.toml")
		defer os.RemoveAll(rootdir)

		if err := os.Chdir(rootdir); err != nil {
			t.Fatal(err)
		}
		defer os.Chdir(wd)
	}

	var f manifest.File
	err := f.Read(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Write(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	updatedManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(updatedManifest)

	if !strings.Contains(content, manifest.SpecIntro) || !strings.Contains(content, manifest.SpecURL) {
		t.Fatal("missing fastly.toml specification reference link")
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
	d := manifest.Data{
		Flag: manifest.Flag{ServiceID: "123"},
		File: manifest.File{ServiceID: "456"},
	}
	_, src := d.ServiceID()
	if src != manifest.SourceFlag {
		t.Fatal("expected SourceFlag")
	}

	// SourceEnv
	d.Flag = manifest.Flag{}
	_, src = d.ServiceID()
	if src != manifest.SourceEnv {
		t.Fatal("expected SourceEnv")
	}

	// SourceFile
	err = os.Setenv(env.ServiceID, "")
	if err != nil {
		t.Fatal(err)
	}
	_, src = d.ServiceID()
	if src != manifest.SourceFile {
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

	var m manifest.File

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
