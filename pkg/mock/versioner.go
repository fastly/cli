package mock

import "fmt"

// AssetVersioner mocks the github.AssetVersioner interface.
type AssetVersioner struct {
	AssetVersion   string
	BinaryFilename string
	DownloadOK     bool
	DownloadedFile string
}

// BinaryName implements github.Versioner interface.
func (av AssetVersioner) BinaryName() string {
	return av.BinaryFilename
}

// DownloadLatest implements github.Versioner interface.
func (av AssetVersioner) DownloadLatest() (string, error) {
	if av.DownloadOK {
		return av.DownloadedFile, nil
	}
	return "", fmt.Errorf("not implemented")
}

// DownloadVersion implements github.Versioner interface.
func (av AssetVersioner) DownloadVersion(_ string) (string, error) {
	return "", nil
}

// Download implements github.Versioner interface.
func (av AssetVersioner) Download(_ string) (string, error) {
	return "", nil
}

// URL implements github.Versioner interface.
func (av AssetVersioner) URL() (string, error) {
	return "", nil
}

// LatestVersion implements github.Versioner interface.
func (av AssetVersioner) LatestVersion() (string, error) {
	return av.AssetVersion, nil
}

// RequestedVersion implements github.Versioner interface.
func (av AssetVersioner) RequestedVersion() (version string) {
	return ""
}

// SetRequestedVersion implements github.Versioner interface.
func (av AssetVersioner) SetRequestedVersion(_ string) {
	// no-op
}
