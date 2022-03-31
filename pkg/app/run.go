package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/commands/version"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v6/fastly"
	"github.com/fastly/kingpin"
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
	ErrLog     fsterr.LogInterface
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
		File:       opts.ConfigFile,
		Path:       opts.ConfigPath,
		Env:        opts.Env,
		Output:     opts.Stdout,
		ErrLog:     opts.ErrLog,
		HTTPClient: opts.HTTPClient,
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
	//
	// NOTE: Global flags, unlike command flags, must be unique. For example, if
	// you try to use a letter that is already taken by any other command flag,
	// then kingpin will trigger a runtime panic 🎉
	//
	// NOTE: Short flags CAN be safely reused across commands.
	tokenHelp := fmt.Sprintf("Fastly API token (or via %s)", env.Token)
	app.Flag("token", tokenHelp).Short('t').StringVar(&globals.Flag.Token)
	app.Flag("profile", "Switch account profile").Short('o').StringVar(&globals.Flag.Profile)
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

	if data.File.Profile != "" {
		if name, p := profile.Get(data.File.Profile, globals.File.Profiles); name != "" {
			token = p.Token
		}
	}

	if globals.Flag.Profile != "" && command.Name() != "configure" {
		if exist := profile.Exist(globals.Flag.Profile, globals.File.Profiles); exist {
			// Persist the permanent switch of profiles.
			var ok bool
			if globals.File.Profiles, ok = profile.Set(globals.Flag.Profile, globals.File.Profiles); ok {
				if err := globals.File.Write(globals.Path); err != nil {
					return err
				}
			}
		} else {
			msg := profile.DoesNotExist

			name, p := profile.Default(globals.File.Profiles)
			if name == "" {
				msg = fmt.Sprintf("%s (no account profiles configured)", msg)
				return fsterr.RemediationError{
					Inner:       fmt.Errorf(msg),
					Remediation: fsterr.ProfileRemediation,
				}
			}

			msg = fmt.Sprintf("%s The default profile '%s' (%s) will be used.", msg, name, p.Email)
			text.Warning(opts.Stdout, msg)

			label := "\nWould you like to continue? [y/N] "
			cont, err := text.AskYesNo(opts.Stdout, label, opts.Stdin)
			if err != nil {
				return err
			}
			if !cont {
				return nil
			}
			token = p.Token
		}
	}

	// If we are using the token from config file, check the files permissions
	// to assert if they are not too open or have been altered outside of the
	// application and warn if so.
	if source == config.SourceFile && name != "configure" {
		if fi, err := os.Stat(config.FilePath); err == nil {
			if mode := fi.Mode().Perm(); mode > config.FilePermissions {
				text.Warning(opts.Stdout, "Unprotected configuration file.")
				fmt.Fprintf(opts.Stdout, "Permissions for '%s' are too open\n", config.FilePath)
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

	globals.APIClient, err = opts.APIClient(token, endpoint)
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
