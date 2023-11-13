package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"
	"github.com/fastly/kingpin"
	"github.com/fatih/color"
	"github.com/hashicorp/cap/oidc"
	"github.com/skratchdot/open-golang/open"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands"
	"github.com/fastly/cli/pkg/commands/compute"
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
	"github.com/fastly/cli/pkg/sync"
	"github.com/fastly/cli/pkg/text"
)

// Run kick starts the CLI application.
func Run(args []string, stdin io.Reader) error {
	data, err := Init(args, stdin)
	if err != nil {
		return fmt.Errorf("failed to initialise application: %w", err)
	}
	return Exec(data)
}

// Init constructs all the required objects and data for Exec().
//
// NOTE: We define as a package level variable so we can mock output for tests.
var Init = func(args []string, stdin io.Reader) (*global.Data, error) {
	// Parse the arguments provided by the user via the command-line interface.
	args = args[1:]

	// Define a HTTP client that will be used for making arbitrary HTTP requests.
	httpClient := &http.Client{Timeout: time.Minute * 2}

	// Define the standard input/output streams.
	var (
		in  io.Reader = stdin
		out io.Writer = sync.NewWriter(color.Output)
	)

	// Read relevant configuration options from the user's environment.
	var e config.Environment
	e.Read(env.Parse(os.Environ()))

	// Identify verbose flag early (before Kingpin parser has executed) so we can
	// print additional output related to the CLI configuration.
	var verboseOutput bool
	for _, seg := range args {
		if seg == "-v" || seg == "--verbose" {
			verboseOutput = true
		}
	}

	// Identify auto-yes/non-interactive flag early (before Kingpin parser has
	// executed) so we can handle the interactive prompts appropriately with
	// regards to processing the CLI configuration.
	var autoYes, nonInteractive bool
	for _, seg := range args {
		if seg == "-y" || seg == "--auto-yes" {
			autoYes = true
		}
		if seg == "-i" || seg == "--non-interactive" {
			nonInteractive = true
		}
	}

	// Extract a subset of configuration options from the local app directory.
	var cfg config.File
	cfg.SetAutoYes(autoYes)
	cfg.SetNonInteractive(nonInteractive)
	if err := cfg.Read(config.FilePath, in, out, fsterr.Log, verboseOutput); err != nil {
		return nil, err
	}

	// Extract user's project configuration from the fastly.toml manifest.
	var md manifest.Data
	md.File.Args = args
	md.File.SetErrLog(fsterr.Log)
	md.File.SetOutput(out)

	// NOTE: We skip handling the error because not all commands relate to Compute.
	_ = md.File.Read(manifest.Filename)

	// Configure authentication inputs.
	// We do this here so that we can mock the values in our test suite.
	req, err := http.NewRequest(http.MethodGet, auth.OIDCMetadata, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request object for OpenID Connect .well-known metadata: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request OpenID Connect .well-known metadata: %w", err)
	}
	openIDConfig, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OpenID Connect .well-known metadata: %w", err)
	}
	_ = resp.Body.Close()
	var wellknown auth.WellKnownEndpoints
	err = json.Unmarshal(openIDConfig, &wellknown)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenID Connect .well-known metadata: %w", err)
	}
	result := make(chan auth.AuthorizationResult)
	router := http.NewServeMux()
	verifier, err := oidc.NewCodeVerifier()
	if err != nil {
		return nil, fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to generate a code verifier for SSO authentication server: %w", err),
			Remediation: auth.Remediation,
		}
	}
	authServer := &auth.Server{
		DebugMode:          e.DebugMode,
		HTTPClient:         httpClient,
		Result:             result,
		Router:             router,
		Verifier:           verifier,
		WellKnownEndpoints: wellknown,
	}
	router.HandleFunc("/callback", authServer.HandleCallback())

	factory := func(token, endpoint string, debugMode bool) (api.Interface, error) {
		client, err := fastly.NewClientForEndpoint(token, endpoint)
		if debugMode {
			client.DebugMode = true
		}
		return client, err
	}

	versioners := global.Versioners{
		CLI: github.New(github.Opts{
			HTTPClient: httpClient,
			Org:        "fastly",
			Repo:       "cli",
			Binary:     "fastly",
		}),
		Viceroy: github.New(github.Opts{
			HTTPClient: httpClient,
			Org:        "fastly",
			Repo:       "viceroy",
			Binary:     "viceroy",
			Version:    md.File.LocalServer.ViceroyVersion,
		}),
		WasmTools: github.New(github.Opts{
			HTTPClient: httpClient,
			Org:        "bytecodealliance",
			Repo:       "wasm-tools",
			Binary:     "wasm-tools",
			External:   true,
			Nested:     true,
		}),
	}

	return &global.Data{
		APIClientFactory: factory,
		Args:             args,
		AuthServer:       authServer,
		Config:           cfg,
		ConfigPath:       config.FilePath,
		Env:              e,
		ErrLog:           fsterr.Log,
		ExecuteWasmTools: compute.ExecuteWasmTools,
		HTTPClient:       httpClient,
		Manifest:         &md,
		Opener:           open.Run,
		Output:           out,
		Versioners:       versioners,
		Input:            in,
	}, nil
}

// Exec constructs the application including all of the subcommands, parses the
// args, invokes the client factory with the token to create a Fastly API
// client, and executes the chosen command, using the provided io.Reader and
// io.Writer for input and output, respectively. In the real CLI, func main is
// just a simple shim to this function; it exists to make end-to-end testing of
// commands easier/possible.
//
// The Exec helper should NOT output any error-related information to the out
// io.Writer. All error-related information should be encoded into an error type
// and returned to the caller. This includes usage text.
func Exec(data *global.Data) error {
	app := configureKingpin(data)
	commands := commands.Define(app, data)
	command, commandName, err := processCommandInput(data, app, commands)
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
	metadataDisable, _ := strconv.ParseBool(data.Env.WasmMetadataDisable)
	if slices.Contains(data.Args, "--metadata-enable") && !metadataDisable && !data.Config.CLI.MetadataNoticeDisplayed && commandCollectsData(commandName) {
		text.Important(data.Output, "The Fastly CLI is configured to collect data related to Wasm builds (e.g. compilation times, resource usage, and other non-identifying data). To learn more about our data & privacy policies visit https://www.fastly.com/trust. Join the conversation https://bit.ly/wasm-metadata")
		text.Break(data.Output)
		data.Config.CLI.MetadataNoticeDisplayed = true
		err := data.Config.Write(data.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to persist change to metadata notice: %w", err)
		}
		time.Sleep(5 * time.Second) // this message is only displayed once so give the user a chance to see it before it possibly scrolls off screen
	}

	if data.Flags.Quiet {
		data.Manifest.File.SetQuiet(true)
	}

	apiEndpoint, endpointSource := data.APIEndpoint()
	if data.Verbose() {
		displayAPIEndpoint(apiEndpoint, endpointSource, data.Output)
	}

	token, err := processToken(commands, commandName, data)
	if err != nil {
		return fmt.Errorf("failed to process token: %w", err)
	}

	data.APIClient, data.RTSClient, err = configureClients(token, apiEndpoint, data.APIClientFactory, data.Flags.Debug)
	if err != nil {
		data.ErrLog.Add(err)
		return fmt.Errorf("error constructing client: %w", err)
	}

	f := checkForUpdates(data.Versioners.CLI, commandName, data.Flags.Quiet)
	defer f(data.Output)

	return command.Exec(data.Input, data.Output)
}

func configureKingpin(data *global.Data) *kingpin.Application {
	// Set up the main application root, including global flags, and then each
	// of the subcommands. Note that we deliberately don't use some of the more
	// advanced features of the kingpin.Application flags, like env var
	// bindings, because we need to do things like track where a config
	// parameter came from.
	app := kingpin.New("fastly", "A tool to interact with the Fastly API")
	app.Writers(data.Output, io.Discard) // don't let kingpin write error output
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
	app.Flag("accept-defaults", "Accept default options for all interactive prompts apart from Yes/No confirmations").Short('d').BoolVar(&data.Flags.AcceptDefaults)
	app.Flag("account", "Fastly Accounts endpoint").Hidden().StringVar(&data.Flags.Account)
	app.Flag("auto-yes", "Answer yes automatically to all Yes/No confirmations. This may suppress security warnings").Short('y').BoolVar(&data.Flags.AutoYes)
	// IMPORTANT: `--debug` is a built-in Kingpin flag so we can't use that.
	app.Flag("debug-mode", "Print API request and response details (NOTE: can disrupt the normal CLI flow output formatting)").BoolVar(&data.Flags.Debug)
	app.Flag("endpoint", "Fastly API endpoint").Hidden().StringVar(&data.Flags.Endpoint)
	app.Flag("non-interactive", "Do not prompt for user input - suitable for CI processes. Equivalent to --accept-defaults and --auto-yes").Short('i').BoolVar(&data.Flags.NonInteractive)
	app.Flag("profile", "Switch account profile for single command execution (see also: 'fastly profile switch')").Short('o').StringVar(&data.Flags.Profile)
	app.Flag("quiet", "Silence all output except direct command output. This won't prevent interactive prompts (see: --accept-defaults, --auto-yes, --non-interactive)").Short('q').BoolVar(&data.Flags.Quiet)
	app.Flag("token", tokenHelp).HintAction(env.Vars).Short('t').StringVar(&data.Flags.Token)
	app.Flag("verbose", "Verbose logging").Short('v').BoolVar(&data.Flags.Verbose)

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
func processToken(commands []cmd.Command, commandName string, data *global.Data) (token string, err error) {
	warningMessage := "No API token could be found"
	var tokenSource lookup.Source
	token, tokenSource = data.Token()

	// Check if token is from fastly.toml [profile] and refresh if expired.
	tokenSource, warningMessage, err = checkProfileToken(tokenSource, commandName, warningMessage, data)
	if err != nil {
		return token, fmt.Errorf("failed to check profile token: %w", err)
	}

	// If there's no token available, and we need one for the invoked command,
	// then we'll trigger the SSO authentication flow.
	if tokenSource == lookup.SourceUndefined && commandRequiresToken(commandName) {
		token, tokenSource, err = ssoAuthentication(tokenSource, token, warningMessage, commands, data)
		if err != nil {
			return token, fmt.Errorf("failed to check profile token: %w", err)
		}
	}

	if data.Verbose() {
		displayToken(tokenSource, data)
	}
	if !data.Flags.Quiet {
		checkConfigPermissions(commandName, tokenSource, data.Output)
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
	data *global.Data,
) (lookup.Source, string, error) {
	if tokenSource == lookup.SourceFile && commandRequiresToken(commandName) {
		var (
			profileData       *config.Profile
			found             bool
			name, profileName string
		)
		switch {
		case data.Flags.Profile != "": // --profile
			profileName = data.Flags.Profile
		case data.Manifest.File.Profile != "": // `profile` field in fastly.toml
			profileName = data.Manifest.File.Profile
		default:
			profileName = "default"
		}
		for name, profileData = range data.Config.Profiles {
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
		if shouldSkipSSO(profileName, profileData, data) {
			return tokenSource, warningMessage, nil
		}

		// If OAuth flow has never been executed for the defined token, then we're
		// dealing with a user with a pre-existing traditional token and they've
		// opted into the OAuth flow.
		if noSSOToken(profileData) {
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

			if data.Flags.Verbose {
				text.Info(data.Output, "Your access token has now expired. We will attempt to refresh it")
			}
			accountEndpoint, _ := data.AccountEndpoint()
			apiEndpoint, _ := data.APIEndpoint()

			updatedJWT, err := auth.RefreshAccessToken(accountEndpoint, profileData.RefreshToken)
			if err != nil {
				return tokenSource, warningMessage, fmt.Errorf("failed to refresh access token: %w", err)
			}

			email, at, err := auth.ValidateAndRetrieveAPIToken(accountEndpoint, apiEndpoint, updatedJWT.AccessToken, data.Env.DebugMode, data.HTTPClient)
			if err != nil {
				return tokenSource, warningMessage, fmt.Errorf("failed to validate JWT and retrieve API token: %w", err)
			}

			// NOTE: The refresh token can sometimes be refreshed along with the access token.
			// This happens all the time in my testing but according to what is
			// spec'd this apparently is something that _might_ happen.
			// So after we get the refreshed access token, we check to see if the
			// refresh token that was returned by the API call has also changed when
			// compared to the refresh token stored in the CLI config file.
			current := profile.Get(profileName, data.Config.Profiles)
			if current == nil {
				return tokenSource, warningMessage, fmt.Errorf("failed to locate '%s' profile", profileName)
			}
			now := time.Now().Unix()
			refreshToken := current.RefreshToken
			refreshTokenCreated := current.RefreshTokenCreated
			refreshTokenTTL := current.RefreshTokenTTL
			if current.RefreshToken != updatedJWT.RefreshToken {
				if data.Flags.Verbose {
					text.Info(data.Output, "Your refresh token was also updated")
					text.Break(data.Output)
				}
				refreshToken = updatedJWT.RefreshToken
				refreshTokenCreated = now
				refreshTokenTTL = updatedJWT.RefreshExpiresIn
			}

			ps, ok := profile.Edit(profileName, data.Config.Profiles, func(p *config.Profile) {
				p.AccessToken = updatedJWT.AccessToken
				p.AccessTokenCreated = now
				p.AccessTokenTTL = updatedJWT.ExpiresIn
				p.Email = email
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
			data.Config.Profiles = ps
			if err := data.Config.Write(data.ConfigPath); err != nil {
				data.ErrLog.Add(err)
				return tokenSource, warningMessage, fmt.Errorf("error saving config file: %w", err)
			}
		}
	}

	return tokenSource, warningMessage, nil
}

// shouldSkipSSO identifies if a config is a pre-v5 config and, if it is,
// informs the user how they can use the SSO flow. It checks if the SSO
// environment variable has been set and enables the SSO flow if so.
func shouldSkipSSO(profileName string, pd *config.Profile, data *global.Data) bool {
	if noSSOToken(pd) {
		return data.Env.UseSSO != "1"
		// FIXME: Put back messaging once SSO is GA.
		// if data.Env.UseSSO == "1" {
		// 	return false // don't skip SSO
		// }
		// if !data.Flags.Quiet {
		// 	if data.Flags.Verbose {
		// 		text.Break(data.Output)
		// 	}
		// 	text.Important(data.Output, "The Fastly API token used by the current '%s' profile is not a Fastly SSO (Single Sign-On) generated token. SSO-based tokens offer more security and convenience. To update your token, set `FASTLY_USE_SSO=1` before invoking the Fastly CLI. This will ensure the current profile is switched to using an SSO generated API token. Once the token has been switched over you no longer need to set `FASTLY_USE_SSO` for this profile (--token and FASTLY_API_TOKEN can still be used as overrides).\n\n", profileName)
		// }
		// return true // skip SSO
	}
	return false // don't skip SSO
}

func noSSOToken(pd *config.Profile) bool {
	// If user has followed SSO flow before, then these will not be zero values.
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
	data *global.Data,
) (string, lookup.Source, error) {
	for _, command := range commands {
		commandName := strings.Split(command.Name(), " ")[0]
		if commandName == "sso" {
			if !data.Flags.AutoYes && !data.Flags.NonInteractive {
				text.Important(data.Output, "%s. We need to open your browser to authenticate you.", warningMessage)
				text.Break(data.Output)
				cont, err := text.AskYesNo(data.Output, text.BoldYellow("Do you want to continue? [y/N]: "), data.Input)
				text.Break(data.Output)
				if err != nil {
					return token, tokenSource, err
				}
				if !cont {
					return token, tokenSource, nil
				}
			}

			data.SkipAuthPrompt = true // skip the same prompt in `sso` command flow
			err := command.Exec(data.Input, data.Output)
			if err != nil {
				return token, tokenSource, fmt.Errorf("failed to authenticate: %w", err)
			}
			text.Break(data.Output)
			break
		}
	}

	// Recheck for token (should be persisted to profile data).
	token, tokenSource = data.Token()
	if tokenSource == lookup.SourceUndefined {
		return token, tokenSource, fsterr.ErrNoToken
	}
	return token, tokenSource, nil
}

func displayToken(tokenSource lookup.Source, data *global.Data) {
	profileSource := determineProfile(data.Manifest.File.Profile, data.Flags.Profile, data.Config.Profiles)

	switch tokenSource {
	case lookup.SourceFlag:
		fmt.Fprintf(data.Output, "Fastly API token provided via --token\n\n")
	case lookup.SourceEnvironment:
		fmt.Fprintf(data.Output, "Fastly API token provided via %s\n\n", env.APIToken)
	case lookup.SourceFile:
		fmt.Fprintf(data.Output, "Fastly API token provided via config file (profile: %s)\n\n", profileSource)
	case lookup.SourceUndefined, lookup.SourceDefault:
		fallthrough
	default:
		fmt.Fprintf(data.Output, "Fastly API token not provided\n\n")
	}
}

// If we are using the token from config file, check the file's permissions
// to assert if they are not too open or have been altered outside of the
// application and warn if so.
func checkConfigPermissions(commandName string, tokenSource lookup.Source, out io.Writer) {
	segs := strings.Split(commandName, " ")
	if tokenSource == lookup.SourceFile && (len(segs) > 0 && segs[0] != "profile") {
		if fi, err := os.Stat(config.FilePath); err == nil {
			if mode := fi.Mode().Perm(); mode > config.FilePermissions {
				text.Warning(out, "Unprotected configuration file.\n\n")
				text.Output(out, "Permissions for '%s' are too open\n\n", config.FilePath)
				text.Output(out, "It is recommended that your configuration file is NOT accessible by others.\n\n")
			}
		}
	}
}

func displayAPIEndpoint(endpoint string, endpointSource lookup.Source, out io.Writer) {
	switch endpointSource {
	case lookup.SourceFlag:
		fmt.Fprintf(out, "Fastly API endpoint (via --endpoint): %s\n", endpoint)
	case lookup.SourceEnvironment:
		fmt.Fprintf(out, "Fastly API endpoint (via %s): %s\n", env.Endpoint, endpoint)
	case lookup.SourceFile:
		fmt.Fprintf(out, "Fastly API endpoint (via config file): %s\n", endpoint)
	case lookup.SourceDefault, lookup.SourceUndefined:
		fallthrough
	default:
		fmt.Fprintf(out, "Fastly API endpoint: %s\n", endpoint)
	}
}

func configureClients(token, apiEndpoint string, acf global.APIClientFactory, debugMode bool) (apiClient api.Interface, rtsClient api.RealtimeStatsInterface, err error) {
	apiClient, err = acf(token, apiEndpoint, debugMode)
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
	switch command {
	case "compute init", "compute metadata":
		// NOTE: Most `compute` commands require a token except init/metadata.
		return false
	}
	command = strings.Split(command, " ")[0]
	switch command {
	case "config", "profile", "sso", "update", "version":
		return false
	}
	return true
}
