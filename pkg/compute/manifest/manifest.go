package manifest

import (
	"os"
	"strconv"
	"strings"

	"github.com/fastly/cli/pkg/errors"
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

// File represents all of the configuration parameters in the fastly.toml
// manifest file schema.
type File struct {
	ManifestVersion string   `toml:"manifest_version"`
	Version         int      `toml:"version"`
	Name            string   `toml:"name"`
	Description     string   `toml:"description"`
	Authors         []string `toml:"authors"`
	Language        string   `toml:"language"`
	ServiceID       string   `toml:"service_id"`

	exists bool
}

// Exists yields whether the manifest exists.
func (f *File) Exists() bool {
	return f.exists
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
	if err == nil {
		f.exists = true

		if f.ManifestVersion == "" {
			return errors.ErrMissingManifestVersion
		}

		// historically before settling on an integer type, the manifest_version
		// was using a form of semantic versioning which never went above 0.1.0
		if strings.Contains(f.ManifestVersion, ".") {
			if f.ManifestVersion != "0.1.0" {
				return errors.ErrUnrecognisedManifestVersion
			}
		} else {
			manifestVersion, err := strconv.Atoi(f.ManifestVersion)
			if err != nil {
				return errors.ErrUnrecognisedManifestVersion
			}

			if manifestVersion < 1 || manifestVersion > ManifestLatestVersion {
				return errors.ErrUnrecognisedManifestVersion
			}
		}
	}

	return err
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
