// Package viceroy carries an optional Viceroy executable embedded directly
// into the fastly CLI binary. When built with the `viceroy_embed` build tag
// and an asset is available for the target platform, the binary is extracted
// to disk on first use and exec'd from there. Without the tag, or on a
// platform that has no shipped Viceroy asset, the package reports
// Supported() == false and callers should fall back to downloading Viceroy
// at runtime.
package viceroy

import (
	_ "embed"
	"errors"
	"strings"
)

//go:embed VICEROY_VERSION
var versionRaw string

// ErrUnsupported is returned by Extract when this build does not carry a
// Viceroy binary for the current platform.
var ErrUnsupported = errors.New("viceroy: no embedded binary for this platform")

// Version reports the Viceroy version pinned at CLI build time. It returns
// the contents of pkg/embedded/viceroy/VICEROY_VERSION regardless of whether
// an asset is actually embedded for the current platform, so callers can use
// it for logging and comparison without first consulting Supported.
func Version() string {
	return strings.TrimSpace(versionRaw)
}

// Supported reports whether this build carries a usable embedded Viceroy
// for the current GOOS/GOARCH. It returns false when the binary was built
// without -tags viceroy_embed, or when the tag is set but no asset exists
// for the target platform.
func Supported() bool {
	return platformSupported && len(binaryZstd) > 0
}
