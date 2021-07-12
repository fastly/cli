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
	"github.com/fastly/cli/pkg/backend"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/configure"
	"github.com/fastly/cli/pkg/domain"
	"github.com/fastly/cli/pkg/edgedictionary"
	"github.com/fastly/cli/pkg/edgedictionaryitem"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/healthcheck"
	"github.com/fastly/cli/pkg/ip"
	"github.com/fastly/cli/pkg/logging"
	"github.com/fastly/cli/pkg/logging/azureblob"
	"github.com/fastly/cli/pkg/logging/bigquery"
	"github.com/fastly/cli/pkg/logging/cloudfiles"
	"github.com/fastly/cli/pkg/logging/datadog"
	"github.com/fastly/cli/pkg/logging/digitalocean"
	"github.com/fastly/cli/pkg/logging/elasticsearch"
	"github.com/fastly/cli/pkg/logging/ftp"
	"github.com/fastly/cli/pkg/logging/gcs"
	"github.com/fastly/cli/pkg/logging/googlepubsub"
	"github.com/fastly/cli/pkg/logging/heroku"
	"github.com/fastly/cli/pkg/logging/honeycomb"
	"github.com/fastly/cli/pkg/logging/https"
	"github.com/fastly/cli/pkg/logging/kafka"
	"github.com/fastly/cli/pkg/logging/kinesis"
	"github.com/fastly/cli/pkg/logging/logentries"
	"github.com/fastly/cli/pkg/logging/loggly"
	"github.com/fastly/cli/pkg/logging/logshuttle"
	"github.com/fastly/cli/pkg/logging/openstack"
	"github.com/fastly/cli/pkg/logging/papertrail"
	"github.com/fastly/cli/pkg/logging/s3"
	"github.com/fastly/cli/pkg/logging/scalyr"
	"github.com/fastly/cli/pkg/logging/sftp"
	"github.com/fastly/cli/pkg/logging/splunk"
	"github.com/fastly/cli/pkg/logging/sumologic"
	"github.com/fastly/cli/pkg/logging/syslog"
	"github.com/fastly/cli/pkg/logs"
	"github.com/fastly/cli/pkg/pop"
	"github.com/fastly/cli/pkg/purge"
	"github.com/fastly/cli/pkg/revision"
	"github.com/fastly/cli/pkg/service"
	"github.com/fastly/cli/pkg/serviceversion"
	"github.com/fastly/cli/pkg/stats"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/cli/pkg/update"
	"github.com/fastly/cli/pkg/vcl"
	"github.com/fastly/cli/pkg/vcl/custom"
	"github.com/fastly/cli/pkg/vcl/snippet"
	"github.com/fastly/cli/pkg/version"
	"github.com/fastly/cli/pkg/whoami"
	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/fastly/kingpin"
)

var (
	completionRegExp = regexp.MustCompile("completion-(?:script-)?(?:bash|zsh)$")
)

// RunOpts represent arguments to Run()
type RunOpts struct {
	APIClient    APIClientFactory
	Args         []string
	ConfigFile   config.File
	ConfigPath   string
	Env          config.Environment
	ErrLog       errors.LogInterface
	HTTPClient   api.HTTPClient
	Stdin        io.Reader
	Stdout       io.Writer
	VersionerCLI update.Versioner
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

	configureRoot := configure.NewRootCommand(app, opts.ConfigPath, configure.APIClientFactory(opts.APIClient), &globals)
	whoamiRoot := whoami.NewRootCommand(app, opts.HTTPClient, &globals)
	versionRoot := version.NewRootCommand(app)
	updateRoot := update.NewRootCommand(app, opts.ConfigPath, opts.VersionerCLI, opts.HTTPClient, &globals)
	ipRoot := ip.NewRootCommand(app, &globals)
	popRoot := pop.NewRootCommand(app, &globals)
	purgeRoot := purge.NewRootCommand(app, &globals)

	serviceRoot := service.NewRootCommand(app, &globals)
	serviceCreate := service.NewCreateCommand(serviceRoot.CmdClause, &globals)
	serviceList := service.NewListCommand(serviceRoot.CmdClause, &globals)
	serviceDescribe := service.NewDescribeCommand(serviceRoot.CmdClause, &globals)
	serviceUpdate := service.NewUpdateCommand(serviceRoot.CmdClause, &globals)
	serviceDelete := service.NewDeleteCommand(serviceRoot.CmdClause, &globals)
	serviceSearch := service.NewSearchCommand(serviceRoot.CmdClause, &globals)

	serviceVersionRoot := serviceversion.NewRootCommand(app, &globals)
	serviceVersionClone := serviceversion.NewCloneCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionList := serviceversion.NewListCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionUpdate := serviceversion.NewUpdateCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionActivate := serviceversion.NewActivateCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionDeactivate := serviceversion.NewDeactivateCommand(serviceVersionRoot.CmdClause, &globals)
	serviceVersionLock := serviceversion.NewLockCommand(serviceVersionRoot.CmdClause, &globals)

	computeRoot := compute.NewRootCommand(app, &globals)
	computeInit := compute.NewInitCommand(computeRoot.CmdClause, opts.HTTPClient, &globals)
	computeBuild := compute.NewBuildCommand(computeRoot.CmdClause, opts.HTTPClient, &globals)
	computeDeploy := compute.NewDeployCommand(computeRoot.CmdClause, opts.HTTPClient, &globals)
	computePublish := compute.NewPublishCommand(computeRoot.CmdClause, &globals, computeBuild, computeDeploy)
	computePack := compute.NewPackCommand(computeRoot.CmdClause, &globals)
	computeUpdate := compute.NewUpdateCommand(computeRoot.CmdClause, opts.HTTPClient, &globals)
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

	dictionaryRoot := edgedictionary.NewRootCommand(app, &globals)
	dictionaryCreate := edgedictionary.NewCreateCommand(dictionaryRoot.CmdClause, &globals)
	dictionaryDescribe := edgedictionary.NewDescribeCommand(dictionaryRoot.CmdClause, &globals)
	dictionaryDelete := edgedictionary.NewDeleteCommand(dictionaryRoot.CmdClause, &globals)
	dictionaryList := edgedictionary.NewListCommand(dictionaryRoot.CmdClause, &globals)
	dictionaryUpdate := edgedictionary.NewUpdateCommand(dictionaryRoot.CmdClause, &globals)

	dictionaryItemRoot := edgedictionaryitem.NewRootCommand(app, &globals)
	dictionaryItemList := edgedictionaryitem.NewListCommand(dictionaryItemRoot.CmdClause, &globals)
	dictionaryItemDescribe := edgedictionaryitem.NewDescribeCommand(dictionaryItemRoot.CmdClause, &globals)
	dictionaryItemCreate := edgedictionaryitem.NewCreateCommand(dictionaryItemRoot.CmdClause, &globals)
	dictionaryItemUpdate := edgedictionaryitem.NewUpdateCommand(dictionaryItemRoot.CmdClause, &globals)
	dictionaryItemDelete := edgedictionaryitem.NewDeleteCommand(dictionaryItemRoot.CmdClause, &globals)
	dictionaryItemBatchModify := edgedictionaryitem.NewBatchCommand(dictionaryItemRoot.CmdClause, &globals)

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

	kinesisRoot := kinesis.NewRootCommand(loggingRoot.CmdClause, &globals)
	kinesisCreate := kinesis.NewCreateCommand(kinesisRoot.CmdClause, &globals)
	kinesisList := kinesis.NewListCommand(kinesisRoot.CmdClause, &globals)
	kinesisDescribe := kinesis.NewDescribeCommand(kinesisRoot.CmdClause, &globals)
	kinesisUpdate := kinesis.NewUpdateCommand(kinesisRoot.CmdClause, &globals)
	kinesisDelete := kinesis.NewDeleteCommand(kinesisRoot.CmdClause, &globals)

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

	splunkRoot := splunk.NewRootCommand(loggingRoot.CmdClause, &globals)
	splunkCreate := splunk.NewCreateCommand(splunkRoot.CmdClause, &globals)
	splunkList := splunk.NewListCommand(splunkRoot.CmdClause, &globals)
	splunkDescribe := splunk.NewDescribeCommand(splunkRoot.CmdClause, &globals)
	splunkUpdate := splunk.NewUpdateCommand(splunkRoot.CmdClause, &globals)
	splunkDelete := splunk.NewDeleteCommand(splunkRoot.CmdClause, &globals)

	scalyrRoot := scalyr.NewRootCommand(loggingRoot.CmdClause, &globals)
	scalyrCreate := scalyr.NewCreateCommand(scalyrRoot.CmdClause, &globals)
	scalyrList := scalyr.NewListCommand(scalyrRoot.CmdClause, &globals)
	scalyrDescribe := scalyr.NewDescribeCommand(scalyrRoot.CmdClause, &globals)
	scalyrUpdate := scalyr.NewUpdateCommand(scalyrRoot.CmdClause, &globals)
	scalyrDelete := scalyr.NewDeleteCommand(scalyrRoot.CmdClause, &globals)

	logglyRoot := loggly.NewRootCommand(loggingRoot.CmdClause, &globals)
	logglyCreate := loggly.NewCreateCommand(logglyRoot.CmdClause, &globals)
	logglyList := loggly.NewListCommand(logglyRoot.CmdClause, &globals)
	logglyDescribe := loggly.NewDescribeCommand(logglyRoot.CmdClause, &globals)
	logglyUpdate := loggly.NewUpdateCommand(logglyRoot.CmdClause, &globals)
	logglyDelete := loggly.NewDeleteCommand(logglyRoot.CmdClause, &globals)

	honeycombRoot := honeycomb.NewRootCommand(loggingRoot.CmdClause, &globals)
	honeycombCreate := honeycomb.NewCreateCommand(honeycombRoot.CmdClause, &globals)
	honeycombList := honeycomb.NewListCommand(honeycombRoot.CmdClause, &globals)
	honeycombDescribe := honeycomb.NewDescribeCommand(honeycombRoot.CmdClause, &globals)
	honeycombUpdate := honeycomb.NewUpdateCommand(honeycombRoot.CmdClause, &globals)
	honeycombDelete := honeycomb.NewDeleteCommand(honeycombRoot.CmdClause, &globals)

	herokuRoot := heroku.NewRootCommand(loggingRoot.CmdClause, &globals)
	herokuCreate := heroku.NewCreateCommand(herokuRoot.CmdClause, &globals)
	herokuList := heroku.NewListCommand(herokuRoot.CmdClause, &globals)
	herokuDescribe := heroku.NewDescribeCommand(herokuRoot.CmdClause, &globals)
	herokuUpdate := heroku.NewUpdateCommand(herokuRoot.CmdClause, &globals)
	herokuDelete := heroku.NewDeleteCommand(herokuRoot.CmdClause, &globals)

	sftpRoot := sftp.NewRootCommand(loggingRoot.CmdClause, &globals)
	sftpCreate := sftp.NewCreateCommand(sftpRoot.CmdClause, &globals)
	sftpList := sftp.NewListCommand(sftpRoot.CmdClause, &globals)
	sftpDescribe := sftp.NewDescribeCommand(sftpRoot.CmdClause, &globals)
	sftpUpdate := sftp.NewUpdateCommand(sftpRoot.CmdClause, &globals)
	sftpDelete := sftp.NewDeleteCommand(sftpRoot.CmdClause, &globals)

	logshuttleRoot := logshuttle.NewRootCommand(loggingRoot.CmdClause, &globals)
	logshuttleCreate := logshuttle.NewCreateCommand(logshuttleRoot.CmdClause, &globals)
	logshuttleList := logshuttle.NewListCommand(logshuttleRoot.CmdClause, &globals)
	logshuttleDescribe := logshuttle.NewDescribeCommand(logshuttleRoot.CmdClause, &globals)
	logshuttleUpdate := logshuttle.NewUpdateCommand(logshuttleRoot.CmdClause, &globals)
	logshuttleDelete := logshuttle.NewDeleteCommand(logshuttleRoot.CmdClause, &globals)

	cloudfilesRoot := cloudfiles.NewRootCommand(loggingRoot.CmdClause, &globals)
	cloudfilesCreate := cloudfiles.NewCreateCommand(cloudfilesRoot.CmdClause, &globals)
	cloudfilesList := cloudfiles.NewListCommand(cloudfilesRoot.CmdClause, &globals)
	cloudfilesDescribe := cloudfiles.NewDescribeCommand(cloudfilesRoot.CmdClause, &globals)
	cloudfilesUpdate := cloudfiles.NewUpdateCommand(cloudfilesRoot.CmdClause, &globals)
	cloudfilesDelete := cloudfiles.NewDeleteCommand(cloudfilesRoot.CmdClause, &globals)

	digitaloceanRoot := digitalocean.NewRootCommand(loggingRoot.CmdClause, &globals)
	digitaloceanCreate := digitalocean.NewCreateCommand(digitaloceanRoot.CmdClause, &globals)
	digitaloceanList := digitalocean.NewListCommand(digitaloceanRoot.CmdClause, &globals)
	digitaloceanDescribe := digitalocean.NewDescribeCommand(digitaloceanRoot.CmdClause, &globals)
	digitaloceanUpdate := digitalocean.NewUpdateCommand(digitaloceanRoot.CmdClause, &globals)
	digitaloceanDelete := digitalocean.NewDeleteCommand(digitaloceanRoot.CmdClause, &globals)

	elasticsearchRoot := elasticsearch.NewRootCommand(loggingRoot.CmdClause, &globals)
	elasticsearchCreate := elasticsearch.NewCreateCommand(elasticsearchRoot.CmdClause, &globals)
	elasticsearchList := elasticsearch.NewListCommand(elasticsearchRoot.CmdClause, &globals)
	elasticsearchDescribe := elasticsearch.NewDescribeCommand(elasticsearchRoot.CmdClause, &globals)
	elasticsearchUpdate := elasticsearch.NewUpdateCommand(elasticsearchRoot.CmdClause, &globals)
	elasticsearchDelete := elasticsearch.NewDeleteCommand(elasticsearchRoot.CmdClause, &globals)

	azureblobRoot := azureblob.NewRootCommand(loggingRoot.CmdClause, &globals)
	azureblobCreate := azureblob.NewCreateCommand(azureblobRoot.CmdClause, &globals)
	azureblobList := azureblob.NewListCommand(azureblobRoot.CmdClause, &globals)
	azureblobDescribe := azureblob.NewDescribeCommand(azureblobRoot.CmdClause, &globals)
	azureblobUpdate := azureblob.NewUpdateCommand(azureblobRoot.CmdClause, &globals)
	azureblobDelete := azureblob.NewDeleteCommand(azureblobRoot.CmdClause, &globals)

	datadogRoot := datadog.NewRootCommand(loggingRoot.CmdClause, &globals)
	datadogCreate := datadog.NewCreateCommand(datadogRoot.CmdClause, &globals)
	datadogList := datadog.NewListCommand(datadogRoot.CmdClause, &globals)
	datadogDescribe := datadog.NewDescribeCommand(datadogRoot.CmdClause, &globals)
	datadogUpdate := datadog.NewUpdateCommand(datadogRoot.CmdClause, &globals)
	datadogDelete := datadog.NewDeleteCommand(datadogRoot.CmdClause, &globals)

	httpsRoot := https.NewRootCommand(loggingRoot.CmdClause, &globals)
	httpsCreate := https.NewCreateCommand(httpsRoot.CmdClause, &globals)
	httpsList := https.NewListCommand(httpsRoot.CmdClause, &globals)
	httpsDescribe := https.NewDescribeCommand(httpsRoot.CmdClause, &globals)
	httpsUpdate := https.NewUpdateCommand(httpsRoot.CmdClause, &globals)
	httpsDelete := https.NewDeleteCommand(httpsRoot.CmdClause, &globals)

	kafkaRoot := kafka.NewRootCommand(loggingRoot.CmdClause, &globals)
	kafkaCreate := kafka.NewCreateCommand(kafkaRoot.CmdClause, &globals)
	kafkaList := kafka.NewListCommand(kafkaRoot.CmdClause, &globals)
	kafkaDescribe := kafka.NewDescribeCommand(kafkaRoot.CmdClause, &globals)
	kafkaUpdate := kafka.NewUpdateCommand(kafkaRoot.CmdClause, &globals)
	kafkaDelete := kafka.NewDeleteCommand(kafkaRoot.CmdClause, &globals)

	googlepubsubRoot := googlepubsub.NewRootCommand(loggingRoot.CmdClause, &globals)
	googlepubsubCreate := googlepubsub.NewCreateCommand(googlepubsubRoot.CmdClause, &globals)
	googlepubsubList := googlepubsub.NewListCommand(googlepubsubRoot.CmdClause, &globals)
	googlepubsubDescribe := googlepubsub.NewDescribeCommand(googlepubsubRoot.CmdClause, &globals)
	googlepubsubUpdate := googlepubsub.NewUpdateCommand(googlepubsubRoot.CmdClause, &globals)
	googlepubsubDelete := googlepubsub.NewDeleteCommand(googlepubsubRoot.CmdClause, &globals)

	openstackRoot := openstack.NewRootCommand(loggingRoot.CmdClause, &globals)
	openstackCreate := openstack.NewCreateCommand(openstackRoot.CmdClause, &globals)
	openstackList := openstack.NewListCommand(openstackRoot.CmdClause, &globals)
	openstackDescribe := openstack.NewDescribeCommand(openstackRoot.CmdClause, &globals)
	openstackUpdate := openstack.NewUpdateCommand(openstackRoot.CmdClause, &globals)
	openstackDelete := openstack.NewDeleteCommand(openstackRoot.CmdClause, &globals)

	logsRoot := logs.NewRootCommand(app, &globals)
	logsTail := logs.NewTailCommand(logsRoot.CmdClause, &globals)

	statsRoot := stats.NewRootCommand(app, &globals)
	statsRegions := stats.NewRegionsCommand(statsRoot.CmdClause, &globals)
	statsHistorical := stats.NewHistoricalCommand(statsRoot.CmdClause, &globals)
	statsRealtime := stats.NewRealtimeCommand(statsRoot.CmdClause, &globals)

	vclRoot := vcl.NewRootCommand(app, &globals)

	vclCustomRoot := custom.NewRootCommand(vclRoot.CmdClause, &globals)
	vclCustomCreate := custom.NewCreateCommand(vclCustomRoot.CmdClause, &globals)
	vclCustomDelete := custom.NewDeleteCommand(vclCustomRoot.CmdClause, &globals)
	vclCustomDescribe := custom.NewDescribeCommand(vclCustomRoot.CmdClause, &globals)
	vclCustomList := custom.NewListCommand(vclCustomRoot.CmdClause, &globals)
	vclCustomUpdate := custom.NewUpdateCommand(vclCustomRoot.CmdClause, &globals)

	vclSnippetRoot := snippet.NewRootCommand(vclRoot.CmdClause, &globals)
	vclSnippetCreate := snippet.NewCreateCommand(vclSnippetRoot.CmdClause, &globals)
	vclSnippetDelete := snippet.NewDeleteCommand(vclSnippetRoot.CmdClause, &globals)
	vclSnippetDescribe := snippet.NewDescribeCommand(vclSnippetRoot.CmdClause, &globals)
	vclSnippetList := snippet.NewListCommand(vclSnippetRoot.CmdClause, &globals)
	vclSnippetUpdate := snippet.NewUpdateCommand(vclSnippetRoot.CmdClause, &globals)

	commands := []cmd.Command{
		configureRoot,
		whoamiRoot,
		versionRoot,
		updateRoot,
		ipRoot,
		popRoot,
		purgeRoot,

		serviceRoot,
		serviceCreate,
		serviceList,
		serviceDescribe,
		serviceUpdate,
		serviceDelete,
		serviceSearch,

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
		computePublish,
		computePack,
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

		dictionaryRoot,
		dictionaryCreate,
		dictionaryDescribe,
		dictionaryDelete,
		dictionaryList,
		dictionaryUpdate,

		dictionaryItemRoot,
		dictionaryItemList,
		dictionaryItemDescribe,
		dictionaryItemCreate,
		dictionaryItemUpdate,
		dictionaryItemDelete,
		dictionaryItemBatchModify,

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

		kinesisRoot,
		kinesisCreate,
		kinesisList,
		kinesisDescribe,
		kinesisUpdate,
		kinesisDelete,

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

		splunkRoot,
		splunkCreate,
		splunkList,
		splunkDescribe,
		splunkUpdate,
		splunkDelete,

		scalyrRoot,
		scalyrCreate,
		scalyrList,
		scalyrDescribe,
		scalyrUpdate,
		scalyrDelete,

		logglyRoot,
		logglyCreate,
		logglyList,
		logglyDescribe,
		logglyUpdate,
		logglyDelete,

		honeycombRoot,
		honeycombCreate,
		honeycombList,
		honeycombDescribe,
		honeycombUpdate,
		honeycombDelete,

		herokuRoot,
		herokuCreate,
		herokuList,
		herokuDescribe,
		herokuUpdate,
		herokuDelete,

		sftpRoot,
		sftpCreate,
		sftpList,
		sftpDescribe,
		sftpUpdate,
		sftpDelete,

		logshuttleRoot,
		logshuttleCreate,
		logshuttleList,
		logshuttleDescribe,
		logshuttleUpdate,
		logshuttleDelete,

		cloudfilesRoot,
		cloudfilesCreate,
		cloudfilesList,
		cloudfilesDescribe,
		cloudfilesUpdate,
		cloudfilesDelete,

		digitaloceanRoot,
		digitaloceanCreate,
		digitaloceanList,
		digitaloceanDescribe,
		digitaloceanUpdate,
		digitaloceanDelete,

		elasticsearchRoot,
		elasticsearchCreate,
		elasticsearchList,
		elasticsearchDescribe,
		elasticsearchUpdate,
		elasticsearchDelete,

		azureblobRoot,
		azureblobCreate,
		azureblobList,
		azureblobDescribe,
		azureblobUpdate,
		azureblobDelete,

		datadogRoot,
		datadogCreate,
		datadogList,
		datadogDescribe,
		datadogUpdate,
		datadogDelete,

		httpsRoot,
		httpsCreate,
		httpsList,
		httpsDescribe,
		httpsUpdate,
		httpsDelete,

		kafkaRoot,
		kafkaCreate,
		kafkaList,
		kafkaDescribe,
		kafkaUpdate,
		kafkaDelete,

		googlepubsubRoot,
		googlepubsubCreate,
		googlepubsubList,
		googlepubsubDescribe,
		googlepubsubUpdate,
		googlepubsubDelete,

		openstackRoot,
		openstackCreate,
		openstackList,
		openstackDescribe,
		openstackUpdate,
		openstackDelete,

		logsRoot,
		logsTail,

		statsRoot,
		statsRegions,
		statsHistorical,
		statsRealtime,

		vclRoot,

		vclCustomRoot,
		vclCustomCreate,
		vclCustomDelete,
		vclCustomDescribe,
		vclCustomList,
		vclCustomUpdate,

		vclSnippetRoot,
		vclSnippetCreate,
		vclSnippetDelete,
		vclSnippetDescribe,
		vclSnippetList,
		vclSnippetUpdate,
	}

	// Handle parse errors and display contextal usage if possible. Due to bugs
	// and an obession for lots of output side-effects in the kingpin.Parse
	// logic, we suppress it from writing any usage or errors to the writer by
	// swapping the writer with a no-op and then restoring the real writer
	// afterwards. This ensures usage text is only written once to the writer
	// and gives us greater control over our error formatting.
	app.Writers(io.Discard, io.Discard)
	name, err := app.Parse(opts.Args)
	if err != nil && !argsIsHelpJSON(opts.Args) { // Ignore error if `help --format json`
		globals.ErrLog.Add(err)
		usage := Usage(opts.Args, app, opts.Stdout, io.Discard)
		return errors.RemediationError{Prefix: usage, Inner: fmt.Errorf("error parsing arguments: %w", err)}
	}
	if ctx, _ := app.ParseContext(opts.Args); contextHasHelpFlag(ctx) {
		usage := Usage(opts.Args, app, opts.Stdout, io.Discard)
		return errors.RemediationError{Prefix: usage}
	}
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

	command, found := cmd.Select(name, commands)
	if !found {
		usage := Usage(opts.Args, app, opts.Stdout, io.Discard)
		return errors.RemediationError{Prefix: usage, Inner: fmt.Errorf("command not found")}
	}

	if opts.VersionerCLI != nil && name != "update" && !version.IsPreRelease(revision.AppVersion) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel() // push cancel on the defer stack first...
		f := update.CheckAsync(ctx, opts.ConfigFile, opts.ConfigPath, revision.AppVersion, opts.VersionerCLI, opts.Stdin, opts.Stdout)
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
