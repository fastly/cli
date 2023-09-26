package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml"

	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
)

const (
	// DirectoryPermissions is the default directory permissions for the config file directory.
	DirectoryPermissions = 0o700

	// FilePermissions is the default file permissions for the config file.
	FilePermissions = 0o600
)

var (
	// CurrentConfigVersion indicates the present config version.
	CurrentConfigVersion int

	// ErrLegacyConfig indicates that the local configuration file is using the
	// legacy format.
	ErrLegacyConfig = errors.New("the configuration file is in the legacy format")

	// ErrInvalidConfig indicates that the configuration file used was invalid.
	ErrInvalidConfig = errors.New("the configuration file is invalid")

	// RemediationManualFix indicates that the configuration file used was invalid
	// and that the user rejected the use of the static config embedded into the
	// compiled CLI binary and so the user must resolve their invalid config.
	RemediationManualFix = "You'll need to manually fix any invalid configuration syntax."
)

// LegacyUser represents the old toml configuration format.
//
// NOTE: this exists to catch situations where an existing CLI user upgrades
// their version of the CLI and ends up trying to use the latest iteration of
// the toml configuration. We don't want them to have to re-enter their email
// or token, so we'll decode the existing config file into the LegacyUser type
// and then extract those details later when constructing the proper File type.
//
// I had tried to make this an unexported type but it seemed the toml decoder
// would fail to unmarshal the configuration unless it was an exported type.
type LegacyUser struct {
	Email string `toml:"email"`
	Token string `toml:"token"`
}

// Fastly represents fastly specific configuration.
type Fastly struct {
	APIEndpoint string `toml:"api_endpoint"`
}

// WasmMetadata represents what telemetry data will be recorded.
type WasmMetadata struct {
	// BuildInfo represents information regarding the time taken for builds and
	// compilation processes, helping us identify bottlenecks and optimize
	// performance (enable/disable).
	BuildInfo string `toml:"build_info"`
	// MachineInfo represents general, non-identifying system specifications (CPU,
	// RAM, operating system) to better understand the hardware landscape our CLI
	// operates in (enable/disable).
	MachineInfo string `toml:"machine_info"`
	// PackageInfo represents packages and libraries utilized in your source code,
	// enabling us to prioritize support for the most commonly used components
	// (enable/disable).
	PackageInfo string `toml:"package_info"`
}

// CLI represents CLI specific configuration.
type CLI struct {
	// TelemetryNoticeDisplayed indicates if the user has been notified of the
	// telemetry behaviours being enabled by default and how they can opt-out.
	TelemetryNoticeDisplayed bool `toml:"telemetry_notice_displayed"`
	// Version indicates the CLI configuration version.
	// It is updated each time a change is made to the config structure.
	Version string `toml:"version"`
}

// Versioner represents GitHub assets configuration.
// e.g. viceroy, wasm-tools etc
type Versioner struct {
	// LastChecked is when the asset version was last checked.
	LastChecked string `toml:"last_checked"`
	// LatestVersion is the latest asset version at the time it is set.
	LatestVersion string `toml:"latest_version"`
	// TTL is how long the CLI waits before considering the asset version stale.
	TTL string `toml:"ttl"`
}

// Language represents Compute language specific configuration.
type Language struct {
	Go   Go   `toml:"go"`
	Rust Rust `toml:"rust"`
}

// Go represents Go Compute language specific configuration.
type Go struct {
	// TinyGoConstraint is the `tinygo` version that we support.
	TinyGoConstraint string `toml:"tinygo_constraint"`

	// TinyGoConstraintFallback is a fallback `tinygo` version for users who have
	// a pre-existing project with a 0.1.x Fastly Go SDK specified.
	TinyGoConstraintFallback string `toml:"tinygo_constraint_fallback"`

	// ToolchainConstraint is the `go` version that we support with WASI.
	ToolchainConstraint string `toml:"toolchain_constraint"`

	// ToolchainConstraintTinyGo is the `go` version that we support with TinyGo.
	//
	// We aim for go versions that support go modules by default.
	// https://go.dev/blog/using-go-modules
	ToolchainConstraintTinyGo string `toml:"toolchain_constraint_tinygo"`
}

// Rust represents Rust Compute language specific configuration.
type Rust struct {
	// ToolchainConstraint is the `rustup` toolchain constraint for the compiler
	// that we support (a range is expected, e.g. >= 1.49.0 < 2.0.0).
	ToolchainConstraint string `toml:"toolchain_constraint"`

	// WasmWasiTarget is the Rust compilation target for Wasi capable Wasm.
	WasmWasiTarget string `toml:"wasm_wasi_target"`
}

// Profiles represents multiple profile accounts.
type Profiles map[string]*Profile

// Profile represents a specific profile account.
type Profile struct {
	Default bool   `toml:"default" json:"default"`
	Email   string `toml:"email" json:"email"`
	Token   string `toml:"token" json:"token"`
}

// StarterKitLanguages represents language specific starter kits.
type StarterKitLanguages struct {
	AssemblyScript []StarterKit `toml:"assemblyscript"`
	Go             []StarterKit `toml:"go"`
	JavaScript     []StarterKit `toml:"javascript"`
	Rust           []StarterKit `toml:"rust"`
}

// StarterKit represents starter kit specific configuration.
type StarterKit struct {
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Path        string `toml:"path"`
	Tag         string `toml:"tag"`
	Branch      string `toml:"branch"`
}

// ensureConfigDirExists creates the application configuration directory if it
// doesn't already exist.
func ensureConfigDirExists(path string) error {
	basePath := filepath.Dir(path)
	return filesystem.MakeDirectoryIfNotExists(basePath)
}

// File represents our application toml configuration.
type File struct {
	// CLI represents CLI specific configuration.
	CLI CLI `toml:"cli"`
	// ConfigVersion is the version of the config.
	ConfigVersion int `toml:"config_version"`
	// Fastly represents fastly specific configuration.
	Fastly Fastly `toml:"fastly"`
	// Language represents C@E language specific configuration.
	Language Language `toml:"language"`
	// Profiles represents multiple profile accounts.
	Profiles Profiles `toml:"profile"`
	// StarterKitLanguages represents language specific starter kits.
	StarterKits StarterKitLanguages `toml:"starter-kits"`
	// Viceroy represents viceroy specific configuration.
	Viceroy Versioner `toml:"viceroy"`
	// WasmMetadata represents what metadata will be collected.
	WasmMetadata WasmMetadata `toml:"wasm-metadata"`
	// WasmTools represents wasm-tools specific configuration.
	WasmTools Versioner `toml:"wasm-tools"`

	// We store off a possible legacy configuration so that we can later extract
	// the relevant email and token values that may pre-exist.
	//
	// NOTE: We set omitempty so when we write the in-memory data back to disk
	// we'll cause the [user] block to be removed. If we didn't do this, then
	// every time we run a command with --verbose we would see a message telling
	// us our config.toml was in a legacy format, even though we would have
	// already migrated the user data to the [profile] section.
	LegacyUser LegacyUser `toml:"user,omitempty"`

	// Store the flag values for --auto-yes/--non-interactive as at the time of
	// the config File construction we need these values and need to be stored so
	// that other callers of File.Read() don't need to have the values passed
	// around in function arguments.
	//
	// NOTE: These fields are private to prevent them being written back to disk,
	// but it means we need to expose Setter methods.
	autoYes        bool
	nonInteractive bool
}

// SetAutoYes sets the associated flag value.
func (f *File) SetAutoYes(v bool) {
	f.autoYes = v
}

// SetNonInteractive sets the associated flag value.
func (f *File) SetNonInteractive(v bool) {
	f.nonInteractive = v
}

// NOTE: Static 👇 is public for the sake of the test suite.

// Static is the embedded configuration file used by the CLI.
//
//go:embed config.toml
var Static []byte

// Read decodes a disk file into an in-memory data structure.
//
// NOTE: If user local configuration can't be read, then we'll ask the user to
// confirm whether to use the static config embedded in the CLI binary. If the
// user local configuration is deemed to be invalid, then we'll automatically
// switch to the static config and migrate the user's profile data (if any).
func (f *File) Read(
	path string,
	in io.Reader,
	out io.Writer,
	errLog fsterr.LogInterface,
	verbose bool,
) error {
	// Ensure the static config is sound. This should never happen (tm).
	// We are checking this earlier to simplify the code later on.
	var staticConfig File
	err := toml.Unmarshal(Static, &staticConfig)
	if err != nil {
		errLog.Add(err)
		return invalidStaticConfigErr(err)
	}

	CurrentConfigVersion = staticConfig.ConfigVersion

	// G304 (CWE-22): Potential file inclusion via variable.
	// gosec flagged this:
	// Disabling as we need to load the config.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	/* #nosec */
	// nosemgrep: trailofbits.go.invalid-usage-of-modified-variable.invalid-usage-of-modified-variable
	data, err := os.ReadFile(path)
	if err != nil {
		data = Static
	}

	unmarshalErr := toml.Unmarshal(data, f)
	if unmarshalErr != nil {
		errLog.Add(unmarshalErr)

		// If the local disk config failed to be unmarshalled, then
		// ask the user if they would like us to replace their config with the
		// version embedded into the CLI binary.

		text.Break(out)

		if !f.autoYes {
			replacement := "Replace it with a valid version? (any existing email/token data will be lost) [y/N] "
			label := fmt.Sprintf("Your configuration file (%s) is invalid. %s", path, replacement)
			cont, err := text.AskYesNo(out, label, in)
			if err != nil {
				return fmt.Errorf("error reading input: %w", err)
			}
			if !cont {
				err := fsterr.RemediationError{
					Inner:       fmt.Errorf("%v: %v", ErrInvalidConfig, unmarshalErr),
					Remediation: RemediationManualFix,
				}
				errLog.Add(err)
				return err
			}
		}
		f = &staticConfig
	}

	err = ensureConfigDirExists(path)
	if err != nil {
		errLog.Add(err)
		return err
	}

	if f.NeedsUpdating(data, out, errLog, verbose) {
		return f.UseStatic(path)
	}

	return nil
}

// MigrateLegacy ensures legacy data is transitioned to config new format.
func (f *File) MigrateLegacy() {
	if f.LegacyUser.Email != "" || f.LegacyUser.Token != "" {
		if f.Profiles == nil {
			f.Profiles = make(Profiles)
		}

		// We keep the assignment separate just in case the user somehow has a
		// config.toml with BOTH a populated [user] + [profile] section, and
		// possibly even already has a default account of "user".

		key := "user"
		if _, ok := f.Profiles[key]; ok {
			key = "legacy" // avoid overriding the default
		}

		f.Profiles[key] = &Profile{
			Default: true,
			Email:   f.LegacyUser.Email,
			Token:   f.LegacyUser.Token,
		}
		f.LegacyUser = LegacyUser{}
	}
}

// NeedsUpdating indicates if the application config needs updating.
func (f *File) NeedsUpdating(data []byte, out io.Writer, errLog fsterr.LogInterface, verbose bool) bool {
	tree, err := toml.LoadBytes(data)
	if err != nil {
		// NOTE: We do not expect this error block to ever be hit because if we've
		// already successfully called toml.Unmarshal, then calling toml.LoadBytes
		// should equally be successful.
		panic("LoadBytes failed but Unmarshal succeeded")
	}

	switch {
	case tree.Get("user") != nil:
		// The top-level 'user' section is what we're using to identify whether the
		// local config.toml file is using a legacy format. If we find that key, then
		// we must delete the file and return an error so that the calling code can
		// take the appropriate action of creating the file anew.
		errLog.Add(ErrLegacyConfig)

		if verbose {
			text.Output(out, `
				Found your local configuration file (required to use the CLI) was using a legacy format.
				File will be updated to the latest format.
			`)
			text.Break(out)
		}
		return true
	case f.ConfigVersion != CurrentConfigVersion:
		// If the ConfigVersion doesn't match, then this suggests a breaking change
		// divergence in either the user's config or the CLI's config.
		if verbose {
			text.Output(out, "Found your local configuration file (required to use the CLI) to be incompatible with the current CLI version. Your configuration will be migrated to a compatible configuration format.")
			text.Break(out)
		}
		return true
	case f.CLI.Version != revision.SemVer(revision.AppVersion):
		// If the CLI.Version doesn't match the CLI binary version, then this suggests
		// a version update. This _might_ include a breaking change in the CLI's
		// logic/implementation, or a new starter kit, for example.
		// In this case we update the config regardless to ensure the
		// CLI.Version is up to date.
		return true
	}

	return false
}

// UseStatic switches the in-memory configuration with the static version
// embedded into the CLI binary and writes it back to disk.
//
// NOTE: We will attempt to migrate the profile data.
func (f *File) UseStatic(path string) error {
	err := toml.Unmarshal(Static, f)
	if err != nil {
		return invalidStaticConfigErr(err)
	}

	f.CLI.Version = revision.SemVer(revision.AppVersion)
	f.MigrateLegacy()

	err = ensureConfigDirExists(path)
	if err != nil {
		return err
	}

	return f.Write(path)
}

// Write encodes in-memory data to disk.
func (f *File) Write(path string) error {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as in most cases the input is determined by our own package.
	// In other cases we want to control the input for testing purposes.
	/* #nosec */
	fp, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, FilePermissions)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	encoder := toml.NewEncoder(fp)
	// Remove leading spaces from the TOML file.
	encoder.Indentation("")
	if err := encoder.Encode(f); err != nil {
		return fmt.Errorf("error writing to config file: %w", err)
	}
	if err := fp.Close(); err != nil {
		return fmt.Errorf("error saving config file changes: %w", err)
	}

	return nil
}

// Environment represents all of the configuration parameters that can come
// from environment variables.
type Environment struct {
	// Token is the env var we look in for the Fastly API token.
	Token string
	// Endpoint is the env var we look in for the API endpoint.
	Endpoint string
	// WasmMetadataDisable is the env var we look in to disable all data
	// collection related to a Wasm binary.
	// Set to "true" to disable all forms of data collection.
	WasmMetadataDisable string
}

// Read populates the fields from the provided environment.
func (e *Environment) Read(state map[string]string) {
	e.Token = state[env.Token]
	e.Endpoint = state[env.Endpoint]
	e.WasmMetadataDisable = state[env.WasmMetadataDisable]
}

// invalidStaticConfigErr generates an error to alert the user to an issue with
// the CLI's internal configuration.
func invalidStaticConfigErr(err error) error {
	return fsterr.RemediationError{
		Inner:       fmt.Errorf("%v: %v", ErrInvalidConfig, err),
		Remediation: fsterr.InvalidStaticConfigRemediation,
	}
}

// FileName is the name of the application configuration file.
const FileName = "config.toml"

// FilePath is the location of the fastly CLI application config file.
var FilePath = func() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "fastly", FileName)
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".fastly", FileName)
	}
	panic("unable to deduce user config dir or user home dir")
}()
