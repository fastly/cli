package app

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/commands/version"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/fastly/kingpin"
)

// FastlyAPIClient is a ClientFactory that returns a real Fastly API client
// using the provided token and endpoint.
func FastlyAPIClient(token, endpoint string) (api.Interface, error) {
	client, err := fastly.NewClientForEndpoint(token, endpoint)
	return client, err
}

// Run constructs the application including all of the subcommands, parses the
// args, invokes the client factory with the token to create a Fastly API
// client, and executes the chosen command, using the provided io.Reader and
// io.Writer for input and output, respectively. In the real CLI, func main is
// just a simple shim to this function; it exists to make end-to-end testing of
// commands easier/possible.
//
// The Run helper should NOT output any error-related information to the out
// io.Writer. All error-related information should be encoded into an error type
// and returned to the caller. This includes usage text.
func Run(opts RunOpts) error {
	var md manifest.Data
	md.File.SetErrLog(opts.ErrLog)
	md.File.SetOutput(opts.Stdout)
	_ = md.File.Read(manifest.Filename)

	// The g will hold generally-applicable configuration parameters
	// from a variety of sources, and is provided to each concrete command.
	g := global.Data{
		Env:        opts.Env,
		ErrLog:     opts.ErrLog,
		Config:     opts.ConfigFile,
		HTTPClient: opts.HTTPClient,
		Manifest:   md,
		Output:     opts.Stdout,
		Path:       opts.ConfigPath,
	}

	// Set up the main application root, including global flags, and then each
	// of the subcommands. Note that we deliberately don't use some of the more
	// advanced features of the kingpin.Application flags, like env var
	// bindings, because we need to do things like track where a config
	// parameter came from.
	app := kingpin.New("fastly", "A tool to interact with the Fastly API")
	app.Writers(opts.Stdout, io.Discard) // don't let kingpin write error output
	app.UsageContext(&kingpin.UsageContext{
		Template: VerboseUsageTemplate,
		Funcs:    UsageTemplateFuncs,
	})

	// Prevent kingpin from calling os.Exit, this gives us greater control over
	// error states and output control flow.
	app.Terminate(nil)

	// WARNING: kingpin has no way of decorating flags as being "global"
	// therefore if you add/remove a global flag you will also need to update
	// the globalFlags map in pkg/app/usage.go which is used for usage rendering.
	// You should also update `IsGlobalFlagsOnly` in ../cmd/cmd.go
	//
	// NOTE: Global flags, unlike command flags, must be unique. This means BOTH
	// the long flag and the short flag identifiers must be unique. If you try to
	// reuse an identifier (long or short), then kingpin will trigger a runtime
	// panic 🎉
	//
	// NOTE: Short flags CAN be safely reused across commands.
	tokenHelp := fmt.Sprintf("Fastly API token (or via %s)", env.Token)
	app.Flag("accept-defaults", "Accept default options for all interactive prompts apart from Yes/No confirmations").Short('d').BoolVar(&g.Flags.AcceptDefaults)
	app.Flag("auto-yes", "Answer yes automatically to all Yes/No confirmations. This may suppress security warnings").Short('y').BoolVar(&g.Flags.AutoYes)
	app.Flag("endpoint", "Fastly API endpoint").Hidden().StringVar(&g.Flags.Endpoint)
	app.Flag("non-interactive", "Do not prompt for user input - suitable for CI processes. Equivalent to --accept-defaults and --auto-yes").Short('i').BoolVar(&g.Flags.NonInteractive)
	app.Flag("profile", "Switch account profile for single command execution (see also: 'fastly profile switch')").Short('o').StringVar(&g.Flags.Profile)
	app.Flag("quiet", "Silence all output except direct command output. This won't prevent interactive prompts (see: --accept-defaults, --auto-yes, --non-interactive)").Short('q').BoolVar(&g.Flags.Quiet)
	app.Flag("token", tokenHelp).Short('t').StringVar(&g.Flags.Token)
	app.Flag("verbose", "Verbose logging").Short('v').BoolVar(&g.Flags.Verbose)

	commands := defineCommands(app, &g, md, opts)
	command, name, err := processCommandInput(opts, app, &g, commands)
	if err != nil {
		return err
	}
	// We short-circuit the execution for specific cases:
	//
	// - cmd.ArgsIsHelpJSON() == true
	// - shell autocompletion flag provided
	switch name {
	case "help--format=json":
		fallthrough
	case "help--formatjson":
		fallthrough
	case "shell-autocomplete":
		return nil
	}

	if g.Flags.Quiet {
		md.File.SetQuiet(true)
	}

	token, source := g.Token()

	if g.Verbose() {
		displayTokenSource(
			source,
			opts.Stdout,
			env.Token,
			determineProfile(md.File.Profile, g.Flags.Profile, g.Config.Profiles),
		)
	}

	token, err = profile.Init(token, &md, &g, opts.Stdin, opts.Stdout)
	if err != nil {
		return err
	}

	// If we are using the token from config file, check the file's permissions
	// to assert if they are not too open or have been altered outside of the
	// application and warn if so.
	segs := strings.Split(name, " ")
	if source == lookup.SourceFile && (len(segs) > 0 && segs[0] != "profile") {
		if fi, err := os.Stat(config.FilePath); err == nil {
			if mode := fi.Mode().Perm(); mode > config.FilePermissions {
				if !g.Flags.Quiet {
					text.Warning(opts.Stdout, "Unprotected configuration file.")
					fmt.Fprintf(opts.Stdout, "Permissions for '%s' are too open\n", config.FilePath)
					fmt.Fprintf(opts.Stdout, "It is recommended that your configuration file is NOT accessible by others.\n")
					fmt.Fprintln(opts.Stdout)
				}
			}
		}
	}

	endpoint, source := g.Endpoint()
	if g.Verbose() {
		switch source {
		case lookup.SourceEnvironment:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint (via %s): %s\n", env.Endpoint, endpoint)
		case lookup.SourceFile:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint (via config file): %s\n", endpoint)
		default:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint: %s\n", endpoint)
		}
	}

	// NOTE: We return error immediately so there's no issue assigning to global.
	// nosemgrep
	g.APIClient, err = opts.APIClient(token, endpoint)
	if err != nil {
		g.ErrLog.Add(err)
		return fmt.Errorf("error constructing Fastly API client: %w", err)
	}

	// NOTE: We return error immediately so there's no issue assigning to global.
	// nosemgrep
	g.RTSClient, err = fastly.NewRealtimeStatsClientForEndpoint(token, fastly.DefaultRealtimeStatsEndpoint)
	if err != nil {
		g.ErrLog.Add(err)
		return fmt.Errorf("error constructing Fastly realtime stats client: %w", err)
	}

	if opts.Versioners.CLI != nil && name != "update" && !version.IsPreRelease(revision.AppVersion) {
		f := update.CheckAsync(
			revision.AppVersion,
			opts.Versioners.CLI,
			g.Flags.Quiet,
		)
		defer f(opts.Stdout) // ...and the printing function second, so we hit the timeout
	}

	return command.Exec(opts.Stdin, opts.Stdout)
}

// RunOpts represent arguments to Run()
type RunOpts struct {
	APIClient  APIClientFactory
	Args       []string
	ConfigFile config.File
	ConfigPath string
	Env        config.Environment
	ErrLog     fsterr.LogInterface
	HTTPClient api.HTTPClient
	Stdin      io.Reader
	Stdout     io.Writer
	Versioners Versioners
}

// APIClientFactory creates a Fastly API client (modeled as an api.Interface)
// from a user-provided API token. It exists as a type in order to parameterize
// the Run helper with it: in the real CLI, we can use NewClient from the Fastly
// API client library via RealClient; in tests, we can provide a mock API
// interface via MockClient.
type APIClientFactory func(token, endpoint string) (api.Interface, error)

// Versioners represents all supported versioner types.
type Versioners struct {
	CLI     github.AssetVersioner
	Viceroy github.AssetVersioner
}

// displayTokenSource prints the token source.
func displayTokenSource(source lookup.Source, out io.Writer, token, profileSource string) {
	switch source {
	case lookup.SourceFlag:
		fmt.Fprintf(out, "Fastly API token provided via --token\n")
	case lookup.SourceEnvironment:
		fmt.Fprintf(out, "Fastly API token provided via %s\n", token)
	case lookup.SourceFile:
		fmt.Fprintf(out, "Fastly API token provided via config file (profile: %s)\n", profileSource)
	default:
		fmt.Fprintf(out, "Fastly API token not provided\n")
	}
}

// determineProfile determines if the provided token was acquired via the
// fastly.toml manifest, the --profile flag, or was a default profile from
// within the config.toml application configuration.
func determineProfile(manifestValue, flagValue string, profiles config.Profiles) string {
	if manifestValue != "" {
		return manifestValue + " -- via fastly.toml"
	}
	if flagValue != "" {
		return flagValue
	}
	name, _ := profile.Default(profiles)
	return name
}
