package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/blang/semver"
	fstruntime "github.com/fastly/cli/pkg/runtime"
	"github.com/google/go-github/v38/github"
)

// TestBinary validates that the Binary method returns the expected binary name.
func TestName(t *testing.T) {
	want := "fastly"
	if fstruntime.Windows {
		want = want + ".exe"
	}

	gh := NewGitHub(GitHubOpts{"org", "repo", "fastly"})

	if have := gh.Binary(); have != want {
		t.Fatalf("want: %s, have: %s", want, have)
	}
}

type mockClient struct {
	name    string
	release *github.RepositoryRelease
}

// Satisfy the GitHubRepoClient interface...

func (c mockClient) GetLatestRelease(_ context.Context, _, _ string) (release *github.RepositoryRelease, response *github.Response, err error) {
	return release, response, err
}

func (c mockClient) GetRelease(_ context.Context, _, _ string, _ int64) (release *github.RepositoryRelease, response *github.Response, err error) {
	return c.release, response, err
}

func (c mockClient) DownloadReleaseAsset(_ context.Context, _, _ string, _ int64, _ *http.Client) (asset io.ReadCloser, redirectURL string, err error) {
	asset, err = os.Open(filepath.Join("testdata", c.name))
	return asset, redirectURL, err
}

func (c mockClient) ListReleases(_ context.Context, _, _ string, _ *github.ListOptions) (releases []*github.RepositoryRelease, response *github.Response, err error) {
	return []*github.RepositoryRelease{c.release}, response, err
}

// Mock some functions called by Download() method...

func (c mockClient) GetReleaseID(_ context.Context, _ semver.Version) (id int64, err error) {
	return id, err
}

// TestDownloadArchiveExtract validates both Windows and Unix release assets.
func TestDownloadArchiveExtract(t *testing.T) {
	scenarios := []struct {
		Platform string
		Arch     string
		Ext      string
	}{
		{
			Platform: "darwin",
			Arch:     "amd64",
			Ext:      ".tar.gz",
		},
		{
			Platform: "windows",
			Arch:     "amd64",
			Ext:      ".zip",
		},
	}

	for _, testcase := range scenarios {
		name := fmt.Sprintf("%s_%s", testcase.Platform, testcase.Arch)

		t.Run(name, func(t *testing.T) {
			// Avoid, for example, running the Windows OS scenario on non Windows OS.
			// Otherwise, the Windows OS scenario would show on Darwin an error like:
			// no asset found for your OS (darwin) and architecture (amd64)
			if runtime.GOOS != testcase.Platform || runtime.GOARCH != testcase.Arch {
				t.Skip()
			}

			latest, err := semver.Parse("0.41.0")
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			version := latest.String()
			id := int64(123)
			asset := fmt.Sprintf(DefaultAssetFormat, "fastly", version, testcase.Platform, testcase.Arch, testcase.Ext)

			binary := "fastly"
			if fstruntime.Windows {
				binary = binary + ".exe"
			}

			gh := GitHub{
				client: mockClient{
					name: asset,
					release: &github.RepositoryRelease{
						Name: &version,
						Assets: []*github.ReleaseAsset{
							{
								Name: &asset,
								ID:   &id,
							},
						},
					},
				},
				org:    "fastly",
				repo:   "cli",
				binary: binary,
			}
			gh.SetAsset(asset)

			bin, err := gh.Download(context.Background(), latest)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if err := os.RemoveAll(bin); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
