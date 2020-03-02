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

// File represents all of the configuration parameters in the fastly.toml
// manifest file schema.
type File struct {
	Version     int      `toml:"version"`
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Authors     []string `toml:"authors"`
	Language    string   `toml:"language"`
	ServiceID   string   `toml:"service_id"`
}

func (f *File) Read(filename string) error {
	_, err := toml.DecodeFile(filename, f)
	return err
}

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
	ServiceID string
}
