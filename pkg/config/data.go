package config

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/filesystem"
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
	DirectoryPermissions = 0700

	// FilePermissions is the default file permissions for the config file.
	FilePermissions = 0600

	// RemoteEndpoint represents the API endpoint where we'll pull the dynamic
	// configuration file from.
	//
	// TODO: ensure this is a production domain like https://api.fastly.com
	RemoteEndpoint = "http://integralist-cli-dynamic-config.com.global.prod.fastly.net/cli/config"

	// UpdateSuccessful represents the message shown to a user when their
	// application configuration has been updated successfully.
	UpdateSuccessful = "Successfully wrote updated application configuration file to disk."

	// Defaults is the default template of our local application config.
	// TODO: once go 1.16 is released move this to a go embedded file.
	Defaults = `[fastly]
api_endpoint = "https://api.fastly.com"

[cli]
remote_config = "%s"
ttl = "1h"
last_checked = "%s"

[language]
  [language.rust]
  toolchain_version = "1.46.0"
  wasm_wasi_target = "wasm32-wasi"
	fastly_sys = "0.3.7"

[[starter-kits]]
  lang = "assemblyscript"
  name = "Default"
  path = "https://github.com/fastly/compute-starter-kit-assemblyscript-default"
  tag = "v0.2.0"
[[starter-kits]]
  lang = "rust"
  name = "Default"
  path = "https://github.com/fastly/compute-starter-kit-rust-default.git"
  branch = "0.6.0"
[[starter-kits]]
  lang = "rust"
  name = "Beacon"
  path = "https://github.com/fastly/compute-starter-kit-rust-beacon-termination.git"
[[starter-kits]]
  lang = "rust"
  name = "Static content (S3/GCS)"
  path = "https://github.com/fastly/compute-starter-kit-rust-static-content.git"
	tag  = "v1"`
)

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
	File File
	Env  Environment
	Flag Flag

	Client    api.Interface
	RTSClient api.RealtimeStatsInterface
}

// Token yields the Fastly API token.
func (d *Data) Token() (string, Source) {
	if d.Flag.Token != "" {
		return d.Flag.Token, SourceFlag
	}

	if d.Env.Token != "" {
		return d.Env.Token, SourceEnvironment
	}

	if d.File.User.Token != "" {
		return d.File.User.Token, SourceFile
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

// FilePath is the location of the fastly CLI application config file.
var FilePath = func() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "fastly", "config.toml")
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".fastly", "config.toml")
	}
	panic("unable to deduce user config dir or user home dir")
}()

// DefaultEndpoint is the default Fastly API endpoint.
const DefaultEndpoint = "https://api.fastly.com"

// File represents our dynamic application toml configuration.
type File struct {
	Fastly      ConfigFastly       `toml:"fastly"`
	CLI         ConfigCLI          `toml:"cli"`
	User        ConfigUser         `toml:"user"`
	Language    ConfigLanguage     `toml:"language"`
	StarterKits []ConfigStarterKit `toml:"starter-kits"`
}

// ConfigFastly represents fastly specific configuration.
type ConfigFastly struct {
	APIEndpoint string `toml:"api_endpoint"`
}

// ConfigCLI represents CLI specific configuration.
type ConfigCLI struct {
	RemoteConfig string `toml:"remote_config"`
	TTL          string `toml:"ttl"`
	LastChecked  string `toml:"last_checked"`
}

// ConfigUser represents user specific configuration.
type ConfigUser struct {
	Token string `toml:"token"`
	Email string `toml:"email"`
}

// ConfigLanguage represents C@E language specific configuration.
type ConfigLanguage struct {
	Rust ConfigRust `toml:"rust"`
}

// ConfigLanguage represents Rust C@E language specific configuration.
type ConfigRust struct {
	// ToolchainVersion is the `rustup` toolchain string for the compiler that we
	// support
	ToolchainVersion string `toml:"toolchain_version"`

	// WasmWasiTarget is the Rust compilation target for Wasi capable Wasm.
	WasmWasiTarget string `toml:"wasm_wasi_target"`

	// FastlySys is the internal Rust ABI version that should be supported.
	FastlySys string `toml:"fastly_sys"`
}

// ConfigStarterKit represents starter kit specific configuration.
type ConfigStarterKit struct {
	Language string `toml:"lang"`
	Name     string `toml:"name"`
	Path     string `toml:"path"`
	Tag      string `toml:"tag"`
	Branch   string `toml:"branch"`
}

// FilterKits filters out kits that don't match the requested language.
func (f *File) FilterKits(lang string) []ConfigStarterKit {
	var kits []ConfigStarterKit

	for _, sk := range f.StarterKits {
		if sk.Language == lang {
			kits = append(kits, sk)
		}
	}

	return kits
}

// Load gets the configuration file from the CLI API endpoint and encodes it
// from memory into config.File.
func (f *File) Load(configEndpoint string, httpClient api.HTTPClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, configEndpoint, nil)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = toml.NewDecoder(resp.Body).Decode(f)
	if err != nil {
		return err
	}

	f.CLI.LastChecked = time.Now().Format(time.RFC3339)

	// Create the destination directory for the config file
	basePath := filepath.Dir(FilePath)
	err = filesystem.MakeDirectoryIfNotExists(basePath)
	if err != nil {
		return err
	}

	// Write the new configuration back to disk.
	return f.Write(FilePath)
}

// Read decodes a toml file from the local disk into config.File.
func (f *File) Read(fpath string) error {
	// G304 (CWE-22): Potential file inclusion via variable.
	// gosec flagged this:
	// Disabling as we need to load the config.toml from the user's file system.
	// This file is decoded into a predefined struct, any unrecognised fields are dropped.
	/* #nosec */
	bs, err := os.ReadFile(fpath)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(bs, f)
	return err
}

// LoadDefaults populates config.File with default values.
//
// This is used in the scenario where the remote endpoint containing dynamic
// configuration data has either failed to load or we've been unsuccessful in
// writing the data back to disk.
//
// TODO: refactor when go 1.16 arrives to use the //go:embed directive
// https://tip.golang.org/pkg/embed/
func (f *File) LoadDefaults(defaults string, remote_config string, last_checked string) error {
	config := fmt.Sprintf(defaults, remote_config, last_checked)
	if err := toml.Unmarshal([]byte(config), f); err != nil {
		return err
	}

	return f.Write(FilePath)
}

// Write the instance of File to a local application config file.
//
// NOTE: the expected workflow for this method is for the caller to have
// modified the public field(s) first so that we can write new content to the
// config file from the receiver object itself.
//
// EXAMPLE:
// file.CLI.LastChecked = time.Now().Format(time.RFC3339)
// file.Write(configFilePath)
func (f *File) Write(filename string) error {
	fp, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, FilePermissions)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	if err := toml.NewEncoder(fp).Encode(f); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	if err := fp.Close(); err != nil {
		return fmt.Errorf("error saving config file: %w", err)
	}
	return nil
}

// Environment represents all of the configuration parameters that can come
// from environment variables.
type Environment struct {
	Token    string
	Endpoint string
}

const (
	// EnvVarToken is the env var we look in for the Fastly API token.
	// gosec flagged this:
	// G101 (CWE-798): Potential hardcoded credentials
	// Disabling as we use the value in the command help output.
	/* #nosec */
	EnvVarToken = "FASTLY_API_TOKEN"

	// EnvVarEndpoint is the env var we look in for the API endpoint.
	EnvVarEndpoint = "FASTLY_API_ENDPOINT"
)

// Read populates the fields from the provided environment.
func (e *Environment) Read(env map[string]string) {
	e.Token = env[EnvVarToken]
	e.Endpoint = env[EnvVarEndpoint]
}

// Flag represents all of the configuration parameters that can be set with
// explicit flags. Consumers should bind their flag values to these fields
// directly.
type Flag struct {
	Token    string
	Verbose  bool
	Endpoint string
}
