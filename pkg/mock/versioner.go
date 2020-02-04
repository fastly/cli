package mock

import (
	"context"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/update"
)

// Versioner mocks the update.Versioner interface.
type Versioner struct {
	Version string
	Error   error
}

// Make sure mock.Versioner implements update.Versioner.
var _ update.Versioner = (*Versioner)(nil)

// LatestVersion returns the parsed version field, or error if it's non-nil.
func (v Versioner) LatestVersion(context.Context) (semver.Version, error) {
	if v.Error != nil {
		return semver.Version{}, v.Error
	}
	return semver.Parse(strings.TrimPrefix(v.Version, "v"))
}

// Download is a no-op.
func (v Versioner) Download(context.Context, semver.Version) (filename string, err error) {
	return filename, fmt.Errorf("not implemented")
}
