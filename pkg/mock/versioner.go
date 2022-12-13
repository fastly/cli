package mock

import (
	"fmt"
)

// Versioner mocks the github.Versioner interface.
type Versioner struct {
	AssetURL       string
	AssetVersion   string
	BinaryFilename string // name of compiled binary
	DownloadOK     bool
	DownloadedFile string
	Error          error
	Local          string // name to use for binary once extracted
}

// Binary implements github.Versioner interface.
func (v Versioner) Binary() string {
	return v.BinaryFilename
}

// Download implements github.Versioner interface.
func (v Versioner) Download() (bin string, err error) {
	if v.DownloadOK {
		return v.DownloadedFile, nil
	}
	return bin, fmt.Errorf("not implemented")
}

// URL implements github.Versioner interface.
func (v Versioner) URL() (string, error) {
	return v.AssetURL, nil
}

// Version implements github.Versioner interface.
func (v Versioner) Version() (string, error) {
	return v.AssetVersion, nil
}
