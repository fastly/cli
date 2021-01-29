package update

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/check"
	"github.com/fastly/cli/pkg/config"
)

// Check if the CLI can be updated.
func Check(ctx context.Context, currentVersion string, v Versioner) (current, latest semver.Version, shouldUpdate bool, err error) {
	current, err = semver.Parse(strings.TrimPrefix(currentVersion, "v"))
	if err != nil {
		return current, latest, false, fmt.Errorf("error reading current version: %w", err)
	}

	latest, err = v.LatestVersion(ctx)
	if err != nil {
		return current, latest, false, fmt.Errorf("error fetching latest version: %w", err)
	}

	return current, latest, latest.GT(current), nil
}

type checkResult struct {
	current      semver.Version
	latest       semver.Version
	shouldUpdate bool
	err          error
}

// CheckAsync is a helper function for Check. If the app config's LastChecked
// time has past the specified TTL, launch a goroutine to perform the Check
// using the provided context. Return a function that will print an informative
// message to the writer if there is a newer version available.
//
// Callers should invoke CheckAsync via
//
//     f := CheckAsync(...)
//     defer f()
//
func CheckAsync(ctx context.Context, file config.ConfigFile, configFilePath string, currentVersion string, v Versioner) (printResults func(io.Writer)) {
	if !check.Stale(file.CLI.LastChecked, file.CLI.TTL) {
		return func(io.Writer) {} // no-op
	}

	results := make(chan checkResult, 1)
	go func() {
		current, latest, shouldUpdate, err := Check(ctx, currentVersion, v)
		results <- checkResult{current, latest, shouldUpdate, err}
	}()

	return func(w io.Writer) {
		result := <-results
		if result.err == nil {
			// `fastly configure` may have changed the file contents.
			if err := file.Read(configFilePath); err == nil {
				file.CLI.LastChecked = time.Now().Format(time.RFC3339)
				file.Write(configFilePath)
			}
		}
		if result.shouldUpdate {
			fmt.Fprintf(w, "\n")
			fmt.Fprintf(w, "A new version of the Fastly CLI is available.\n")
			fmt.Fprintf(w, "Current version: %s\n", result.current)
			fmt.Fprintf(w, "Latest version: %s\n", result.latest)
			fmt.Fprintf(w, "Run `fastly update` to get the latest version.\n")
			fmt.Fprintf(w, "\n")
		}
	}
}
