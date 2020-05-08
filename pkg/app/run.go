package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/backend"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/configure"
	"github.com/fastly/cli/pkg/domain"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/healthcheck"
	"github.com/fastly/cli/pkg/logging"
	"github.com/fastly/cli/pkg/logging/bigquery"
	"github.com/fastly/cli/pkg/logging/ftp"
	"github.com/fastly/cli/pkg/logging/gcs"
	"github.com/fastly/cli/pkg/logging/logentries"
	"github.com/fastly/cli/pkg/logging/papertrail"
	"github.com/fastly/cli/pkg/logging/s3"
	"github.com/fastly/cli/pkg/logging/sumologic"
	"github.com/fastly/cli/pkg/logging/syslog"
	"github.com/fastly/cli/pkg/service"
	"github.com/fastly/cli/pkg/serviceversion"
	"github.com/fastly/cli/pkg/stats"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/cli/pkg/version"
	"github.com/fastly/cli/pkg/whoami"
	"github.com/fastly/go-fastly/fastly"
	"gopkg.in/alecthomas/kingpin.v3-unstable"
)

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
func Run(args []string, env config.Environment, file config.File, configFilePath string, cf APIClientFactory, httpClient api.HTTPClient, versioner update.Versioner, in io.Reader, out io.Writer) error {
	// The globals will hold generally-applicable configuration parameters
	// from a variety of sources, and is provided to each concrete command.
	globals := config.Data{
		File: file,
		Env:  env,
	}

	// Set up the main application root, including global flags, and then each
	// of the subcommands. Note that we deliberately don't use some of the more
	// advanced features of the kingpin.Application flags, like env var
	// bindings, because we need to do things like track where a config
	// parameter came from.
	app := kingpin.New("fastly", "A tool to interact with the Fastly API")
	app.Terminate(nil)               // don't let kingpin call os.Exit
	app.Writers(out, ioutil.Discard) // don't let kingpin write error output
	app.UsageContext(&kingpin.UsageContext{
		Template: VerboseUsageTemplate,
		Funcs:    UsageTemplateFuncs,
	})

	// WARNING: kingping has no way of decorating flags as being "global"
	// therefore if you add/remove a global flag you will also need to update
	// the globalFlag map in pkg/app/usage.go which is used for usage rendering.
	tokenHelp := fmt.Sprintf("Fastly API token (or via %s)", config.EnvVarToken)
	app.Flag("token", tokenHelp).Short('t').StringVar(&globals.Flag.Token)
	app.Flag("verbose", "Verbose logging").Short('v').BoolVar(&globals.Flag.Verbose)
	app.Flag("endpoint", "Fastly API endpoint").Hidden().StringVar(&globals.Flag.Endpoint)

	configureRoot := configure.NewRootCommand(app, configFilePath, configure.APIClientFactory(cf), &globals)
	whoamiRoot := whoami.NewRootCommand(app, httpClient, &globals)
	versionRoot := version.NewRootCommand(app)
	updateRoot := update.NewRootCommand(app, versioner, httpClient)

	serviceRoot := service.NewRootCommand(app, &globals)
	serviceCreate := service.NewCreateCommand(serviceRoot.CmdClause, &globals)
	serviceList := service.NewListCommand(serviceRoot.CmdClause, &globals)
	serviceDescribe := service.NewDescribeCommand(serviceRoot.CmdClause, &globals)
	serviceUpdate := service.NewUpdateCommand(serviceRoot.CmdClause, &globals)
	serviceDelete := service.NewDeleteCommand(serviceRoot.CmdClause, &globals)

	serviceVersionRoot := serviceversion.NewRootCommand(app, &globals)
	serviceVersionClone := serviceversion.NewCloneCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionList := serviceversion.NewListCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionUpdate := serviceversion.NewUpdateCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionActivate := serviceversion.NewActivateCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionDeactivate := serviceversion.NewDeactivateCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionLock := serviceversion.NewLockCommand(serviceVersionRoot.CmdClause, &globals)

	computeRoot := compute.NewRootCommand(app, &globals)
	computeInit := compute.NewInitCommand(computeRoot.CmdClause, &globals)
	computeBuild := compute.NewBuildCommand(computeRoot.CmdClause, httpClient, &globals)
	computeDeploy := compute.NewDeployCommand(computeRoot.CmdClause, httpClient, &globals)
	computeUpdate := compute.NewUpdateCommand(computeRoot.CmdClause, httpClient, &globals)
	computeValidate := compute.NewValidateCommand(computeRoot.CmdClause, &globals)

	domainRoot := domain.NewRootCommand(app, &globals)
	domainCreate := domain.NewCreateCommand(domainRoot.CmdClause, &globals)
	domainList := domain.NewListCommand(domainRoot.CmdClause, &globals)
	domainDescribe := domain.NewDescribeCommand(domainRoot.CmdClause, &globals)
	domainUpdate := domain.NewUpdateCommand(domainRoot.CmdClause, &globals)
	domainDelete := domain.NewDeleteCommand(domainRoot.CmdClause, &globals)

	backendRoot := backend.NewRootCommand(app, &globals)
	backendCreate := backend.NewCreateCommand(backendRoot.CmdClause, &globals)
	backendList := backend.NewListCommand(backendRoot.CmdClause, &globals)
	backendDescribe := backend.NewDescribeCommand(backendRoot.CmdClause, &globals)
	backendUpdate := backend.NewUpdateCommand(backendRoot.CmdClause, &globals)
	backendDelete := backend.NewDeleteCommand(backendRoot.CmdClause, &globals)

	healthcheckRoot := healthcheck.NewRootCommand(app, &globals)
	healthcheckCreate := healthcheck.NewCreateCommand(healthcheckRoot.CmdClause, &globals)
	healthcheckList := healthcheck.NewListCommand(healthcheckRoot.CmdClause, &globals)
	healthcheckDescribe := healthcheck.NewDescribeCommand(healthcheckRoot.CmdClause, &globals)
	healthcheckUpdate := healthcheck.NewUpdateCommand(healthcheckRoot.CmdClause, &globals)
	healthcheckDelete := healthcheck.NewDeleteCommand(healthcheckRoot.CmdClause, &globals)

	loggingRoot := logging.NewRootCommand(app, &globals)

	bigQueryRoot := bigquery.NewRootCommand(loggingRoot.CmdClause, &globals)
	bigQueryCreate := bigquery.NewCreateCommand(bigQueryRoot.CmdClause, &globals)
	bigQueryList := bigquery.NewListCommand(bigQueryRoot.CmdClause, &globals)
	bigQueryDescribe := bigquery.NewDescribeCommand(bigQueryRoot.CmdClause, &globals)
	bigQueryUpdate := bigquery.NewUpdateCommand(bigQueryRoot.CmdClause, &globals)
	bigQueryDelete := bigquery.NewDeleteCommand(bigQueryRoot.CmdClause, &globals)

	s3Root := s3.NewRootCommand(loggingRoot.CmdClause, &globals)
	s3Create := s3.NewCreateCommand(s3Root.CmdClause, &globals)
	s3List := s3.NewListCommand(s3Root.CmdClause, &globals)
	s3Describe := s3.NewDescribeCommand(s3Root.CmdClause, &globals)
	s3Update := s3.NewUpdateCommand(s3Root.CmdClause, &globals)
	s3Delete := s3.NewDeleteCommand(s3Root.CmdClause, &globals)

	syslogRoot := syslog.NewRootCommand(loggingRoot.CmdClause, &globals)
	syslogCreate := syslog.NewCreateCommand(syslogRoot.CmdClause, &globals)
	syslogList := syslog.NewListCommand(syslogRoot.CmdClause, &globals)
	syslogDescribe := syslog.NewDescribeCommand(syslogRoot.CmdClause, &globals)
	syslogUpdate := syslog.NewUpdateCommand(syslogRoot.CmdClause, &globals)
	syslogDelete := syslog.NewDeleteCommand(syslogRoot.CmdClause, &globals)

	logentriesRoot := logentries.NewRootCommand(loggingRoot.CmdClause, &globals)
	logentriesCreate := logentries.NewCreateCommand(logentriesRoot.CmdClause, &globals)
	logentriesList := logentries.NewListCommand(logentriesRoot.CmdClause, &globals)
	logentriesDescribe := logentries.NewDescribeCommand(logentriesRoot.CmdClause, &globals)
	logentriesUpdate := logentries.NewUpdateCommand(logentriesRoot.CmdClause, &globals)
	logentriesDelete := logentries.NewDeleteCommand(logentriesRoot.CmdClause, &globals)

	papertrailRoot := papertrail.NewRootCommand(loggingRoot.CmdClause, &globals)
	papertrailCreate := papertrail.NewCreateCommand(papertrailRoot.CmdClause, &globals)
	papertrailList := papertrail.NewListCommand(papertrailRoot.CmdClause, &globals)
	papertrailDescribe := papertrail.NewDescribeCommand(papertrailRoot.CmdClause, &globals)
	papertrailUpdate := papertrail.NewUpdateCommand(papertrailRoot.CmdClause, &globals)
	papertrailDelete := papertrail.NewDeleteCommand(papertrailRoot.CmdClause, &globals)

	sumologicRoot := sumologic.NewRootCommand(loggingRoot.CmdClause, &globals)
	sumologicCreate := sumologic.NewCreateCommand(sumologicRoot.CmdClause, &globals)
	sumologicList := sumologic.NewListCommand(sumologicRoot.CmdClause, &globals)
	sumologicDescribe := sumologic.NewDescribeCommand(sumologicRoot.CmdClause, &globals)
	sumologicUpdate := sumologic.NewUpdateCommand(sumologicRoot.CmdClause, &globals)
	sumologicDelete := sumologic.NewDeleteCommand(sumologicRoot.CmdClause, &globals)

	gcsRoot := gcs.NewRootCommand(loggingRoot.CmdClause, &globals)
	gcsCreate := gcs.NewCreateCommand(gcsRoot.CmdClause, &globals)
	gcsList := gcs.NewListCommand(gcsRoot.CmdClause, &globals)
	gcsDescribe := gcs.NewDescribeCommand(gcsRoot.CmdClause, &globals)
	gcsUpdate := gcs.NewUpdateCommand(gcsRoot.CmdClause, &globals)
	gcsDelete := gcs.NewDeleteCommand(gcsRoot.CmdClause, &globals)

	ftpRoot := ftp.NewRootCommand(loggingRoot.CmdClause, &globals)
	ftpCreate := ftp.NewCreateCommand(ftpRoot.CmdClause, &globals)
	ftpList := ftp.NewListCommand(ftpRoot.CmdClause, &globals)
	ftpDescribe := ftp.NewDescribeCommand(ftpRoot.CmdClause, &globals)
	ftpUpdate := ftp.NewUpdateCommand(ftpRoot.CmdClause, &globals)
	ftpDelete := ftp.NewDeleteCommand(ftpRoot.CmdClause, &globals)

	statsRoot := stats.NewRootCommand(app, &globals)
	statsRegions := stats.NewRegionsCommand(statsRoot.CmdClause, &globals)
	statsHistorical := stats.NewHistoricalCommand(statsRoot.CmdClause, &globals)
	statsRealtime := stats.NewRealtimeCommand(statsRoot.CmdClause, &globals)

	commands := []common.Command{
		configureRoot,
		whoamiRoot,
		versionRoot,
		updateRoot,

		serviceRoot,
		serviceCreate,
		serviceList,
		serviceDescribe,
		serviceUpdate,
		serviceDelete,

		serviceVersionRoot,
		serviceVersionClone,
		serviceVersionList,
		serviceVersionUpdate,
		serviceVersionActivate,
		serviceVersionDeactivate,
		serviceVersionLock,

		computeRoot,
		computeInit,
		computeBuild,
		computeDeploy,
		computeUpdate,
		computeValidate,

		domainRoot,
		domainCreate,
		domainList,
		domainDescribe,
		domainUpdate,
		domainDelete,

		backendRoot,
		backendCreate,
		backendList,
		backendDescribe,
		backendUpdate,
		backendDelete,

		healthcheckRoot,
		healthcheckCreate,
		healthcheckList,
		healthcheckDescribe,
		healthcheckUpdate,
		healthcheckDelete,

		loggingRoot,

		bigQueryRoot,
		bigQueryCreate,
		bigQueryList,
		bigQueryDescribe,
		bigQueryUpdate,
		bigQueryDelete,

		s3Root,
		s3Create,
		s3List,
		s3Describe,
		s3Update,
		s3Delete,

		syslogRoot,
		syslogCreate,
		syslogList,
		syslogDescribe,
		syslogUpdate,
		syslogDelete,

		logentriesRoot,
		logentriesCreate,
		logentriesList,
		logentriesDescribe,
		logentriesUpdate,
		logentriesDelete,

		papertrailRoot,
		papertrailCreate,
		papertrailList,
		papertrailDescribe,
		papertrailUpdate,
		papertrailDelete,

		sumologicRoot,
		sumologicCreate,
		sumologicList,
		sumologicDescribe,
		sumologicUpdate,
		sumologicDelete,

		gcsRoot,
		gcsCreate,
		gcsList,
		gcsDescribe,
		gcsUpdate,
		gcsDelete,

		ftpRoot,
		ftpCreate,
		ftpList,
		ftpDescribe,
		ftpUpdate,
		ftpDelete,

		statsRoot,
		statsRegions,
		statsHistorical,
		statsRealtime,
	}

	// Handle parse errors and display contextal usage if possible. Due to bugs
	// and an obession for lots of output side-effects in the kingpin.Parse
	// logic, we suppress it from writing any usage or errors to the writer by
	// swapping the writer with a no-op and then restoring the real writer
	// afterwards. This ensures usage text is only written once to the writer
	// and gives us greater control over our error formatting.
	app.Writers(ioutil.Discard, ioutil.Discard)
	name, err := app.Parse(args)
	if err != nil {
		usage := Usage(args, app, out, ioutil.Discard)
		return errors.RemediationError{Prefix: usage, Inner: fmt.Errorf("error parsing arguments: %w", err)}
	}
	if ctx, _ := app.ParseContext(args); contextHasHelpFlag(ctx) {
		usage := Usage(args, app, out, ioutil.Discard)
		return errors.RemediationError{Prefix: usage}
	}
	app.Writers(out, ioutil.Discard)

	// A side-effect of suppressing app.Parse from writing output is the usage
	// isn't printed for the default `help` command. Therefore we capture it
	// here by calling Parse, again swapping the Writers. This also ensures the
	// larger and more verbose help formatting is used.
	if name == "help" {
		var buf bytes.Buffer
		app.Writers(&buf, ioutil.Discard)
		app.Parse(args)
		app.Writers(out, ioutil.Discard)

		// The full-fat output of `fastly help` should have a hint at the bottom
		// for more specific help. Unfortunately I don't know of a better way to
		// distinguish `fastly help` from e.g. `fastly help configure` than this
		// check.
		if len(args) > 0 && args[len(args)-1] == "help" {
			fmt.Fprintln(&buf, "For help on a specific command, try e.g.")
			fmt.Fprintln(&buf, "")
			fmt.Fprintln(&buf, "\tfastly help configure")
			fmt.Fprintln(&buf, "\tfastly configure --help")
			fmt.Fprintln(&buf, "")
		}

		return errors.RemediationError{Prefix: buf.String()}
	}

	token, source := globals.Token()
	if globals.Verbose() {
		switch source {
		case config.SourceFlag:
			fmt.Fprintf(out, "Fastly API token provided via --token\n")
		case config.SourceEnvironment:
			fmt.Fprintf(out, "Fastly API token provided via %s\n", config.EnvVarToken)
		case config.SourceFile:
			fmt.Fprintf(out, "Fastly API token provided via config file\n")
		default:
			fmt.Fprintf(out, "Fastly API token not provided\n")
		}
	}

	// If we are using the token from config file, check the files permissions
	// to assert if they are not too open or have been altered outside of the
	// application and warn if so.
	if source == config.SourceFile && name != "configure" {
		if fi, err := os.Stat(config.FilePath); err == nil {
			if mode := fi.Mode().Perm(); mode > config.FilePermissions {
				text.Warning(out, "Unprotected configuration file.")
				fmt.Fprintf(out, "Permissions %04o for '%s' are too open\n", mode, config.FilePath)
				fmt.Fprintf(out, "It is recommended that your configuration file is NOT accessible by others.\n")
				fmt.Fprintln(out)
			}
		}
	}

	endpoint, source := globals.Endpoint()
	if globals.Verbose() {
		switch source {
		case config.SourceEnvironment:
			fmt.Fprintf(out, "Fastly API endpoint (via %s): %s\n", config.EnvVarEndpoint, endpoint)
		case config.SourceFile:
			fmt.Fprintf(out, "Fastly API endpoint (via config file): %s\n", endpoint)
		default:
			fmt.Fprintf(out, "Fastly API endpoint: %s\n", endpoint)
		}
	}

	globals.Client, err = cf(token, endpoint)
	if err != nil {
		return fmt.Errorf("error constructing Fastly API client: %w", err)
	}

	globals.RTSClient, err = fastly.NewRealtimeStatsClientForEndpoint(token, fastly.DefaultRealtimeStatsEndpoint)
	if err != nil {
		return fmt.Errorf("error constructing Fastly realtime stats client: %w", err)
	}

	command, found := common.SelectCommand(name, commands)
	if !found {
		usage := Usage(args, app, out, ioutil.Discard)
		return errors.RemediationError{Prefix: usage, Inner: fmt.Errorf("command not found")}
	}

	if versioner != nil && name != "update" && version.AppVersion != version.None {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel() // push cancel on the defer stack first...
		f := update.CheckAsync(ctx, file, configFilePath, version.AppVersion, versioner)
		defer f(out) // ...and the printing function second, so we hit the timeout
	}

	return command.Exec(in, out)
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

// contextHasHelpFlag asserts whether a given kingpin.ParseContext contains a
// `help` flag.
func contextHasHelpFlag(ctx *kingpin.ParseContext) bool {
	_, ok := ctx.Elements.FlagMap()["help"]
	return ok
}
