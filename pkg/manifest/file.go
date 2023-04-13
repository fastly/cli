package manifest

import (
	"fmt"
	"io"
	"os"

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
	tree, err := toml.LoadFile(path)
	if err != nil {
		f.logErr(err)
		return err
	}

	err = tree.Unmarshal(f)
	if err != nil {
		f.logErr(err)
		return err
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

	if f.ManifestVersion < ManifestLatestVersion {
		return fsterr.ErrIncompatibleManifestVersion
	}

	if dt := tree.Get("setup.dictionaries"); dt != nil {
		text.Warning(f.output, "Your fastly.toml manifest contains `[setup.dictionaries]`, which should be updated to `[setup.config_stores]`. Refer to the documentation at https://developer.fastly.com/reference/compute/fastly-toml/")
		text.Break(f.output)
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
