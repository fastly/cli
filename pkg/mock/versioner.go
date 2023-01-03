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

// Download implements github.Versioner interface.
func (av AssetVersioner) Download() (string, error) {
	if av.DownloadOK {
		return av.DownloadedFile, nil
	}
	return "", fmt.Errorf("not implemented")
}

// URL implements github.Versioner interface.
func (av AssetVersioner) URL() (string, error) {
	return "", nil
}

// Version implements github.Versioner interface.
func (av AssetVersioner) Version() (string, error) {
	return av.AssetVersion, nil
}
