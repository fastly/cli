package app

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"
	"github.com/fastly/kingpin"

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
)

// FastlyAPIClient is a ClientFactory that returns a real Fastly API client
// using the provided token and endpoint.
func FastlyAPIClient(token, endpoint string, debugMode bool) (api.Interface, error) {
	client, err := fastly.NewClientForEndpoint(token, endpoint)
	if debugMode {
		client.DebugMode = true
	}
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
	// The g will hold generally-applicable configuration parameters
	// from a variety of sources, and is provided to each concrete command.
	g := global.Data{
		Config:           opts.ConfigFile,
		ConfigPath:       opts.ConfigPath,
		Env:              opts.Env,
		ErrLog:           opts.ErrLog,
		ExecuteWasmTools: opts.ExecuteWasmTools,
		HTTPClient:       opts.HTTPClient,
		Manifest:         *opts.Manifest,
		Output:           opts.Stdout,
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
	// NOTE: Global flags (long and short) MUST be unique.
	// A subcommand can't define a flag that is already global.
	// Kingpin will otherwise trigger a runtime panic 🎉
	// Interestingly, short flags can be reused but only across subcommands.
	tokenHelp := fmt.Sprintf("Fastly API token (or via %s)", env.Token)
	app.Flag("accept-defaults", "Accept default options for all interactive prompts apart from Yes/No confirmations").Short('d').BoolVar(&g.Flags.AcceptDefaults)
	app.Flag("auto-yes", "Answer yes automatically to all Yes/No confirmations. This may suppress security warnings").Short('y').BoolVar(&g.Flags.AutoYes)
	// IMPORTANT: `--debug` is a built-in Kingpin flag so we can't use that.
	app.Flag("debug-mode", "Print API request and response details (NOTE: can disrupt the normal CLI flow output formatting)").BoolVar(&g.Flags.Debug)
	app.Flag("endpoint", "Fastly API endpoint").Hidden().StringVar(&g.Flags.Endpoint)
	app.Flag("non-interactive", "Do not prompt for user input - suitable for CI processes. Equivalent to --accept-defaults and --auto-yes").Short('i').BoolVar(&g.Flags.NonInteractive)
	app.Flag("profile", "Switch account profile for single command execution (see also: 'fastly profile switch')").Short('o').StringVar(&g.Flags.Profile)
	app.Flag("quiet", "Silence all output except direct command output. This won't prevent interactive prompts (see: --accept-defaults, --auto-yes, --non-interactive)").Short('q').BoolVar(&g.Flags.Quiet)
	app.Flag("token", tokenHelp).HintAction(env.Vars).Short('t').StringVar(&g.Flags.Token)
	app.Flag("verbose", "Verbose logging").Short('v').BoolVar(&g.Flags.Verbose)

	commands := defineCommands(app, &g, *opts.Manifest, opts)
	command, commandName, err := processCommandInput(opts, app, &g, commands)
	if err != nil {
		return err
	}
	// We short-circuit the execution for specific cases:
	//
	// - cmd.ArgsIsHelpJSON() == true
	// - shell autocompletion flag provided
	switch commandName {
	case "help--format=json":
		fallthrough
	case "help--formatjson":
		fallthrough
	case "shell-autocomplete":
		return nil
	}

	// FIXME: Tweak messaging before for 10.7.0
	// To learn more about what data is being collected, why, and how to disable it: https://developer.fastly.com/reference/cli/
	metadataDisable, _ := strconv.ParseBool(g.Env.WasmMetadataDisable)
	if slices.Contains(opts.Args, "--metadata-enable") && !metadataDisable && !g.Config.CLI.MetadataNoticeDisplayed && commandCollectsData(commandName) {
		text.Important(g.Output, "The Fastly CLI is configured to collect data related to Wasm builds (e.g. compilation times, resource usage, and other non-identifying data). To learn more about our data & privacy policies visit https://www.fastly.com/trust. Join the conversation https://bit.ly/wasm-metadata")
		text.Break(g.Output)
		g.Config.CLI.MetadataNoticeDisplayed = true
		err := g.Config.Write(g.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to persist change to metadata notice: %w", err)
		}
		time.Sleep(5 * time.Second) // this message is only displayed once so give the user a chance to see it before it possibly scrolls off screen
	}

	if g.Flags.Quiet {
		opts.Manifest.File.SetQuiet(true)
	}

	token, source := g.Token()

	// Ensure the user has configured an API token, otherwise trigger the
	// authentication flow (unless calling one of the profile commands).
	if source == lookup.SourceUndefined && !allowNoToken(commandName) {
		for _, command := range commands {
			if command.Name() == "authenticate" {
				text.Warning(opts.Stdout, "No API token could be found. We need to open your browser to authenticate you.")
				text.Break(opts.Stdout)
				cont, err := text.AskYesNo(opts.Stdout, "Are you sure you want to continue? [yes/no]: ", opts.Stdin)
				if err != nil {
					return err
				}
				if !cont {
					return nil
				}
				text.Break(opts.Stdout)

				err = command.Exec(opts.Stdin, opts.Stdout)
				if err != nil {
					return fmt.Errorf("failed to authenticate: %w", err)
				}
				break
			}
		}

		// Recheck for token (should be persisted to profile data).
		token, source = g.Token()
		if source == lookup.SourceUndefined {
			return fsterr.ErrNoToken
		}
	}

	if g.Verbose() {
		displayTokenSource(
			source,
			opts.Stdout,
			env.Token,
			determineProfile(opts.Manifest.File.Profile, g.Flags.Profile, g.Config.Profiles),
		)
	}

	token, err = profile.Init(token, opts.Manifest, &g, opts.Stdin, opts.Stdout)
	if err != nil {
		return err
	}

	// If we are using the token from config file, check the file's permissions
	// to assert if they are not too open or have been altered outside of the
	// application and warn if so.
	segs := strings.Split(commandName, " ")
	if source == lookup.SourceFile && (len(segs) > 0 && segs[0] != "profile") {
		if fi, err := os.Stat(config.FilePath); err == nil {
			if mode := fi.Mode().Perm(); mode > config.FilePermissions {
				if !g.Flags.Quiet {
					text.Warning(opts.Stdout, "Unprotected configuration file.\n\n")
					text.Output(opts.Stdout, "Permissions for '%s' are too open\n\n", config.FilePath)
					text.Output(opts.Stdout, "It is recommended that your configuration file is NOT accessible by others.\n\n")
				}
			}
		}
	}

	endpoint, source := g.Endpoint()
	if g.Verbose() {
		switch source {
		case lookup.SourceEnvironment:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint (via %s): %s\n\n", env.Endpoint, endpoint)
		case lookup.SourceFile:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint (via config file): %s\n\n", endpoint)
		case lookup.SourceFlag:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint provided via --endpoint\n\n")
		case lookup.SourceDefault, lookup.SourceUndefined:
			fallthrough
		default:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint: %s\n\n", endpoint)
		}
	}

	// NOTE: We return error immediately so there's no issue assigning to global.
	// nosemgrep
	g.APIClient, err = opts.APIClient(token, endpoint, g.Flags.Debug)
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

	if opts.Versioners.CLI != nil && commandName != "update" && !version.IsPreRelease(revision.AppVersion) {
		f := update.CheckAsync(
			revision.AppVersion,
			opts.Versioners.CLI,
			g.Flags.Quiet,
		)
		defer f(opts.Stdout) // ...and the printing function second, so we hit the timeout
	}

	return command.Exec(opts.Stdin, opts.Stdout)
}

// RunOpts represent arguments to Run().
type RunOpts struct {
	APIClient        APIClientFactory
	Args             []string
	ConfigFile       config.File
	ConfigPath       string
	Env              config.Environment
	ErrLog           fsterr.LogInterface
	ExecuteWasmTools func(bin string, args []string) error
	HTTPClient       api.HTTPClient
	Manifest         *manifest.Data
	Stdin            io.Reader
	Stdout           io.Writer
	Versioners       Versioners
}

// APIClientFactory creates a Fastly API client (modeled as an api.Interface)
// from a user-provided API token. It exists as a type in order to parameterize
// the Run helper with it: in the real CLI, we can use NewClient from the Fastly
// API client library via RealClient; in tests, we can provide a mock API
// interface via MockClient.
type APIClientFactory func(token, endpoint string, debugMode bool) (api.Interface, error)

// Versioners represents all supported versioner types.
type Versioners struct {
	CLI       github.AssetVersioner
	Viceroy   github.AssetVersioner
	WasmTools github.AssetVersioner
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
	case lookup.SourceDefault, lookup.SourceUndefined:
		fallthrough
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

// commandCollectsData determines if the command to be executed is one that
// collects data related to a Wasm binary.
func commandCollectsData(command string) bool {
	switch command {
	case "compute build", "compute hashsum", "compute hash-files", "compute publish", "compute serve":
		return true
	}
	return false
}

// allowNoToken determines if the command to be executed is one that should work
// even if there is no prior API token available.
func allowNoToken(command string) bool {
	if command == "version" || command == "whoami" || strings.HasPrefix(command, "profile ") {
		return true
	}
	return false
}
