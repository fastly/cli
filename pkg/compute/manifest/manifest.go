package manifest

import (
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
)

// Filename is the name of the package manifest file.
// It is expected to be a project specific configuration file.
const Filename = "fastly.toml"

// ManifestLatestVersion represents the latest known manifest schema version
// supported by the CLI.
const ManifestLatestVersion = 1

// Source enumerates where a manifest parameter is taken from.
type Source uint8

const (
	// SourceUndefined indicates the parameter isn't provided in any of the
	// available sources, similar to "not found".
	SourceUndefined Source = iota

	// SourceFile indicates the parameter came from a manifest file.
	SourceFile

	// SourceFlag indicates the parameter came from an explicit flag.
	SourceFlag
)

var once sync.Once

// Data holds global-ish manifest data from manifest files, and flag sources.
// It has methods to give each parameter to the components that need it,
// including the place the parameter came from, which is a requirement.
//
// If the same parameter is defined in multiple places, it is resolved according
// to the following priority order: the manifest file (lowest priority) and then
// explicit flags (highest priority).
type Data struct {
	File File
	Flag Flag
}

// Name yields a Name.
func (d *Data) Name() (string, Source) {
	if d.Flag.Name != "" {
		return d.Flag.Name, SourceFlag
	}

	if d.File.Name != "" {
		return d.File.Name, SourceFile
	}

	return "", SourceUndefined
}

// ServiceID yields a ServiceID.
func (d *Data) ServiceID() (string, Source) {
	if d.Flag.ServiceID != "" {
		return d.Flag.ServiceID, SourceFlag
	}

	if d.File.ServiceID != "" {
		return d.File.ServiceID, SourceFile
	}

	return "", SourceUndefined
}

// Description yields a Description.
func (d *Data) Description() (string, Source) {
	if d.Flag.Description != "" {
		return d.Flag.Description, SourceFlag
	}

	if d.File.Description != "" {
		return d.File.Description, SourceFile
	}

	return "", SourceUndefined
}

// Authors yields an Authors.
func (d *Data) Authors() ([]string, Source) {
	if len(d.Flag.Authors) > 0 {
		return d.Flag.Authors, SourceFlag
	}

	if len(d.File.Authors) > 0 {
		return d.File.Authors, SourceFile
	}

	return []string{}, SourceUndefined
}

// Version represents the currently supported schema for the fastly.toml
// manifest file that determines the configuration for a compute@edge service.
//
// NOTE: the File object has a field called ManifestVersion which this type is
// assigned. The reason we don't name this type ManifestVersion is to appease
// the static analysis linter which complains re: stutter in the import
// manifest.ManifestVersion.
type Version int

// UnmarshalText manages multiple scenarios where historically the manifest
// version was a string value and not an integer.
//
// Example mappings:
//
// "0.1.0" -> 1
// "1"     -> 1
// 1       -> 1
// "1.0.0" -> 1
// 0.1     -> 1
// "0.2.0" -> 1
// "2.0.0" -> 2
//
// We also constrain the version so that if a user has a manifest_version
// defined as "99.0.0" then we won't accidentally store it as the integer 99
// but instead will return an error because it exceeds the current
// ManifestLatestVersion version of 1.
func (v *Version) UnmarshalText(text []byte) error {
	s := string(text)

	if i, err := strconv.Atoi(s); err == nil {
		*v = Version(i)
		return nil
	}

	if f, err := strconv.ParseFloat(s, 32); err == nil {
		intfl := int(f)
		if intfl == 0 {
			*v = 1
		} else {
			*v = Version(intfl)
		}
		return nil
	}

	if strings.Contains(s, ".") {
		segs := strings.Split(s, ".")

		// A length of 3 presumes a semver (e.g. 0.1.0)
		if len(segs) == 3 {
			if segs[0] != "0" {
				if i, err := strconv.Atoi(segs[0]); err == nil {
					if i > ManifestLatestVersion {
						return errors.ErrUnrecognisedManifestVersion
					}
					*v = Version(i)
					return nil
				}
			} else {
				*v = 1
				return nil
			}
		}
	}

	return errors.ErrUnrecognisedManifestVersion
}

// File represents all of the configuration parameters in the fastly.toml
// manifest file schema.
type File struct {
	ManifestVersion Version  `toml:"manifest_version"`
	Name            string   `toml:"name"`
	Description     string   `toml:"description"`
	Authors         []string `toml:"authors"`
	Language        string   `toml:"language"`
	ServiceID       string   `toml:"service_id"`

	exists bool
	output io.Writer
}

// Exists yields whether the manifest exists.
func (f *File) Exists() bool {
	return f.exists
}

// SetOutput sets the output stream for any messages.
func (f *File) SetOutput(output io.Writer) {
	f.output = output
}

// Read loads the manifest file content from disk.
func (f *File) Read(fpath string) error {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable.
	// Disabling as we need to load the fastly.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	/* #nosec */
	bs, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}

	err = toml.Unmarshal(bs, f)
	if err != nil {
		return err
	}

	f.exists = true

	if f.ManifestVersion == 0 {
		f.ManifestVersion = 1

		// TODO: Provide link to v1 schema once published publicly.
		//
		// NOTE: the use of once is a quick-fix to side-step duplicate outputs.
		// To fix this properly will require a refactor of the structure of how our
		// global output is passed around.
		once.Do(func() {
			text.Warning(f.output, "The fastly.toml was missing a `manifest_version` field. A default schema version of `1` will be used.")
			text.Break(f.output)
			f.Write(fpath)
		})
	}

	return nil
}

// Write persists the manifest content to disk.
func (f *File) Write(filename string) error {
	fp, err := os.Create(filename)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(fp).Encode(f); err != nil {
		return err
	}
	if err := fp.Sync(); err != nil {
		return err
	}
	if err := fp.Close(); err != nil {
		return err
	}
	return nil
}

// Flag represents all of the manifest parameters that can be set with explicit
// flags. Consumers should bind their flag values to these fields directly.
type Flag struct {
	Name        string
	Description string
	Authors     []string
	ServiceID   string
}
