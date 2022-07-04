package config

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/useragent"
	toml "github.com/pelletier/go-toml"
	"github.com/sethvargo/go-retry"
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

// ErrInvalidConfig indicates that the configuration file used was the static
// one embedded into the compiled CLI binary and that failed to be unmarshalled.
var ErrInvalidConfig = errors.New("the configuration file is invalid")

// RemediationManualFix indicates that the configuration file used was invalid
// and that the user rejected the use of the static config embedded into the
// compiled CLI binary and so the user must resolve their invalid config.
var RemediationManualFix = "You'll need to manually fix any invalid configuration syntax."

// Mutex provides synchronisation for any WRITE operations on the CLI config.
// This includes calls to the Write method (which affects the disk
// representation) as well as in-memory data structure updates.
//
// NOTE: Historically the CLI has only had to write to the CLI config from
// within the `main` function but now the `fastly update` command accepts a
// flag that allows explicitly updating the CLI config, which means we need to
// ensure there isn't a race condition with writing the config to disk.
//
// We also need to be careful with low-level mutex usage because it's easy to
// cause a deadlock! For example, if we lock around a section of code that
// can potentially return out of the function early, then we'd never call the
// mutex's Unlock method.
//
// We also can't rely on using the defer statement either because in the case
// of the UseStatic() method, the last line of the method is a call to .Write()
// which itself also tries to acquire a lock. So if we deferred the Unlock(),
// we'd eventually call out to .Write() and deadlock waiting to acquire the lock
// that we never actually released from within the caller.
//
// Because of this we should be mindful whenever using the defer statement, and
// possibly opt for multiple calls to Lock/Unlock around specific portions of
// the code accessing the in-memory data (while also checking for early returns).
//
// NOTE: Ideally we wouldn't need to sprinkle mutex references all over the
// code, and instead have it abstracted away inside the config.File type, but
// this requires exposing lots of setter methods for each field on each nested
// subfield type (this would be a lot of extra code). If we start to experience
// issues with mutex usage, then this decision can be revisited.
var Mutex = &sync.Mutex{}

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
	static []byte `toml:",omitempty"`
}

// Fastly represents fastly specific configuration.
type Fastly struct {
	APIEndpoint string `toml:"api_endpoint"`
}

// CLI represents CLI specific configuration.
type CLI struct {
	RemoteConfig string `toml:"remote_config"`
	TTL          string `toml:"ttl"`
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
	Rust Rust `toml:"rust"`
}

// Rust represents Rust C@E language specific configuration.
type Rust struct {
	// ToolchainVersion is the `rustup` toolchain string for the compiler that we
	// support
	//
	// DEPRECATED in favour of ToolchainConstraint
	ToolchainVersion string `toml:"toolchain_version"`

	// ToolchainConstrain is the `rustup` toolchain constraint for the compiler
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
	Default bool   `toml:"default"`
	Email   string `toml:"email"`
	Token   string `toml:"token"`
}

// StarterKitLanguages represents language specific starter kits.
type StarterKitLanguages struct {
	AssemblyScript []StarterKit `toml:"assemblyscript"`
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

// SetStatic sets the embedded config into the File for backup purposes.
//
// NOTE: The reason we have a setter method is because the File struct is
// expected to be marshalled back into a toml file and we don't want the
// contents of f.static to be persisted to disk (which happens when a field is
// defined as public, so we make it private instead and expose a getter/setter).
func (f *File) SetStatic(static []byte) {
	f.static = static
}

// Static returns the embedded backup config.
func (f *File) Static() []byte {
	return f.static
}

// Load decodes a network response body into an in-memory data structure.
func (f *File) Load(endpoint, path string, c api.HTTPClient) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", useragent.Name)
	resp, err := c.Do(req)
	if err != nil {
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return fsterr.RemediationError{
				Inner:       err,
				Remediation: fsterr.NetworkRemediation,
			}
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching remote configuration: expected '200 OK', received '%s'", resp.Status)
	}

	// NOTE: Decoding does not cause existing field data in f to be reset (e.g.
	// Profiles that are set in-memory continue to be set after reading in the
	// remote configuration data even though that data source doesn't define any
	// such data). This can be manually validated using an io.TeeReader.
	Mutex.Lock()
	err = toml.NewDecoder(resp.Body).Decode(f)
	Mutex.Unlock()

	if err != nil {
		return err
	}

	Mutex.Lock()
	{
		f.CLI.Version = revision.SemVer(revision.AppVersion)
		f.CLI.LastChecked = time.Now().Format(time.RFC3339)
		migrateLegacyData(f)
	}
	Mutex.Unlock()

	err = createConfigDir(path)
	if err != nil {
		return err
	}

	return f.Write(path)
}

// migrateLegacyData ensures legacy data is transitioned to config new format.
func migrateLegacyData(f *File) {
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

// Read decodes a disk file into an in-memory data structure.
//
// If reading from disk fails, then we'll use the static config embedded into
// the CLI binary (which we expect to be valid). If an attempt to unmarshal
// the static config fails then we have to consider something fundamental has
// gone wrong and subsequently expect the caller to exit the program.
//
// NOTE: Some users have noticed that their profile data is deleted after
// certain (nondeterministic) operations. The fallback to static config may be
// contributing to this behaviour because if a read of the user's config fails
// and we start to use the static version as a fallback, then that will contain
// no profile data and that in-memory representation will be persisted back to
// disk, causing the user's profile data to be lost.
func (f *File) Read(path string, in io.Reader, out io.Writer, errLog fsterr.LogInterface) error {
	var (
		attempts int
		data     []byte
		retryErr error
		readErr  error
	)

	// To help mitigate any loss of profile data within the user's config we will
	// retry the file READ operation if it fails.
	ctx := context.Background()
	b := retry.NewConstant(1 * time.Second)
	b = retry.WithMaxRetries(2, b) // attempts == 3 (first attempt + two retries)

	retryErr = retry.Do(ctx, b, func(_ context.Context) error {
		attempts++

		// G304 (CWE-22): Potential file inclusion via variable.
		// gosec flagged this:
		// Disabling as we need to load the config.toml from the user's file system.
		// This file is decoded into a predefined struct, any unrecognised fields are dropped.
		/* #nosec */
		data, readErr = os.ReadFile(path)
		if readErr != nil {
			errLog.Add(readErr)

			// We don't retry the error if the file doesn't exist. Although in
			// principle we shouldn't see this error as we .Stat() at the start of
			// the file.Read() method.
			if errors.Is(readErr, os.ErrNotExist) {
				return readErr
			}
			return retry.RetryableError(readErr)
		}

		return nil
	})

	// NOTE: retryErr is when we reached the max number of retries, while the
	// readErr is for catching when we didn't want to retry the READ operation.
	if retryErr != nil || readErr != nil {
		errLog.AddWithContext(retryErr, map[string]interface{}{
			"attempts": attempts,
		})
		data = f.static
	}

	Mutex.Lock()
	unmarshalErr := toml.Unmarshal(data, f)
	Mutex.Unlock()

	if unmarshalErr != nil {
		errLog.Add(unmarshalErr)

		// The embedded config is unexpectedly invalid.
		if readErr != nil {
			return invalidConfigErr(unmarshalErr)
		}

		// Otherwise if the local disk config failed to be unmarshalled, then
		// ask the user if they would like us to replace their config with the
		// version embedded into the CLI binary.

		text.Break(out)
		label := fmt.Sprintf("Your configuration file (%s) is invalid. Replace it with a valid version? (any existing email/token data will be lost) [y/N] ", path)
		cont, err := text.Input(out, label, in)
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		contl := strings.ToLower(cont)
		if contl == "y" || contl == "yes" {
			data = f.static

			Mutex.Lock()
			err = toml.Unmarshal(data, f)
			Mutex.Unlock()

			if err != nil {
				errLog.Add(err)
				return invalidConfigErr(err)
			}
		} else {
			err := fsterr.RemediationError{
				Inner:       fmt.Errorf("%v: %v", ErrInvalidConfig, unmarshalErr),
				Remediation: RemediationManualFix,
			}
			errLog.Add(err)
			return err
		}
	}

	// We expect LastChecked/Version to not be set if coming from the static config.
	Mutex.Lock()
	{
		if f.CLI.LastChecked == "" {
			f.CLI.LastChecked = time.Now().Format(time.RFC3339)
		}
		if f.CLI.Version == "" {
			f.CLI.Version = revision.SemVer(revision.AppVersion)
		}
	}
	Mutex.Unlock()

	err := createConfigDir(path)
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
	if user := tree.Get("user"); user != nil {
		errLog.Add(ErrLegacyConfig)
		return ErrLegacyConfig
	}

	return nil
}

// UseStatic allow us to switch the in-memory configuration with the static
// version embedded into the CLI binary and writes it back to disk.
func (f *File) UseStatic(cfg []byte, path string) (err error) {
	Mutex.Lock()
	err = toml.Unmarshal(cfg, f)
	Mutex.Unlock()

	if err != nil {
		return invalidConfigErr(err)
	}

	Mutex.Lock()
	{
		f.CLI.LastChecked = time.Now().Format(time.RFC3339)
		f.CLI.Version = revision.SemVer(revision.AppVersion)
		migrateLegacyData(f)
	}
	Mutex.Unlock()

	err = createConfigDir(path)
	if err != nil {
		return err
	}

	return f.Write(path)
}

// Write encodes in-memory data to disk.
func (f *File) Write(path string) (err error) {
	Mutex.Lock()
	defer Mutex.Unlock()

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
