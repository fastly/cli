package update

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"
)

// TestBinary validates that the Binary method returns the expected binary name.
func TestName(t *testing.T) {
	want := "binary"
	gh := NewGitHub(GitHubOpts{"org", "repo", want})

	if have := gh.Binary(); have != want {
		t.Fatalf("want: %s, have: %s", want, have)
	}
}

// TestDownloadArchiveExtract validates both Windows and Unix release assets.
func TestDownloadArchiveExtract(t *testing.T) {
	opts := GitHubOpts{
		Org:    "fastly",
		Repo:   "cli",
		Binary: "fastly",
	}
	gh := NewGitHub(opts)

	scenarios := []struct {
		Platform string
		Arch     string
	}{
		{
			Platform: "darwin",
			Arch:     "amd64",
		},
		{
			Platform: "windows",
			Arch:     "amd64",
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
			latest, err := gh.LatestVersion(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			asset := fmt.Sprintf(DefaultAssetFormat, opts.Org, latest.String(), testcase.Platform, testcase.Arch)
			gh.SetAsset(asset)
			fmt.Printf("\n\nasset: %+v\n\n", asset)

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
