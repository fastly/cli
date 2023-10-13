// Package main is the entry point for the Fastly CLI.
package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/sync"
	"github.com/fastly/cli/pkg/text"
)

func main() {
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
		httpClient              = &http.Client{Timeout: time.Minute * 2}
		in            io.Reader = os.Stdin
		out           io.Writer = sync.NewWriter(color.Output)
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
	var cfg config.File
	cfg.SetAutoYes(autoYes)
	cfg.SetNonInteractive(nonInteractive)

	// The CLI relies on a valid configuration, otherwise we can't continue.
	err := cfg.Read(config.FilePath, in, out, fsterr.Log, verboseOutput)
	if err != nil {
		fsterr.Deduce(err).Print(color.Error)

		// WARNING: os.Exit will exit, and any `defer` calls will not be run.
		os.Exit(1)
	}

	var md manifest.Data
	md.File.Args = args
	md.File.SetErrLog(fsterr.Log)
	md.File.SetOutput(out)

	// NOTE: We skip handling the error because not all commands relate to Compute.
	_ = md.File.Read(manifest.Filename)

	// Main is basically just a shim to call Run, so we do that here.
	opts := app.RunOpts{
		APIClient:  clientFactory,
		Args:       args,
		ConfigFile: cfg,
		ConfigPath: config.FilePath,
		Env:        env,
		ErrLog:     fsterr.Log,
		HTTPClient: httpClient,
		Manifest:   &md,
		Stdin:      in,
		Stdout:     out,
		Versioners: app.Versioners{
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
		text.Break(out)
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
