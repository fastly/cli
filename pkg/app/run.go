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

	"github.com/fatih/color"
	"github.com/hashicorp/cap/oidc"
	"github.com/skratchdot/open-golang/open"

	"github.com/fastly/go-fastly/v13/fastly"

	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/auth"
	"github.com/fastly/cli/pkg/commands"
	authcmd "github.com/fastly/cli/pkg/commands/auth"
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
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/sync"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/useragent"
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
		in            = stdin
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

	// Identify debug-mode flag early (before Kingpin parser has executed) so we
	// can inform the github versioners that we're in debug mode.
	var debugMode bool
	for _, seg := range args {
		if seg == "--debug-mode" {
			debugMode = true
		}
	}

	versioners := global.Versioners{
		CLI: github.New(github.Opts{
			DebugMode:  debugMode,
			HTTPClient: httpClient,
			Org:        "fastly",
			Repo:       "cli",
			Binary:     "fastly",
		}),
		Viceroy: github.New(github.Opts{
			DebugMode:  debugMode,
			HTTPClient: httpClient,
			Org:        "fastly",
			Repo:       "viceroy",
			Binary:     "viceroy",
			Version:    md.File.LocalServer.ViceroyVersion,
		}),
		WasmTools: github.New(github.Opts{
			DebugMode:  debugMode,
			HTTPClient: httpClient,
			Org:        "bytecodealliance",
			Repo:       "wasm-tools",
			Binary:     "wasm-tools",
			External:   true,
			Nested:     true,
		}),
	}

	// If a UserAgent extension has been set in the environment,
	// apply it
	if e.UserAgentExtension != "" {
		useragent.SetExtension(e.UserAgentExtension)
	}
	// Override the go-fastly UserAgent value by prepending the CLI version.
	//
	// Results in a header similar to:
	// User-Agent: FastlyCLI/v11.3.0, FastlyGo/10.5.0 (+github.com/fastly/go-fastly; go1.24.3)
	// (with any extension supplied above between the FastlyCLI and FastlyGo values)
	fastly.UserAgent = fmt.Sprintf("%s, %s", useragent.Name, fastly.UserAgent)

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

	// Check for --json flag early and set quiet mode if found.
	if slices.Contains(data.Args, "--json") {
		data.Flags.Quiet = true
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
	if !slices.Contains(data.Args, "--metadata-disable") && !metadataDisable && !data.Config.CLI.MetadataNoticeDisplayed && commandCollectsData(commandName) && !data.Flags.Quiet {
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

	// Migrate legacy profiles to [auth] section.
	// MigrateProfilesToAuth merges without overwriting existing auth entries.
	if len(data.Config.Profiles) > 0 {
		data.Config.MigrateProfilesToAuth()
		data.Config.Profiles = nil
		if err := data.Config.Write(data.ConfigPath); err != nil {
			data.ErrLog.Add(err)
		}
	}

	apiEndpoint, endpointSource := data.APIEndpoint()
	if data.Verbose() && !commandSuppressesVerbose(command) {
		displayAPIEndpoint(apiEndpoint, endpointSource, data.Output)
	}

	// User can set env.DebugMode env var or the --debug-mode boolean flag.
	// This will prioritise the flag over the env var.
	if data.Flags.Debug {
		data.Env.DebugMode = "true"
	}

	// NOTE: Some commands need just the auth server to be running
	// but not necessarily need to process an existing token.
	needsAuthServer := commandRequiresAuthServer(commandName, data.Args)
	if !commandRequiresToken(command) && needsAuthServer {
		// NOTE: Checking for nil allows our test suite to mock the server.
		// i.e. it'll be nil whenever the CLI is run by a user but not `go test`.
		if data.AuthServer == nil {
			authServer, err := configureAuth(apiEndpoint, data.Args, data.Config, data.HTTPClient, data.Env)
			if err != nil {
				// Non-fatal: SSO flows will detect the nil auth server and
				// report a clear error.
				data.ErrLog.Add(err)
			} else {
				data.AuthServer = authServer
			}
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

		if !data.Flags.Quiet && data.Flags.Token == "" && data.Env.APIToken == "" && data.Manifest != nil && data.Manifest.File.Profile != "" {
			if data.Config.GetAuthToken(data.Manifest.File.Profile) == nil {
				if defaultName, _ := data.Config.GetDefaultAuthToken(); defaultName != "" {
					text.Warning(data.Output, "fastly.toml profile %q not found in auth config, using default token %q.\n", data.Manifest.File.Profile, defaultName)
				} else {
					text.Warning(data.Output, "fastly.toml profile %q not found in auth config and no default token is configured.\n", data.Manifest.File.Profile)
				}
			}
		}

		token, tokenSource, err := processToken(data)
		if err != nil {
			if errors.Is(err, fsterr.ErrDontContinue) {
				return nil // we shouldn't exit 1 if user chooses to stop
			}
			return fmt.Errorf("failed to process token: %w", err)
		}

		if token == "" && tokenSource == lookup.SourceUndefined {
			token, tokenSource, err = promptForAuth(data)
			if err != nil {
				if errors.Is(err, fsterr.ErrDontContinue) {
					return nil
				}
				return fmt.Errorf("failed to process token: %w", err)
			}
		}

		if data.Verbose() && !commandSuppressesVerbose(command) {
			displayToken(tokenSource, data)
		}
		if !data.Flags.Quiet {
			checkConfigPermissions(tokenSource, data.Output)
		}

		data.APIClient, data.RTSClient, err = configureClients(token, apiEndpoint, data.APIClientFactory, data.Flags.Debug)
		if err != nil {
			data.ErrLog.Add(err)
			return fmt.Errorf("error constructing client: %w", err)
		}
	}

	checkTokenExpirationWarning(data, commandName)

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
	app.Flag("accept-defaults", "Accept default options for all interactive prompts apart from Yes/No confirmations").Short('d').BoolVar(&data.Flags.AcceptDefaults)
	app.Flag("account", "Fastly Accounts endpoint").Hidden().StringVar(&data.Flags.AccountEndpoint)
	app.Flag("api", "Fastly API endpoint").Hidden().StringVar(&data.Flags.APIEndpoint)
	app.Flag("auto-yes", "Answer yes automatically to all Yes/No confirmations. This may suppress security warnings").Short('y').BoolVar(&data.Flags.AutoYes)
	// IMPORTANT: `--debug` is a built-in Kingpin flag so we must use `debug-mode`.
	app.Flag("debug-mode", "Print API request and response details (NOTE: can disrupt the normal CLI flow output formatting)").BoolVar(&data.Flags.Debug)
	// IMPORTANT: `--sso` causes a Kingpin runtime panic 🤦 so we use `enable-sso`.
	app.Flag("enable-sso", "[DEPRECATED: use 'fastly auth login --sso --token <name>'] Enable SSO for current profile").Hidden().BoolVar(&data.Flags.SSO)
	app.Flag("non-interactive", "Do not prompt for user input - suitable for CI processes. Equivalent to --accept-defaults and --auto-yes").Short('i').BoolVar(&data.Flags.NonInteractive)
	app.Flag("profile", "[DEPRECATED: use 'fastly auth use'] Switch account profile for single command execution").Hidden().Short('o').StringVar(&data.Flags.Profile)
	app.Flag("quiet", "Silence all output except direct command output. This won't prevent interactive prompts (see: --accept-defaults, --auto-yes, --non-interactive)").Short('q').BoolVar(&data.Flags.Quiet)
	if !env.AuthCommandDisabled() {
		tokenHelp := fmt.Sprintf("Fastly API token, or name of a stored auth token (use 'default' for the default token). Falls back to %s env var", env.APIToken)
		app.Flag("token", tokenHelp).HintAction(env.Vars).Short('t').StringVar(&data.Flags.Token)
	}
	app.Flag("verbose", "Verbose logging").Short('v').BoolVar(&data.Flags.Verbose)

	return app
}

// processToken handles all aspects related to the required API token.
//
// For [auth] SSO tokens, we check freshness and attempt to refresh if expired.
// If both access and refresh tokens are expired, we trigger the SSO flow.
//
// Tokens from --token (raw, unavailable when FASTLY_DISABLE_AUTH_COMMAND is
// set) or FASTLY_API_TOKEN are assumed to be valid.
func processToken(data *global.Data) (token string, tokenSource lookup.Source, err error) {
	token, tokenSource = data.Token()

	switch tokenSource {
	case lookup.SourceUndefined:
		if data.Flags.NonInteractive || data.Flags.AutoYes || data.Flags.AcceptDefaults {
			return "", tokenSource, fsterr.ErrNonInteractiveNoToken()
		}
		if data.Env.UseSSO == "1" {
			return ssoAuthentication("No API token could be found", data, false)
		}
		return "", tokenSource, nil
	case lookup.SourceAuth:
		name := data.AuthTokenName()
		if name == "" {
			break
		}
		at := data.Config.GetAuthToken(name)
		if at != nil && at.Type == config.AuthTokenTypeSSO && at.RefreshToken != "" {
			reauth, err := checkAndRefreshAuthSSOToken(name, at, data)
			if err != nil {
				if errors.Is(err, auth.ErrInvalidGrant) {
					return ssoAuthentication("We can't refresh your token", data, true)
				}
				return token, tokenSource, fmt.Errorf("failed to refresh auth token %q: %w", name, err)
			}
			if reauth {
				return ssoAuthentication("Your auth token has expired and needs re-authentication", data, false)
			}
			token = at.Token
		}
	case lookup.SourceEnvironment, lookup.SourceFlag, lookup.SourceDefault, lookup.SourceFile:
		// no-op
	}

	return token, tokenSource, nil
}

// checkAndRefreshAuthSSOToken refreshes an SSO-type [auth] token if expired.
func checkAndRefreshAuthSSOToken(name string, at *config.AuthToken, data *global.Data) (reauth bool, err error) {
	if at.AccessExpiresAt == "" {
		return false, nil // no expiry info, assume still valid
	}

	accessExpires, err := time.Parse(time.RFC3339, at.AccessExpiresAt)
	if err != nil {
		return false, fmt.Errorf("invalid access_expires_at %q: %w", at.AccessExpiresAt, err)
	}

	// Access token still valid.
	if time.Now().Before(accessExpires) {
		return false, nil
	}

	// Access token expired; check if refresh token is also expired.
	if at.RefreshExpiresAt != "" {
		refreshExpires, err := time.Parse(time.RFC3339, at.RefreshExpiresAt)
		if err != nil {
			return false, fmt.Errorf("invalid refresh_expires_at %q: %w", at.RefreshExpiresAt, err)
		}
		if time.Now().After(refreshExpires) {
			return true, nil // both expired, needs full re-auth
		}
	}

	if data.AuthServer == nil {
		return true, nil // can't refresh without auth server
	}

	if data.Flags.Verbose {
		text.Info(data.Output, "\nYour access token has now expired. We will attempt to refresh it")
	}

	updatedJWT, err := data.AuthServer.RefreshAccessToken(at.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidGrant) {
			return true, nil // refresh token rejected, needs full re-auth
		}
		return false, fmt.Errorf("failed to refresh access token: %w", err)
	}

	email, apiToken, err := data.AuthServer.ValidateAndRetrieveAPIToken(updatedJWT.AccessToken)
	if err != nil {
		return false, fmt.Errorf("failed to validate JWT and retrieve API token: %w", err)
	}

	now := time.Now()
	at.Token = apiToken.AccessToken
	at.Email = email
	at.AccessToken = updatedJWT.AccessToken
	at.AccessExpiresAt = now.Add(time.Duration(updatedJWT.ExpiresIn) * time.Second).Format(time.RFC3339)

	// Refresh token may also be rotated.
	if at.RefreshToken != updatedJWT.RefreshToken {
		if data.Flags.Verbose {
			text.Info(data.Output, "Your refresh token was also updated")
			text.Break(data.Output)
		}
		at.RefreshToken = updatedJWT.RefreshToken
		at.RefreshExpiresAt = now.Add(time.Duration(updatedJWT.RefreshExpiresIn) * time.Second).Format(time.RFC3339)
	}

	authcmd.EnrichWithTokenSelf(data, at)

	data.Config.SetAuthToken(name, at)
	if err := data.Config.Write(data.ConfigPath); err != nil {
		data.ErrLog.Add(err)
		return false, fmt.Errorf("error saving config file: %w", err)
	}

	return false, nil
}

// authRelatedCommands lists the top-level command families related to
// authentication. Expiry warnings are suppressed for these commands to avoid
// noise during login, token management, and identity flows.
// This matches the set hidden by FASTLY_DISABLE_AUTH_COMMAND (pkg/env/env.go).
var authRelatedCommands = []string{"auth", "auth-token", "sso", "profile", "whoami"}

// checkTokenExpirationWarning prints a warning if the active stored token is
// about to expire. Only fires for SourceAuth tokens; env/flag tokens are opaque.
// Suppressed for auth-related commands and when --quiet or --json is active.
func checkTokenExpirationWarning(data *global.Data, commandName string) {
	if data.Flags.Quiet {
		return
	}
	if isAuthRelatedCommand(commandName) {
		return
	}

	_, src := data.Token()
	if src != lookup.SourceAuth {
		return
	}

	name := data.AuthTokenName()
	if name == "" {
		name = data.Config.Auth.Default
	}
	at := data.Config.GetAuthToken(name)
	if at == nil {
		return
	}

	status, expires, err := authcmd.GetExpirationStatus(at, time.Now())
	if err != nil && data.ErrLog != nil {
		data.ErrLog.Add(err)
	}
	if status != authcmd.StatusExpiringSoon {
		return
	}

	summary := authcmd.ExpirationSummary(status, expires, time.Now())
	remediation := authcmd.ExpirationRemediation(at.Type)
	text.Warning(data.Output, "Your active token %s. %s\n", summary, remediation)
}

// isAuthRelatedCommand reports whether commandName belongs to an auth-related
// command family. Matches both bare commands ("auth") and subcommands ("auth list").
func isAuthRelatedCommand(commandName string) bool {
	for _, prefix := range authRelatedCommands {
		if commandName == prefix || strings.HasPrefix(commandName, prefix+" ") {
			return true
		}
	}
	return false
}

// ssoAuthentication invokes the SSO runner to handle authentication.
func ssoAuthentication(outputMessage string, data *global.Data, forceReAuth bool) (token string, tokenSource lookup.Source, err error) {
	if data.SSORunner == nil {
		return "", lookup.SourceUndefined, fmt.Errorf("SSO runner is not configured")
	}

	skipPrompt := false
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
		skipPrompt = true
	}

	if err := data.SSORunner(data.Input, data.Output, forceReAuth, skipPrompt); err != nil {
		return token, tokenSource, fmt.Errorf("failed to authenticate: %w", err)
	}
	text.Break(data.Output)

	token, tokenSource = data.Token()
	if tokenSource == lookup.SourceUndefined {
		return token, tokenSource, fsterr.ErrNoToken()
	}
	return token, tokenSource, nil
}

func promptForAuth(data *global.Data) (string, lookup.Source, error) {
	text.Important(data.Output, "This command requires authentication to access your Fastly account.")
	text.Break(data.Output)
	if !env.AuthCommandDisabled() {
		text.Output(data.Output, "If you prefer SSO, run: fastly auth login --sso --token <name>\n")
	}
	text.Output(data.Output, "Otherwise, paste an API token now. It will be stored as your default auth token.\n")
	if env.AuthCommandDisabled() {
		text.Output(data.Output, "You can also set %s.\n", env.APIToken)
	} else {
		text.Output(data.Output, "You can also pass --token or set %s.\n", env.APIToken)
	}
	text.Output(data.Output, "An API token can be generated at: https://manage.fastly.com/account/personal/tokens\n")
	text.Output(data.Output, "Learn more: fastly.help/cli/cli-auth\n\n")

	token, err := text.InputSecure(data.Output, "Paste your API token: ", data.Input)
	if err != nil {
		return "", lookup.SourceUndefined, fmt.Errorf("error reading token input: %w", err)
	}
	if token == "" {
		return "", lookup.SourceUndefined, fsterr.ErrNoToken()
	}

	name, md, err := authcmd.StoreStaticToken(data, token)
	if err != nil {
		return "", lookup.SourceUndefined, err
	}

	text.Success(data.Output, "Authenticated as %s (token stored as %q)", md.Email, name)
	text.Info(data.Output, "Token saved to %s", data.ConfigPath)
	return token, lookup.SourceAuth, nil
}

func displayToken(tokenSource lookup.Source, data *global.Data) {
	switch tokenSource {
	case lookup.SourceFlag:
		fmt.Fprintf(data.Output, "Fastly API token provided via --token\n\n")
	case lookup.SourceEnvironment:
		fmt.Fprintf(data.Output, "Fastly API token provided via %s\n\n", env.APIToken)
	case lookup.SourceAuth:
		name := data.AuthTokenName()
		if name != "" {
			fmt.Fprintf(data.Output, "Fastly API token provided via config file (auth: %s)\n\n", name)
		} else {
			fmt.Fprintf(data.Output, "Fastly API token provided via config file (auth)\n\n")
		}
	case lookup.SourceUndefined, lookup.SourceDefault, lookup.SourceFile:
		fallthrough
	default:
		fmt.Fprintf(data.Output, "Fastly API token not provided\n\n")
	}
}

// If we are using the token from config file, check the file's permissions
// to assert if they are not too open or have been altered outside of the
// application and warn if so.
func checkConfigPermissions(tokenSource lookup.Source, out io.Writer) {
	if tokenSource == lookup.SourceAuth {
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
	case lookup.SourceDefault, lookup.SourceUndefined, lookup.SourceAuth:
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

// commandCollectsData determines if the command to be executed is one that
// collects data related to a Wasm binary.
func commandCollectsData(command string) bool {
	switch command {
	case "compute build", "compute hash-files", "compute publish", "compute serve":
		return true
	}
	return false
}

// commandRequiresAuthServer determines if the command to be executed is one that
// requires just the authentication server to be running.
func commandRequiresAuthServer(command string, args []string) bool {
	switch command {
	case "auth login":
		return slices.Contains(args, "--sso")
	case "profile create", "profile switch", "profile update", "sso":
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
	case "auth", "config", "profile", "sso", "update", "version":
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
	// Set a more meaningful error message when Fastly servers are unresponsive
	// check if the response code is a 500 or above
	if resp.StatusCode >= http.StatusInternalServerError {
		var body []byte
		body, _ = io.ReadAll(resp.Body) // default to empty string if we fail to read the body
		return nil, fmt.Errorf("the Fastly servers are unresponsive, please check the Fastly Status page (https://fastlystatus.com) and reach out to support if the error persists (HTTP Status Code: %d, Error Message: %s)", resp.StatusCode, body)
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

// commandSuppressesVerbose checks if the given command suppresses verbose output.
// This uses type assertion to check if the command has an embedded Base struct with SuppressVerbose set.
func commandSuppressesVerbose(command argparser.Command) bool {
	// Try to access the SuppressesVerbose method which is available on commands that embed argparser.Base
	type verboseSuppressor interface {
		SuppressesVerbose() bool
	}
	if vs, ok := command.(verboseSuppressor); ok {
		return vs.SuppressesVerbose()
	}

	return false
}
