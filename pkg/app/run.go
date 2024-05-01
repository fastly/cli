package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/kingpin"
	"github.com/fatih/color"
	"github.com/hashicorp/cap/oidc"
	"github.com/skratchdot/open-golang/open"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/auth"
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
	cmds := commands.Define(app, data)
	command, commandName, err := processCommandInput(data, app, cmds)
	if err != nil {
		return err
	}

	// We short-circuit the execution for specific cases:
	//
	// - argparser.ArgsIsHelpJSON() == true
	// - shell autocompletion flag provided
	switch commandName {
	case "help--format=json":
		fallthrough
	case "help--formatjson":
		fallthrough
	case "shell-autocomplete":
		return nil
	}

	metadataDisable, _ := strconv.ParseBool(data.Env.WasmMetadataDisable)
	if !slices.Contains(data.Args, "--metadata-disable") && !metadataDisable && !data.Config.CLI.MetadataNoticeDisplayed && commandCollectsData(commandName) {
		text.Important(data.Output, "The Fastly CLI is configured to collect data related to Wasm builds (e.g. compilation times, resource usage, and other non-identifying data). To learn more about what data is being collected, why, and how to disable it: https://www.fastly.com/documentation/reference/cli")
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

	// User can set env.DebugMode env var or the --debug-mode boolean flag.
	// This will prioritise the flag over the env var.
	if data.Flags.Debug {
		data.Env.DebugMode = "true"
	}

	// NOTE: Some commands need just the auth server to be running.
	// But not necessarily need to process an existing token.
	// e.g. `profile create example_sso_user --sso`
	// Which needs the auth server so it can start up an OAuth flow.
	if !commandRequiresToken(command) && commandRequiresAuthServer(commandName) {
		// NOTE: Checking for nil allows our test suite to mock the server.
		// i.e. it'll be nil whenever the CLI is run by a user but not `go test`.
		if data.AuthServer == nil {
			authServer, err := configureAuth(apiEndpoint, data.Args, data.Config, data.HTTPClient, data.Env)
			if err != nil {
				return fmt.Errorf("failed to configure authentication processes: %w", err)
			}
			data.AuthServer = authServer
		}
	}

	if commandRequiresToken(command) {
		// NOTE: Checking for nil allows our test suite to mock the server.
		// i.e. it'll be nil whenever the CLI is run by a user but not `go test`.
		if data.AuthServer == nil {
			authServer, err := configureAuth(apiEndpoint, data.Args, data.Config, data.HTTPClient, data.Env)
			if err != nil {
				return fmt.Errorf("failed to configure authentication processes: %w", err)
			}
			data.AuthServer = authServer
		}

		token, tokenSource, err := processToken(cmds, data)
		if err != nil {
			if errors.Is(err, fsterr.ErrDontContinue) {
				return nil // we shouldn't exit 1 if user chooses to stop
			}
			return fmt.Errorf("failed to process token: %w", err)
		}

		if data.Verbose() {
			displayToken(tokenSource, data)
		}
		if !data.Flags.Quiet {
			checkConfigPermissions(commandName, tokenSource, data.Output)
		}

		data.APIClient, data.RTSClient, err = configureClients(token, apiEndpoint, data.APIClientFactory, data.Flags.Debug)
		if err != nil {
			data.ErrLog.Add(err)
			return fmt.Errorf("error constructing client: %w", err)
		}
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

	// IMPORTANT: Kingpin doesn't support global flags.
	// Any flags defined below must also be added to two other places:
	// 1. ./usage.go (`globalFlags` map).
	// 2. ../cmd/argparser.go (`IsGlobalFlagsOnly` function).
	//
	// NOTE: Global flags (long and short) MUST be unique.
	// A subcommand can't define a flag that is already global.
	// Kingpin will otherwise trigger a runtime panic 🎉
	// Interestingly, short flags can be reused but only across subcommands.
	tokenHelp := fmt.Sprintf("Fastly API token (or via %s)", env.APIToken)
	app.Flag("accept-defaults", "Accept default options for all interactive prompts apart from Yes/No confirmations").Short('d').BoolVar(&data.Flags.AcceptDefaults)
	app.Flag("account", "Fastly Accounts endpoint").Hidden().StringVar(&data.Flags.AccountEndpoint)
	app.Flag("api", "Fastly API endpoint").Hidden().StringVar(&data.Flags.APIEndpoint)
	app.Flag("auto-yes", "Answer yes automatically to all Yes/No confirmations. This may suppress security warnings").Short('y').BoolVar(&data.Flags.AutoYes)
	// IMPORTANT: `--debug` is a built-in Kingpin flag so we must use `debug-mode`.
	app.Flag("debug-mode", "Print API request and response details (NOTE: can disrupt the normal CLI flow output formatting)").BoolVar(&data.Flags.Debug)
	// IMPORTANT: `--sso` causes a Kingpin runtime panic 🤦 so we use `enable-sso`.
	app.Flag("enable-sso", "Enable Single-Sign On (SSO) for current profile execution (see also: 'fastly sso')").Hidden().BoolVar(&data.Flags.SSO)
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
func processToken(cmds []argparser.Command, data *global.Data) (token string, tokenSource lookup.Source, err error) {
	token, tokenSource = data.Token()

	// Check if token is from a profile.
	// e.g. --profile, fastly.toml override, or config default profile.
	// If it is, then we'll check if it is expired.
	//
	// NOTE: tokens via FASTLY_API_TOKEN or --token aren't checked for a TTL.
	// This is because we don't have them persisted to disk.
	// Meaning we can't check for a TTL or access an access/refresh token.
	// So we have to presume those overrides are using a long-lived token.
	switch tokenSource {
	case lookup.SourceFile:
		profileName, profileData, err := data.Profile()
		if err != nil {
			return "", tokenSource, err
		}
		// User with long-lived token will skip SSO if they've not enabled it.
		if shouldSkipSSO(profileName, profileData, data) {
			return token, tokenSource, nil
		}
		// User now either has an existing SSO-based token or they want to migrate.
		// If a long-lived token, then trigger SSO.
		if auth.IsLongLivedToken(profileData) {
			return ssoAuthentication("You've not authenticated via OAuth before", cmds, data)
		}
		// Otherwise, for an existing SSO token, check its freshness.
		reauth, err := checkAndRefreshSSOToken(profileData, profileName, data)
		if err != nil {
			return token, tokenSource, fmt.Errorf("failed to check access/refresh token: %w", err)
		}
		if reauth {
			return ssoAuthentication("Your access token has expired and so has your refresh token", cmds, data)
		}
	case lookup.SourceUndefined:
		// If there's no token available, then trigger SSO authentication flow.
		//
		// FIXME: Remove this conditional when SSO is GA.
		// Once put back, it means "no token" == "automatic SSO".
		// For now, a brand new CLI user will have to manually create long-lived
		// tokens via the UI in order to use the Fastly CLI.
		if data.Env.UseSSO != "1" && !data.Flags.SSO {
			return "", tokenSource, nil
		}
		return ssoAuthentication("No API token could be found", cmds, data)
	case lookup.SourceEnvironment, lookup.SourceFlag, lookup.SourceDefault:
		// no-op
	}

	return token, tokenSource, nil
}

// checkAndRefreshSSOToken refreshes the access/refresh tokens if expired.
func checkAndRefreshSSOToken(profileData *config.Profile, profileName string, data *global.Data) (reauth bool, err error) {
	// Access Token has expired
	if auth.TokenExpired(profileData.AccessTokenTTL, profileData.AccessTokenCreated) {
		// Refresh Token has expired
		if auth.TokenExpired(profileData.RefreshTokenTTL, profileData.RefreshTokenCreated) {
			return true, nil
		}

		if data.Flags.Verbose {
			text.Info(data.Output, "\nYour access token has now expired. We will attempt to refresh it")
		}

		updatedJWT, err := data.AuthServer.RefreshAccessToken(profileData.RefreshToken)
		if err != nil {
			return false, fmt.Errorf("failed to refresh access token: %w", err)
		}

		email, at, err := data.AuthServer.ValidateAndRetrieveAPIToken(updatedJWT.AccessToken)
		if err != nil {
			return false, fmt.Errorf("failed to validate JWT and retrieve API token: %w", err)
		}

		// NOTE: The refresh token can sometimes be refreshed along with the access token.
		// This happens all the time in my testing but according to what is
		// spec'd this apparently is something that _might_ happen.
		// So after we get the refreshed access token, we check to see if the
		// refresh token that was returned by the API call has also changed when
		// compared to the refresh token stored in the CLI config file.
		current := profile.Get(profileName, data.Config.Profiles)
		if current == nil {
			return false, fmt.Errorf("failed to locate '%s' profile", profileName)
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
			return false, fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to update '%s' profile with new token data", profileName),
				Remediation: "Run `fastly sso` to retry.",
			}
		}
		data.Config.Profiles = ps
		if err := data.Config.Write(data.ConfigPath); err != nil {
			data.ErrLog.Add(err)
			return false, fmt.Errorf("error saving config file: %w", err)
		}
	}

	return false, nil
}

// shouldSkipSSO identifies if a config is a pre-v5 config and, if it is,
// informs the user how they can use the SSO flow. It checks if the SSO
// environment variable (or flag) has been set and enables the SSO flow if so.
func shouldSkipSSO(_ string, profileData *config.Profile, data *global.Data) bool {
	if auth.IsLongLivedToken(profileData) {
		// Skip SSO if user hasn't indicated they want to migrate.
		return data.Env.UseSSO != "1" && !data.Flags.SSO
		// FIXME: Put back messaging once SSO is GA.
		// if data.Env.UseSSO == "1" || data.Flags.SSO {
		// 	return false // don't skip SSO
		// }
		// if !data.Flags.Quiet {
		// 	if data.Flags.Verbose {
		// 		text.Break(data.Output)
		// 	}
		// 	text.Important(data.Output, "The Fastly API token used by the current '%s' profile is not a Fastly SSO (Single Sign-On) generated token. SSO-based tokens offer more security and convenience. To update your token, either set `FASTLY_USE_SSO=1` or pass `--enable-sso` before invoking the Fastly CLI. This will ensure the current profile is switched to using an SSO generated API token. Once the token has been switched over you no longer need to set `FASTLY_USE_SSO` for this profile (--token and FASTLY_API_TOKEN can still be used as overrides).\n\n", profileName)
		// }
		// return true // skip SSO
	}
	return false // don't skip SSO
}

// ssoAuthentication executes the `sso` command to handle authentication.
func ssoAuthentication(outputMessage string, cmds []argparser.Command, data *global.Data) (token string, tokenSource lookup.Source, err error) {
	for _, command := range cmds {
		commandName := strings.Split(command.Name(), " ")[0]
		if commandName == "sso" {
			if !data.Flags.AutoYes && !data.Flags.NonInteractive {
				if data.Verbose() {
					text.Break(data.Output)
				}
				text.Important(data.Output, "%s. We need to open your browser to authenticate you.", outputMessage)
				text.Break(data.Output)
				cont, err := text.AskYesNo(data.Output, text.BoldYellow("Do you want to continue? [y/N]: "), data.Input)
				text.Break(data.Output)
				if err != nil {
					return token, tokenSource, err
				}
				if !cont {
					return token, tokenSource, fsterr.ErrDontContinue
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

	// Updated token should be persisted to disk after command.Exec() completes.
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
		fmt.Fprintf(out, "Fastly API endpoint (via --api): %s\n", endpoint)
	case lookup.SourceEnvironment:
		fmt.Fprintf(out, "Fastly API endpoint (via %s): %s\n", env.APIEndpoint, endpoint)
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

// commandRequiresAuthServer determines if the command to be executed is one that
// requires just the authentication server to be running.
func commandRequiresAuthServer(command string) bool {
	switch command {
	case "profile create", "profile update":
		return true
	}
	return false
}

// commandRequiresToken determines if the command to be executed is one that
// requires an API token.
func commandRequiresToken(command argparser.Command) bool {
	commandName := command.Name()
	switch commandName {
	case "compute init":
		if initCmd, ok := command.(*compute.InitCommand); ok {
			return text.IsFastlyID(initCmd.CloneFrom)
		}
		return false
	case "compute build", "compute hash-files", "compute metadata", "compute serve":
		return false
	}
	commandName = strings.Split(commandName, " ")[0]
	switch commandName {
	case "config", "profile", "update", "version":
		return false
	}
	return true
}

// configureAuth processes authentication tasks.
//
// 1. Acquire .well-known configuration data.
// 2. Instantiate authentication server.
// 3. Start up request multiplexer.
func configureAuth(apiEndpoint string, args []string, f config.File, c api.HTTPClient, e config.Environment) (*auth.Server, error) {
	metadataEndpoint := fmt.Sprintf(auth.OIDCMetadata, accountEndpoint(args, e, f))
	req, err := http.NewRequest(http.MethodGet, metadataEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request object for OpenID Connect .well-known metadata: %w", err)
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request OpenID Connect .well-known metadata (%s): %w", metadataEndpoint, err)
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
		APIEndpoint:        apiEndpoint,
		DebugMode:          e.DebugMode,
		HTTPClient:         c,
		Result:             result,
		Router:             router,
		Verifier:           verifier,
		WellKnownEndpoints: wellknown,
	}

	router.HandleFunc("/callback", authServer.HandleCallback())

	return authServer, nil
}

// accountEndpoint parses the account endpoint from multiple locations.
func accountEndpoint(args []string, e config.Environment, cfg config.File) string {
	// Check for flag override.
	for i, a := range args {
		if a == "--account" && i+1 < len(args) {
			return args[i+1]
		}
	}
	// Check for environment override.
	if e.AccountEndpoint != "" {
		return e.AccountEndpoint
	}
	// Check for internal config override.
	if cfg.Fastly.AccountEndpoint != global.DefaultAccountEndpoint && cfg.Fastly.AccountEndpoint != "" {
		return cfg.Fastly.AccountEndpoint
	}
	// Otherwise return the default account endpoint.
	return global.DefaultAccountEndpoint
}
