package manifest

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
)

// Source enumerates where a manifest parameter is taken from.
type Source uint8

const (
	// Filename is the name of the package manifest file.
	// It is expected to be a project specific configuration file.
	Filename = "fastly.toml"

	// ManifestLatestVersion represents the latest known manifest schema version
	// supported by the CLI.
	ManifestLatestVersion = 2

	// FilePermissions represents a read/write file mode.
	FilePermissions = 0666

	// SourceUndefined indicates the parameter isn't provided in any of the
	// available sources, similar to "not found".
	SourceUndefined Source = iota

	// SourceFile indicates the parameter came from a manifest file.
	SourceFile

	// SourceEnv indicates the parameter came from the user's shell environment.
	SourceEnv

	// SourceFlag indicates the parameter came from an explicit flag.
	SourceFlag

	// SpecIntro informs the user of what the manifest file is for.
	SpecIntro = "This file describes a Fastly Compute@Edge package. To learn more visit:"

	// SpecURL points to the fastly.toml manifest specification reference.
	SpecURL = "https://developer.fastly.com/reference/fastly-toml/"
)

var once sync.Once

// Data holds global-ish manifest data from manifest files, and flag sources.
// It has methods to give each parameter to the components that need it,
// including the place the parameter came from, which is a requirement.
//
// If the same parameter is defined in multiple places, it is resolved according
// to the following priority order: the manifest file (lowest priority) and then
// explicit flags (highest priority).
type Data struct {
	File File
	Flag Flag
}

// Name yields a Name.
func (d *Data) Name() (string, Source) {
	if d.Flag.Name != "" {
		return d.Flag.Name, SourceFlag
	}

	if d.File.Name != "" {
		return d.File.Name, SourceFile
	}

	return "", SourceUndefined
}

// ServiceID yields a ServiceID.
func (d *Data) ServiceID() (string, Source) {
	if d.Flag.ServiceID != "" {
		return d.Flag.ServiceID, SourceFlag
	}

	if sid := os.Getenv(env.ServiceID); sid != "" {
		return sid, SourceEnv
	}

	if d.File.ServiceID != "" {
		return d.File.ServiceID, SourceFile
	}

	return "", SourceUndefined
}

// Description yields a Description.
func (d *Data) Description() (string, Source) {
	if d.Flag.Description != "" {
		return d.Flag.Description, SourceFlag
	}

	if d.File.Description != "" {
		return d.File.Description, SourceFile
	}

	return "", SourceUndefined
}

// Authors yields an Authors.
func (d *Data) Authors() ([]string, Source) {
	if len(d.Flag.Authors) > 0 {
		return d.Flag.Authors, SourceFlag
	}

	if len(d.File.Authors) > 0 {
		return d.File.Authors, SourceFile
	}

	return []string{}, SourceUndefined
}

// Version represents the currently supported schema for the fastly.toml
// manifest file that determines the configuration for a compute@edge service.
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
func (v *Version) UnmarshalText(text []byte) error {
	s := string(text)

	if i, err := strconv.Atoi(s); err == nil {
		*v = Version(i)
		return nil
	}

	if f, err := strconv.ParseFloat(s, 32); err == nil {
		intfl := int(f)
		if intfl == 0 {
			*v = 1
		} else {
			*v = Version(intfl)
		}
		return nil
	}

	if strings.Contains(s, ".") {
		segs := strings.Split(s, ".")

		// A length of 3 presumes a semver (e.g. 0.1.0)
		if len(segs) == 3 {
			if segs[0] != "0" {
				if i, err := strconv.Atoi(segs[0]); err == nil {
					if i > ManifestLatestVersion {
						return fsterr.ErrUnrecognisedManifestVersion
					}
					*v = Version(i)
					return nil
				}
			} else {
				*v = 1
				return nil
			}
		}
	}

	return fsterr.ErrUnrecognisedManifestVersion
}

// File represents all of the configuration parameters in the fastly.toml
// manifest file schema.
type File struct {
	ManifestVersion Version     `toml:"manifest_version"`
	Name            string      `toml:"name"`
	Description     string      `toml:"description"`
	Authors         []string    `toml:"authors"`
	Language        string      `toml:"language"`
	ServiceID       string      `toml:"service_id"`
	LocalServer     LocalServer `toml:"local_server,omitempty"`
	Setup           Setup       `toml:"setup,omitempty"`

	exists bool
	output io.Writer
}

// Setup represents a set of service configuration that works with the code in
// the package. See https://developer.fastly.com/reference/fastly-toml/.
type Setup struct {
	Backends     map[string]*SetupBackend   `toml:"backends,omitempty"`
	Dictionaries map[string]SetupDictionary `toml:"dictionaries,omitempty"`
}

// SetupBackend represents a '[setup.backends.<T>]' instance.
type SetupBackend struct {
	Address string
	Port    uint
	Prompt  string
	Exists  bool // not defined in fastly.toml but used internally by the CLI.
}

// SetupDictionary represents a '[setup.dictionaries.<T>]' instance.
type SetupDictionary struct {
	Items  map[string]SetupDictionaryItems
	Prompt string
}

// SetupDictionaryItems represents a '[setup.dictionaries.<T>.items]' instance.
type SetupDictionaryItems struct {
	Type  string
	Value string
}

// LocalServer represents a list of backends that should be mocked as per the
// configuration values.
type LocalServer struct {
	Backends     map[string]LocalBackend    `toml:"backends"`
	Dictionaries map[string]LocalDictionary `toml:"dictionaries,omitempty"`
}

// LocalBackend represents a backend to be mocked by the local testing server.
type LocalBackend struct {
	URL          string `toml:"url"`
	OverrideHost string `toml:"override_host,omitempty"`
}

// LocalDictionary represents a dictionary to be mocked by the local testing server.
type LocalDictionary struct {
	File   string `toml:"file"`
	Format string `toml:"format"`
}

// Exists yields whether the manifest exists.
func (f *File) Exists() bool {
	return f.exists
}

// SetOutput sets the output stream for any messages.
func (f *File) SetOutput(output io.Writer) {
	f.output = output
}

// Read loads the manifest file content from disk.
func (f *File) Read(fpath string) error {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable.
	// Disabling as we need to load the fastly.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	/* #nosec */
	bs, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}

	// NOTE: temporary fix needed because of a bug that appeared in v0.25.0 where
	// the manifest_version was stored in fastly.toml as a 'section', e.g.
	// `[manifest_version]`.
	//
	// This subsequently would cause errors when trying to unmarshal the data, so
	// we need to identify if it exists in the file (as a section) and remove it.
	//
	// We do this before trying to unmarshal the toml data into a go data
	// structure otherwise we'll see errors from the toml library.
	manifestSection, err := containsManifestSection(bs)
	if err != nil {
		return fmt.Errorf("failed to parse the fastly.toml manifest: %w", err)
	}

	if manifestSection {
		buf, err := stripManifestSection(bytes.NewReader(bs), fpath)
		if err != nil {
			return fsterr.ErrInvalidManifestVersion
		}
		bs = buf.Bytes()
	}

	manifestVersion, err := getManifestVersion(bs)
	if err != nil || manifestVersion != ManifestLatestVersion {
		return errors.ErrIncompatibleManifestVersion
	}

	err = toml.Unmarshal(bs, f)
	if err != nil {
		// NOTE: The toml library messes with the returned error so when we use
		// fsterrors.Deduce(err).Print(os.Stderr) to determine the remediation
		// error we actually fail to find a match and end up using a BugRemediation
		// as a default (losing important information for the user).
		//
		// To work around this we type assert the underlying error, and if a match
		// is found we return the specific remediation error, otherwise we return
		// whatever error the toml library gave us.
		//
		// We also need to declare the errtype variable explicitly as being of an
		// `error` type for errors.As() to correctly report the underlying error
		// type. This is because our remediation errors are actually a struct type
		// that implement the error interface.
		var errtype error = fsterr.ErrUnrecognisedManifestVersion
		if errors.As(err, &errtype) {
			return fsterr.ErrUnrecognisedManifestVersion
		}

		return err
	}

	f.exists = true

	if f.ManifestVersion == 0 {
		f.ManifestVersion = ManifestLatestVersion

		// NOTE: the use of once is a quick-fix to side-step duplicate outputs.
		// To fix this properly will require a refactor of the structure of how our
		// global output is passed around.
		once.Do(func() {
			text.Warning(f.output, fmt.Sprintf("The fastly.toml was missing a `manifest_version` field. A default schema version of `%d` will be used.", ManifestLatestVersion))
			text.Break(f.output)
			text.Output(f.output, fmt.Sprintf("Refer to the fastly.toml package manifest format: %s", SpecURL))
			text.Break(f.output)
			f.Write(fpath)
		})
	}

	return nil
}

// containsManifestSection loads the slice of bytes into a toml tree structure
// before checking if the manifest_version is defined as a toml section block.
func containsManifestSection(bs []byte) (bool, error) {
	tree, err := toml.LoadBytes(bs)
	if err != nil {
		return false, err
	}

	if _, ok := tree.GetArray("manifest_version").(*toml.Tree); ok {
		return true, nil
	}

	return false, nil
}

// stripManifestSection reads the manifest line-by-line storing the lines that
// don't contain `[manifest_version]` into a buffer to be written back to disk.
//
// It would've been better if we could have relied on the toml library to delete
// the section but unfortunately that means it would end up deleting the entire
// block and not just the key specified. Meaning if the manifest_version key
// was in the middle of the manifest with other keys below it, deleting the
// manifest_version would cause all keys below it to be deleted as they would
// all be considered part of that section block.
func stripManifestSection(r io.Reader, fpath string) (*bytes.Buffer, error) {
	var bs []byte
	buf := bytes.NewBuffer(bs)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if scanner.Text() != "[manifest_version]" {
			_, err := buf.Write(scanner.Bytes())
			if err != nil {
				return buf, err
			}
			_, err = buf.WriteString("\n")
			if err != nil {
				return buf, err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return buf, err
	}

	err := os.WriteFile(fpath, buf.Bytes(), FilePermissions)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

// getManifestVersion returns the version number.
func getManifestVersion(bs []byte) (int64, error) {
	tree, err := toml.LoadBytes(bs)
	if err != nil {
		return 0, err
	}

	if version, ok := tree.GetArray("manifest_version").(int64); ok {
		return version, nil
	}

	return 0, nil
}

// Write persists the manifest content to disk.
func (f *File) Write(fpath string) error {
	fp, err := os.Create(fpath)
	if err != nil {
		return err
	}

	if err := appendSpecRef(fp); err != nil {
		return err
	}

	if err := toml.NewEncoder(fp).Encode(f); err != nil {
		return err
	}

	if err := fp.Sync(); err != nil {
		return err
	}

	if err := fp.Close(); err != nil {
		return err
	}

	return nil
}

// appendSpecRef appends the fastly.toml specification URL to the manifest.
func appendSpecRef(w io.Writer) error {
	s := fmt.Sprintf("# %s\n# %s\n\n", SpecIntro, SpecURL)
	_, err := io.WriteString(w, s)
	if err != nil {
		return err
	}
	return nil
}

// Flag represents all of the manifest parameters that can be set with explicit
// flags. Consumers should bind their flag values to these fields directly.
type Flag struct {
	Name        string
	Description string
	Authors     []string
	ServiceID   string
}
