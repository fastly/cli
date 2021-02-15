package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/check"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/update"
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
		args                     = os.Args[1:]
		configFilePath           = config.FilePath // write-only for `fastly configure`
		clientFactory            = app.FastlyAPIClient
		httpClient               = http.DefaultClient
		versioner                = update.NewGitHub(context.Background())
		in             io.Reader = os.Stdin
		out            io.Writer = common.NewSyncWriter(os.Stdout)
	)

	// We have to manually handle the inclusion of the verbose flag here because
	// Kingpin doesn't evaluate the provided arguments until app.Run which
	// happens later in the file and yet we need to know if we should be printing
	// output related to the application configuration file in this file.
	var verboseOutput bool
	for _, seg := range args {
		if seg == "-v" || seg == "-verbose" || seg == "--verbose" {
			verboseOutput = true
		}
	}

	// Extract a subset of configuration options from the local application directory.
	var file config.File
	err := file.Read(config.FilePath)
	if err != nil {
		if verboseOutput {
			text.Output(out, `
				We were unable to locate a local configuration file required to use the CLI.
				We will create that file for you now.
			`)
			text.Break(out)
		}

		err := file.Load(config.RemoteEndpoint, httpClient)
		if err != nil {
			fmt.Println(err)
			os.Exit(1) // TODO: offer a clearer remediation step
		}
	}

	// When the local configuration file is stale we'll need to acquire the
	// latest version and write that back to disk. To ensure the CLI program
	// doesn't complete before the write has finished, we block via a channel.
	waitForWrite := make(chan bool)
	wait := false

	// Validate if configuration is older than its TTL
	if check.Stale(file.CLI.LastChecked, file.CLI.TTL) {
		if verboseOutput {
			text.Warning(out, `
Compatibility and versioning information for the Fastly CLI is being updated in the background.  The updated data will be used next time you execute a fastly command.
			`)
		}

		wait = true
		go func() {
			// NOTE: we no longer use the hardcoded config.RemoteEndpoint constant.
			// Instead we rely on the values inside of the application
			// configuration file to determine where to load the config from.
			err := file.Load(file.CLI.RemoteConfig, httpClient)
			if err != nil {
				fmt.Println(err)
				os.Exit(1) // TODO: offer a clearer remediation step
			}

			waitForWrite <- true
		}()
	}

	// Main is basically just a shim to call Run, so we do that here.
	if err := app.Run(args, env, file, configFilePath, clientFactory, httpClient, versioner, in, out); err != nil {
		errors.Deduce(err).Print(os.Stderr)

		// NOTE: if we have an error processing the command, then we should be sure
		// to wait for the async file write to complete (otherwise we'll end up in
		// a situation where there is a local application configuration file but
		// with incomplete contents).
		//
		// It would have been nice to just do something like...
		//
		// if wait {
		//   defer func(){
		//     <-waitForWrite
		//     text.Info(out, configUpdateSuccessful)
		//   }()
		// }
		//
		// ...and to have this a bit further up the script, as it would have meant
		// I could avoid duplicating the following if statement in two places.
		//
		// As it is, I have to wait for the async write operation here and also at
		// the end of the main function.
		//
		// The problem with defer is that it doesn't work when os.Exit() is
		// encountered, so you either use something like runtime.Goexit() which is
		// pretty hairy and introduces other changes like `defer os.Exit(0)` at the
		// top of the main() function OR we re-architecture the call flow which
		// isn't ideal either.
		//
		// So I've opted for duplication.
		//
		if wait {
			<-waitForWrite
			if verboseOutput {
				text.Info(out, config.UpdateSuccessful)
			}
		}

		os.Exit(1)
	}

	// If the command being run finishes before the latest config is written back
	// to disk, then wait for the write operation to complete.
	//
	// I use a variable instead of calling check.Stale() again, incase the file
	// object has indeed been updated already and is no longer considered stale!
	if wait {
		<-waitForWrite
		if verboseOutput {
			text.Info(out, config.UpdateSuccessful)
		}
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
