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
}

// GitHub is a versioner that uses GitHub releases.
type GitHub struct {
	client *github.Client
	org    string
	repo   string
}

// NewGitHub returns a usable GitHub versioner utilizing the provided token.
func NewGitHub(ctx context.Context, org string, repo string) *GitHub {
	var (
		githubClient = github.NewClient(nil)
	)
	return &GitHub{
		client: githubClient,
		org:    org,
		repo:   repo,
	}
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

	archivePath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s.tgz", g.org, version))
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

	binaryPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s_%d", g.org, version, time.Now().UnixNano()))
	if err := archiver.NewTarGz().Extract(archivePath, g.org, binaryPath); err != nil {
		return filename, fmt.Errorf("error extracting binary: %w", err)
	}

	return filepath.Join(binaryPath, g.org), nil
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
	target := fmt.Sprintf("%s_v%s_%s-%s.tar.gz", g.org, version, runtime.GOOS, runtime.GOARCH)
	for _, asset := range assets {
		if asset.GetName() == target {
			return asset.GetID(), nil
		}
	}
	return id, fmt.Errorf("no asset found for your OS (%s) and architecture (%s)", runtime.GOOS, runtime.GOARCH)
}
