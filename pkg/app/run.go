package app

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/acl"
	"github.com/fastly/cli/pkg/commands/aclentry"
	"github.com/fastly/cli/pkg/commands/authtoken"
	"github.com/fastly/cli/pkg/commands/backend"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/commands/configure"
	"github.com/fastly/cli/pkg/commands/domain"
	"github.com/fastly/cli/pkg/commands/edgedictionary"
	"github.com/fastly/cli/pkg/commands/edgedictionaryitem"
	"github.com/fastly/cli/pkg/commands/healthcheck"
	"github.com/fastly/cli/pkg/commands/ip"
	"github.com/fastly/cli/pkg/commands/logging"
	"github.com/fastly/cli/pkg/commands/logging/azureblob"
	"github.com/fastly/cli/pkg/commands/logging/bigquery"
	"github.com/fastly/cli/pkg/commands/logging/cloudfiles"
	"github.com/fastly/cli/pkg/commands/logging/datadog"
	"github.com/fastly/cli/pkg/commands/logging/digitalocean"
	"github.com/fastly/cli/pkg/commands/logging/elasticsearch"
	"github.com/fastly/cli/pkg/commands/logging/ftp"
	"github.com/fastly/cli/pkg/commands/logging/gcs"
	"github.com/fastly/cli/pkg/commands/logging/googlepubsub"
	"github.com/fastly/cli/pkg/commands/logging/heroku"
	"github.com/fastly/cli/pkg/commands/logging/honeycomb"
	"github.com/fastly/cli/pkg/commands/logging/https"
	"github.com/fastly/cli/pkg/commands/logging/kafka"
	"github.com/fastly/cli/pkg/commands/logging/kinesis"
	"github.com/fastly/cli/pkg/commands/logging/logentries"
	"github.com/fastly/cli/pkg/commands/logging/loggly"
	"github.com/fastly/cli/pkg/commands/logging/logshuttle"
	"github.com/fastly/cli/pkg/commands/logging/newrelic"
	"github.com/fastly/cli/pkg/commands/logging/openstack"
	"github.com/fastly/cli/pkg/commands/logging/papertrail"
	"github.com/fastly/cli/pkg/commands/logging/s3"
	"github.com/fastly/cli/pkg/commands/logging/scalyr"
	"github.com/fastly/cli/pkg/commands/logging/sftp"
	"github.com/fastly/cli/pkg/commands/logging/splunk"
	"github.com/fastly/cli/pkg/commands/logging/sumologic"
	"github.com/fastly/cli/pkg/commands/logging/syslog"
	"github.com/fastly/cli/pkg/commands/logs"
	"github.com/fastly/cli/pkg/commands/pop"
	"github.com/fastly/cli/pkg/commands/purge"
	"github.com/fastly/cli/pkg/commands/service"
	"github.com/fastly/cli/pkg/commands/serviceversion"
	"github.com/fastly/cli/pkg/commands/stats"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/commands/user"
	"github.com/fastly/cli/pkg/commands/vcl"
	"github.com/fastly/cli/pkg/commands/vcl/custom"
	"github.com/fastly/cli/pkg/commands/vcl/snippet"
	"github.com/fastly/cli/pkg/commands/version"
	"github.com/fastly/cli/pkg/commands/whoami"
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
	data.File.SetOutput(globals.Output)
	data.File.Read(manifest.Filename)

	aclCmdRoot := acl.NewRootCommand(app, &globals)
	aclCreate := acl.NewCreateCommand(aclCmdRoot.CmdClause, &globals, data)
	aclDelete := acl.NewDeleteCommand(aclCmdRoot.CmdClause, &globals, data)
	aclDescribe := acl.NewDescribeCommand(aclCmdRoot.CmdClause, &globals, data)
	aclList := acl.NewListCommand(aclCmdRoot.CmdClause, &globals, data)
	aclUpdate := acl.NewUpdateCommand(aclCmdRoot.CmdClause, &globals, data)
	aclEntryCmdRoot := aclentry.NewRootCommand(app, &globals)
	aclEntryCreate := aclentry.NewCreateCommand(aclEntryCmdRoot.CmdClause, &globals, data)
	aclEntryDelete := aclentry.NewDeleteCommand(aclEntryCmdRoot.CmdClause, &globals, data)
	aclEntryDescribe := aclentry.NewDescribeCommand(aclEntryCmdRoot.CmdClause, &globals, data)
	aclEntryList := aclentry.NewListCommand(aclEntryCmdRoot.CmdClause, &globals, data)
	aclEntryUpdate := aclentry.NewUpdateCommand(aclEntryCmdRoot.CmdClause, &globals, data)
	authtokenCmdRoot := authtoken.NewRootCommand(app, &globals)
	authtokenCreate := authtoken.NewCreateCommand(authtokenCmdRoot.CmdClause, &globals, data)
	authtokenDelete := authtoken.NewDeleteCommand(authtokenCmdRoot.CmdClause, &globals, data)
	authtokenDescribe := authtoken.NewDescribeCommand(authtokenCmdRoot.CmdClause, &globals, data)
	authtokenList := authtoken.NewListCommand(authtokenCmdRoot.CmdClause, &globals, data)
	backendCmdRoot := backend.NewRootCommand(app, &globals)
	backendCreate := backend.NewCreateCommand(backendCmdRoot.CmdClause, &globals, data)
	backendDelete := backend.NewDeleteCommand(backendCmdRoot.CmdClause, &globals, data)
	backendDescribe := backend.NewDescribeCommand(backendCmdRoot.CmdClause, &globals, data)
	backendList := backend.NewListCommand(backendCmdRoot.CmdClause, &globals, data)
	backendUpdate := backend.NewUpdateCommand(backendCmdRoot.CmdClause, &globals, data)
	computeCmdRoot := compute.NewRootCommand(app, &globals)
	computeBuild := compute.NewBuildCommand(computeCmdRoot.CmdClause, opts.HTTPClient, &globals)
	computeDeploy := compute.NewDeployCommand(computeCmdRoot.CmdClause, opts.HTTPClient, &globals, data)
	computeInit := compute.NewInitCommand(computeCmdRoot.CmdClause, opts.HTTPClient, &globals, data)
	computePack := compute.NewPackCommand(computeCmdRoot.CmdClause, &globals, data)
	computePublish := compute.NewPublishCommand(computeCmdRoot.CmdClause, &globals, computeBuild, computeDeploy, data)
	computeServe := compute.NewServeCommand(computeCmdRoot.CmdClause, &globals, computeBuild, opts.Versioners.Viceroy, data)
	computeUpdate := compute.NewUpdateCommand(computeCmdRoot.CmdClause, opts.HTTPClient, &globals, data)
	computeValidate := compute.NewValidateCommand(computeCmdRoot.CmdClause, &globals)
	configureCmdRoot := configure.NewRootCommand(app, opts.ConfigPath, configure.APIClientFactory(opts.APIClient), &globals)
	dictionaryCmdRoot := edgedictionary.NewRootCommand(app, &globals)
	dictionaryCreate := edgedictionary.NewCreateCommand(dictionaryCmdRoot.CmdClause, &globals, data)
	dictionaryDelete := edgedictionary.NewDeleteCommand(dictionaryCmdRoot.CmdClause, &globals, data)
	dictionaryDescribe := edgedictionary.NewDescribeCommand(dictionaryCmdRoot.CmdClause, &globals, data)
	dictionaryItemCmdRoot := edgedictionaryitem.NewRootCommand(app, &globals)
	dictionaryItemBatchModify := edgedictionaryitem.NewBatchCommand(dictionaryItemCmdRoot.CmdClause, &globals, data)
	dictionaryItemCreate := edgedictionaryitem.NewCreateCommand(dictionaryItemCmdRoot.CmdClause, &globals, data)
	dictionaryItemDelete := edgedictionaryitem.NewDeleteCommand(dictionaryItemCmdRoot.CmdClause, &globals, data)
	dictionaryItemDescribe := edgedictionaryitem.NewDescribeCommand(dictionaryItemCmdRoot.CmdClause, &globals, data)
	dictionaryItemList := edgedictionaryitem.NewListCommand(dictionaryItemCmdRoot.CmdClause, &globals, data)
	dictionaryItemUpdate := edgedictionaryitem.NewUpdateCommand(dictionaryItemCmdRoot.CmdClause, &globals, data)
	dictionaryList := edgedictionary.NewListCommand(dictionaryCmdRoot.CmdClause, &globals, data)
	dictionaryUpdate := edgedictionary.NewUpdateCommand(dictionaryCmdRoot.CmdClause, &globals, data)
	domainCmdRoot := domain.NewRootCommand(app, &globals)
	domainCreate := domain.NewCreateCommand(domainCmdRoot.CmdClause, &globals, data)
	domainDelete := domain.NewDeleteCommand(domainCmdRoot.CmdClause, &globals, data)
	domainDescribe := domain.NewDescribeCommand(domainCmdRoot.CmdClause, &globals, data)
	domainList := domain.NewListCommand(domainCmdRoot.CmdClause, &globals, data)
	domainUpdate := domain.NewUpdateCommand(domainCmdRoot.CmdClause, &globals)
	domainValidate := domain.NewValidateCommand(domainCmdRoot.CmdClause, &globals, data)
	healthcheckCmdRoot := healthcheck.NewRootCommand(app, &globals)
	healthcheckCreate := healthcheck.NewCreateCommand(healthcheckCmdRoot.CmdClause, &globals)
	healthcheckDelete := healthcheck.NewDeleteCommand(healthcheckCmdRoot.CmdClause, &globals, data)
	healthcheckDescribe := healthcheck.NewDescribeCommand(healthcheckCmdRoot.CmdClause, &globals, data)
	healthcheckList := healthcheck.NewListCommand(healthcheckCmdRoot.CmdClause, &globals, data)
	healthcheckUpdate := healthcheck.NewUpdateCommand(healthcheckCmdRoot.CmdClause, &globals, data)
	ipCmdRoot := ip.NewRootCommand(app, &globals)
	loggingCmdRoot := logging.NewRootCommand(app, &globals)
	loggingAzureblobCmdRoot := azureblob.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingAzureblobCreate := azureblob.NewCreateCommand(loggingAzureblobCmdRoot.CmdClause, &globals)
	loggingAzureblobDelete := azureblob.NewDeleteCommand(loggingAzureblobCmdRoot.CmdClause, &globals, data)
	loggingAzureblobDescribe := azureblob.NewDescribeCommand(loggingAzureblobCmdRoot.CmdClause, &globals, data)
	loggingAzureblobList := azureblob.NewListCommand(loggingAzureblobCmdRoot.CmdClause, &globals, data)
	loggingAzureblobUpdate := azureblob.NewUpdateCommand(loggingAzureblobCmdRoot.CmdClause, &globals)
	loggingBigQueryCmdRoot := bigquery.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingBigQueryCreate := bigquery.NewCreateCommand(loggingBigQueryCmdRoot.CmdClause, &globals)
	loggingBigQueryDelete := bigquery.NewDeleteCommand(loggingBigQueryCmdRoot.CmdClause, &globals, data)
	loggingBigQueryDescribe := bigquery.NewDescribeCommand(loggingBigQueryCmdRoot.CmdClause, &globals, data)
	loggingBigQueryList := bigquery.NewListCommand(loggingBigQueryCmdRoot.CmdClause, &globals, data)
	loggingBigQueryUpdate := bigquery.NewUpdateCommand(loggingBigQueryCmdRoot.CmdClause, &globals)
	loggingCloudfilesCmdRoot := cloudfiles.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingCloudfilesCreate := cloudfiles.NewCreateCommand(loggingCloudfilesCmdRoot.CmdClause, &globals, data)
	loggingCloudfilesDelete := cloudfiles.NewDeleteCommand(loggingCloudfilesCmdRoot.CmdClause, &globals, data)
	loggingCloudfilesDescribe := cloudfiles.NewDescribeCommand(loggingCloudfilesCmdRoot.CmdClause, &globals, data)
	loggingCloudfilesList := cloudfiles.NewListCommand(loggingCloudfilesCmdRoot.CmdClause, &globals, data)
	loggingCloudfilesUpdate := cloudfiles.NewUpdateCommand(loggingCloudfilesCmdRoot.CmdClause, &globals, data)
	loggingDatadogCmdRoot := datadog.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingDatadogCreate := datadog.NewCreateCommand(loggingDatadogCmdRoot.CmdClause, &globals)
	loggingDatadogDelete := datadog.NewDeleteCommand(loggingDatadogCmdRoot.CmdClause, &globals, data)
	loggingDatadogDescribe := datadog.NewDescribeCommand(loggingDatadogCmdRoot.CmdClause, &globals, data)
	loggingDatadogList := datadog.NewListCommand(loggingDatadogCmdRoot.CmdClause, &globals, data)
	loggingDatadogUpdate := datadog.NewUpdateCommand(loggingDatadogCmdRoot.CmdClause, &globals)
	loggingDigitaloceanCmdRoot := digitalocean.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingDigitaloceanCreate := digitalocean.NewCreateCommand(loggingDigitaloceanCmdRoot.CmdClause, &globals, data)
	loggingDigitaloceanDelete := digitalocean.NewDeleteCommand(loggingDigitaloceanCmdRoot.CmdClause, &globals, data)
	loggingDigitaloceanDescribe := digitalocean.NewDescribeCommand(loggingDigitaloceanCmdRoot.CmdClause, &globals, data)
	loggingDigitaloceanList := digitalocean.NewListCommand(loggingDigitaloceanCmdRoot.CmdClause, &globals, data)
	loggingDigitaloceanUpdate := digitalocean.NewUpdateCommand(loggingDigitaloceanCmdRoot.CmdClause, &globals)
	loggingElasticsearchCmdRoot := elasticsearch.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingElasticsearchCreate := elasticsearch.NewCreateCommand(loggingElasticsearchCmdRoot.CmdClause, &globals)
	loggingElasticsearchDelete := elasticsearch.NewDeleteCommand(loggingElasticsearchCmdRoot.CmdClause, &globals, data)
	loggingElasticsearchDescribe := elasticsearch.NewDescribeCommand(loggingElasticsearchCmdRoot.CmdClause, &globals, data)
	loggingElasticsearchList := elasticsearch.NewListCommand(loggingElasticsearchCmdRoot.CmdClause, &globals, data)
	loggingElasticsearchUpdate := elasticsearch.NewUpdateCommand(loggingElasticsearchCmdRoot.CmdClause, &globals)
	loggingFtpCmdRoot := ftp.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingFtpCreate := ftp.NewCreateCommand(loggingFtpCmdRoot.CmdClause, &globals, data)
	loggingFtpDelete := ftp.NewDeleteCommand(loggingFtpCmdRoot.CmdClause, &globals, data)
	loggingFtpDescribe := ftp.NewDescribeCommand(loggingFtpCmdRoot.CmdClause, &globals, data)
	loggingFtpList := ftp.NewListCommand(loggingFtpCmdRoot.CmdClause, &globals, data)
	loggingFtpUpdate := ftp.NewUpdateCommand(loggingFtpCmdRoot.CmdClause, &globals, data)
	loggingGcsCmdRoot := gcs.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingGcsCreate := gcs.NewCreateCommand(loggingGcsCmdRoot.CmdClause, &globals)
	loggingGcsDelete := gcs.NewDeleteCommand(loggingGcsCmdRoot.CmdClause, &globals, data)
	loggingGcsDescribe := gcs.NewDescribeCommand(loggingGcsCmdRoot.CmdClause, &globals, data)
	loggingGcsList := gcs.NewListCommand(loggingGcsCmdRoot.CmdClause, &globals, data)
	loggingGcsUpdate := gcs.NewUpdateCommand(loggingGcsCmdRoot.CmdClause, &globals)
	loggingGooglepubsubCmdRoot := googlepubsub.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingGooglepubsubCreate := googlepubsub.NewCreateCommand(loggingGooglepubsubCmdRoot.CmdClause, &globals)
	loggingGooglepubsubDelete := googlepubsub.NewDeleteCommand(loggingGooglepubsubCmdRoot.CmdClause, &globals, data)
	loggingGooglepubsubDescribe := googlepubsub.NewDescribeCommand(loggingGooglepubsubCmdRoot.CmdClause, &globals, data)
	loggingGooglepubsubList := googlepubsub.NewListCommand(loggingGooglepubsubCmdRoot.CmdClause, &globals, data)
	loggingGooglepubsubUpdate := googlepubsub.NewUpdateCommand(loggingGooglepubsubCmdRoot.CmdClause, &globals)
	loggingHerokuCmdRoot := heroku.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingHerokuCreate := heroku.NewCreateCommand(loggingHerokuCmdRoot.CmdClause, &globals, data)
	loggingHerokuDelete := heroku.NewDeleteCommand(loggingHerokuCmdRoot.CmdClause, &globals, data)
	loggingHerokuDescribe := heroku.NewDescribeCommand(loggingHerokuCmdRoot.CmdClause, &globals, data)
	loggingHerokuList := heroku.NewListCommand(loggingHerokuCmdRoot.CmdClause, &globals, data)
	loggingHerokuUpdate := heroku.NewUpdateCommand(loggingHerokuCmdRoot.CmdClause, &globals, data)
	loggingHoneycombCmdRoot := honeycomb.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingHoneycombCreate := honeycomb.NewCreateCommand(loggingHoneycombCmdRoot.CmdClause, &globals)
	loggingHoneycombDelete := honeycomb.NewDeleteCommand(loggingHoneycombCmdRoot.CmdClause, &globals, data)
	loggingHoneycombDescribe := honeycomb.NewDescribeCommand(loggingHoneycombCmdRoot.CmdClause, &globals, data)
	loggingHoneycombList := honeycomb.NewListCommand(loggingHoneycombCmdRoot.CmdClause, &globals, data)
	loggingHoneycombUpdate := honeycomb.NewUpdateCommand(loggingHoneycombCmdRoot.CmdClause, &globals)
	loggingHTTPSCmdRoot := https.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingHTTPSCreate := https.NewCreateCommand(loggingHTTPSCmdRoot.CmdClause, &globals, data)
	loggingHTTPSDelete := https.NewDeleteCommand(loggingHTTPSCmdRoot.CmdClause, &globals, data)
	loggingHTTPSDescribe := https.NewDescribeCommand(loggingHTTPSCmdRoot.CmdClause, &globals, data)
	loggingHTTPSList := https.NewListCommand(loggingHTTPSCmdRoot.CmdClause, &globals, data)
	loggingHTTPSUpdate := https.NewUpdateCommand(loggingHTTPSCmdRoot.CmdClause, &globals, data)
	loggingKafkaCmdRoot := kafka.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingKafkaCreate := kafka.NewCreateCommand(loggingKafkaCmdRoot.CmdClause, &globals)
	loggingKafkaDelete := kafka.NewDeleteCommand(loggingKafkaCmdRoot.CmdClause, &globals, data)
	loggingKafkaDescribe := kafka.NewDescribeCommand(loggingKafkaCmdRoot.CmdClause, &globals, data)
	loggingKafkaList := kafka.NewListCommand(loggingKafkaCmdRoot.CmdClause, &globals, data)
	loggingKafkaUpdate := kafka.NewUpdateCommand(loggingKafkaCmdRoot.CmdClause, &globals)
	loggingKinesisCmdRoot := kinesis.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingKinesisCreate := kinesis.NewCreateCommand(loggingKinesisCmdRoot.CmdClause, &globals, data)
	loggingKinesisDelete := kinesis.NewDeleteCommand(loggingKinesisCmdRoot.CmdClause, &globals, data)
	loggingKinesisDescribe := kinesis.NewDescribeCommand(loggingKinesisCmdRoot.CmdClause, &globals, data)
	loggingKinesisList := kinesis.NewListCommand(loggingKinesisCmdRoot.CmdClause, &globals, data)
	loggingKinesisUpdate := kinesis.NewUpdateCommand(loggingKinesisCmdRoot.CmdClause, &globals, data)
	loggingLogentriesCmdRoot := logentries.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingLogentriesCreate := logentries.NewCreateCommand(loggingLogentriesCmdRoot.CmdClause, &globals, data)
	loggingLogentriesDelete := logentries.NewDeleteCommand(loggingLogentriesCmdRoot.CmdClause, &globals, data)
	loggingLogentriesDescribe := logentries.NewDescribeCommand(loggingLogentriesCmdRoot.CmdClause, &globals, data)
	loggingLogentriesList := logentries.NewListCommand(loggingLogentriesCmdRoot.CmdClause, &globals, data)
	loggingLogentriesUpdate := logentries.NewUpdateCommand(loggingLogentriesCmdRoot.CmdClause, &globals, data)
	loggingLogglyCmdRoot := loggly.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingLogglyCreate := loggly.NewCreateCommand(loggingLogglyCmdRoot.CmdClause, &globals)
	loggingLogglyDelete := loggly.NewDeleteCommand(loggingLogglyCmdRoot.CmdClause, &globals, data)
	loggingLogglyDescribe := loggly.NewDescribeCommand(loggingLogglyCmdRoot.CmdClause, &globals, data)
	loggingLogglyList := loggly.NewListCommand(loggingLogglyCmdRoot.CmdClause, &globals, data)
	loggingLogglyUpdate := loggly.NewUpdateCommand(loggingLogglyCmdRoot.CmdClause, &globals)
	loggingLogshuttleCmdRoot := logshuttle.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingLogshuttleCreate := logshuttle.NewCreateCommand(loggingLogshuttleCmdRoot.CmdClause, &globals)
	loggingLogshuttleDelete := logshuttle.NewDeleteCommand(loggingLogshuttleCmdRoot.CmdClause, &globals, data)
	loggingLogshuttleDescribe := logshuttle.NewDescribeCommand(loggingLogshuttleCmdRoot.CmdClause, &globals, data)
	loggingLogshuttleList := logshuttle.NewListCommand(loggingLogshuttleCmdRoot.CmdClause, &globals, data)
	loggingLogshuttleUpdate := logshuttle.NewUpdateCommand(loggingLogshuttleCmdRoot.CmdClause, &globals)
	loggingNewRelicCmdRoot := newrelic.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingNewRelicCreate := newrelic.NewCreateCommand(loggingNewRelicCmdRoot.CmdClause, &globals, data)
	loggingNewRelicDelete := newrelic.NewDeleteCommand(loggingNewRelicCmdRoot.CmdClause, &globals, data)
	loggingNewRelicDescribe := newrelic.NewDescribeCommand(loggingNewRelicCmdRoot.CmdClause, &globals, data)
	loggingNewRelicList := newrelic.NewListCommand(loggingNewRelicCmdRoot.CmdClause, &globals, data)
	loggingNewRelicUpdate := newrelic.NewUpdateCommand(loggingNewRelicCmdRoot.CmdClause, &globals, data)
	loggingOpenstackCmdRoot := openstack.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingOpenstackCreate := openstack.NewCreateCommand(loggingOpenstackCmdRoot.CmdClause, &globals, data)
	loggingOpenstackDelete := openstack.NewDeleteCommand(loggingOpenstackCmdRoot.CmdClause, &globals, data)
	loggingOpenstackDescribe := openstack.NewDescribeCommand(loggingOpenstackCmdRoot.CmdClause, &globals, data)
	loggingOpenstackList := openstack.NewListCommand(loggingOpenstackCmdRoot.CmdClause, &globals, data)
	loggingOpenstackUpdate := openstack.NewUpdateCommand(loggingOpenstackCmdRoot.CmdClause, &globals, data)
	loggingPapertrailCmdRoot := papertrail.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingPapertrailCreate := papertrail.NewCreateCommand(loggingPapertrailCmdRoot.CmdClause, &globals)
	loggingPapertrailDelete := papertrail.NewDeleteCommand(loggingPapertrailCmdRoot.CmdClause, &globals, data)
	loggingPapertrailDescribe := papertrail.NewDescribeCommand(loggingPapertrailCmdRoot.CmdClause, &globals, data)
	loggingPapertrailList := papertrail.NewListCommand(loggingPapertrailCmdRoot.CmdClause, &globals, data)
	loggingPapertrailUpdate := papertrail.NewUpdateCommand(loggingPapertrailCmdRoot.CmdClause, &globals)
	loggingS3CmdRoot := s3.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingS3Create := s3.NewCreateCommand(loggingS3CmdRoot.CmdClause, &globals, data)
	loggingS3Delete := s3.NewDeleteCommand(loggingS3CmdRoot.CmdClause, &globals, data)
	loggingS3Describe := s3.NewDescribeCommand(loggingS3CmdRoot.CmdClause, &globals, data)
	loggingS3List := s3.NewListCommand(loggingS3CmdRoot.CmdClause, &globals, data)
	loggingS3Update := s3.NewUpdateCommand(loggingS3CmdRoot.CmdClause, &globals, data)
	loggingScalyrCmdRoot := scalyr.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingScalyrCreate := scalyr.NewCreateCommand(loggingScalyrCmdRoot.CmdClause, &globals)
	loggingScalyrDelete := scalyr.NewDeleteCommand(loggingScalyrCmdRoot.CmdClause, &globals, data)
	loggingScalyrDescribe := scalyr.NewDescribeCommand(loggingScalyrCmdRoot.CmdClause, &globals, data)
	loggingScalyrList := scalyr.NewListCommand(loggingScalyrCmdRoot.CmdClause, &globals, data)
	loggingScalyrUpdate := scalyr.NewUpdateCommand(loggingScalyrCmdRoot.CmdClause, &globals)
	loggingSftpCmdRoot := sftp.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingSftpCreate := sftp.NewCreateCommand(loggingSftpCmdRoot.CmdClause, &globals)
	loggingSftpDelete := sftp.NewDeleteCommand(loggingSftpCmdRoot.CmdClause, &globals, data)
	loggingSftpDescribe := sftp.NewDescribeCommand(loggingSftpCmdRoot.CmdClause, &globals, data)
	loggingSftpList := sftp.NewListCommand(loggingSftpCmdRoot.CmdClause, &globals, data)
	loggingSftpUpdate := sftp.NewUpdateCommand(loggingSftpCmdRoot.CmdClause, &globals)
	loggingSplunkCmdRoot := splunk.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingSplunkCreate := splunk.NewCreateCommand(loggingSplunkCmdRoot.CmdClause, &globals)
	loggingSplunkDelete := splunk.NewDeleteCommand(loggingSplunkCmdRoot.CmdClause, &globals, data)
	loggingSplunkDescribe := splunk.NewDescribeCommand(loggingSplunkCmdRoot.CmdClause, &globals, data)
	loggingSplunkList := splunk.NewListCommand(loggingSplunkCmdRoot.CmdClause, &globals, data)
	loggingSplunkUpdate := splunk.NewUpdateCommand(loggingSplunkCmdRoot.CmdClause, &globals)
	loggingSumologicCmdRoot := sumologic.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingSumologicCreate := sumologic.NewCreateCommand(loggingSumologicCmdRoot.CmdClause, &globals)
	loggingSumologicDelete := sumologic.NewDeleteCommand(loggingSumologicCmdRoot.CmdClause, &globals, data)
	loggingSumologicDescribe := sumologic.NewDescribeCommand(loggingSumologicCmdRoot.CmdClause, &globals, data)
	loggingSumologicList := sumologic.NewListCommand(loggingSumologicCmdRoot.CmdClause, &globals, data)
	loggingSumologicUpdate := sumologic.NewUpdateCommand(loggingSumologicCmdRoot.CmdClause, &globals)
	loggingSyslogCmdRoot := syslog.NewRootCommand(loggingCmdRoot.CmdClause, &globals)
	loggingSyslogCreate := syslog.NewCreateCommand(loggingSyslogCmdRoot.CmdClause, &globals, data)
	loggingSyslogDelete := syslog.NewDeleteCommand(loggingSyslogCmdRoot.CmdClause, &globals, data)
	loggingSyslogDescribe := syslog.NewDescribeCommand(loggingSyslogCmdRoot.CmdClause, &globals, data)
	loggingSyslogList := syslog.NewListCommand(loggingSyslogCmdRoot.CmdClause, &globals, data)
	loggingSyslogUpdate := syslog.NewUpdateCommand(loggingSyslogCmdRoot.CmdClause, &globals, data)
	logsCmdRoot := logs.NewRootCommand(app, &globals)
	logsTail := logs.NewTailCommand(logsCmdRoot.CmdClause, &globals, data)
	popCmdRoot := pop.NewRootCommand(app, &globals)
	purgeCmdRoot := purge.NewRootCommand(app, &globals, data)
	serviceCmdRoot := service.NewRootCommand(app, &globals)
	serviceCreate := service.NewCreateCommand(serviceCmdRoot.CmdClause, &globals)
	serviceDelete := service.NewDeleteCommand(serviceCmdRoot.CmdClause, &globals, data)
	serviceDescribe := service.NewDescribeCommand(serviceCmdRoot.CmdClause, &globals, data)
	serviceList := service.NewListCommand(serviceCmdRoot.CmdClause, &globals)
	serviceSearch := service.NewSearchCommand(serviceCmdRoot.CmdClause, &globals, data)
	serviceUpdate := service.NewUpdateCommand(serviceCmdRoot.CmdClause, &globals, data)
	serviceVersionCmdRoot := serviceversion.NewRootCommand(app, &globals)
	serviceVersionActivate := serviceversion.NewActivateCommand(serviceVersionCmdRoot.CmdClause, &globals, data)
	serviceVersionClone := serviceversion.NewCloneCommand(serviceVersionCmdRoot.CmdClause, &globals, data)
	serviceVersionDeactivate := serviceversion.NewDeactivateCommand(serviceVersionCmdRoot.CmdClause, &globals, data)
	serviceVersionList := serviceversion.NewListCommand(serviceVersionCmdRoot.CmdClause, &globals, data)
	serviceVersionLock := serviceversion.NewLockCommand(serviceVersionCmdRoot.CmdClause, &globals, data)
	serviceVersionUpdate := serviceversion.NewUpdateCommand(serviceVersionCmdRoot.CmdClause, &globals, data)
	statsCmdRoot := stats.NewRootCommand(app, &globals)
	statsHistorical := stats.NewHistoricalCommand(statsCmdRoot.CmdClause, &globals)
	statsRealtime := stats.NewRealtimeCommand(statsCmdRoot.CmdClause, &globals)
	statsRegions := stats.NewRegionsCommand(statsCmdRoot.CmdClause, &globals)
	updateRoot := update.NewRootCommand(app, opts.ConfigPath, opts.Versioners.CLI, opts.HTTPClient, &globals)
	userCmdRoot := user.NewRootCommand(app, &globals)
	userCreate := user.NewCreateCommand(userCmdRoot.CmdClause, &globals, data)
	userDelete := user.NewDeleteCommand(userCmdRoot.CmdClause, &globals, data)
	userDescribe := user.NewDescribeCommand(userCmdRoot.CmdClause, &globals, data)
	userList := user.NewListCommand(userCmdRoot.CmdClause, &globals, data)
	userUpdate := user.NewUpdateCommand(userCmdRoot.CmdClause, &globals, data)
	vclCmdRoot := vcl.NewRootCommand(app, &globals)
	vclCustomCmdRoot := custom.NewRootCommand(vclCmdRoot.CmdClause, &globals)
	vclCustomCreate := custom.NewCreateCommand(vclCustomCmdRoot.CmdClause, &globals, data)
	vclCustomDelete := custom.NewDeleteCommand(vclCustomCmdRoot.CmdClause, &globals, data)
	vclCustomDescribe := custom.NewDescribeCommand(vclCustomCmdRoot.CmdClause, &globals, data)
	vclCustomList := custom.NewListCommand(vclCustomCmdRoot.CmdClause, &globals, data)
	vclCustomUpdate := custom.NewUpdateCommand(vclCustomCmdRoot.CmdClause, &globals, data)
	vclSnippetCmdRoot := snippet.NewRootCommand(vclCmdRoot.CmdClause, &globals)
	vclSnippetCreate := snippet.NewCreateCommand(vclSnippetCmdRoot.CmdClause, &globals, data)
	vclSnippetDelete := snippet.NewDeleteCommand(vclSnippetCmdRoot.CmdClause, &globals, data)
	vclSnippetDescribe := snippet.NewDescribeCommand(vclSnippetCmdRoot.CmdClause, &globals, data)
	vclSnippetList := snippet.NewListCommand(vclSnippetCmdRoot.CmdClause, &globals, data)
	vclSnippetUpdate := snippet.NewUpdateCommand(vclSnippetCmdRoot.CmdClause, &globals, data)
	versionCmdRoot := version.NewRootCommand(app)
	whoamiCmdRoot := whoami.NewRootCommand(app, opts.HTTPClient, &globals)

	commands := []cmd.Command{
		aclCmdRoot,
		aclCreate,
		aclDelete,
		aclDescribe,
		aclList,
		aclUpdate,
		aclEntryCmdRoot,
		aclEntryCreate,
		aclEntryDelete,
		aclEntryDescribe,
		aclEntryList,
		aclEntryUpdate,
		authtokenCmdRoot,
		authtokenCreate,
		authtokenDelete,
		authtokenDescribe,
		authtokenList,
		backendCmdRoot,
		backendCreate,
		backendDelete,
		backendDescribe,
		backendList,
		backendUpdate,
		computeBuild,
		computeCmdRoot,
		computeDeploy,
		computeInit,
		computePack,
		computePublish,
		computeServe,
		computeUpdate,
		computeValidate,
		configureCmdRoot,
		dictionaryCmdRoot,
		dictionaryCreate,
		dictionaryDelete,
		dictionaryDescribe,
		dictionaryItemBatchModify,
		dictionaryItemCmdRoot,
		dictionaryItemCreate,
		dictionaryItemDelete,
		dictionaryItemDescribe,
		dictionaryItemList,
		dictionaryItemUpdate,
		dictionaryList,
		dictionaryUpdate,
		domainCmdRoot,
		domainCreate,
		domainDelete,
		domainDescribe,
		domainList,
		domainUpdate,
		domainValidate,
		healthcheckCmdRoot,
		healthcheckCreate,
		healthcheckDelete,
		healthcheckDescribe,
		healthcheckList,
		healthcheckUpdate,
		ipCmdRoot,
		loggingAzureblobCmdRoot,
		loggingAzureblobCreate,
		loggingAzureblobDelete,
		loggingAzureblobDescribe,
		loggingAzureblobList,
		loggingAzureblobUpdate,
		loggingBigQueryCmdRoot,
		loggingBigQueryCreate,
		loggingBigQueryDelete,
		loggingBigQueryDescribe,
		loggingBigQueryList,
		loggingBigQueryUpdate,
		loggingCloudfilesCmdRoot,
		loggingCloudfilesCreate,
		loggingCloudfilesDelete,
		loggingCloudfilesDescribe,
		loggingCloudfilesList,
		loggingCloudfilesUpdate,
		loggingCmdRoot,
		loggingDatadogCmdRoot,
		loggingDatadogCreate,
		loggingDatadogDelete,
		loggingDatadogDescribe,
		loggingDatadogList,
		loggingDatadogUpdate,
		loggingDigitaloceanCmdRoot,
		loggingDigitaloceanCreate,
		loggingDigitaloceanDelete,
		loggingDigitaloceanDescribe,
		loggingDigitaloceanList,
		loggingDigitaloceanUpdate,
		loggingElasticsearchCmdRoot,
		loggingElasticsearchCreate,
		loggingElasticsearchDelete,
		loggingElasticsearchDescribe,
		loggingElasticsearchList,
		loggingElasticsearchUpdate,
		loggingFtpCmdRoot,
		loggingFtpCreate,
		loggingFtpDelete,
		loggingFtpDescribe,
		loggingFtpList,
		loggingFtpUpdate,
		loggingGcsCmdRoot,
		loggingGcsCreate,
		loggingGcsDelete,
		loggingGcsDescribe,
		loggingGcsList,
		loggingGcsUpdate,
		loggingGooglepubsubCmdRoot,
		loggingGooglepubsubCreate,
		loggingGooglepubsubDelete,
		loggingGooglepubsubDescribe,
		loggingGooglepubsubList,
		loggingGooglepubsubUpdate,
		loggingHerokuCmdRoot,
		loggingHerokuCreate,
		loggingHerokuDelete,
		loggingHerokuDescribe,
		loggingHerokuList,
		loggingHerokuUpdate,
		loggingHoneycombCmdRoot,
		loggingHoneycombCreate,
		loggingHoneycombDelete,
		loggingHoneycombDescribe,
		loggingHoneycombList,
		loggingHoneycombUpdate,
		loggingHTTPSCmdRoot,
		loggingHTTPSCreate,
		loggingHTTPSDelete,
		loggingHTTPSDescribe,
		loggingHTTPSList,
		loggingHTTPSUpdate,
		loggingKafkaCmdRoot,
		loggingKafkaCreate,
		loggingKafkaDelete,
		loggingKafkaDescribe,
		loggingKafkaList,
		loggingKafkaUpdate,
		loggingKinesisCmdRoot,
		loggingKinesisCreate,
		loggingKinesisDelete,
		loggingKinesisDescribe,
		loggingKinesisList,
		loggingKinesisUpdate,
		loggingLogentriesCmdRoot,
		loggingLogentriesCreate,
		loggingLogentriesDelete,
		loggingLogentriesDescribe,
		loggingLogentriesList,
		loggingLogentriesUpdate,
		loggingLogglyCmdRoot,
		loggingLogglyCreate,
		loggingLogglyDelete,
		loggingLogglyDescribe,
		loggingLogglyList,
		loggingLogglyUpdate,
		loggingLogshuttleCmdRoot,
		loggingLogshuttleCreate,
		loggingLogshuttleDelete,
		loggingLogshuttleDescribe,
		loggingLogshuttleList,
		loggingLogshuttleUpdate,
		loggingNewRelicCmdRoot,
		loggingNewRelicCreate,
		loggingNewRelicDelete,
		loggingNewRelicDescribe,
		loggingNewRelicList,
		loggingNewRelicUpdate,
		loggingOpenstackCmdRoot,
		loggingOpenstackCreate,
		loggingOpenstackDelete,
		loggingOpenstackDescribe,
		loggingOpenstackList,
		loggingOpenstackUpdate,
		loggingPapertrailCmdRoot,
		loggingPapertrailCreate,
		loggingPapertrailDelete,
		loggingPapertrailDescribe,
		loggingPapertrailList,
		loggingPapertrailUpdate,
		loggingS3CmdRoot,
		loggingS3Create,
		loggingS3Delete,
		loggingS3Describe,
		loggingS3List,
		loggingS3Update,
		loggingScalyrCmdRoot,
		loggingScalyrCreate,
		loggingScalyrDelete,
		loggingScalyrDescribe,
		loggingScalyrList,
		loggingScalyrUpdate,
		loggingSftpCmdRoot,
		loggingSftpCreate,
		loggingSftpDelete,
		loggingSftpDescribe,
		loggingSftpList,
		loggingSftpUpdate,
		loggingSplunkCmdRoot,
		loggingSplunkCreate,
		loggingSplunkDelete,
		loggingSplunkDescribe,
		loggingSplunkList,
		loggingSplunkUpdate,
		loggingSumologicCmdRoot,
		loggingSumologicCreate,
		loggingSumologicDelete,
		loggingSumologicDescribe,
		loggingSumologicList,
		loggingSumologicUpdate,
		loggingSyslogCmdRoot,
		loggingSyslogCreate,
		loggingSyslogDelete,
		loggingSyslogDescribe,
		loggingSyslogList,
		loggingSyslogUpdate,
		logsCmdRoot,
		logsTail,
		popCmdRoot,
		purgeCmdRoot,
		serviceCmdRoot,
		serviceCreate,
		serviceDelete,
		serviceDescribe,
		serviceList,
		serviceSearch,
		serviceUpdate,
		serviceVersionActivate,
		serviceVersionClone,
		serviceVersionCmdRoot,
		serviceVersionDeactivate,
		serviceVersionList,
		serviceVersionLock,
		serviceVersionUpdate,
		statsCmdRoot,
		statsHistorical,
		statsRealtime,
		statsRegions,
		updateRoot,
		userCmdRoot,
		userCreate,
		userDelete,
		userDescribe,
		userList,
		userUpdate,
		vclCmdRoot,
		vclCustomCmdRoot,
		vclCustomCreate,
		vclCustomDelete,
		vclCustomDescribe,
		vclCustomList,
		vclCustomUpdate,
		vclSnippetCmdRoot,
		vclSnippetCreate,
		vclSnippetDelete,
		vclSnippetDescribe,
		vclSnippetList,
		vclSnippetUpdate,
		versionCmdRoot,
		whoamiCmdRoot,
	}

	// Get information about the incoming command/args.
	ctx, _ := app.ParseContext(opts.Args)

	// Identify the command to be executed
	command, found := cmd.Select(ctx.SelectedCommand.FullCommand(), commands)
	var examples string
	if found {
		examples = command.Examples()
	}

	// Usage scenario: command has --help flag
	if contextHasHelpFlag(ctx) {
		usage := fmt.Sprintf("%s%s", Usage(opts.Args, app, opts.Stdout, io.Discard), examples)
		return errors.RemediationError{Prefix: usage}
	}

	// Usage scenario: subcommand not recognised (e.g. typo when writing the command)
	if !found {
		usage := fmt.Sprintf("%s%s", Usage(opts.Args, app, opts.Stdout, io.Discard), examples)
		return errors.RemediationError{Prefix: usage, Inner: fmt.Errorf("command not found")}
	}

	// Handle parse errors and display contextual usage if possible. Due to bugs
	// and an obession for lots of output side-effects in the kingpin.Parse
	// logic, we suppress it from writing any usage or errors to the writer by
	// swapping the writer with a no-op and then restoring the real writer
	// afterwards. This ensures usage text is only written once to the writer
	// and gives us greater control over our error formatting.
	app.Writers(io.Discard, io.Discard)

	// NOTE: `name` is the command including any subcommands.
	// e.g. `pops` or `backend create`
	//
	// The Parse() method is different from the earlier ParseContext() in that it
	// also validates the flags and will error, for example, if the caller missed
	// a required flag.
	name, err := app.Parse(opts.Args)

	// Usage scenario: unable to parse command (e.g. required flag not provided)
	if err != nil && !argsIsHelpJSON(opts.Args) { // Ignore error if `help --format json`
		globals.ErrLog.Add(err)
		usage := Usage(opts.Args, app, opts.Stdout, io.Discard)
		return errors.RemediationError{Prefix: usage, Inner: fmt.Errorf("error parsing arguments: %w", err)}
	}

	// Restore output writers
	app.Writers(opts.Stdout, io.Discard)

	// As the `help` command model gets privately added as a side-effect of
	// kingping.Parse, we cannot add the `--format json` flag to the model.
	// Therefore, we have to manually parse the args slice here to check for the
	// existence of `help --format json`, if present we print usage JSON and
	// exit early.
	if argsIsHelpJSON(opts.Args) {
		json, err := UsageJSON(app)
		if err != nil {
			globals.ErrLog.Add(err)
			return err
		}
		fmt.Fprintf(opts.Stdout, "%s", json)
		return nil
	}

	// A side-effect of suppressing app.Parse from writing output is the usage
	// isn't printed for the default `help` command. Therefore we capture it
	// here by calling Parse, again swapping the Writers. This also ensures the
	// larger and more verbose help formatting is used.
	if name == "help" {
		var buf bytes.Buffer
		app.Writers(&buf, io.Discard)
		app.Parse(opts.Args)
		app.Writers(opts.Stdout, io.Discard)

		// The full-fat output of `fastly help` should have a hint at the bottom
		// for more specific help. Unfortunately I don't know of a better way to
		// distinguish `fastly help` from e.g. `fastly help configure` than this
		// check.
		if len(opts.Args) > 0 && opts.Args[len(opts.Args)-1] == "help" {
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

// contextHasHelpFlag asserts whether a given kingpin.ParseContext contains a
// `help` flag.
func contextHasHelpFlag(ctx *kingpin.ParseContext) bool {
	_, ok := ctx.Elements.FlagMap()["help"]
	return ok
}

// argsIsHelpJSON determines whether the supplied command arguments are exactly
// `help --format json`.
func argsIsHelpJSON(args []string) bool {
	return (len(args) == 3 &&
		args[0] == "help" &&
		args[1] == "--format" &&
		args[2] == "json")
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
