package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/mholt/archiver"

	"github.com/fastly/cli/pkg/api"
	fstruntime "github.com/fastly/cli/pkg/runtime"
)

const (
	// metadataURL takes a GitHub repo (e.g. cli or viceroy), an OS (e.g. darwin or linux), and an arch (e.g. amd64 or arm64).
	metadataURL = "https://developer.fastly.com/api/internal/releases/meta/%s/%s/%s"
)

// New returns a usable asset.
func New(opts Opts) *Asset {
	binary := opts.Binary
	if fstruntime.Windows && filepath.Ext(binary) == "" {
		binary += ".exe"
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

// BinaryName returns the configured binary output name.
//
// NOTE: For some operating systems this might include a file extension, such
// as .exe for Windows.
func (g Asset) BinaryName() string {
	return g.binary
}

// DownloadLatest retrieves the latest binary version.
func (g *Asset) DownloadLatest() (bin string, err error) {
	endpoint, err := g.URL()
	if err != nil {
		return "", err
	}
	return g.Download(endpoint)
}

// DownloadVersion retrieves the specified binary version.
func (g *Asset) DownloadVersion(version string) (bin string, err error) {
	_, err = semver.Parse(version)
	if err != nil {
		return "", err
	}

	endpoint, err := g.URL()
	if err != nil {
		return "", err
	}

	endpoint = strings.ReplaceAll(endpoint, g.version, version)

	return g.Download(endpoint)
}

// Download retrieves the binary archive format from the specified endpoint.
func (g *Asset) Download(endpoint string) (bin string, err error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create a HTTP request: %w", err)
	}

	if g.httpClient == nil {
		g.httpClient = http.DefaultClient
	}
	res, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request GitHub release asset: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to request GitHub release asset: %s", res.Status)
	}
	defer res.Body.Close() // #nosec G307

	tmpDir, err := os.MkdirTemp("", "fastly-download")
	if err != nil {
		return "", fmt.Errorf("failed to create temp release directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archive, err := createArchive(filepath.Base(endpoint), tmpDir, res.Body)
	if err != nil {
		return "", err
	}

	extractedBinary, err := extractBinary(archive, g.binary, tmpDir)
	if err != nil {
		return "", err
	}

	return moveExtractedBinary(g.binary, extractedBinary)
}

// URL returns the downloadable asset URL if set, otherwise calls the API metadata endpoint.
func (g *Asset) URL() (url string, err error) {
	if g.url != "" {
		return g.url, nil
	}

	m, err := g.metadata()
	if err != nil {
		return "", err
	}

	g.url = m.URL
	g.version = m.Version

	return g.url, nil
}

// LatestVersion returns the asset LatestVersion if set, otherwise calls the API metadata endpoint.
func (g *Asset) LatestVersion() (version string, err error) {
	if g.version != "" {
		return g.version, nil
	}

	m, err := g.metadata()
	if err != nil {
		return "", err
	}

	g.url = m.URL
	g.version = m.Version

	return g.version, nil
}

// metadata acquires GitHub metadata.
func (g *Asset) metadata() (m Metadata, err error) {
	endpoint := fmt.Sprintf(metadataURL, g.repo, runtime.GOOS, runtime.GOARCH)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return m, fmt.Errorf("failed to create a HTTP request: %w", err)
	}

	if g.httpClient == nil {
		g.httpClient = http.DefaultClient
	}
	res, err := g.httpClient.Do(req)
	if err != nil {
		return m, fmt.Errorf("failed to request GitHub metadata: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return m, fmt.Errorf("failed to request GitHub metadata: %s", res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return m, fmt.Errorf("failed to read GitHub's metadata response: %w", err)
	}

	err = json.Unmarshal(data, &m)
	if err != nil {
		return m, fmt.Errorf("failed to parse GitHub's metadata: %w", err)
	}

	return m, nil
}

// Metadata represents the DevHub API response for software metadata.
type Metadata struct {
	// URL is the endpoint for downloading the release asset.
	URL string `json:"url"`
	// Version is the release version of the asset.
	Version string `json:"version"`
}

// AssetVersioner describes a source of CLI release artifacts.
type AssetVersioner interface {
	// BinaryName returns the configured binary output name.
	BinaryName() string
	// Download downloads the asset from the specified endpoint.
	Download(endpoint string) (bin string, err error)
	// DownloadLatest downloads the latest version of the asset.
	DownloadLatest() (bin string, err error)
	// DownloadVersion downloads the specified version of the asset.
	DownloadVersion(version string) (bin string, err error)
	// URL returns the asset URL if set, otherwise calls the API metadata endpoint.
	URL() (url string, err error)
	// LatestVersion returns the latest version.
	LatestVersion() (version string, err error)
}

// createArchive copies the DevHub response body data into a temporary archive
// file and returns the path to the file.
func createArchive(assetBase, tmpDir string, data io.ReadCloser) (path string, err error) {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as the inputs need to be dynamically determined.
	/* #nosec */
	archive, err := os.Create(filepath.Join(tmpDir, assetBase))
	if err != nil {
		return "", fmt.Errorf("failed to create a temporary file: %w", err)
	}

	_, err = io.Copy(archive, data)
	if err != nil {
		return "", fmt.Errorf("failed to copy the release asset response body: %w", err)
	}

	if err := archive.Close(); err != nil {
		return "", fmt.Errorf("failed to close release asset file: %w", err)
	}

	return archive.Name(), nil
}

// extractBinary extracts the executable binary (e.g. fastly or viceroy) from
// the specified archive file, modifies its permissions and returns the path.
func extractBinary(archive, filename, dst string) (bin string, err error) {
	if err := archiver.Extract(archive, filename, dst); err != nil {
		return "", fmt.Errorf("failed to extract binary: %w", err)
	}
	extractedBinary := filepath.Join(dst, filename)

	// G302 (CWE-276): Expect file permissions to be 0600 or less
	// gosec flagged this:
	// Disabling as the file was not executable without it and we need all users
	// to be able to execute the binary.
	/* #nosec */
	err = os.Chmod(extractedBinary, 0o777)
	if err != nil {
		return "", fmt.Errorf("failed to modify permissions on extracted binary: %w", err)
	}

	return extractedBinary, nil
}

// moveExtractedBinary creates a temporary file (representing the final
// executable binary) and moves the oldpath to it and returns its path.
func moveExtractedBinary(binName, oldpath string) (path string, err error) {
	tmpBin, err := os.CreateTemp("", binName)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	defer func(name string) {
		if err != nil {
			_ = os.Remove(name)
		}
	}(tmpBin.Name())

	if err := tmpBin.Close(); err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(oldpath, tmpBin.Name()); err != nil {
		return "", fmt.Errorf("failed to rename release asset file: %w", err)
	}

	return tmpBin.Name(), nil
}
