package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/mholt/archiver/v3"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/debug"
	fstruntime "github.com/fastly/cli/pkg/runtime"
)

const (
	// metadataURL takes a GitHub repo (e.g. cli or viceroy), an OS (e.g. darwin or linux), and an arch (e.g. amd64 or arm64).
	metadataURL = "https://developer.fastly.com/api/internal/releases/meta/%s/%s/%s"
)

// InstallDir represents the directory where the assets should be installed.
//
// NOTE: This is a package level variable as it makes testing the behaviour of
// the package easier because the test code can replace the value when running
// the test suite.
var InstallDir = func() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "fastly")
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".fastly")
	}
	panic("unable to deduce user config dir or user home dir")
}()

// New returns a usable asset.
func New(opts Opts) *Asset {
	binary := opts.Binary
	if fstruntime.Windows && filepath.Ext(binary) == "" {
		binary += ".exe"
	}

	return &Asset{
		binary:           binary,
		debug:            opts.DebugMode,
		external:         opts.External,
		httpClient:       opts.HTTPClient,
		nested:           opts.Nested,
		org:              opts.Org,
		repo:             opts.Repo,
		versionRequested: opts.Version,
	}
}

// Opts represents options to be passed to NewGitHub.
type Opts struct {
	// Binary is the name of the executable binary.
	Binary string
	// DebugMode indicates the user has set debug-mode.
	DebugMode bool
	// External indicates the repository is a non-Fastly repo.
	// This means we need a custom metadata fetcher (i.e. dont use metadataURL).
	External bool
	// HTTPClient is able to make HTTP requests.
	HTTPClient api.HTTPClient
	// Nested indicates if the binary is at the root of the archive or not.
	// e.g. wasm-tools archive contains a folder which contains the binary.
	// Where as Viceroy and CLI archives directly contain the binary.
	Nested bool
	// Org is a GitHub organisation.
	Org string
	// Repo is a GitHub repository.
	Repo string
	// Version is the asset's release version to download.
	// The value is the semver format (example: "0.1.0").
	// If not set, then the latest version is implied.
	Version string
}

// Asset is a versioner that uses Asset releases.
type Asset struct {
	// binary is the name of the executable binary.
	binary string
	// debug indicates if the user is running in debug-mode.
	debug bool
	// external indicates the repository is a non-Fastly repo.
	external bool
	// httpClient is able to make HTTP requests.
	httpClient api.HTTPClient
	// nested indicates if the binary is at the root of the archive or not.
	nested bool
	// org is a GitHub organisation.
	org string
	// repo is a GitHub repository.
	repo string
	// url is the endpoint for downloading the release asset.
	url string
	// version is the release version of the asset.
	version string
	// versionRequested is the requested release version of the asset.
	versionRequested string
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
	if g.debug {
		debug.DumpHTTPRequest(req)
	}
	res, err := g.httpClient.Do(req)
	if g.debug {
		debug.DumpHTTPResponse(res)
	}
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

	assetBase := filepath.Base(endpoint)
	archive, err := createArchive(assetBase, tmpDir, res.Body)
	if err != nil {
		return "", err
	}

	extractedBinary, err := extractBinary(archive, g.binary, tmpDir, assetBase, g.nested)
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

// RequestedVersion returns the version of the asset defined in the fastly.toml.
// NOTE: This is only relevant for `compute serve` with viceroy_version pinning.
func (g *Asset) RequestedVersion() string {
	return g.versionRequested
}

// SetRequestedVersion sets the version of the asset to be downloaded.
// This is typically used by `compute serve` when an `--env` flag is set.
func (g *Asset) SetRequestedVersion(version string) {
	g.versionRequested = version
}

// metadata acquires GitHub metadata.
func (g *Asset) metadata() (m DevHubMetadata, err error) {
	endpoint := fmt.Sprintf(metadataURL, g.repo, runtime.GOOS, runtime.GOARCH)
	if g.external {
		endpoint = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", g.org, g.repo)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return m, fmt.Errorf("failed to create a HTTP request: %w", err)
	}

	if g.httpClient == nil {
		g.httpClient = http.DefaultClient
	}
	if g.debug {
		debug.DumpHTTPRequest(req)
	}
	res, err := g.httpClient.Do(req)
	if g.debug {
		debug.DumpHTTPResponse(res)
	}
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

	if g.external {
		return g.parseExternalMetadata(data)
	}

	err = json.Unmarshal(data, &m)
	if err != nil {
		return m, fmt.Errorf("failed to parse GitHub's metadata: %w", err)
	}

	return m, nil
}

// InstallPath returns the location of where the asset should be installed.
func (g *Asset) InstallPath() string {
	return filepath.Join(InstallDir, g.BinaryName())
}

// DevHubMetadata represents the DevHub API response for software metadata.
type DevHubMetadata struct {
	// URL is the endpoint for downloading the release asset.
	URL string `json:"url"`
	// Version is the release version of the asset (e.g. 10.1.0).
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
	// InstallPath returns the location of where the binary should be installed.
	InstallPath() string
	// RequestedVersion returns the version defined in the fastly.toml file.
	RequestedVersion() (version string)
	// SetRequestedVersion sets the version of the asset to be downloaded.
	SetRequestedVersion(version string)
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
	// #nosec
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

// extractBinary extracts the executable binary (e.g. fastly, viceroy,
// wasm-tools) from the specified archive file, modifies its permissions and
// returns the path.
//
// NOTE: wasm-tools binary is within a nested directory.
// So we have to account for that by extracting the directory from the archive
// and then correct the path before attempting to modify the permissions.
func extractBinary(archive, binaryName, dst, assetBase string, nested bool) (bin string, err error) {
	extractPath := binaryName
	if nested {
		extension := ".tar.gz"
		if fstruntime.Windows {
			extension = ".zip"
		}
		// e.g. extract the nested directory "wasm-tools-1.0.42-aarch64-macos"
		// which itself contains the `wasm-tools` binary
		extractPath = strings.TrimSuffix(assetBase, extension)
	}
	if err := archiver.Extract(archive, extractPath, dst); err != nil {
		return "", fmt.Errorf("failed to extract binary: %w", err)
	}

	extractedBinary := filepath.Join(dst, binaryName)
	if nested {
		// e.g. reference the binary from within the nested directory
		extractedBinary = filepath.Join(dst, extractPath, binaryName)
	}

	// G302 (CWE-276): Expect file permissions to be 0600 or less
	// gosec flagged this:
	// Disabling as the file was not executable without it and we need all users
	// to be able to execute the binary.
	/* #nosec */
	err = os.Chmod(extractedBinary, 0o755)
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

// SetBinPerms ensures 0777 perms are set on the binary.
func SetBinPerms(bin string) error {
	// G302 (CWE-276): Expect file permissions to be 0600 or less
	// gosec flagged this:
	// Disabling as the file was not executable without it and we need all users
	// to be able to execute the binary.
	// #nosec
	err := os.Chmod(bin, 0o777)
	if err != nil {
		return fmt.Errorf("error setting executable permissions for %s: %w", bin, err)
	}
	return nil
}

// RawAsset represents a GitHub release asset.
type RawAsset struct {
	// BrowserDownloadURL is a fully qualified URL to download the release asset.
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Metadata represents the GitHub API metadata response for releases.
type Metadata struct {
	// Name is the release name.
	Name string `json:"name"`
	// Assets a list of all available assets within the release.
	Assets []RawAsset `json:"assets"`

	org, repo, binary string
}

// Version parses a semver from the name field.
func (m Metadata) Version() string {
	r := regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+(-(.*))?`)
	return r.FindString(m.Name)
}

// URL filters the assets for a platform correct asset.
//
// NOTE: This only works with wasm-tools naming conventions.
// If we add more tools to download in future then we can abstract as necessary.
func (m Metadata) URL() string {
	platform := runtime.GOOS
	if platform == "darwin" {
		platform = "macos"
	}

	arch := runtime.GOARCH
	switch arch {
	case "arm64":
		arch = "aarch64"
	case "amd64":
		arch = "x86_64"
	}

	extension := "tar.gz"
	if fstruntime.Windows {
		extension = "zip"
	}

	for _, a := range m.Assets {
		version := m.Version()
		// NOTE: We use `m.repo` for wasm-tools instead of `m.binary`.
		// This is because we append `.exe` to `m.binary` on Windows.
		// Instead of filtering the extension we just use `m.repo` instead.
		pattern := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s-%s-%s-%s.%s", m.org, m.repo, version, m.repo, version, arch, platform, extension)
		if matched, _ := regexp.MatchString(pattern, a.BrowserDownloadURL); matched {
			return a.BrowserDownloadURL
		}
	}

	return ""
}

// parseExternalMetadata takes the raw GitHub metadata and coerces it into a
// DevHub specific metadata format.
func (g *Asset) parseExternalMetadata(data []byte) (DevHubMetadata, error) {
	var (
		dhm DevHubMetadata
		m   Metadata
	)

	err := json.Unmarshal(data, &m)
	if err != nil {
		return dhm, fmt.Errorf("failed to parse GitHub's metadata: %w", err)
	}

	m.org = g.org
	m.repo = g.repo
	m.binary = g.binary

	dhm.Version = m.Version()
	dhm.URL = m.URL()

	return dhm, nil
}
