package mock

import "fmt"

// Versioner mocks the github.Versioner interface.
type Versioner struct {
	AssetVersion   string
	BinaryFilename string
	DownloadOK     bool
	DownloadedFile string
}

// BinaryName implements github.Versioner interface.
func (v Versioner) BinaryName() string {
	return v.BinaryFilename
}

// Download implements github.Versioner interface.
func (v Versioner) Download() (string, error) {
	if v.DownloadOK {
		return v.DownloadedFile, nil
	}
	return "", fmt.Errorf("not implemented")
}

// URL implements github.Versioner interface.
func (v Versioner) URL() (string, error) {
	return "", nil
}

// Version implements github.Versioner interface.
func (v Versioner) Version() (string, error) {
	return v.AssetVersion, nil
}
