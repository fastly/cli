package app

import (
	"errors"
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
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/cmd"
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

	app := configureKingpin(opts.Stdout, &g)
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

	endpoint, endpointSource := g.Endpoint()
	if g.Verbose() {
		displayAPIEndpoint(endpoint, endpointSource, opts.Stdout)
	}

	token, err := processToken(commands, commandName, opts.Manifest, &g, opts.Stdin, opts.Stdout)
	if err != nil {
		return fmt.Errorf("failed to process token: %w", err)
	}

	g.APIClient, g.RTSClient, err = configureClients(token, endpoint, opts.APIClient, g.Flags.Debug)
	if err != nil {
		g.ErrLog.Add(err)
		return fmt.Errorf("error constructing client: %w", err)
	}

	f := checkForUpdates(opts.Versioners.CLI, commandName, g.Flags.Quiet)
	defer f(opts.Stdout)

	return command.Exec(opts.Stdin, opts.Stdout)
}

// RunOpts represent arguments to Run().
type RunOpts struct {
	// APIClient is a factory function for creating an api.Interface type.
	APIClient APIClientFactory
	// Args are the command line arguments provided by the user.
	Args []string
	// AuthServer is an instance of the authentication server type.
	// Used for interacting with Fastly's SSO/OAuth authentication provider.
	AuthServer auth.Starter
	// ConfigFile is an instance of the CLI application config.
	ConfigFile config.File
	// ConfigPath is the location of the CLI application config.
	ConfigPath string
	// Env is an instance of the supported environment variables.
	Env config.Environment
	// ErrLog is an instance of a error log recorder.
	ErrLog fsterr.LogInterface
	// ExecuteWasmTools is a function that calls wasm-tools executable.
	// Designed to be used for mocking in the CLI test suite.
	ExecuteWasmTools func(bin string, args []string) error
	// HTTPClient is a standard HTTP client.
	HTTPClient api.HTTPClient
	// Manifest is the fastly.toml manifest file.
	Manifest *manifest.Data
	// Opener is a function that can open a browser window.
	Opener func(string) error
	// Stdin is the standard input destination.
	Stdin io.Reader
	// Stdout is the standard output destination.
	Stdout io.Writer
	// Versioners contains multiple software versioning checkers.
	// e.g. Check for latest CLI or Viceroy version.
	Versioners Versioners
}

func configureKingpin(out io.Writer, g *global.Data) *kingpin.Application {
	// Set up the main application root, including global flags, and then each
	// of the subcommands. Note that we deliberately don't use some of the more
	// advanced features of the kingpin.Application flags, like env var
	// bindings, because we need to do things like track where a config
	// parameter came from.
	app := kingpin.New("fastly", "A tool to interact with the Fastly API")
	app.Writers(out, io.Discard) // don't let kingpin write error output
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
	tokenHelp := fmt.Sprintf("Fastly API token (or via %s)", env.APIToken)
	app.Flag("accept-defaults", "Accept default options for all interactive prompts apart from Yes/No confirmations").Short('d').BoolVar(&g.Flags.AcceptDefaults)
	app.Flag("account", "Fastly Accounts endpoint").Hidden().StringVar(&g.Flags.Account)
	app.Flag("auto-yes", "Answer yes automatically to all Yes/No confirmations. This may suppress security warnings").Short('y').BoolVar(&g.Flags.AutoYes)
	// IMPORTANT: `--debug` is a built-in Kingpin flag so we can't use that.
	app.Flag("debug-mode", "Print API request and response details (NOTE: can disrupt the normal CLI flow output formatting)").BoolVar(&g.Flags.Debug)
	app.Flag("endpoint", "Fastly API endpoint").Hidden().StringVar(&g.Flags.Endpoint)
	app.Flag("non-interactive", "Do not prompt for user input - suitable for CI processes. Equivalent to --accept-defaults and --auto-yes").Short('i').BoolVar(&g.Flags.NonInteractive)
	app.Flag("profile", "Switch account profile for single command execution (see also: 'fastly profile switch')").Short('o').StringVar(&g.Flags.Profile)
	app.Flag("quiet", "Silence all output except direct command output. This won't prevent interactive prompts (see: --accept-defaults, --auto-yes, --non-interactive)").Short('q').BoolVar(&g.Flags.Quiet)
	app.Flag("token", tokenHelp).HintAction(env.Vars).Short('t').StringVar(&g.Flags.Token)
	app.Flag("verbose", "Verbose logging").Short('v').BoolVar(&g.Flags.Verbose)

	return app
}

// processToken handles all aspects related to the required API token.
//
// First we check if a profile token is defined in config and if so, we will
// validate if it has expired, and if it has we will attempt to refresh it.
//
// If both the access token and the refresh token has expired we will trigger
// the `fastly sso` command to execute.
//
// Either way, the CLI config is updated to reflect the token that was either
// refreshed or regenerated from the authentication process.
//
// Next, we check the config file's permissions.
//
// Finally, we check if there is a profile override in place (e.g. set via the
// --profile flag or using the `profile` field in the fastly.toml manifest).
func processToken(commands []cmd.Command, commandName string, m *manifest.Data, g *global.Data, in io.Reader, out io.Writer) (token string, err error) {
	warningMessage := "No API token could be found"
	var tokenSource lookup.Source
	token, tokenSource = g.Token()

	// Check if token is from fastly.toml [profile] and refresh if expired.
	tokenSource, warningMessage, err = checkProfileToken(tokenSource, commandName, warningMessage, out, g)
	if err != nil {
		return token, fmt.Errorf("failed to check profile token: %w", err)
	}

	// If there's no token available, and we need one for the invoked command,
	// then we'll trigger the SSO authentication flow.
	if tokenSource == lookup.SourceUndefined && commandRequiresToken(commandName) {
		token, tokenSource, err = ssoAuthentication(tokenSource, token, warningMessage, commands, in, out, g)
		if err != nil {
			return token, fmt.Errorf("failed to check profile token: %w", err)
		}
	}

	checkConfigPermissions(g.Flags.Quiet, commandName, tokenSource, out)

	// Check for a profile override.
	token, err = profile.Init(token, m, g, in, out)
	if err != nil {
		return "", err
	}

	if g.Verbose() {
		displayToken(tokenSource, m.File.Profile, *g, out)
	}

	return token, nil
}

// checkProfileToken can potentially modify `tokenSource` to trigger a re-auth.
// It can also return a modified `warningMessage` depending on user case.
//
// NOTE: tokens via FASTLY_API_TOKEN or --token aren't checked for a TTL.
func checkProfileToken(
	tokenSource lookup.Source,
	commandName, warningMessage string,
	out io.Writer,
	g *global.Data,
) (lookup.Source, string, error) {
	if tokenSource == lookup.SourceFile && commandRequiresToken(commandName) {
		var (
			profileData       *config.Profile
			found             bool
			name, profileName string
		)
		switch {
		case g.Flags.Profile != "": // --profile
			profileName = g.Flags.Profile
		case g.Manifest.File.Profile != "": // `profile` field in fastly.toml
			profileName = g.Manifest.File.Profile
		default:
			profileName = "default"
		}
		for name, profileData = range g.Config.Profiles {
			if (profileName == "default" && profileData.Default) || name == profileName {
				// Once we find the default profile we can update the variable to be the
				// associated profile name so later on we can use that information to
				// update the specific profile.
				if profileName == "default" {
					profileName = name
				}
				found = true
				break
			}
		}
		if !found {
			return tokenSource, warningMessage, fmt.Errorf("failed to locate '%s' profile", profileName)
		}

		// Allow user to opt-in to SSO/OAuth so they can replace their long-lived token.
		if shouldSkipSSO(profileName, profileData, out, g) {
			return tokenSource, warningMessage, nil
		}

		// If OAuth flow has never been executed for the defined token, then we're
		// dealing with a user with a pre-existing traditional token and they've
		// opted into the OAuth flow.
		if noOAuthToken(profileData) {
			warningMessage = "You've not authenticated via OAuth before"
			tokenSource = forceReAuth()
			return tokenSource, warningMessage, nil
		}

		// Access Token has expired
		if auth.TokenExpired(profileData.AccessTokenTTL, profileData.AccessTokenCreated) {
			// Refresh Token has expired
			if auth.TokenExpired(profileData.RefreshTokenTTL, profileData.RefreshTokenCreated) {
				warningMessage = "Your access token has expired and so has your refresh token"
				tokenSource = forceReAuth()
				return tokenSource, warningMessage, nil
			}

			if g.Flags.Verbose {
				text.Info(out, "Your access token has now expired. We will attempt to refresh it")
			}

			account, _ := g.Account()
			updated, err := auth.RefreshAccessToken(account, profileData.RefreshToken)
			if err != nil {
				return tokenSource, warningMessage, fmt.Errorf("failed to refresh access token: %w", err)
			}
			claims, err := auth.VerifyJWTSignature(account, updated.AccessToken)
			if err != nil {
				return tokenSource, warningMessage, fmt.Errorf("failed to verify refreshed JWT: %w", err)
			}
			email, ok := claims["email"]
			if !ok {
				return tokenSource, warningMessage, errors.New("failed to extract email from JWT claims")
			}

			// Exchange the access token for a Fastly API token
			apiEndpoint, _ := g.Endpoint()
			at, err := auth.ExchangeAccessToken(updated.AccessToken, apiEndpoint, g.HTTPClient, g.Env.DebugMode)
			if err != nil {
				return tokenSource, warningMessage, fmt.Errorf("failed to exchange access token for an API token: %w", err)
			}

			// NOTE: The refresh token can sometimes be refreshed along with the access token.
			// This happens all the time in my testing but according to what is
			// spec'd this apparently is something that _might_ happen.
			// So after we get the refreshed access token, we check to see if the
			// refresh token that was returned by the API call has also changed when
			// compared to the refresh token stored in the CLI config file.
			current := profile.Get(profileName, g.Config.Profiles)
			if current == nil {
				return tokenSource, warningMessage, fmt.Errorf("failed to locate '%s' profile", profileName)
			}
			now := time.Now().Unix()
			refreshToken := current.RefreshToken
			refreshTokenCreated := current.RefreshTokenCreated
			refreshTokenTTL := current.RefreshTokenTTL
			if current.RefreshToken != updated.RefreshToken {
				if g.Flags.Verbose {
					text.Info(out, "Your refresh token was also updated")
					text.Break(out)
				}
				refreshToken = updated.RefreshToken
				refreshTokenCreated = now
				refreshTokenTTL = updated.RefreshExpiresIn
			}

			ps, ok := profile.Edit(profileName, g.Config.Profiles, func(p *config.Profile) {
				p.AccessToken = updated.AccessToken
				p.AccessTokenCreated = now
				p.AccessTokenTTL = updated.ExpiresIn
				p.Email = email.(string)
				p.RefreshToken = refreshToken
				p.RefreshTokenCreated = refreshTokenCreated
				p.RefreshTokenTTL = refreshTokenTTL
				p.Token = at.AccessToken
			})
			if !ok {
				return tokenSource, warningMessage, fsterr.RemediationError{
					Inner:       fmt.Errorf("failed to update '%s' profile with new token data", profileName),
					Remediation: "Run `fastly sso` to retry.",
				}
			}
			g.Config.Profiles = ps
			if err := g.Config.Write(g.ConfigPath); err != nil {
				g.ErrLog.Add(err)
				return tokenSource, warningMessage, fmt.Errorf("error saving config file: %w", err)
			}
		}
	}

	return tokenSource, warningMessage, nil
}

// shouldSkipSSO identifies if a config is a pre-v4 config and, if it is,
// informs the user how they can use the OAuth flow. It checks if the OAuth
// environment variable has been set and enables the OAuth flow if so.
func shouldSkipSSO(profileName string, pd *config.Profile, out io.Writer, g *global.Data) bool {
	if noOAuthToken(pd) {
		if g.Env.UseSSO == "1" {
			return false // don't skip OAuth
		}
		if !g.Flags.Quiet {
			text.Important(out, "The Fastly API token used by the current '%s' profile is not a Fastly SSO (Single Sign-On) generated token. SSO-based tokens offer more security and convenience. To update your token, set `FASTLY_USE_SSO=1` before invoking the Fastly CLI. This will ensure the current profile is switched to using an SSO generated API token. Once the token has been switched over you no longer need to set `FASTLY_USE_SSO` for this profile (--token and FASTLY_API_TOKEN can still be used as overrides).", profileName)
			text.Break(out)
		}
		return true // skip OAuth
	}
	return false // don't skip OAuth
}

func noOAuthToken(pd *config.Profile) bool {
	// If user has followed OAuth flow before, then these will not be zero values.
	return pd.AccessToken == "" && pd.RefreshToken == "" && pd.AccessTokenCreated == 0 && pd.RefreshTokenCreated == 0
}

// To re-authenticate we simply reset the tokenSource variable.
// A later conditional block catches it and trigger a re-auth.
func forceReAuth() lookup.Source {
	return lookup.SourceUndefined
}

func ssoAuthentication(
	tokenSource lookup.Source,
	token, warningMessage string,
	commands []cmd.Command,
	in io.Reader,
	out io.Writer,
	g *global.Data,
) (string, lookup.Source, error) {
	for _, command := range commands {
		commandName := strings.Split(command.Name(), " ")[0]
		if commandName == "sso" {
			if !g.Flags.AutoYes && !g.Flags.NonInteractive {
				text.Important(out, "%s. We need to open your browser to authenticate you.", warningMessage)
				text.Break(out)
				cont, err := text.AskYesNo(out, text.BoldYellow("Do you want to continue? [y/N]: "), in)
				text.Break(out)
				if err != nil {
					return token, tokenSource, err
				}
				if !cont {
					return token, tokenSource, nil
				}
			}

			g.SkipAuthPrompt = true // skip the same prompt in `sso` command flow
			err := command.Exec(in, out)
			if err != nil {
				return token, tokenSource, fmt.Errorf("failed to authenticate: %w", err)
			}
			text.Break(out)
			break
		}
	}

	// Recheck for token (should be persisted to profile data).
	token, tokenSource = g.Token()
	if tokenSource == lookup.SourceUndefined {
		return token, tokenSource, fsterr.ErrNoToken
	}
	return token, tokenSource, nil
}

func displayToken(tokenSource lookup.Source, profileName string, g global.Data, out io.Writer) {
	profileSource := determineProfile(profileName, g.Flags.Profile, g.Config.Profiles)

	switch tokenSource {
	case lookup.SourceFlag:
		fmt.Fprintf(out, "Fastly API token provided via --token\n\n")
	case lookup.SourceEnvironment:
		fmt.Fprintf(out, "Fastly API token provided via %s\n\n", env.APIToken)
	case lookup.SourceFile:
		fmt.Fprintf(out, "Fastly API token provided via config file (profile: %s)\n\n", profileSource)
	case lookup.SourceUndefined, lookup.SourceDefault:
		fallthrough
	default:
		fmt.Fprintf(out, "Fastly API token not provided\n\n")
	}
}

// If we are using the token from config file, check the file's permissions
// to assert if they are not too open or have been altered outside of the
// application and warn if so.
func checkConfigPermissions(quietMode bool, commandName string, tokenSource lookup.Source, out io.Writer) {
	segs := strings.Split(commandName, " ")
	if tokenSource == lookup.SourceFile && (len(segs) > 0 && segs[0] != "profile") {
		if fi, err := os.Stat(config.FilePath); err == nil {
			if mode := fi.Mode().Perm(); mode > config.FilePermissions {
				if !quietMode {
					text.Warning(out, "Unprotected configuration file.\n\n")
					text.Output(out, "Permissions for '%s' are too open\n\n", config.FilePath)
					text.Output(out, "It is recommended that your configuration file is NOT accessible by others.\n\n")
				}
			}
		}
	}
}

func displayAPIEndpoint(endpoint string, endpointSource lookup.Source, verboseMode bool, out io.Writer) {
	if verboseMode {
		switch endpointSource {
		case lookup.SourceFlag:
			fmt.Fprintf(out, "Fastly API endpoint (via --endpoint): %s\n\n", endpoint)
		case lookup.SourceEnvironment:
			fmt.Fprintf(out, "Fastly API endpoint (via %s): %s\n\n", env.Endpoint, endpoint)
		case lookup.SourceFile:
			fmt.Fprintf(out, "Fastly API endpoint (via config file): %s\n\n", endpoint)
		case lookup.SourceDefault, lookup.SourceUndefined:
			fallthrough
		default:
			fmt.Fprintf(out, "Fastly API endpoint: %s\n\n", endpoint)
		}
	}
}

func configureClients(token, endpoint string, acf APIClientFactory, debugMode bool) (apiClient api.Interface, rtsClient api.RealtimeStatsInterface, err error) {
	apiClient, err = acf(token, endpoint, debugMode)
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing Fastly API client: %w", err)
	}

	rtsClient, err = fastly.NewRealtimeStatsClientForEndpoint(token, fastly.DefaultRealtimeStatsEndpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing Fastly realtime stats client: %w", err)
	}

	return apiClient, rtsClient, nil
}

func checkForUpdates(av github.AssetVersioner, commandName string, quietMode bool) func(io.Writer) {
	if av != nil && commandName != "update" && !version.IsPreRelease(revision.AppVersion) {
		return update.CheckAsync(revision.AppVersion, av, quietMode)
	}
	return func(_ io.Writer) {
		// no-op
	}
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

// commandRequiresToken determines if the command to be executed is one that
// requires an API token.
func commandRequiresToken(command string) bool {
	if command == "compute init" {
		return false
	}
	command = strings.Split(command, " ")[0]
	switch command {
	case "config", "profile", "sso", "update", "version":
		return false
	}
	return true
}
