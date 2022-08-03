package config

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
	toml "github.com/pelletier/go-toml"
)

// Source enumerates where a config parameter is taken from.
type Source uint8

const (
	// SourceUndefined indicates the parameter isn't provided in any of the
	// available sources, similar to "not found".
	SourceUndefined Source = iota

	// SourceFile indicates the parameter came from a config file.
	SourceFile

	// SourceEnvironment indicates the parameter came from an env var.
	SourceEnvironment

	// SourceFlag indicates the parameter came from an explicit flag.
	SourceFlag

	// SourceDefault indicates the parameter came from a program default.
	SourceDefault

	// DirectoryPermissions is the default directory permissions for the config file directory.
	DirectoryPermissions = 0o700

	// FilePermissions is the default file permissions for the config file.
	FilePermissions = 0o600

	// RemoteEndpoint represents the API endpoint where we'll pull the dynamic
	// configuration file from.
	//
	// NOTE: once the configuration is stored locally, it will allow for
	// overriding this default endpoint.
	RemoteEndpoint = "https://developer.fastly.com/api/internal/cli-config"

	// UpdateSuccessful represents the message shown to a user when their
	// application configuration has been updated successfully.
	UpdateSuccessful = "Successfully updated platform compatibility and versioning information."
)

// ErrLegacyConfig indicates that the local configuration file is using the
// legacy format.
var ErrLegacyConfig = errors.New("the configuration file is in the legacy format")

// ErrInvalidConfig indicates that the configuration file used was invalid.
var ErrInvalidConfig = errors.New("the configuration file is invalid")

// RemediationManualFix indicates that the configuration file used was invalid
// and that the user rejected the use of the static config embedded into the
// compiled CLI binary and so the user must resolve their invalid config.
var RemediationManualFix = "You'll need to manually fix any invalid configuration syntax."

// Data holds global-ish configuration data from all sources: environment
// variables, config files, and flags. It has methods to give each parameter to
// the components that need it, including the place the parameter came from,
// which is a requirement.
//
// If the same parameter is defined in multiple places, it is resolved according
// to the following priority order: the config file (lowest priority), env vars,
// and then explicit flags (highest priority).
//
// This package and its types are only meant for parameters that are applicable
// to most/all subcommands (e.g. API token) and are consistent for a given user
// (e.g. an email address). Otherwise, parameters should be defined in specific
// command structs, and parsed as flags.
type Data struct {
	Env      Environment
	File     File
	Flag     Flag
	Manifest manifest.Data
	Output   io.Writer
	Path     string

	// Custom interfaces
	ErrLog     fsterr.LogInterface
	APIClient  api.Interface
	HTTPClient api.HTTPClient
	RTSClient  api.RealtimeStatsInterface
}

// Token yields the Fastly API token.
func (d *Data) Token() (string, Source) {
	if d.Flag.Token != "" {
		return d.Flag.Token, SourceFlag
	}

	if d.Env.Token != "" {
		return d.Env.Token, SourceEnvironment
	}

	if d.Manifest.File.Profile != "" {
		for k, v := range d.File.Profiles {
			if k == d.Manifest.File.Profile {
				return v.Token, SourceFile
			}
		}
	}

	if d.Flag.Profile != "" {
		for k, v := range d.File.Profiles {
			if k == d.Flag.Profile {
				return v.Token, SourceFile
			}
		}
	}

	for _, v := range d.File.Profiles {
		if v.Default {
			return v.Token, SourceFile
		}
	}

	return "", SourceUndefined
}

// Verbose yields the verbose flag, which can only be set via flags.
func (d *Data) Verbose() bool {
	return d.Flag.Verbose
}

// Endpoint yields the API endpoint.
func (d *Data) Endpoint() (string, Source) {
	if d.Flag.Endpoint != "" {
		return d.Flag.Endpoint, SourceFlag
	}

	if d.Env.Endpoint != "" {
		return d.Env.Endpoint, SourceEnvironment
	}

	if d.File.Fastly.APIEndpoint != DefaultEndpoint && d.File.Fastly.APIEndpoint != "" {
		return d.File.Fastly.APIEndpoint, SourceFile
	}

	return DefaultEndpoint, SourceDefault // this method should not fail
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

// DefaultEndpoint is the default Fastly API endpoint.
const DefaultEndpoint = "https://api.fastly.com"

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

// File represents our dynamic application toml configuration.
type File struct {
	CLI           CLI                 `toml:"cli"`
	ConfigVersion int                 `toml:"config_version"`
	Fastly        Fastly              `toml:"fastly"`
	Language      Language            `toml:"language"`
	Profiles      Profiles            `toml:"profile"`
	StarterKits   StarterKitLanguages `toml:"starter-kits"`
	Viceroy       Viceroy             `toml:"viceroy"`

	// We store off a possible legacy configuration so that we can later extract
	// the relevant email and token values that may pre-exist.
	//
	// NOTE: We set omitempty so when we write the in-memory data back to disk
	// we'll cause the [user] block to be removed. If we didn't do this, then
	// every time we run a command with --verbose we would see a message telling
	// us our config.toml was in a legacy format, even though we would have
	// already migrated the user data to the [profile] section.
	LegacyUser LegacyUser `toml:"user,omitempty"`

	// Store off copy of the static application configuration that has been
	// embedded into the compiled CLI binary.
	//
	// NOTE: This field is private to prevent it being written back to disk.
	static []byte `toml:",omitempty"`
}

// Fastly represents fastly specific configuration.
type Fastly struct {
	APIEndpoint string `toml:"api_endpoint"`
}

// CLI represents CLI specific configuration.
type CLI struct {
	RemoteConfig string `toml:"remote_config"`
	LastChecked  string `toml:"last_checked"`
	Version      string `toml:"version"`
}

// User represents user specific configuration.
type User struct {
	Token string `toml:"token"`
	Email string `toml:"email"`
}

// Viceroy represents viceroy specific configuration.
type Viceroy struct {
	LastChecked   string `toml:"last_checked"`
	LatestVersion string `toml:"latest_version"`
	TTL           string `toml:"ttl"`
}

// Language represents C@E language specific configuration.
type Language struct {
	Go   Go   `toml:"go"`
	Rust Rust `toml:"rust"`
}

// Go represents Go C@E language specific configuration.
type Go struct {
	// TinyGoConstraint is the `tinygo` version that we support.
	TinyGoConstraint string `toml:"tinygo_constraint"`

	// ToolchainConstraint is the `go` version that we support.
	//
	// We aim for go versions that support go modules by default.
	// https://go.dev/blog/using-go-modules
	ToolchainConstraint string `toml:"toolchain_constraint"`
}

// Rust represents Rust C@E language specific configuration.
type Rust struct {
	// ToolchainVersion is the `rustup` toolchain string for the compiler that we
	// support
	//
	// Deprecated: Use ToolchainConstraint instead
	ToolchainVersion string `toml:"toolchain_version"`

	// ToolchainConstraint is the `rustup` toolchain constraint for the compiler
	// that we support (a range is expected, e.g. >= 1.49.0 < 2.0.0).
	ToolchainConstraint string `toml:"toolchain_constraint"`

	// WasmWasiTarget is the Rust compilation target for Wasi capable Wasm.
	WasmWasiTarget string `toml:"wasm_wasi_target"`

	// FastlySysConstraint is a free-form semver constraint for the internal Rust
	// ABI version that should be supported.
	FastlySysConstraint string `toml:"fastly_sys_constraint"`

	// RustupConstraint is a free-form semver constraint for the rustup version
	// that should be installed.
	RustupConstraint string `toml:"rustup_constraint"`
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

// MigrateLegacy ensures legacy data is transitioned to config new format.
func (f *File) MigrateLegacy() {
	if f.LegacyUser.Email != "" || f.LegacyUser.Token != "" {
		if f.Profiles == nil {
			f.Profiles = make(Profiles)
		}
		// NOTE: We keep the assignment separate just in case the user somehow has
		// a config.toml with BOTH a populated [user] + [profile] section.
		f.Profiles["user"] = &Profile{
			Default: true,
			Email:   f.LegacyUser.Email,
			Token:   f.LegacyUser.Token,
		}
		f.LegacyUser = LegacyUser{}
	}
}

// createConfigDir creates the application configuration directory if it
// doesn't already exist.
func createConfigDir(path string) error {
	basePath := filepath.Dir(path)
	err := filesystem.MakeDirectoryIfNotExists(basePath)
	if err != nil {
		return err
	}
	return nil
}

// ValidConfig checks the current config version isn't different from the
// config statically embedded into the CLI binary. If it is then we consider
// the config not valid and we'll fallback to the embedded config.
func (f *File) ValidConfig(verbose bool, out io.Writer) bool {
	var cfg File
	err := toml.Unmarshal(f.static, &cfg)
	if err != nil {
		return false
	}

	if f.ConfigVersion != cfg.ConfigVersion {
		if verbose {
			text.Output(out, `
				Found your local configuration file (required to use the CLI) to be incompatible with the current CLI version.
				Your configuration will be migrated to a compatible configuration format.
			`)
			text.Break(out)
		}
		return false
	}
	return true
}

// NOTE: Static 👇 is public for the sake of the test suite.

//go:embed config.toml
var Static []byte

// Read decodes a disk file into an in-memory data structure.
func (f *File) Read(
	path string,
	in io.Reader,
	out io.Writer,
	errLog fsterr.LogInterface,
	verbose bool,
) error {
	var useStatic bool
	f.static = Static

	const replacement = "Replace it with a valid version? (any existing email/token data will be lost) [y/N] "

	// G304 (CWE-22): Potential file inclusion via variable.
	// gosec flagged this:
	// Disabling as we need to load the config.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	/* #nosec */
	data, err := os.ReadFile(path)
	if err != nil {
		errLog.Add(err)
		data = f.static
		useStatic = true

		msg := "unable to load your configuration data"
		label := fmt.Sprintf("We were %s. %s", msg, replacement)

		cont, err := text.AskYesNo(out, label, in)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		if !cont {
			err := fsterr.RemediationError{
				Inner:       fmt.Errorf(msg),
				Remediation: RemediationManualFix,
			}
			errLog.Add(err)
			return err
		}
	}

	unmarshalErr := toml.Unmarshal(data, f)

	if unmarshalErr != nil {
		errLog.Add(unmarshalErr)

		// If the static embedded config failed to be unmarshalled, then that's
		// unexpected and we can't recover from that.
		if useStatic {
			return invalidConfigErr(unmarshalErr)
		}

		// Otherwise if the local disk config failed to be unmarshalled, then
		// ask the user if they would like us to replace their config with the
		// version embedded into the CLI binary.

		text.Break(out)

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

		data = f.static

		err = toml.Unmarshal(data, f)

		if err != nil {
			errLog.Add(err)
			return invalidConfigErr(err)
		}
	}

	// NOTE: When using the embedded config the LastChecked/Version fields won't be set.
	if f.CLI.LastChecked == "" {
		f.CLI.LastChecked = time.Now().Format(time.RFC3339)
	}
	if f.CLI.Version == "" {
		f.CLI.Version = revision.SemVer(revision.AppVersion)
	}

	err = createConfigDir(path)
	if err != nil {
		errLog.Add(err)
		return err
	}

	// The reason we're writing data back to disk, although we've only just read
	// data from the same file, is because we've already potentially mutated some
	// fields like [cli.last_checked] and [cli.version].
	err = f.Write(path)
	if err != nil {
		errLog.Add(err)
		return err
	}

	// The top-level 'user' section is what we're using to identify whether the
	// local config.toml file is using a legacy format. If we find that key, then
	// we must delete the file and return an error so that the calling code can
	// take the appropriate action of creating the file anew.
	tree, err := toml.LoadBytes(data)
	if err != nil {
		errLog.Add(err)

		// NOTE: We do not expect this error block to ever be hit because if we've
		// already successfully called toml.Unmarshal, then calling toml.LoadBytes
		// should equally be successful, but we'll code defensively nonetheless.
		return invalidConfigErr(err)
	}

	var legacyFormat bool

	if user := tree.Get("user"); user != nil {
		errLog.Add(ErrLegacyConfig)

		if verbose {
			text.Output(out, `
        Found your local configuration file (required to use the CLI) was using a legacy format.
        File will be updated to the latest format.
      `)
			text.Break(out)
		}

		legacyFormat = true
	}

	if legacyFormat || !f.ValidConfig(verbose, out) {
		err = f.UseStatic(f.static, path)
		if err != nil {
			return err
		}
	}

	return nil
}

// UseStatic allow us to switch the in-memory configuration with the static
// version embedded into the CLI binary and writes it back to disk.
func (f *File) UseStatic(cfg []byte, path string) (err error) {
	err = toml.Unmarshal(cfg, f)
	if err != nil {
		return invalidConfigErr(err)
	}

	f.CLI.LastChecked = time.Now().Format(time.RFC3339)
	f.CLI.Version = revision.SemVer(revision.AppVersion)
	f.MigrateLegacy()

	err = createConfigDir(path)
	if err != nil {
		return err
	}

	return f.Write(path)
}

var mutex = &sync.Mutex{}

// Write encodes in-memory data to disk.
//
// NOTE: pkg/commands/update/check.go contains a CheckAsync function which
// asynchronously calls the config's file.Read() method, followed by calling the
// config's file.Write() method. Because of this we use a mutex to prevent a
// race condition writing content to disk, in case the user command invoked was
// one of the profile commands (which are expected to trigger a change in data).
func (f *File) Write(path string) (err error) {
	mutex.Lock()
	defer mutex.Unlock()

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
	if err := toml.NewEncoder(fp).Encode(f); err != nil {
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
	Token    string
	Endpoint string
}

// Read populates the fields from the provided environment.
func (e *Environment) Read(state map[string]string) {
	e.Token = state[env.Token]
	e.Endpoint = state[env.Endpoint]
}

// Flag represents all of the configuration parameters that can be set with
// explicit flags. Consumers should bind their flag values to these fields
// directly.
type Flag struct {
	AcceptDefaults bool
	AutoYes        bool
	Endpoint       string
	NonInteractive bool
	Profile        string
	Token          string
	Verbose        bool
}

// This suggests our embedded config is unexpectedly faulty and so we should
// fail with a bug remediation.
func invalidConfigErr(err error) error {
	return fsterr.RemediationError{
		Inner:       fmt.Errorf("%v: %v", ErrInvalidConfig, err),
		Remediation: fsterr.BugRemediation,
	}
}
