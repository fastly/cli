package manifest

import (
	"os"

	"github.com/fastly/cli/v10/pkg/env"
)

// Data holds global-ish manifest data from manifest files, and flag sources.
// It has methods to give each parameter to the components that need it,
// including the place the parameter came from, which is a requirement.
//
// If the same parameter is defined in multiple places, it is resolved according
// to the following priority order: the manifest file (lowest priority) and then
// environment variables (where applicable), and explicit flags (highest priority).
type Data struct {
	File File
	Flag Flag
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

// Description yields a Description.
func (d *Data) Description() (string, Source) {
	if d.File.Description != "" {
		return d.File.Description, SourceFile
	}

	return "", SourceUndefined
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

	if sid := os.Getenv(env.ServiceID); sid != "" {
		return sid, SourceEnv
	}

	if d.File.ServiceID != "" {
		return d.File.ServiceID, SourceFile
	}

	return "", SourceUndefined
}
