package manifest

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Filename is the name of the package manifest file.
const Filename = "fastly.toml"

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
	Version     int      `toml:"version"`
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Authors     []string `toml:"authors"`
	Language    string   `toml:"language"`
	ServiceID   string   `toml:"service_id"`

	exists bool
}

// Exists yeilds whether the manifest exists.
func (f *File) Exists() bool {
	return f.exists
}

// Read loads the manifest file content from disk.
func (f *File) Read(filename string) error {
	_, err := toml.DecodeFile(filename, f)
	if err == nil {
		f.exists = true
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
