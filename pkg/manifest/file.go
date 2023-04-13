package manifest

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
)

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

	quiet     bool
	errLog    fsterr.LogInterface
	exists    bool
	output    io.Writer
	readError error
}

// AutoMigrateVersion updates the manifest_version value to
// ManifestLatestVersion if the current version is less than the latest
// supported and only if there is no [setup] or [local_server] configuration defined.
//
// NOTE: It contains similar conversions to the custom Version.UnmarshalText().
// Specifically, it type switches the any into various types before
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
			return nil, fmt.Errorf("error marshalling modified manifest_version fastly.toml: %w", err)
		}

		// NOTE: The scenario will end up triggering two calls to toml.Unmarshal().
		// The first call here, then a second call inside of the File.Read() caller.
		// This only happens once. All future file reads result in one Unmarshal.
		err = toml.Unmarshal(data, f)
		if err != nil {
			return data, fmt.Errorf("error unmarshaling fastly.toml: %w", err)
		}

		if err = f.Write(path); err != nil {
			return data, fsterr.ErrIncompatibleManifestVersion
		}

		return data, nil
	}

	return data, fsterr.ErrIncompatibleManifestVersion
}

// Exists yields whether the manifest exists.
//
// Specifically, it indicates that a toml.Unmarshal() of the toml disk content
// to data in memory was successful without error.
func (f *File) Exists() bool {
	return f.exists
}

// Load parses the input data into the File struct and persists it to disk.
//
// NOTE: This is used by the `compute build` command logic.
// Which has to modify the toml tree for supporting a v4.0.0 migration path.
// e.g. if user manifest is missing [scripts.build] then add a default value.
func (f *File) Load(data []byte) error {
	err := toml.Unmarshal(data, f)
	if err != nil {
		return fmt.Errorf("error unmarshaling fastly.toml: %w", err)
	}
	return f.Write(Filename)
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
		f.logErr(err)
		return err
	}

	// The AutoMigrateVersion() method will either return the []byte unmodified or
	// it will have updated the manifest_version field to reflect the latest
	// version supported by the Fastly CLI.
	data, err = f.AutoMigrateVersion(data, path)
	if err != nil {
		f.logErr(err)
		return err
	}

	err = toml.Unmarshal(data, f)
	if err != nil {
		f.logErr(err)
		return fsterr.ErrParsingManifest
	}

	if f.ManifestVersion == 0 {
		f.ManifestVersion = ManifestLatestVersion

		if !f.quiet {
			text.Warning(f.output, fmt.Sprintf("The fastly.toml was missing a `manifest_version` field. A default schema version of `%d` will be used.", ManifestLatestVersion))
			text.Break(f.output)
			text.Output(f.output, fmt.Sprintf("Refer to the fastly.toml package manifest format: %s", SpecURL))
			text.Break(f.output)
		}
		err = f.Write(path)
		if err != nil {
			f.logErr(err)
			return fmt.Errorf("unable to save fastly.toml manifest change: %w", err)
		}
	}

	f.exists = true

	return nil
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

// SetQuiet sets the associated flag value.
func (f *File) SetQuiet(v bool) {
	f.quiet = v
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

	return fp.Close()
}

func (f *File) logErr(err error) {
	if f.errLog != nil {
		f.errLog.Add(err)
	}
}

// appendSpecRef appends the fastly.toml specification URL to the manifest.
func appendSpecRef(w io.Writer) error {
	s := fmt.Sprintf("# %s\n# %s\n\n", SpecIntro, SpecURL)
	_, err := io.WriteString(w, s)
	return err
}

// Scripts represents build configuration.
type Scripts struct {
	Build     string `toml:"build,omitempty"`
	PostBuild string `toml:"post_build,omitempty"`
}
