// Package revision defines variables that will be populated with values from
// the Makefile at build time via LDFLAGS.
package revision

import "strings"

var (
	// AppVersion is the semver for this version of the client, or
	// "v0.0.0-unknown". Set by `make release`.
	AppVersion string

	// GitCommit is the short git SHA associated with this build, or
	// "unknown". Set by `make release`.
	GitCommit string

	// GoVersion is the output of `go version` associated with this build, or
	// "go version unknown". Set by `make release`.
	GoVersion string
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
	if GoVersion == "" {
		GoVersion = "go version unknown"
	}
}

// SemVer accepts the application revision version, which is prefixed with a
// `v` and also has a commit hash following the semantic version, and returns
// just the semantic version.
//
// e.g. v1.0.0-xyz --> 1.0.0
func SemVer(av string) string {
	av = strings.TrimPrefix(av, "v")
	seg := strings.Split(av, "-")

	return seg[0]
}
