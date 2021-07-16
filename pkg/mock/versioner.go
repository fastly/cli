package mock

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver"
)

// Versioner mocks the update.Versioner interface.
type Versioner struct {
	Version        string
	Error          error
	BinaryName     string // name of compiled binary
	Local          string // name to use for binary once extracted
	DownloadOK     bool
	DownloadedFile string
}

// LatestVersion returns the parsed version field, or error if it's non-nil.
func (v Versioner) LatestVersion(context.Context) (semver.Version, error) {
	if v.Error != nil {
		return semver.Version{}, v.Error
	}
	return semver.Parse(strings.TrimPrefix(v.Version, "v"))
}

// Download is a no-op.
func (v Versioner) Download(context.Context, semver.Version) (filename string, err error) {
	if v.DownloadOK {
		return v.DownloadedFile, nil
	}
	return filename, fmt.Errorf("not implemented")
}

// Name will return the name of the binary.
func (v Versioner) Name() string {
	if v.Local != "" {
		return v.Local
	}
	return v.BinaryName
}

// Binary will return the configured name of the binary.
//
// NOTE: This is different from Name() in that it takes into account the local
// field that allows renaming of a binary.
func (v Versioner) Binary() string {
	return v.BinaryName
}

// RenameLocalBinary will rename the downloaded binary.
func (v Versioner) RenameLocalBinary(binName string) error {
	// NoOp
	return nil
}

// SetAsset allows configuring the release asset format.
func (v Versioner) SetAsset(name string) {
	// NoOp
}
