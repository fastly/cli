package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/v28/github"
	"github.com/mholt/archiver"
)

// Versioner describes a source of CLI release artifacts.
type Versioner interface {
	LatestVersion(context.Context) (semver.Version, error)
	Download(context.Context, semver.Version) (filename string, err error)
	Name() string
}

// GitHub is a versioner that uses GitHub releases.
type GitHub struct {
	client *github.Client
	org    string
	repo   string
	binary string // name of compiled binary
	local  string // name to use for binary once extracted
}

// GitHubOpts represents options to be passed to NewGitHub.
type GitHubOpts struct {
	Org    string
	Repo   string
	Binary string
}

// NewGitHub returns a usable GitHub versioner utilizing the provided token.
func NewGitHub(opts GitHubOpts) *GitHub {
	return &GitHub{
		client: github.NewClient(nil),
		org:    opts.Org,
		repo:   opts.Repo,
		binary: opts.Binary,
	}
}

// RenameLocalBinary will rename the downloaded binary.
//
// NOTE: this exists so that we can, for example, rename a binary such as
// 'viceroy' to something less ambiguous like 'fastly-localtesting'.
func (g *GitHub) RenameLocalBinary(s string) error {
	g.local = s
	return nil
}

// Name will return the name of the binary.
func (g GitHub) Name() string {
	if g.local != "" {
		return g.local
	}
	return g.binary
}

// LatestVersion implements the Versioner interface.
func (g GitHub) LatestVersion(ctx context.Context) (semver.Version, error) {
	release, _, err := g.client.Repositories.GetLatestRelease(ctx, g.org, g.repo)
	if err != nil {
		return semver.Version{}, err
	}
	return semver.Parse(strings.TrimPrefix(release.GetName(), "v"))
}

// Download implements the Versioner interface.
func (g GitHub) Download(ctx context.Context, version semver.Version) (filename string, err error) {
	releaseID, err := g.getReleaseID(ctx, version)
	if err != nil {
		return filename, err
	}

	release, _, err := g.client.Repositories.GetRelease(ctx, g.org, g.repo, releaseID)
	if err != nil {
		return filename, fmt.Errorf("error fetching release: %w", err)
	}

	assetID, err := g.getAssetID(release.Assets, version)
	if err != nil {
		return filename, err
	}

	rc, redir, err := g.client.Repositories.DownloadReleaseAsset(ctx, g.org, g.repo, assetID)
	if err != nil {
		return filename, err
	}
	if redir != "" {
		// gosec flagged this:
		// G107 (CWE-88): Potential HTTP request made with variable url.
		// Disabling as we trust the source of the URL variable.
		/* #nosec */
		resp, err := http.Get(redir)
		if err != nil {
			return filename, err
		}
		if resp.StatusCode != http.StatusOK {
			return filename, fmt.Errorf("GitHub gave %s", resp.Status)
		}
		rc = resp.Body
	}
	defer rc.Close()

	archivePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s.tgz", g.binary, version))
	dst, err := os.Create(archivePath)
	if err != nil {
		return filename, fmt.Errorf("error creating temp file: %w", err)
	}
	defer os.RemoveAll(archivePath)

	_, err = io.Copy(dst, rc)
	if err != nil {
		return filename, fmt.Errorf("error downloading release: %w", err)
	}

	if err := dst.Close(); err != nil {
		return filename, fmt.Errorf("error closing release file: %w", err)
	}

	binaryPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s_%d", g.binary, version, time.Now().UnixNano()))
	if err := archiver.NewTarGz().Extract(archivePath, g.binary, binaryPath); err != nil {
		return filename, fmt.Errorf("error extracting binary: %w", err)
	}

	latestPath := filepath.Join(binaryPath, g.binary)

	if g.local != "" {
		if err := os.Rename(latestPath, filepath.Join(binaryPath, g.local)); err != nil {
			return filename, fmt.Errorf("error renaming binary: %w", err)
		}
		latestPath = filepath.Join(binaryPath, g.local)
	}

	return latestPath, nil
}

func (g GitHub) getReleaseID(ctx context.Context, version semver.Version) (id int64, err error) {
	var (
		page        int
		versionStr  = version.String()
		vVersionStr = "v" + versionStr
	)
	for {
		releases, resp, err := g.client.Repositories.ListReleases(ctx, g.org, g.repo, &github.ListOptions{
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

func (g GitHub) getAssetID(assets []github.ReleaseAsset, version semver.Version) (id int64, err error) {
	target := fmt.Sprintf("%s_v%s_%s-%s.tar.gz", g.binary, version, runtime.GOOS, runtime.GOARCH)
	for _, asset := range assets {
		if asset.GetName() == target {
			return asset.GetID(), nil
		}
	}
	return id, fmt.Errorf("no asset found for your OS (%s) and architecture (%s)", runtime.GOOS, runtime.GOARCH)
}
