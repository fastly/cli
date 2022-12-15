package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fastly/cli/pkg/api"
	fstruntime "github.com/fastly/cli/pkg/runtime"
	"github.com/mholt/archiver"
)

const (
	// metadataURL takes a GitHub repo (e.g. cli or viceroy), an OS (e.g. darwin or linux), and an arch (e.g. amd64 or arm64).
	metadataURL = "https://developer.fastly.com/api/internal/releases/meta/%s/%s/%s"
)

// New returns a usable asset.
func New(opts Opts) *Asset {
	binary := opts.Binary
	if fstruntime.Windows && filepath.Ext(binary) == "" {
		binary = binary + ".exe"
	}

	return &Asset{
		httpClient: opts.HTTPClient,
		org:        opts.Org,
		repo:       opts.Repo,
		binary:     binary,
	}
}

// Opts represents options to be passed to NewGitHub.
type Opts struct {
	// Binary is the name of the executable binary.
	Binary string
	// HTTPClient is able to make HTTP requests.
	HTTPClient api.HTTPClient
	// Org is a GitHub organisation.
	Org string
	// Repo is a GitHub repository.
	Repo string
}

// Asset is a versioner that uses Asset releases.
type Asset struct {
	// binary is the name of the executable binary.
	binary string
	// httpClient is able to make HTTP requests.
	httpClient api.HTTPClient
	// org is a GitHub organisation.
	org string
	// repo is a GitHub repository.
	repo string
	// url is the endpoint for downloading the release asset.
	url string
	// version is the release version of the asset.
	version string
}

// Binary returns the configured binary output name.
//
// NOTE: For some operating systems this might include a file extension, such
// as .exe for Windows.
func (g Asset) Binary() string {
	return g.binary
}

// URL returns the downloadable asset URL if set, otherwise calls the API metadata endpoint.
func (g Asset) URL() (url string, err error) {
	if g.url != "" {
		return g.url, nil
	}

	// The metadata method mutates the instance `url` field.
	_, err = g.metadata()
	if err != nil {
		return url, err
	}

	return g.url, nil
}

// Version returns the asset Version if set, otherwise calls the API metadata endpoint.
func (g Asset) Version() (version string, err error) {
	if g.version != "" {
		return g.version, nil
	}

	// The metadata method mutates the instance `url` field.
	_, err = g.metadata()
	if err != nil {
		return version, err
	}

	return g.version, nil
}

// Download retrieves the binary archive format from GitHub.
func (g *Asset) Download() (bin string, err error) {
	endpoint, err := g.URL()
	if err != nil {
		return bin, err
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return bin, err
	}

	if g.httpClient == nil {
		g.httpClient = http.DefaultClient
	}
	res, err := g.httpClient.Do(req)
	if err != nil {
		return bin, err
	}
	if res.StatusCode != http.StatusOK {
		return bin, fmt.Errorf("unsuccessful status: %s", res.Status)
	}
	defer res.Body.Close()

	dir, err := os.MkdirTemp("", "fastly-download")
	if err != nil {
		return bin, fmt.Errorf("error creating temp release directory: %w", err)
	}
	defer os.RemoveAll(dir)

	assetBase := filepath.Base(endpoint)

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as the inputs need to be dynamically determined.
	/* #nosec */
	archive, err := os.Create(filepath.Join(dir, assetBase))
	if err != nil {
		return bin, fmt.Errorf("error creating release asset file: %w", err)
	}

	_, err = io.Copy(archive, res.Body)
	if err != nil {
		return bin, fmt.Errorf("error downloading release asset: %w", err)
	}

	if err := archive.Close(); err != nil {
		return bin, fmt.Errorf("error closing release asset file: %w", err)
	}

	if err := archiver.Extract(archive.Name(), g.binary, dir); err != nil {
		return bin, fmt.Errorf("error extracting binary: %w", err)
	}
	extractedBinary := filepath.Join(dir, g.binary)

	// G302 (CWE-276): Expect file permissions to be 0600 or less
	// gosec flagged this:
	// Disabling as the file was not executable without it and we need all users
	// to be able to execute the binary.
	/* #nosec */
	err = os.Chmod(extractedBinary, 0o777)
	if err != nil {
		return "", err
	}

	tmpBin, err := os.CreateTemp("", g.binary)
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}

	defer func(name string) {
		if err != nil {
			_ = os.Remove(name)
		}
	}(tmpBin.Name())

	if err := tmpBin.Close(); err != nil {
		return "", fmt.Errorf("error closing temp file: %w", err)
	}

	if err := os.Rename(extractedBinary, tmpBin.Name()); err != nil {
		return "", fmt.Errorf("error renaming release asset file: %w", err)
	}

	return tmpBin.Name(), nil
}

// metadata acquires GitHub metadata and stores the asset URL and Version.
func (g *Asset) metadata() (Metadata, error) {
	var m Metadata
	endpoint := fmt.Sprintf(metadataURL, g.repo, runtime.GOOS, runtime.GOARCH)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return m, err
	}

	if g.httpClient == nil {
		g.httpClient = http.DefaultClient
	}
	res, err := g.httpClient.Do(req)
	if err != nil {
		return m, err
	}
	if res.StatusCode != http.StatusOK {
		return m, fmt.Errorf("unsuccessful status: %s", res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return m, err
	}

	err = json.Unmarshal(data, &m)
	if err != nil {
		return m, err
	}

	g.url = m.URL
	g.version = m.Version

	return m, nil
}

// Metadata represents the DevHub API response for software metadata.
type Metadata struct {
	// URL is the endpoint for downloading the release asset.
	URL string `json:"url"`
	// Version is the release version of the asset.
	Version string `json:"version"`
}

// Versioner describes a source of CLI release artifacts.
type Versioner interface {
	// Binary returns the configured binary output name.
	Binary() string
	// Download implements the Versioner interface.
	Download() (bin string, err error)
	// URL returns the asset URL if set, otherwise calls the API metadata endpoint.
	URL() (url string, err error)
	// Version returns the asset Version if set, otherwise calls the API metadata endpoint.
	Version() (version string, err error)
}
