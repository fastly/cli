package manifest

import (
	"fmt"
	"strconv"
	"strings"

	fsterr "github.com/fastly/cli/pkg/errors"
)

// Version represents the currently supported schema for the fastly.toml
// manifest file that determines the configuration for a Compute service.
//
// NOTE: the File object has a field called ManifestVersion which this type is
// assigned. The reason we don't name this type ManifestVersion is to appease
// the static analysis linter which complains re: stutter in the import
// manifest.ManifestVersion.
type Version int

// UnmarshalText manages multiple scenarios where historically the manifest
// version was a string value and not an integer.
//
// Example mappings:
//
// "0.1.0" -> 1
// "1"     -> 1
// 1       -> 1
// "1.0.0" -> 1
// 0.1     -> 1
// "0.2.0" -> 1
// "2.0.0" -> 2
//
// We also constrain the version so that if a user has a manifest_version
// defined as "99.0.0" then we won't accidentally store it as the integer 99
// but instead will return an error because it exceeds the current
// ManifestLatestVersion version.
func (v *Version) UnmarshalText(txt []byte) error {
	s := string(txt)

	// Presumes semver value (e.g. 1.0.0, 0.1.0 or 0.1)
	// Major is converted to integer if != zero.
	// Otherwise if Major == zero, then ignore Minor/Patch and set to latest version.
	var (
		err     error
		version int
	)
	if strings.Contains(s, ".") {
		segs := strings.Split(s, ".")
		s = segs[0]
		if s == "0" {
			s = strconv.Itoa(ManifestLatestVersion)
		}
	}

	version, err = strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("error parsing manifest_version '%s': %w", s, err)
	}

	if version > ManifestLatestVersion {
		return fsterr.ErrUnrecognisedManifestVersion
	}

	*v = Version(version)
	return nil
}
