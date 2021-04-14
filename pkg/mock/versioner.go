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
	Binary         string // name of compiled binary
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
	return v.Binary
}
