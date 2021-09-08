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

	"github.com/blang/semver"
	"github.com/google/go-github/v38/github"
	"github.com/mholt/archiver"
)

// DefaultAssetFormat represents the standard GitHub release asset name format.
const DefaultAssetFormat = "%s_v%s_%s-%s.tar.gz"

// Versioner describes a source of CLI release artifacts.
type Versioner interface {
	Binary() string
	Download(context.Context, semver.Version) (filename string, err error)
	LatestVersion(context.Context) (semver.Version, error)
	SetAsset(name string)
}

// GitHub is a versioner that uses GitHub releases.
type GitHub struct {
	client       *github.Client
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
	return &GitHub{
		client: github.NewClient(nil),
		org:    opts.Org,
		repo:   opts.Repo,
		binary: opts.Binary,
	}
}

// Binary returns the configured binary output name.
func (g *GitHub) Binary() string {
	return g.binary
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

	assetID, err := g.getAssetID(release.Assets)
	if err != nil {
		return filename, err
	}

	rc, _, err := g.client.Repositories.DownloadReleaseAsset(ctx, g.org, g.repo, assetID, http.DefaultClient)
	if err != nil {
		return filename, err
	}
	defer rc.Close()

	var extension string

	// TODO: We might need to also account for Window users by also checking for
	// the .zip extension that goreleaser generates:
	// https://github.com/fastly/cli/blob/26588cfd2d00d18643bac5cc18242b2d5ee84b34/.goreleaser.yml#L51
	//
	// Ideally the formats would be the same, but if that's not possible then
	// we can look to use a genericised method such as
	// https://pkg.go.dev/github.com/mholt/archiver#Extract for handling the
	// extraction of a binary from the asset file instead of using the current
	// archiver.NewTarGz().Extract() method.
	if strings.HasSuffix(g.releaseAsset, ".tar.gz") {
		extension = ".tar.gz"
	}

	tmp := os.TempDir()
	dst, err := os.CreateTemp(tmp, fmt.Sprintf("%s_%s_*%s", g.binary, version, extension))
	if err != nil {
		return filename, fmt.Errorf("error creating temp release asset file: %w", err)
	}

	_, err = io.Copy(dst, rc)
	if err != nil {
		return filename, fmt.Errorf("error downloading release asset: %w", err)
	}

	if err := dst.Close(); err != nil {
		return filename, fmt.Errorf("error closing release asset file: %w", err)
	}

	assetFile := dst.Name()

	// If we need to perform an extraction, create a temporary directory and do it there,
	// then copy the extracted file back to the temporary one so we just track one file.
	if strings.HasSuffix(g.releaseAsset, ".tar.gz") {
		dir, err := os.MkdirTemp(tmp, "extract")
		if err != nil {
			return filename, fmt.Errorf("error creating temp extraction directory: %w", err)
		}
		defer os.RemoveAll(dir)
		if err := archiver.NewTarGz().Extract(assetFile, g.binary, dir); err != nil {
			return filename, fmt.Errorf("error extracting binary: %w", err)
		}
		if err := os.Rename(filepath.Join(dir, g.binary), assetFile); err != nil {
			return filename, fmt.Errorf("error renaming binary: %w", err)
		}
	}

	// G302 (CWE-276): Expect file permissions to be 0600 or less
	// gosec flagged this:
	// Disabling as the file was not executable without it and we need all users
	// to be able to execute the binary.
	/* #nosec */
	err = os.Chmod(assetFile, 0777)
	if err != nil {
		return filename, err
	}

	return assetFile, nil
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

func (g GitHub) getAssetID(assets []*github.ReleaseAsset) (id int64, err error) {
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
