// Package main is the entry point for the Fastly CLI.
package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/sync"
	"github.com/fastly/cli/pkg/text"
)

func main() {
	// Parse the arguments provided by the user via the command-line interface.
	args := os.Args[1:]

	// Define a HTTP client that will be used for making arbitrary HTTP requests.
	httpClient := &http.Client{Timeout: time.Minute * 2}

	// Define the standard input/output streams.
	var (
		in  io.Reader = os.Stdin
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
	// The CLI relies on a valid configuration, otherwise we can't continue.
	err := cfg.Read(config.FilePath, in, out, fsterr.Log, verboseOutput)
	if err != nil {
		fsterr.Deduce(err).Print(color.Error)
		// WARNING: os.Exit will exit, and any `defer` calls will not be run.
		os.Exit(1)
	}

	// Extract user's project configuration from the fastly.toml manifest.
	var md manifest.Data
	md.File.Args = args
	md.File.SetErrLog(fsterr.Log)
	md.File.SetOutput(out)

	// NOTE: We skip handling the error because not all commands relate to Compute.
	_ = md.File.Read(manifest.Filename)

	// The `main` function is a shim for calling `app.Run()`.
	err = app.Run(app.RunOpts{
		APIClient: func(token, endpoint string, debugMode bool) (api.Interface, error) {
			client, err := fastly.NewClientForEndpoint(token, endpoint)
			if debugMode {
				client.DebugMode = true
			}
			return client, err
		},
		Args:             args,
		ConfigFile:       cfg,
		ConfigPath:       config.FilePath,
		Env:              e,
		ErrLog:           fsterr.Log,
		ExecuteWasmTools: compute.ExecuteWasmTools,
		HTTPClient:       httpClient,
		Manifest:         &md,
		Opener:           open.Run,
		Stdin:            in,
		Stdout:           out,
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
			WasmTools: github.New(github.Opts{
				HTTPClient: httpClient,
				Org:        "bytecodealliance",
				Repo:       "wasm-tools",
				Binary:     "wasm-tools",
				External:   true,
				Nested:     true,
			}),
		},
	})

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
