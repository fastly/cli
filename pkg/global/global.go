package global

import (
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/manifest"
)

// DefaultEndpoint is the default Fastly API endpoint.
const DefaultEndpoint = "https://api.fastly.com"

// DefaultAccount is the default Fastly Accounts endpoint.
const DefaultAccount = "https://accounts.fastly.com"

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
	// Env is all the data that is provided by the environment.
	Env config.Environment
	// Config is an instance of the CLI configuration data.
	Config config.File
	// ConfigPath is the path to the CLI's application configuration.
	ConfigPath string
	// ExecuteWasmTools is a function that executes the wasm-tools binary.
	ExecuteWasmTools func(bin string, args []string) error
	// Flags are all the global CLI flags.
	Flags Flags
	// Manifest is the fastly.toml manifest file.
	Manifest manifest.Data
	// Output is the output for displaying information (typically os.Stdout)
	Output io.Writer
	// SkipAuthPrompt is used to indicate to the `sso` command that the
	// interactive prompt can be skipped. This is for scenarios where the command
	// is executed directly by the user.
	SkipAuthPrompt bool

	// Custom interfaces

	// ErrLog provides an interface for recording errors to disk.
	ErrLog fsterr.LogInterface
	// APIClient is a Fastly API client instance.
	APIClient api.Interface
	// HTTPClient is a HTTP client.
	HTTPClient api.HTTPClient
	// RTSClient is a Fastly API client instance for the Real Time Stats endpoints.
	RTSClient api.RealtimeStatsInterface
}

// Token yields the Fastly API token.
//
// Order of precedence:
//   - The --token flag.
//   - The FASTLY_API_TOKEN environment variable.
//   - The --profile flag's associated token.
//   - The `profile` manifest field's associated profile token.
//   - The 'default' profile associated token (if there is one).
func (d *Data) Token() (string, lookup.Source) {
	// --token
	if d.Flags.Token != "" {
		return d.Flags.Token, lookup.SourceFlag
	}

	// FASTLY_API_TOKEN
	if d.Env.APIToken != "" {
		return d.Env.APIToken, lookup.SourceEnvironment
	}

	// --profile
	if d.Flags.Profile != "" {
		for k, v := range d.Config.Profiles {
			if k == d.Flags.Profile {
				return v.Token, lookup.SourceFile
			}
		}
	}

	// `profile` field in fastly.toml
	if d.Manifest.File.Profile != "" {
		for k, v := range d.Config.Profiles {
			if k == d.Manifest.File.Profile {
				return v.Token, lookup.SourceFile
			}
		}
	}

	// [profile] section in app config
	for _, v := range d.Config.Profiles {
		if v.Default {
			return v.Token, lookup.SourceFile
		}
	}

	return "", lookup.SourceUndefined
}

// Verbose yields the verbose flag, which can only be set via flags.
func (d *Data) Verbose() bool {
	return d.Flags.Verbose
}

// Endpoint yields the API endpoint.
func (d *Data) Endpoint() (string, lookup.Source) {
	if d.Flags.Endpoint != "" {
		return d.Flags.Endpoint, lookup.SourceFlag
	}

	if d.Env.Endpoint != "" {
		return d.Env.Endpoint, lookup.SourceEnvironment
	}

	if d.Config.Fastly.APIEndpoint != DefaultEndpoint && d.Config.Fastly.APIEndpoint != "" {
		return d.Config.Fastly.APIEndpoint, lookup.SourceFile
	}

	return DefaultEndpoint, lookup.SourceDefault // this method should not fail
}

// Account yields the Accounts endpoint.
func (d *Data) Account() (string, lookup.Source) {
	if d.Flags.Account != "" {
		return d.Flags.Account, lookup.SourceFlag
	}

	if d.Env.Account != "" {
		return d.Env.Account, lookup.SourceEnvironment
	}

	if d.Config.Fastly.AccountEndpoint != DefaultAccount && d.Config.Fastly.AccountEndpoint != "" {
		return d.Config.Fastly.AccountEndpoint, lookup.SourceFile
	}

	return DefaultAccount, lookup.SourceDefault // this method should not fail
}

// Flags represents all of the configuration parameters that can be set with
// explicit flags. Consumers should bind their flag values to these fields
// directly.
type Flags struct {
	AcceptDefaults bool
	Account        string
	AutoYes        bool
	Debug          bool
	Endpoint       string
	NonInteractive bool
	Profile        string
	Quiet          bool
	Token          string
	Verbose        bool
}
