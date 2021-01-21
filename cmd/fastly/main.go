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

// configEndpoint represents the API endpoint where we'll pull the dynamic
// configuration file from.
const configEndpoint = "http://integralist-cli-dynamic-config.com.global.prod.fastly.net/"

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

	// Extract a subset of configuration options from the local application directory.
	var file config.File
	err := file.Read(config.FilePath)
	if err != nil {
		text.Output(out, `
			We were unable to locate a local configuration file required to use the CLI.
			We will create that file for you now.
		`)
		text.Break(out)

		err := file.Load(configEndpoint, httpClient)
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

	// Validate if configuration is older than 24hrs
	if check.Stale(file.LastVersionCheck) {
		text.Warning(out, `
Your local application configuration is out-of-date.
We'll refresh this for you in the background and it'll be used next time.
		`)

		wait = true
		go func() {
			err := file.Load(configEndpoint, httpClient)
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
		os.Exit(1)
	}

	// If the command being run finishes before the latest config is written back
	// to disk, then wait for the write operation to complete.
	//
	// I use a variable instead of calling check.Stale() again, incase the file
	// object has indeed been updated already and is no longer considered stale!
	if wait {
		<-waitForWrite
		text.Info(out, "Successfully wrote updated application configuration file to disk.")
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
