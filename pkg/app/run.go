package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/commands/version"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
	"github.com/fastly/kingpin"
)

var (
	completionRegExp = regexp.MustCompile("completion-(?:script-)?(?:bash|zsh)$")
)

// Versioners represents all supported versioner types.
type Versioners struct {
	CLI     update.Versioner
	Viceroy update.Versioner
}

// RunOpts represent arguments to Run()
type RunOpts struct {
	APIClient  APIClientFactory
	Args       []string
	ConfigFile config.File
	ConfigPath string
	Env        config.Environment
	ErrLog     errors.LogInterface
	HTTPClient api.HTTPClient
	Stdin      io.Reader
	Stdout     io.Writer
	Versioners Versioners
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
	// The globals will hold generally-applicable configuration parameters
	// from a variety of sources, and is provided to each concrete command.
	globals := config.Data{
		File:   opts.ConfigFile,
		Env:    opts.Env,
		Output: opts.Stdout,
		ErrLog: opts.ErrLog,
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

	// As kingpin generates bash completion as a side-effect of kingpin.Parse we
	// allow it to call os.Exit, only if a completetion flag is present.
	if isCompletion(opts.Args) {
		app.Terminate(os.Exit)
	}

	// WARNING: kingping has no way of decorating flags as being "global"
	// therefore if you add/remove a global flag you will also need to update
	// the globalFlag map in pkg/app/usage.go which is used for usage rendering.
	tokenHelp := fmt.Sprintf("Fastly API token (or via %s)", env.Token)
	app.Flag("token", tokenHelp).Short('t').StringVar(&globals.Flag.Token)
	app.Flag("verbose", "Verbose logging").Short('v').BoolVar(&globals.Flag.Verbose)
	app.Flag("endpoint", "Fastly API endpoint").Hidden().StringVar(&globals.Flag.Endpoint)

	var data manifest.Data
	data.File.SetErrLog(opts.ErrLog)
	data.File.SetOutput(globals.Output)
	data.File.Read(manifest.Filename)

	commands := defineCommands(app, &globals, data, opts)
	command, name, err := processCommandInput(opts, app, &globals, commands)
	if err != nil {
		return err
	}
	// We add a special case for when cmd.ArgsIsHelpJSON() is true.
	if name == "help--format=json" || name == "help--formatjson" {
		return nil
	}

	token, source := globals.Token()
	if globals.Verbose() {
		switch source {
		case config.SourceFlag:
			fmt.Fprintf(opts.Stdout, "Fastly API token provided via --token\n")
		case config.SourceEnvironment:
			fmt.Fprintf(opts.Stdout, "Fastly API token provided via %s\n", env.Token)
		case config.SourceFile:
			fmt.Fprintf(opts.Stdout, "Fastly API token provided via config file\n")
		default:
			fmt.Fprintf(opts.Stdout, "Fastly API token not provided\n")
		}
	}

	// If we are using the token from config file, check the files permissions
	// to assert if they are not too open or have been altered outside of the
	// application and warn if so.
	if source == config.SourceFile && name != "configure" {
		if fi, err := os.Stat(config.FilePath); err == nil {
			if mode := fi.Mode().Perm(); mode > config.FilePermissions {
				text.Warning(opts.Stdout, "Unprotected configuration file.")
				fmt.Fprintf(opts.Stdout, "Permissions %04o for '%s' are too open\n", mode, config.FilePath)
				fmt.Fprintf(opts.Stdout, "It is recommended that your configuration file is NOT accessible by others.\n")
				fmt.Fprintln(opts.Stdout)
			}
		}
	}

	endpoint, source := globals.Endpoint()
	if globals.Verbose() {
		switch source {
		case config.SourceEnvironment:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint (via %s): %s\n", env.Endpoint, endpoint)
		case config.SourceFile:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint (via config file): %s\n", endpoint)
		default:
			fmt.Fprintf(opts.Stdout, "Fastly API endpoint: %s\n", endpoint)
		}
	}

	globals.Client, err = opts.APIClient(token, endpoint)
	if err != nil {
		globals.ErrLog.Add(err)
		return fmt.Errorf("error constructing Fastly API client: %w", err)
	}

	globals.RTSClient, err = fastly.NewRealtimeStatsClientForEndpoint(token, fastly.DefaultRealtimeStatsEndpoint)
	if err != nil {
		globals.ErrLog.Add(err)
		return fmt.Errorf("error constructing Fastly realtime stats client: %w", err)
	}

	if opts.Versioners.CLI != nil && name != "update" && !version.IsPreRelease(revision.AppVersion) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel() // push cancel on the defer stack first...
		f := update.CheckAsync(ctx, opts.ConfigFile, opts.ConfigPath, revision.AppVersion, opts.Versioners.CLI, opts.Stdin, opts.Stdout)
		defer f(opts.Stdout) // ...and the printing function second, so we hit the timeout
	}

	return command.Exec(opts.Stdin, opts.Stdout)
}

// APIClientFactory creates a Fastly API client (modeled as an api.Interface)
// from a user-provided API token. It exists as a type in order to parameterize
// the Run helper with it: in the real CLI, we can use NewClient from the Fastly
// API client library via RealClient; in tests, we can provide a mock API
// interface via MockClient.
type APIClientFactory func(token, endpoint string) (api.Interface, error)

// FastlyAPIClient is a ClientFactory that returns a real Fastly API client
// using the provided token and endpoint.
func FastlyAPIClient(token, endpoint string) (api.Interface, error) {
	client, err := fastly.NewClientForEndpoint(token, endpoint)
	return client, err
}

// isCompletion determines whether the supplied command arguments are for
// bash/zsh completion output.
func isCompletion(args []string) bool {
	var found bool
	for _, arg := range args {
		if completionRegExp.MatchString(arg) {
			found = true
		}
	}
	return found
}
