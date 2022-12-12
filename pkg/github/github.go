package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/blang/semver"
	fstruntime "github.com/fastly/cli/pkg/runtime"
	"github.com/google/go-github/v38/github"
	"github.com/mholt/archiver"
)

// DefaultAssetFormat represents the standard GitHub release asset name format.
//
// Interpolation placeholders:
// - binary name
// - semantic version
// - os
// - arch
// - archive file extension (e.g. ".tar.gz" or ".zip")
const DefaultAssetFormat = "%s_v%s_%s-%s%s"

// Versioner describes a source of CLI release artifacts.
type Versioner interface {
	Binary() string
	BinaryName() string
	Download(context.Context, semver.Version) (filename string, err error)
	LatestVersion(context.Context) (semver.Version, error)
	SetAsset(name string)
}

// GitHubRepoClient describes the GitHub client behaviours we need.
type GitHubRepoClient interface {
	GetLatestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, *github.Response, error)
	GetRelease(ctx context.Context, owner, repo string, id int64) (*github.RepositoryRelease, *github.Response, error)
	DownloadReleaseAsset(ctx context.Context, owner, repo string, id int64, followRedirectsClient *http.Client) (rc io.ReadCloser, redirectURL string, err error)
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
}

// GitHub is a /* versioner */ that uses GitHub releases.
type GitHub struct {
	client       GitHubRepoClient
	org          string
	repo         string
	binary       string // name of compiled binary
	releaseAsset string // name of the release asset file to download
}

// GitHubOpts represents options to be passed to NewGitHub.
type GitHubOpts struct {
	Org    string
	Repo   string
	Binary string
}

// NewGitHub returns a usable GitHub versioner utilizing the provided token.
func NewGitHub(opts GitHubOpts) *GitHub {
	binary := opts.Binary
	if fstruntime.Windows && filepath.Ext(binary) == "" {
		binary = binary + ".exe"
	}

	return &GitHub{
		client: github.NewClient(nil).Repositories,
		org:    opts.Org,
		repo:   opts.Repo,
		binary: binary,
	}
}

// Binary returns the configured binary output name.
//
// NOTE: For some operating systems this might include a file extension, such
// as .exe for Windows.
func (g *GitHub) Binary() string {
	return g.binary
}

// BinaryName returns the binary name minus any extensions.
func (g *GitHub) BinaryName() string {
	return strings.Split(g.binary, ".")[0]
}

// SetAsset allows configuring the release asset format.
//
// NOTE: This existed because the CLI project was originally using a different
// release asset name format to the Viceroy project. Although the two projects
// are now aligned we've kept this feature in case there are any changes
// between the two projects in the future, or if we have to call out to more
// external binaries from within the CLI.
func (g *GitHub) SetAsset(name string) {
	g.releaseAsset = name
}

// LatestVersion calls the GitHub API to return the latest release as a semver.
func (g GitHub) LatestVersion(ctx context.Context) (semver.Version, error) {
	release, _, err := g.client.GetLatestRelease(ctx, g.org, g.repo)
	if err != nil {
		return semver.Version{}, err
	}
	return semver.Parse(strings.TrimPrefix(release.GetName(), "v"))
}

// Download implements the Versioner interface.
//
// Downloading, unarchiving and changing the file modes is done inside a temporary
// directory within $TMPDIR.
// On success, the resulting file is renamed to a temporary one within $TMPDIR, and
// returned. The temporary directory and its content are always removed.
func (g GitHub) Download(ctx context.Context, version semver.Version) (string, error) {
	releaseID, err := g.GetReleaseID(ctx, version)
	if err != nil {
		return "", err
	}

	release, _, err := g.client.GetRelease(ctx, g.org, g.repo, releaseID)
	if err != nil {
		return "", fmt.Errorf("error fetching release: %w", err)
	}

	assetID, err := g.GetAssetID(release.Assets)
	if err != nil {
		return "", err
	}

	asset, _, err := g.client.DownloadReleaseAsset(ctx, g.org, g.repo, assetID, http.DefaultClient)
	if err != nil {
		return "", err
	}
	defer asset.Close()

	dir, err := os.MkdirTemp("", "fastly-download")
	if err != nil {
		return "", fmt.Errorf("error creating temp release directory: %w", err)
	}
	defer os.RemoveAll(dir)

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as the inputs need to be dynamically determined.
	/* #nosec */
	archive, err := os.Create(filepath.Join(dir, g.releaseAsset))
	if err != nil {
		return "", fmt.Errorf("error creating release asset file: %w", err)
	}

	_, err = io.Copy(archive, asset)
	if err != nil {
		return "", fmt.Errorf("error downloading release asset: %w", err)
	}

	if err := archive.Close(); err != nil {
		return "", fmt.Errorf("error closing release asset file: %w", err)
	}

	if err := archiver.Extract(archive.Name(), g.binary, dir); err != nil {
		return "", fmt.Errorf("error extracting binary: %w", err)
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

	bin, err := os.CreateTemp("", g.binary)
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}

	defer func(name string) {
		if err != nil {
			os.Remove(name)
		}
	}(bin.Name())

	if err := bin.Close(); err != nil {
		return "", fmt.Errorf("error closing temp file: %w", err)
	}

	if err := os.Rename(extractedBinary, bin.Name()); err != nil {
		return "", fmt.Errorf("error renaming release asset file: %w", err)
	}

	return bin.Name(), nil
}

// GetReleaseID returns the release ID.
func (g GitHub) GetReleaseID(ctx context.Context, version semver.Version) (id int64, err error) {
	var (
		page        int
		versionStr  = version.String()
		vVersionStr = "v" + versionStr
	)
	for {
		releases, resp, err := g.client.ListReleases(ctx, g.org, g.repo, &github.ListOptions{
			Page:    page,
			PerPage: 100,
		})
		if err != nil {
			return id, err
		}
		for _, release := range releases {
			if name := release.GetName(); name == versionStr || name == vVersionStr {
				return release.GetID(), nil
			}
		}
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}
	return id, fmt.Errorf("no matching release found")
}

// GetAssetID returns the asset ID.
func (g GitHub) GetAssetID(assets []*github.ReleaseAsset) (id int64, err error) {
	if g.releaseAsset == "" {
		return id, fmt.Errorf("no release asset specified")
	}
	for _, asset := range assets {
		if asset.GetName() == g.releaseAsset {
			return asset.GetID(), nil
		}
	}
	return id, fmt.Errorf("no asset found for your OS (%s) and architecture (%s)", runtime.GOOS, runtime.GOARCH)
}
