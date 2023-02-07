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
	Env      config.Environment
	Config   config.File
	Flags    Flags
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
//
// Order of precedence:
//   - The --token flag.
//   - The FASTLY_API_TOKEN environment variable.
//   - The --profile flag's associated token.
//   - The `profile` manifest field's associated profile token.
//   - The 'default' profile associated token (if there is one).
func (d *Data) Token() (string, lookup.Source) {
	if d.Flags.Token != "" {
		return d.Flags.Token, lookup.SourceFlag
	}

	if d.Env.Token != "" {
		return d.Env.Token, lookup.SourceEnvironment
	}

	if d.Flags.Profile != "" {
		for k, v := range d.Config.Profiles {
			if k == d.Flags.Profile {
				return v.Token, lookup.SourceFile
			}
		}
	}

	if d.Manifest.File.Profile != "" {
		for k, v := range d.Config.Profiles {
			if k == d.Manifest.File.Profile {
				return v.Token, lookup.SourceFile
			}
		}
	}

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

// Flags represents all of the configuration parameters that can be set with
// explicit flags. Consumers should bind their flag values to these fields
// directly.
type Flags struct {
	AcceptDefaults bool
	AutoYes        bool
	Endpoint       string
	NonInteractive bool
	Profile        string
	Quiet          bool
	Token          string
	Verbose        bool
}
