// Package revision defines variables that will be populated with values
// specified at build time via LDFLAGS. goreleaser will prompt for missing env
// variables.
// For more details on LDFLAGS:
// https://github.com/golang/go/wiki/GcToolchainTricks#including-build-information-in-the-executable
package revision

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	// AppVersion is the semver for this version of the client, or
	// "v0.0.0-unknown". Handled by goreleaser.
	AppVersion string

	// GitCommit is the short git SHA associated with this build, or
	// "unknown". Handled by goreleaser.
	GitCommit string

	// GoVersion - Prefer letting the code handle this and set GoHostOS and
	// GoHostArc instead. It can be set to the build host's `go version` output.
	GoVersion string

	// GoHostOS is the output of `go env GOHOSTOS` Passed to goreleaser by
	// `make fastly` or the GHA workflow.
	GoHostOS string

	// GoHostArch is the output of `go env GOHOSTARCH` Passed to goreleaser by
	// `make fastly` or the GHA workflow.
	GoHostArch string

	// Environment is set to either "development" (when working locally) or
	// "release" when the code being executed is from a published release.
	// Handled by goreleaser.
	Environment string
)

// None is the AppVersion string for local (unversioned) builds.
const None = "v0.0.0-unknown"

func init() {
	if AppVersion == "" {
		AppVersion = None
	}
	if GitCommit == "" {
		GitCommit = "unknown"
	}
	if GoHostOS == "" {
		GoHostOS = "unknown"
	}
	if GoHostArch == "" {
		GoHostArch = "unknown"
	}
	if GoVersion == "" {
		// runtime.Version() provides the Go tree's version string at build time
		// the other values like OS and Arch aren't accessible and are passed in
		// separately
		GoVersion = fmt.Sprintf("go version %s %s/%s", runtime.Version(), GoHostOS, GoHostArch)
	}
	if Environment == "" {
		Environment = "development"
	}
}

// SemVer accepts the application revision version, which is prefixed with a
// `v` and also has a commit hash following the semantic version, and returns
// just the semantic version.
//
// e.g. `v1.0.0-xyz` --> `1.0.0`.
func SemVer(av string) string {
	av = strings.TrimPrefix(av, "v")
	seg := strings.Split(av, "-")

	return seg[0]
}
