package manifest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// File represents all of the configuration parameters in the fastly.toml
// manifest file schema.
type File struct {
	// Args is necessary to track the subcommand called (see: File.Read method).
	Args []string `toml:"-"`
	// Authors is a list of project authors (typically an email).
	Authors []string `toml:"authors"`
	// ClonedFrom indicates the GitHub repo the starter kit was cloned from.
	// This could be an empty value if the user doesn't use `compute init`.
	ClonedFrom string `toml:"cloned_from,omitempty"`
	// Description is the project description.
	Description string `toml:"description"`
	// Language is the programming language used for the project.
	Language string `toml:"language"`
	// Profile is the name of the profile account the Fastly CLI should use to make API requests.
	Profile string `toml:"profile,omitempty"`
	// LocalServer describes the configuration for the local server built into the Fastly CLI.
	LocalServer LocalServer `toml:"local_server,omitempty"`
	// ManifestVersion is the manifest schema version number.
	ManifestVersion Version `toml:"manifest_version"`
	// Name is the package name.
	Name string `toml:"name"`
	// Scripts describes customisation options for the Fastly CLI build step.
	Scripts Scripts `toml:"scripts,omitempty"`
	// ServiceID is the Fastly Service ID to deploy the package to.
	ServiceID string `toml:"service_id"`
	// Setup describes a set of service configuration that works with the code in the package.
	Setup Setup `toml:"setup,omitempty"`

	quiet     bool
	errLog    fsterr.LogInterface
	exists    bool
	output    io.Writer
	readError error
}

// MarshalTOML performs custom marshalling to TOML for objects of File type.
func (f *File) MarshalTOML() ([]byte, error) {
	localServer := make(map[string]any)

	if f.LocalServer.Backends != nil {
		localServer["backends"] = f.LocalServer.Backends
	}

	if f.LocalServer.ConfigStores != nil {
		localServer["config_stores"] = f.LocalServer.ConfigStores
	}

	if f.LocalServer.KVStores != nil {
		kvStores := make(map[string]any)
		for key, entry := range f.LocalServer.KVStores {
			if entry.External != nil {
				kvStores[key] = map[string]any{
					"file":   entry.External.File,
					"format": entry.External.Format,
				}
			} else {
				items := make([]map[string]any, 0, len(entry.Array))
				for _, e := range entry.Array {
					obj := map[string]any{"key": e.Key}
					if e.File != "" {
						obj["file"] = e.File
					}
					if e.Data != "" {
						obj["data"] = e.Data
					}
					if e.Metadata != "" {
						obj["metadata"] = e.Metadata
					}
					items = append(items, obj)
				}
				kvStores[key] = items
			}
		}
		localServer["kv_stores"] = kvStores
	}

	if f.LocalServer.Pushpin != nil {
		pushpin := make(map[string]any)
		if f.LocalServer.Pushpin.EnablePushpin != nil {
			pushpin["enable"] = *f.LocalServer.Pushpin.EnablePushpin
		}
		if f.LocalServer.Pushpin.PushpinPath != nil {
			pushpin["pushpin_path"] = *f.LocalServer.Pushpin.PushpinPath
		}
		if f.LocalServer.Pushpin.PushpinProxyPort != nil {
			pushpin["proxy_port"] = *f.LocalServer.Pushpin.PushpinProxyPort
		}
		if f.LocalServer.Pushpin.PushpinPublishPort != nil {
			pushpin["publish_port"] = *f.LocalServer.Pushpin.PushpinPublishPort
		}
		localServer["pushpin"] = pushpin
	}

	if f.LocalServer.SecretStores != nil {
		secretStores := make(map[string]any)
		for key, entry := range f.LocalServer.SecretStores {
			if entry.External != nil {
				secretStores[key] = map[string]any{
					"file":   entry.External.File,
					"format": entry.External.Format,
				}
			} else {
				items := make([]map[string]any, 0, len(entry.Array))
				for _, e := range entry.Array {
					obj := map[string]any{"key": e.Key}
					if e.File != "" {
						obj["file"] = e.File
					}
					if e.Data != "" {
						obj["data"] = e.Data
					}
					if e.Env != "" {
						obj["env"] = e.Env
					}
					items = append(items, obj)
				}
				secretStores[key] = items
			}
		}
		localServer["secret_stores"] = secretStores
	}

	if f.LocalServer.ViceroyVersion != "" {
		localServer["viceroy_version"] = f.LocalServer.ViceroyVersion
	}

	out := struct {
		Authors         []string `toml:"authors"`
		ClonedFrom      string   `toml:"cloned_from,omitempty"`
		Description     string   `toml:"description"`
		Language        string   `toml:"language"`
		Profile         string   `toml:"profile,omitempty"`
		LocalServer     any      `toml:"local_server"` // override this field
		ManifestVersion Version  `toml:"manifest_version"`
		Name            string   `toml:"name"`
		Scripts         Scripts  `toml:"scripts,omitempty"`
		ServiceID       string   `toml:"service_id"`
		Setup           Setup    `toml:"setup,omitempty"`
	}{
		Authors:         f.Authors,
		ClonedFrom:      f.ClonedFrom,
		Description:     f.Description,
		Language:        f.Language,
		Profile:         f.Profile,
		LocalServer:     localServer,
		ManifestVersion: f.ManifestVersion,
		Name:            f.Name,
		Scripts:         f.Scripts,
		ServiceID:       f.ServiceID,
		Setup:           f.Setup,
	}

	var buf bytes.Buffer
	err := toml.NewEncoder(&buf).Encode(out)
	return buf.Bytes(), err
}

// Exists yields whether the manifest exists.
//
// Specifically, it indicates that a toml.Unmarshal() of the toml disk content
// to data in memory was successful without error.
func (f *File) Exists() bool {
	return f.exists
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
	// #nosec
	tree, err := toml.LoadFile(path)
	if err != nil {
		// IMPORTANT: Only `fastly compute` references the fastly.toml file.
		if len(f.Args) > 0 && f.Args[0] == "compute" {
			f.logErr(err) // only log error if user executed `compute` subcommand.
		}
		return err
	}

	err = tree.Unmarshal(f)
	if err != nil {
		// IMPORTANT: go-toml consumes our error type within its own.
		//
		// This means we need to manually parse the return error to see if it
		// contains our specific error message. If we don't do this, then the
		// remediation information we pass back will be lost and a generic 'bug'
		// remediation (which is set by logic in main.go) is used instead.
		if strings.Contains(err.Error(), fsterr.ErrUnrecognisedManifestVersion.Inner.Error()) {
			err = fsterr.ErrUnrecognisedManifestVersion
		}
		f.logErr(err)
		return err
	}

	if f.Scripts.EnvFile != "" {
		if err := f.ParseEnvFile(); err != nil {
			return err
		}
	}

	if f.ManifestVersion == 0 {
		f.ManifestVersion = ManifestLatestVersion

		if !f.quiet {
			text.Warning(f.output, fmt.Sprintf("The fastly.toml was missing a `manifest_version` field. A default schema version of `%d` will be used.\n\n", ManifestLatestVersion))
			text.Output(f.output, fmt.Sprintf("Refer to the fastly.toml package manifest format: %s\n\n", SpecURL))
		}
		err = f.Write(path)
		if err != nil {
			f.logErr(err)
			return fmt.Errorf("unable to save fastly.toml manifest change: %w", err)
		}
	}

	if dt := tree.Get("setup.dictionaries"); dt != nil {
		text.Warning(f.output, "Your fastly.toml manifest contains `[setup.dictionaries]`, which should be updated to `[setup.config_stores]`. Refer to the documentation at https://www.fastly.com/documentation/reference/compute/fastly-toml\n\n")
	}

	f.exists = true
	return nil
}

// ParseEnvFile reads the environment file `env_file` and appends all KEY=VALUE
// pairs to the existing `f.Scripts.EnvVars`.
func (f *File) ParseEnvFile() error {
	// IMPORTANT: Avoid persisting potentially secret values to disk.
	// We do this by keeping a copy of EnvVars before they're appended to.
	// Inside of File.Write() we'll reassign EnvVars the original values.
	manifestDefinedEnvVars := make([]string, len(f.Scripts.EnvVars))
	copy(manifestDefinedEnvVars, f.Scripts.EnvVars)
	f.Scripts.manifestDefinedEnvVars = manifestDefinedEnvVars

	path, err := filepath.Abs(f.Scripts.EnvFile)
	if err != nil {
		return fmt.Errorf("failed to generate absolute path for '%s': %w", f.Scripts.EnvFile, err)
	}
	r, err := os.Open(path) // #nosec G304 (CWE-22)
	if err != nil {
		return fmt.Errorf("failed to open path '%s': %w", path, err)
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), "=")
		if len(parts) != 2 {
			return fmt.Errorf("failed to scan env_file '%s': invalid KEY=VALUE format: %#v", path, parts)
		}
		parts[1] = strings.Trim(parts[1], `"'`)
		f.Scripts.EnvVars = append(f.Scripts.EnvVars, strings.Join(parts, "="))
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan env_file '%s': %w", path, err)
	}
	return nil
}

// ReadError yields the error returned from Read().
//
// NOTE: We no longer call Read() from every command. We only call it once
// within app.Run() but we don't handle any errors that are returned from the
// Read() method. This is because failing to read the manifest is fine if the
// error is caused by the file not existing in a directory where the user is
// working on a non-Compute project. This will enable code elsewhere in the CLI to
// understand why the Read() failed. For example, we can use errors.Is() to
// allow returning a specific remediation error from a Compute related command.
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
	fp, err := os.Create(path) // #nosec G304 (CWE-22)
	if err != nil {
		return err
	}
	if err := appendSpecRef(fp); err != nil {
		return err
	}

	// IMPORTANT: Avoid persisting potentially secret values to disk.
	// We do this by keeping a copy of EnvVars before they're appended to.
	// i.e. f.Scripts.manifestDefinedEnvVars
	// We now reassign EnvVars the original values (pre-EnvFile modification).
	// But we also need to account for the in-memory representation.
	//
	// i.e. we call File.Write() at different times but still need EnvVars data.
	//
	// So once we've persisted the correct data back to disk, we can then revert
	// the in-memory data for EnvVars to include the contents from EnvFile
	// i.e. combinedEnvVars
	// just in case the CLI process is still running and needs to do things with
	// environment variables.
	if f.Scripts.EnvFile != "" {
		combinedEnvVars := make([]string, len(f.Scripts.EnvVars))
		copy(combinedEnvVars, f.Scripts.EnvVars)
		f.Scripts.EnvVars = f.Scripts.manifestDefinedEnvVars
		defer func() {
			f.Scripts.EnvVars = combinedEnvVars
		}()
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
	// Build is a custom build script.
	Build string `toml:"build,omitempty"`
	// EnvFile is a path to a file containing build related environment variables.
	// Each line should contain a KEY=VALUE.
	// Reading the contents of this file will populate the `EnvVars` field.
	EnvFile string `toml:"env_file,omitempty"`
	// EnvVars contains build related environment variables.
	EnvVars []string `toml:"env_vars,omitempty"`
	// PostBuild is executed after the build step.
	PostBuild string `toml:"post_build,omitempty"`
	// PostInit is executed after the init step.
	PostInit string `toml:"post_init,omitempty"`

	// Private field used to revert modifications to EnvVars from EnvFile.
	// See File.ParseEnvFile() and File.Write() methods for details.
	// This will contain the environment variables defined in the manifest file.
	manifestDefinedEnvVars []string
}
