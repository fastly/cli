package update

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/github"
	fstruntime "github.com/fastly/cli/pkg/runtime"
)

// Check if the CLI can be updated.
func Check(ctx context.Context, currentVersion string, cliVersioner github.Versioner) (current, latest semver.Version, shouldUpdate bool) {
	current, err := semver.Parse(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		return current, latest, false
	}

	latest, err = cliVersioner.LatestVersion(ctx)
	if err != nil {
		return current, latest, false
	}

	// TODO: change goreleaser to produce .tar.gz for CLI on Windows
	archiveFormat := ".tar.gz"
	if fstruntime.Windows {
		archiveFormat = ".zip"
	}
	asset := fmt.Sprintf(github.DefaultAssetFormat, cliVersioner.BinaryName(), latest, runtime.GOOS, runtime.GOARCH, archiveFormat)
	cliVersioner.SetAsset(asset)

	return current, latest, latest.GT(current)
}

type checkResult struct {
	current      semver.Version
	latest       semver.Version
	shouldUpdate bool
}

// CheckAsync is a helper function for running Check asynchronously.
//
// Launches a goroutine to perform a check for the latest CLI version using the
// provided context and return a function that will print an informative message
// to the writer if there is a newer version available.
//
// Callers should invoke CheckAsync via
//
//	f := CheckAsync(...)
//	defer f()
func CheckAsync(
	ctx context.Context,
	currentVersion string,
	cliVersioner github.Versioner,
	quietMode bool,
) (printResults func(io.Writer)) {
	results := make(chan checkResult, 1)
	go func() {
		current, latest, shouldUpdate := Check(ctx, currentVersion, cliVersioner)
		results <- checkResult{current, latest, shouldUpdate}
	}()

	return func(w io.Writer) {
		result := <-results
		if result.shouldUpdate && !quietMode {
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "A new version of the Fastly CLI is available.\n")
			fmt.Fprintf(w, "Current version: %s\n", result.current)
			fmt.Fprintf(w, "Latest version: %s\n", result.latest)
			fmt.Fprintf(w, "Run `fastly update` to get the latest version.\n")
			fmt.Fprintf(w, "\n")
		}
	}
}
