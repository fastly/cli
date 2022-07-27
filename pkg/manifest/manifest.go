package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

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
	//
	// TODO: The filename needs to be referenced outside of the compute package
	// so consider moving this constant to a different location in the top-level
	// pkg directory instead or even moving the whole manifest package up to the
	// top-level pkg directory as finding a suitable package for just the
	// manifest filename could be tricky.
	Filename = "fastly.toml"

	// ManifestLatestVersion represents the latest known manifest schema version
	// supported by the CLI.
	//
	// NOTE: The CLI is the primary consumer of the fastly.toml manifest so its
	// code is typically coupled to the specification.
	ManifestLatestVersion = 2

	// FilePermissions represents a read/write file mode.
	FilePermissions = 0o666

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
			*v = ManifestLatestVersion
		} else {
			*v = Version(intfl)
		}
		return nil
	}

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
		return fmt.Errorf("error parsing manifest_version: %w", err)
	}

	if version > ManifestLatestVersion {
		return fsterr.ErrUnrecognisedManifestVersion
	}
	*v = Version(version)
	return nil
}

// File represents all of the configuration parameters in the fastly.toml
// manifest file schema.
type File struct {
	Authors         []string    `toml:"authors"`
	Description     string      `toml:"description"`
	Language        string      `toml:"language"`
	Profile         string      `toml:"profile,omitempty"`
	LocalServer     LocalServer `toml:"local_server,omitempty"`
	ManifestVersion Version     `toml:"manifest_version"`
	Name            string      `toml:"name"`
	Scripts         Scripts     `toml:"scripts,omitempty"`
	ServiceID       string      `toml:"service_id"`
	Setup           Setup       `toml:"setup,omitempty"`

	errLog    fsterr.LogInterface
	exists    bool
	output    io.Writer
	readError error
}

// Scripts represents custom operations.
type Scripts struct {
	Build     string `toml:"build,omitempty"`
	PostBuild string `toml:"post_build,omitempty"`
}

// Setup represents a set of service configuration that works with the code in
// the package. See https://developer.fastly.com/reference/fastly-toml/.
type Setup struct {
	Backends     map[string]*SetupBackend    `toml:"backends,omitempty"`
	Dictionaries map[string]*SetupDictionary `toml:"dictionaries,omitempty"`
	Loggers      map[string]*SetupLogger     `toml:"log_endpoints,omitempty"`
}

// SetupBackend represents a '[setup.backends.<T>]' instance.
type SetupBackend struct {
	Address     string `toml:"address,omitempty"`
	Port        uint   `toml:"port,omitempty"`
	Description string `toml:"description,omitempty"`
}

// SetupDictionary represents a '[setup.dictionaries.<T>]' instance.
type SetupDictionary struct {
	Items       map[string]SetupDictionaryItems `toml:"items,omitempty"`
	Description string                          `toml:"description,omitempty"`
}

// SetupDictionaryItems represents a '[setup.dictionaries.<T>.items]' instance.
type SetupDictionaryItems struct {
	Value       string `toml:"value,omitempty"`
	Description string `toml:"description,omitempty"`
}

// SetupLogger represents a '[setup.log_endpoints.<T>]' instance.
type SetupLogger struct {
	Provider string `toml:"provider,omitempty"`
}

// LocalServer represents a list of mocked Viceroy resources.
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
	File     string            `toml:"file,omitempty"`
	Format   string            `toml:"format"`
	Contents map[string]string `toml:"contents,omitempty"`
}

// Exists yields whether the manifest exists.
//
// Specifically, it indicates that a toml.Unmarshal() of the toml disk content
// to data in memory was successful without error.
func (f *File) Exists() bool {
	return f.exists
}

// ReadError yields the error returned from Read().
//
// NOTE: We no longer call Read() from every command. We only call it once
// within app.Run() but we don't handle any errors that are returned from the
// Read() method. This is because failing to read the manifest is fine if the
// error is caused by the file not existing in a directory where the user is
// working on a non-C@E project. This will enable code elsewhere in the CLI to
// understand why the Read() failed. For example, we can use errors.Is() to
// allow returning a specific remediation error from a C@E related command.
func (f *File) ReadError() error {
	return f.readError
}

// SetErrLog sets an instance of errors.LogInterface.
func (f *File) SetErrLog(errLog fsterr.LogInterface) {
	f.errLog = errLog
}

// SetOutput sets the output stream for any messages.
func (f *File) SetOutput(output io.Writer) {
	f.output = output
}

// Read loads the manifest file content from disk.
func (f *File) Read(path string) (err error) {
	defer func() {
		if err != nil {
			f.readError = err
		}
	}()

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable.
	// Disabling as we need to load the fastly.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	/* #nosec */
	data, err := os.ReadFile(path)
	if err != nil {
		f.errLog.Add(err)
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
	manifestSection, err := containsManifestSection(data)
	if err != nil {
		f.errLog.Add(err)
		return fmt.Errorf("failed to parse the fastly.toml manifest: %w", err)
	}

	if manifestSection {
		buf, err := stripManifestSection(bytes.NewReader(data), path)
		if err != nil {
			f.errLog.Add(err)
			return fsterr.ErrInvalidManifestVersion
		}
		data = buf.Bytes()
	}

	// The AutoMigrateVersion() method will either return the []byte unmodified or
	// it will have updated the manifest_version field to reflect the latest
	// version supported by the Fastly CLI.
	data, err = f.AutoMigrateVersion(data, path)
	if err != nil {
		f.errLog.Add(err)
		return err
	}

	err = toml.Unmarshal(data, f)
	if err != nil {
		f.errLog.Add(err)
		return fsterr.ErrParsingManifest
	}

	f.exists = true

	if f.ManifestVersion == 0 {
		f.ManifestVersion = ManifestLatestVersion

		text.Warning(f.output, fmt.Sprintf("The fastly.toml was missing a `manifest_version` field. A default schema version of `%d` will be used.", ManifestLatestVersion))
		text.Break(f.output)
		text.Output(f.output, fmt.Sprintf("Refer to the fastly.toml package manifest format: %s", SpecURL))
		text.Break(f.output)
		f.Write(path)
	}

	return nil
}

// AutoMigrateVersion updates the manifest_version value to
// ManifestLatestVersion if the current version is less than the latest
// supported and only if there is no [setup] configuration defined.
//
// NOTE: It contains similar conversions to the custom Version.UnmarshalText().
// Specifically, it type switches the interface{} into various types before
// attempting to convert the underlying value into an integer.
func (f *File) AutoMigrateVersion(data []byte, path string) ([]byte, error) {
	tree, err := toml.LoadBytes(data)
	if err != nil {
		return data, err
	}

	// If there is no manifest_version set then we return the fastly.toml content
	// unmodified, along with a nil error, so that logic further down the .Read()
	// method will pick up that the unmarshalled data structure will have a zero
	// value of 0 for the ManifestVersion field and so will display a message to
	// the user to inform them that we'll default to setting a manifest_version to
	// the ManifestLatestVersion value.
	i := tree.GetArray("manifest_version")
	if i == nil {
		return data, nil
	}

	setup := tree.GetArray("setup")

	var version int
	switch v := i.(type) {
	case int64:
		version = int(v)
	case float64:
		version = int(v)
	case string:
		if strings.Contains(v, ".") {
			// Presumes semver value (e.g. 1.0.0, 0.1.0 or 0.1)
			// Major is converted to integer if != zero.
			// Otherwise if Major == zero, then ignore Minor/Patch and set to latest version.
			segs := strings.Split(v, ".")
			v = segs[0]
		}
		version, err = strconv.Atoi(v)
		if err != nil {
			return data, fmt.Errorf("error parsing manifest_version: %w", err)
		}
	default:
		return data, fmt.Errorf("error parsing manifest_version: unrecognised type")
	}

	// User is on the latest version supported by the CLI, so we'll return the
	// []byte with the manifest_version field unmodified.
	if version == ManifestLatestVersion {
		return data, nil
	}

	// User has an unrecognised manifest_version specified.
	if version > ManifestLatestVersion {
		return data, fsterr.ErrUnrecognisedManifestVersion
	}

	// User has manifest_version less than latest supported by CLI, but as they
	// don't have a [setup] configuration block defined, it means we can
	// automatically update their fastly.toml file's manifest_version field.
	//
	// NOTE: Inside this block we also update the data variable so it contains the
	// updated manifest_version field too, and that is returned at the end of
	// the function block.
	if setup == nil {
		tree.Set("manifest_version", int64(ManifestLatestVersion))

		data, err = tree.Marshal()
		if err != nil {
			return data, fmt.Errorf("error marshalling modified manifest_version fastly.toml: %w", err)
		}

		// NOTE: The scenario will end up triggering two calls to toml.Unmarshal().
		// The first call here, then a second call inside of the File.Read() caller.
		// This only happens once. All future file reads result in one Unmarshal.
		err = toml.Unmarshal(data, f)
		if err != nil {
			return data, fmt.Errorf("error unmarshalling fastly.toml: %w", err)
		}

		if err = f.Write(path); err != nil {
			return data, fsterr.ErrIncompatibleManifestVersion
		}

		return data, nil
	}

	return data, fsterr.ErrIncompatibleManifestVersion
}

// containsManifestSection loads the slice of bytes into a toml tree structure
// before checking if the manifest_version is defined as a toml section block.
func containsManifestSection(data []byte) (bool, error) {
	tree, err := toml.LoadBytes(data)
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
func stripManifestSection(r io.Reader, path string) (*bytes.Buffer, error) {
	var data []byte
	buf := bytes.NewBuffer(data)

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

	err := os.WriteFile(path, buf.Bytes(), FilePermissions)
	if err != nil {
		return buf, err
	}

	return buf, nil
}

// Write persists the manifest content to disk.
func (f *File) Write(path string) error {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as in most cases this is provided by a static constant embedded
	// from the 'manifest' package, and in other cases we want the user to be
	// able to provide a custom path to their fastly.toml manifest.
	/* #nosec */
	fp, err := os.Create(path)
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
