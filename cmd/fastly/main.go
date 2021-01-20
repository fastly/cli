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
	"github.com/fastly/cli/pkg/update"
)

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

	// Some configuration options can come from a config file.
	var file config.File
	err := file.Read(config.FilePath) // ignore error
	if err != nil {
		// Acquire global CLI configuration file
		err := file.Load(configEndpoint, httpClient)
		if err != nil {
			fmt.Println(err)
			// TODO: offer remediation step
			os.Exit(1)
		}
	}

	doneWithUpdate := make(chan bool)
	stale := false

	// Validate if configuration is older than 24hrs
	if check.Stale(file.LastVersionCheck) {
		fmt.Println("STALE!")
		stale = true
		go func() {
			file.Load(configEndpoint, httpClient)
			doneWithUpdate <- true
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
	if stale {
		fmt.Println("Writing updated configuration to disk.") // TODO: only print for verbose logging
		<-doneWithUpdate
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
