// Package main is the entry point for the Fastly CLI.
package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/sync"
	"github.com/fatih/color"
	"github.com/getsentry/sentry-go"
)

const sentryTimeout = 2 * time.Second

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Debug:       false,
		Dsn:         "https://6e390df3d7924f7bbe521299bbd4f8dc@o1025883.ingest.sentry.io/6423223",
		Environment: revision.Environment,
		Release:     revision.AppVersion,
		IgnoreErrors: []string{
			`required flag \-\-[^\s]+ not provided`,
			`error reading service: no service ID found`,
			`error matching service name with available services`,
			`open fastly.toml: no such file or directory`,
		},
		BeforeSend: func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
			for i, e := range event.Exception {
				event.Exception[i].Value = fsterr.FilterToken(e.Value)
			}
			return event
		},
		BeforeBreadcrumb: func(breadcrumb *sentry.Breadcrumb, _ *sentry.BreadcrumbHint) *sentry.Breadcrumb {
			breadcrumb.Message = fsterr.FilterToken(breadcrumb.Message)
			return breadcrumb
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sentry.Flush(sentryTimeout)

	// Some configuration options can come from env vars.
	var env config.Environment
	env.Read(parseEnv(os.Environ()))

	// All of the work of building the set of commands and subcommands, wiring
	// them together, picking which one to call, and executing it, occurs in a
	// helper function, Run. We parameterize all of the dependencies so we can
	// test it more easily. Here, we declare all of the dependencies, using
	// the "real" versions that pull e.g. actual commandline arguments, the
	// user's real environment, etc.
	var (
		args                    = os.Args[1:]
		clientFactory           = app.FastlyAPIClient
		httpClient              = &http.Client{Timeout: time.Second * 5}
		in            io.Reader = os.Stdin
		out           io.Writer = sync.NewWriter(color.Output)
		versionerCLI            = update.NewGitHub(update.GitHubOpts{
			Org:    "fastly",
			Repo:   "cli",
			Binary: "fastly",
		})
		versionerViceroy = update.NewGitHub(update.GitHubOpts{
			Org:    "fastly",
			Repo:   "viceroy",
			Binary: "viceroy",
		})
	)

	// We have to manually handle the inclusion of the verbose flag here because
	// Kingpin doesn't evaluate the provided arguments until app.Run which
	// happens later in the file and yet we need to know if we should be printing
	// output related to the application configuration file in this file.
	var verboseOutput bool
	for _, seg := range args {
		if seg == "-v" || seg == "--verbose" {
			verboseOutput = true
		}
	}

	// Similarly for the --auto-yes/--non-interactive flags, we need access to
	// these for handling interactive error prompts to the user, in case the CLI
	// is being run in a CI environment.
	var autoYes, nonInteractive bool
	for _, seg := range args {
		if seg == "-y" || seg == "--auto-yes" {
			autoYes = true
		}
		if seg == "-i" || seg == "--non-interactive" {
			nonInteractive = true
		}
	}

	// Extract a subset of configuration options from the local application directory.
	var file config.File
	file.SetAutoYes(autoYes)
	file.SetNonInteractive(nonInteractive)

	// The CLI relies on a valid configuration, otherwise we can't continue.
	err = file.Read(config.FilePath, in, out, fsterr.Log, verboseOutput)
	if err != nil {
		fsterr.Deduce(err).Print(color.Error)
		os.Exit(1)
	}

	// Main is basically just a shim to call Run, so we do that here.
	opts := app.RunOpts{
		APIClient:  clientFactory,
		Args:       args,
		ConfigFile: file,
		ConfigPath: config.FilePath,
		Env:        env,
		ErrLog:     fsterr.Log,
		HTTPClient: httpClient,
		Stdin:      in,
		Stdout:     out,
		Versioners: app.Versioners{
			CLI:     versionerCLI,
			Viceroy: versionerViceroy,
		},
	}
	err = app.Run(opts)

	// NOTE: We persist any error log entries to disk before attempting to handle
	// a possible error response from app.Run as there could be errors recorded
	// during the execution flow but were otherwise handled without bubbling an
	// error back the call stack, and so if the user still experiences something
	// unexpected we will have a record of any errors that happened along the way.
	logErr := fsterr.Log.Persist(fsterr.LogPath, args)
	if logErr != nil {
		fsterr.Deduce(logErr).Print(color.Error)
	}

	if err != nil {
		// NOTE: os.Exit doesn't honour any deferred calls so we have to manually
		// flush the Sentry buffer here (as well as the deferred call at the top of
		// the main function).
		sentry.Flush(sentryTimeout)

		fsterr.Deduce(err).Print(color.Error)

		exitError := fsterr.SkipExitError{}
		if errors.As(err, &exitError) {
			if exitError.Skip {
				return // skip returning an error for 'help' output
			}
		}

		os.Exit(1)
	}
}

func parseEnv(environ []string) map[string]string {
	env := map[string]string{}
	for _, kv := range environ {
		toks := strings.SplitN(kv, "=", 2)
		if len(toks) != 2 {
			continue
		}
		k, v := toks[0], toks[1]
		env[k] = v
	}
	return env
}
