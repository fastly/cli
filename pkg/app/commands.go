package app

import (
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/acl"
	"github.com/fastly/cli/pkg/commands/aclentry"
	"github.com/fastly/cli/pkg/commands/authtoken"
	"github.com/fastly/cli/pkg/commands/backend"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/commands/config"
	"github.com/fastly/cli/pkg/commands/dictionary"
	"github.com/fastly/cli/pkg/commands/dictionaryitem"
	"github.com/fastly/cli/pkg/commands/domain"
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
	"github.com/fastly/cli/pkg/commands/logtail"
	"github.com/fastly/cli/pkg/commands/pop"
	"github.com/fastly/cli/pkg/commands/profile"
	"github.com/fastly/cli/pkg/commands/purge"
	"github.com/fastly/cli/pkg/commands/service"
	"github.com/fastly/cli/pkg/commands/serviceversion"
	"github.com/fastly/cli/pkg/commands/shellcomplete"
	"github.com/fastly/cli/pkg/commands/stats"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/commands/user"
	"github.com/fastly/cli/pkg/commands/vcl"
	"github.com/fastly/cli/pkg/commands/vcl/custom"
	"github.com/fastly/cli/pkg/commands/vcl/snippet"
	"github.com/fastly/cli/pkg/commands/version"
	"github.com/fastly/cli/pkg/commands/whoami"
	cfg "github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/kingpin"
)

func defineCommands(
	app *kingpin.Application,
	globals *cfg.Data,
	data manifest.Data,
	opts RunOpts,
) []cmd.Command {
	shellcompleteCmdRoot := shellcomplete.NewRootCommand(app, globals)
	aclCmdRoot := acl.NewRootCommand(app, globals)
	aclCreate := acl.NewCreateCommand(aclCmdRoot.CmdClause, globals, data)
	aclDelete := acl.NewDeleteCommand(aclCmdRoot.CmdClause, globals, data)
	aclDescribe := acl.NewDescribeCommand(aclCmdRoot.CmdClause, globals, data)
	aclList := acl.NewListCommand(aclCmdRoot.CmdClause, globals, data)
	aclUpdate := acl.NewUpdateCommand(aclCmdRoot.CmdClause, globals, data)
	aclEntryCmdRoot := aclentry.NewRootCommand(app, globals)
	aclEntryCreate := aclentry.NewCreateCommand(aclEntryCmdRoot.CmdClause, globals, data)
	aclEntryDelete := aclentry.NewDeleteCommand(aclEntryCmdRoot.CmdClause, globals, data)
	aclEntryDescribe := aclentry.NewDescribeCommand(aclEntryCmdRoot.CmdClause, globals, data)
	aclEntryList := aclentry.NewListCommand(aclEntryCmdRoot.CmdClause, globals, data)
	aclEntryUpdate := aclentry.NewUpdateCommand(aclEntryCmdRoot.CmdClause, globals, data)
	authtokenCmdRoot := authtoken.NewRootCommand(app, globals)
	authtokenCreate := authtoken.NewCreateCommand(authtokenCmdRoot.CmdClause, globals, data)
	authtokenDelete := authtoken.NewDeleteCommand(authtokenCmdRoot.CmdClause, globals, data)
	authtokenDescribe := authtoken.NewDescribeCommand(authtokenCmdRoot.CmdClause, globals, data)
	authtokenList := authtoken.NewListCommand(authtokenCmdRoot.CmdClause, globals, data)
	backendCmdRoot := backend.NewRootCommand(app, globals)
	backendCreate := backend.NewCreateCommand(backendCmdRoot.CmdClause, globals, data)
	backendDelete := backend.NewDeleteCommand(backendCmdRoot.CmdClause, globals, data)
	backendDescribe := backend.NewDescribeCommand(backendCmdRoot.CmdClause, globals, data)
	backendList := backend.NewListCommand(backendCmdRoot.CmdClause, globals, data)
	backendUpdate := backend.NewUpdateCommand(backendCmdRoot.CmdClause, globals, data)
	computeCmdRoot := compute.NewRootCommand(app, globals)
	computeBuild := compute.NewBuildCommand(computeCmdRoot.CmdClause, globals, data)
	computeDeploy := compute.NewDeployCommand(computeCmdRoot.CmdClause, globals, data)
	computeInit := compute.NewInitCommand(computeCmdRoot.CmdClause, globals, data)
	computePack := compute.NewPackCommand(computeCmdRoot.CmdClause, globals, data)
	computePublish := compute.NewPublishCommand(computeCmdRoot.CmdClause, globals, computeBuild, computeDeploy, data)
	computeServe := compute.NewServeCommand(computeCmdRoot.CmdClause, globals, computeBuild, opts.Versioners.Viceroy, data)
	computeUpdate := compute.NewUpdateCommand(computeCmdRoot.CmdClause, globals, data)
	computeValidate := compute.NewValidateCommand(computeCmdRoot.CmdClause, globals)
	configCmdRoot := config.NewRootCommand(app, globals)
	dictionaryCmdRoot := dictionary.NewRootCommand(app, globals)
	dictionaryCreate := dictionary.NewCreateCommand(dictionaryCmdRoot.CmdClause, globals, data)
	dictionaryDelete := dictionary.NewDeleteCommand(dictionaryCmdRoot.CmdClause, globals, data)
	dictionaryDescribe := dictionary.NewDescribeCommand(dictionaryCmdRoot.CmdClause, globals, data)
	dictionaryItemCmdRoot := dictionaryitem.NewRootCommand(app, globals)
	dictionaryItemCreate := dictionaryitem.NewCreateCommand(dictionaryItemCmdRoot.CmdClause, globals, data)
	dictionaryItemDelete := dictionaryitem.NewDeleteCommand(dictionaryItemCmdRoot.CmdClause, globals, data)
	dictionaryItemDescribe := dictionaryitem.NewDescribeCommand(dictionaryItemCmdRoot.CmdClause, globals, data)
	dictionaryItemList := dictionaryitem.NewListCommand(dictionaryItemCmdRoot.CmdClause, globals, data)
	dictionaryItemUpdate := dictionaryitem.NewUpdateCommand(dictionaryItemCmdRoot.CmdClause, globals, data)
	dictionaryList := dictionary.NewListCommand(dictionaryCmdRoot.CmdClause, globals, data)
	dictionaryUpdate := dictionary.NewUpdateCommand(dictionaryCmdRoot.CmdClause, globals, data)
	domainCmdRoot := domain.NewRootCommand(app, globals)
	domainCreate := domain.NewCreateCommand(domainCmdRoot.CmdClause, globals, data)
	domainDelete := domain.NewDeleteCommand(domainCmdRoot.CmdClause, globals, data)
	domainDescribe := domain.NewDescribeCommand(domainCmdRoot.CmdClause, globals, data)
	domainList := domain.NewListCommand(domainCmdRoot.CmdClause, globals, data)
	domainUpdate := domain.NewUpdateCommand(domainCmdRoot.CmdClause, globals, data)
	domainValidate := domain.NewValidateCommand(domainCmdRoot.CmdClause, globals, data)
	healthcheckCmdRoot := healthcheck.NewRootCommand(app, globals)
	healthcheckCreate := healthcheck.NewCreateCommand(healthcheckCmdRoot.CmdClause, globals, data)
	healthcheckDelete := healthcheck.NewDeleteCommand(healthcheckCmdRoot.CmdClause, globals, data)
	healthcheckDescribe := healthcheck.NewDescribeCommand(healthcheckCmdRoot.CmdClause, globals, data)
	healthcheckList := healthcheck.NewListCommand(healthcheckCmdRoot.CmdClause, globals, data)
	healthcheckUpdate := healthcheck.NewUpdateCommand(healthcheckCmdRoot.CmdClause, globals, data)
	ipCmdRoot := ip.NewRootCommand(app, globals)
	logtailCmdRoot := logtail.NewRootCommand(app, globals, data)
	loggingCmdRoot := logging.NewRootCommand(app, globals)
	loggingAzureblobCmdRoot := azureblob.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingAzureblobCreate := azureblob.NewCreateCommand(loggingAzureblobCmdRoot.CmdClause, globals, data)
	loggingAzureblobDelete := azureblob.NewDeleteCommand(loggingAzureblobCmdRoot.CmdClause, globals, data)
	loggingAzureblobDescribe := azureblob.NewDescribeCommand(loggingAzureblobCmdRoot.CmdClause, globals, data)
	loggingAzureblobList := azureblob.NewListCommand(loggingAzureblobCmdRoot.CmdClause, globals, data)
	loggingAzureblobUpdate := azureblob.NewUpdateCommand(loggingAzureblobCmdRoot.CmdClause, globals, data)
	loggingBigQueryCmdRoot := bigquery.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingBigQueryCreate := bigquery.NewCreateCommand(loggingBigQueryCmdRoot.CmdClause, globals, data)
	loggingBigQueryDelete := bigquery.NewDeleteCommand(loggingBigQueryCmdRoot.CmdClause, globals, data)
	loggingBigQueryDescribe := bigquery.NewDescribeCommand(loggingBigQueryCmdRoot.CmdClause, globals, data)
	loggingBigQueryList := bigquery.NewListCommand(loggingBigQueryCmdRoot.CmdClause, globals, data)
	loggingBigQueryUpdate := bigquery.NewUpdateCommand(loggingBigQueryCmdRoot.CmdClause, globals, data)
	loggingCloudfilesCmdRoot := cloudfiles.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingCloudfilesCreate := cloudfiles.NewCreateCommand(loggingCloudfilesCmdRoot.CmdClause, globals, data)
	loggingCloudfilesDelete := cloudfiles.NewDeleteCommand(loggingCloudfilesCmdRoot.CmdClause, globals, data)
	loggingCloudfilesDescribe := cloudfiles.NewDescribeCommand(loggingCloudfilesCmdRoot.CmdClause, globals, data)
	loggingCloudfilesList := cloudfiles.NewListCommand(loggingCloudfilesCmdRoot.CmdClause, globals, data)
	loggingCloudfilesUpdate := cloudfiles.NewUpdateCommand(loggingCloudfilesCmdRoot.CmdClause, globals, data)
	loggingDatadogCmdRoot := datadog.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingDatadogCreate := datadog.NewCreateCommand(loggingDatadogCmdRoot.CmdClause, globals, data)
	loggingDatadogDelete := datadog.NewDeleteCommand(loggingDatadogCmdRoot.CmdClause, globals, data)
	loggingDatadogDescribe := datadog.NewDescribeCommand(loggingDatadogCmdRoot.CmdClause, globals, data)
	loggingDatadogList := datadog.NewListCommand(loggingDatadogCmdRoot.CmdClause, globals, data)
	loggingDatadogUpdate := datadog.NewUpdateCommand(loggingDatadogCmdRoot.CmdClause, globals, data)
	loggingDigitaloceanCmdRoot := digitalocean.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingDigitaloceanCreate := digitalocean.NewCreateCommand(loggingDigitaloceanCmdRoot.CmdClause, globals, data)
	loggingDigitaloceanDelete := digitalocean.NewDeleteCommand(loggingDigitaloceanCmdRoot.CmdClause, globals, data)
	loggingDigitaloceanDescribe := digitalocean.NewDescribeCommand(loggingDigitaloceanCmdRoot.CmdClause, globals, data)
	loggingDigitaloceanList := digitalocean.NewListCommand(loggingDigitaloceanCmdRoot.CmdClause, globals, data)
	loggingDigitaloceanUpdate := digitalocean.NewUpdateCommand(loggingDigitaloceanCmdRoot.CmdClause, globals, data)
	loggingElasticsearchCmdRoot := elasticsearch.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingElasticsearchCreate := elasticsearch.NewCreateCommand(loggingElasticsearchCmdRoot.CmdClause, globals, data)
	loggingElasticsearchDelete := elasticsearch.NewDeleteCommand(loggingElasticsearchCmdRoot.CmdClause, globals, data)
	loggingElasticsearchDescribe := elasticsearch.NewDescribeCommand(loggingElasticsearchCmdRoot.CmdClause, globals, data)
	loggingElasticsearchList := elasticsearch.NewListCommand(loggingElasticsearchCmdRoot.CmdClause, globals, data)
	loggingElasticsearchUpdate := elasticsearch.NewUpdateCommand(loggingElasticsearchCmdRoot.CmdClause, globals, data)
	loggingFtpCmdRoot := ftp.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingFtpCreate := ftp.NewCreateCommand(loggingFtpCmdRoot.CmdClause, globals, data)
	loggingFtpDelete := ftp.NewDeleteCommand(loggingFtpCmdRoot.CmdClause, globals, data)
	loggingFtpDescribe := ftp.NewDescribeCommand(loggingFtpCmdRoot.CmdClause, globals, data)
	loggingFtpList := ftp.NewListCommand(loggingFtpCmdRoot.CmdClause, globals, data)
	loggingFtpUpdate := ftp.NewUpdateCommand(loggingFtpCmdRoot.CmdClause, globals, data)
	loggingGcsCmdRoot := gcs.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingGcsCreate := gcs.NewCreateCommand(loggingGcsCmdRoot.CmdClause, globals, data)
	loggingGcsDelete := gcs.NewDeleteCommand(loggingGcsCmdRoot.CmdClause, globals, data)
	loggingGcsDescribe := gcs.NewDescribeCommand(loggingGcsCmdRoot.CmdClause, globals, data)
	loggingGcsList := gcs.NewListCommand(loggingGcsCmdRoot.CmdClause, globals, data)
	loggingGcsUpdate := gcs.NewUpdateCommand(loggingGcsCmdRoot.CmdClause, globals, data)
	loggingGooglepubsubCmdRoot := googlepubsub.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingGooglepubsubCreate := googlepubsub.NewCreateCommand(loggingGooglepubsubCmdRoot.CmdClause, globals, data)
	loggingGooglepubsubDelete := googlepubsub.NewDeleteCommand(loggingGooglepubsubCmdRoot.CmdClause, globals, data)
	loggingGooglepubsubDescribe := googlepubsub.NewDescribeCommand(loggingGooglepubsubCmdRoot.CmdClause, globals, data)
	loggingGooglepubsubList := googlepubsub.NewListCommand(loggingGooglepubsubCmdRoot.CmdClause, globals, data)
	loggingGooglepubsubUpdate := googlepubsub.NewUpdateCommand(loggingGooglepubsubCmdRoot.CmdClause, globals, data)
	loggingHerokuCmdRoot := heroku.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingHerokuCreate := heroku.NewCreateCommand(loggingHerokuCmdRoot.CmdClause, globals, data)
	loggingHerokuDelete := heroku.NewDeleteCommand(loggingHerokuCmdRoot.CmdClause, globals, data)
	loggingHerokuDescribe := heroku.NewDescribeCommand(loggingHerokuCmdRoot.CmdClause, globals, data)
	loggingHerokuList := heroku.NewListCommand(loggingHerokuCmdRoot.CmdClause, globals, data)
	loggingHerokuUpdate := heroku.NewUpdateCommand(loggingHerokuCmdRoot.CmdClause, globals, data)
	loggingHoneycombCmdRoot := honeycomb.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingHoneycombCreate := honeycomb.NewCreateCommand(loggingHoneycombCmdRoot.CmdClause, globals, data)
	loggingHoneycombDelete := honeycomb.NewDeleteCommand(loggingHoneycombCmdRoot.CmdClause, globals, data)
	loggingHoneycombDescribe := honeycomb.NewDescribeCommand(loggingHoneycombCmdRoot.CmdClause, globals, data)
	loggingHoneycombList := honeycomb.NewListCommand(loggingHoneycombCmdRoot.CmdClause, globals, data)
	loggingHoneycombUpdate := honeycomb.NewUpdateCommand(loggingHoneycombCmdRoot.CmdClause, globals, data)
	loggingHTTPSCmdRoot := https.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingHTTPSCreate := https.NewCreateCommand(loggingHTTPSCmdRoot.CmdClause, globals, data)
	loggingHTTPSDelete := https.NewDeleteCommand(loggingHTTPSCmdRoot.CmdClause, globals, data)
	loggingHTTPSDescribe := https.NewDescribeCommand(loggingHTTPSCmdRoot.CmdClause, globals, data)
	loggingHTTPSList := https.NewListCommand(loggingHTTPSCmdRoot.CmdClause, globals, data)
	loggingHTTPSUpdate := https.NewUpdateCommand(loggingHTTPSCmdRoot.CmdClause, globals, data)
	loggingKafkaCmdRoot := kafka.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingKafkaCreate := kafka.NewCreateCommand(loggingKafkaCmdRoot.CmdClause, globals, data)
	loggingKafkaDelete := kafka.NewDeleteCommand(loggingKafkaCmdRoot.CmdClause, globals, data)
	loggingKafkaDescribe := kafka.NewDescribeCommand(loggingKafkaCmdRoot.CmdClause, globals, data)
	loggingKafkaList := kafka.NewListCommand(loggingKafkaCmdRoot.CmdClause, globals, data)
	loggingKafkaUpdate := kafka.NewUpdateCommand(loggingKafkaCmdRoot.CmdClause, globals, data)
	loggingKinesisCmdRoot := kinesis.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingKinesisCreate := kinesis.NewCreateCommand(loggingKinesisCmdRoot.CmdClause, globals, data)
	loggingKinesisDelete := kinesis.NewDeleteCommand(loggingKinesisCmdRoot.CmdClause, globals, data)
	loggingKinesisDescribe := kinesis.NewDescribeCommand(loggingKinesisCmdRoot.CmdClause, globals, data)
	loggingKinesisList := kinesis.NewListCommand(loggingKinesisCmdRoot.CmdClause, globals, data)
	loggingKinesisUpdate := kinesis.NewUpdateCommand(loggingKinesisCmdRoot.CmdClause, globals, data)
	loggingLogentriesCmdRoot := logentries.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingLogentriesCreate := logentries.NewCreateCommand(loggingLogentriesCmdRoot.CmdClause, globals, data)
	loggingLogentriesDelete := logentries.NewDeleteCommand(loggingLogentriesCmdRoot.CmdClause, globals, data)
	loggingLogentriesDescribe := logentries.NewDescribeCommand(loggingLogentriesCmdRoot.CmdClause, globals, data)
	loggingLogentriesList := logentries.NewListCommand(loggingLogentriesCmdRoot.CmdClause, globals, data)
	loggingLogentriesUpdate := logentries.NewUpdateCommand(loggingLogentriesCmdRoot.CmdClause, globals, data)
	loggingLogglyCmdRoot := loggly.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingLogglyCreate := loggly.NewCreateCommand(loggingLogglyCmdRoot.CmdClause, globals, data)
	loggingLogglyDelete := loggly.NewDeleteCommand(loggingLogglyCmdRoot.CmdClause, globals, data)
	loggingLogglyDescribe := loggly.NewDescribeCommand(loggingLogglyCmdRoot.CmdClause, globals, data)
	loggingLogglyList := loggly.NewListCommand(loggingLogglyCmdRoot.CmdClause, globals, data)
	loggingLogglyUpdate := loggly.NewUpdateCommand(loggingLogglyCmdRoot.CmdClause, globals, data)
	loggingLogshuttleCmdRoot := logshuttle.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingLogshuttleCreate := logshuttle.NewCreateCommand(loggingLogshuttleCmdRoot.CmdClause, globals, data)
	loggingLogshuttleDelete := logshuttle.NewDeleteCommand(loggingLogshuttleCmdRoot.CmdClause, globals, data)
	loggingLogshuttleDescribe := logshuttle.NewDescribeCommand(loggingLogshuttleCmdRoot.CmdClause, globals, data)
	loggingLogshuttleList := logshuttle.NewListCommand(loggingLogshuttleCmdRoot.CmdClause, globals, data)
	loggingLogshuttleUpdate := logshuttle.NewUpdateCommand(loggingLogshuttleCmdRoot.CmdClause, globals, data)
	loggingNewRelicCmdRoot := newrelic.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingNewRelicCreate := newrelic.NewCreateCommand(loggingNewRelicCmdRoot.CmdClause, globals, data)
	loggingNewRelicDelete := newrelic.NewDeleteCommand(loggingNewRelicCmdRoot.CmdClause, globals, data)
	loggingNewRelicDescribe := newrelic.NewDescribeCommand(loggingNewRelicCmdRoot.CmdClause, globals, data)
	loggingNewRelicList := newrelic.NewListCommand(loggingNewRelicCmdRoot.CmdClause, globals, data)
	loggingNewRelicUpdate := newrelic.NewUpdateCommand(loggingNewRelicCmdRoot.CmdClause, globals, data)
	loggingOpenstackCmdRoot := openstack.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingOpenstackCreate := openstack.NewCreateCommand(loggingOpenstackCmdRoot.CmdClause, globals, data)
	loggingOpenstackDelete := openstack.NewDeleteCommand(loggingOpenstackCmdRoot.CmdClause, globals, data)
	loggingOpenstackDescribe := openstack.NewDescribeCommand(loggingOpenstackCmdRoot.CmdClause, globals, data)
	loggingOpenstackList := openstack.NewListCommand(loggingOpenstackCmdRoot.CmdClause, globals, data)
	loggingOpenstackUpdate := openstack.NewUpdateCommand(loggingOpenstackCmdRoot.CmdClause, globals, data)
	loggingPapertrailCmdRoot := papertrail.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingPapertrailCreate := papertrail.NewCreateCommand(loggingPapertrailCmdRoot.CmdClause, globals, data)
	loggingPapertrailDelete := papertrail.NewDeleteCommand(loggingPapertrailCmdRoot.CmdClause, globals, data)
	loggingPapertrailDescribe := papertrail.NewDescribeCommand(loggingPapertrailCmdRoot.CmdClause, globals, data)
	loggingPapertrailList := papertrail.NewListCommand(loggingPapertrailCmdRoot.CmdClause, globals, data)
	loggingPapertrailUpdate := papertrail.NewUpdateCommand(loggingPapertrailCmdRoot.CmdClause, globals, data)
	loggingS3CmdRoot := s3.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingS3Create := s3.NewCreateCommand(loggingS3CmdRoot.CmdClause, globals, data)
	loggingS3Delete := s3.NewDeleteCommand(loggingS3CmdRoot.CmdClause, globals, data)
	loggingS3Describe := s3.NewDescribeCommand(loggingS3CmdRoot.CmdClause, globals, data)
	loggingS3List := s3.NewListCommand(loggingS3CmdRoot.CmdClause, globals, data)
	loggingS3Update := s3.NewUpdateCommand(loggingS3CmdRoot.CmdClause, globals, data)
	loggingScalyrCmdRoot := scalyr.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingScalyrCreate := scalyr.NewCreateCommand(loggingScalyrCmdRoot.CmdClause, globals, data)
	loggingScalyrDelete := scalyr.NewDeleteCommand(loggingScalyrCmdRoot.CmdClause, globals, data)
	loggingScalyrDescribe := scalyr.NewDescribeCommand(loggingScalyrCmdRoot.CmdClause, globals, data)
	loggingScalyrList := scalyr.NewListCommand(loggingScalyrCmdRoot.CmdClause, globals, data)
	loggingScalyrUpdate := scalyr.NewUpdateCommand(loggingScalyrCmdRoot.CmdClause, globals, data)
	loggingSftpCmdRoot := sftp.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingSftpCreate := sftp.NewCreateCommand(loggingSftpCmdRoot.CmdClause, globals, data)
	loggingSftpDelete := sftp.NewDeleteCommand(loggingSftpCmdRoot.CmdClause, globals, data)
	loggingSftpDescribe := sftp.NewDescribeCommand(loggingSftpCmdRoot.CmdClause, globals, data)
	loggingSftpList := sftp.NewListCommand(loggingSftpCmdRoot.CmdClause, globals, data)
	loggingSftpUpdate := sftp.NewUpdateCommand(loggingSftpCmdRoot.CmdClause, globals, data)
	loggingSplunkCmdRoot := splunk.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingSplunkCreate := splunk.NewCreateCommand(loggingSplunkCmdRoot.CmdClause, globals, data)
	loggingSplunkDelete := splunk.NewDeleteCommand(loggingSplunkCmdRoot.CmdClause, globals, data)
	loggingSplunkDescribe := splunk.NewDescribeCommand(loggingSplunkCmdRoot.CmdClause, globals, data)
	loggingSplunkList := splunk.NewListCommand(loggingSplunkCmdRoot.CmdClause, globals, data)
	loggingSplunkUpdate := splunk.NewUpdateCommand(loggingSplunkCmdRoot.CmdClause, globals, data)
	loggingSumologicCmdRoot := sumologic.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingSumologicCreate := sumologic.NewCreateCommand(loggingSumologicCmdRoot.CmdClause, globals, data)
	loggingSumologicDelete := sumologic.NewDeleteCommand(loggingSumologicCmdRoot.CmdClause, globals, data)
	loggingSumologicDescribe := sumologic.NewDescribeCommand(loggingSumologicCmdRoot.CmdClause, globals, data)
	loggingSumologicList := sumologic.NewListCommand(loggingSumologicCmdRoot.CmdClause, globals, data)
	loggingSumologicUpdate := sumologic.NewUpdateCommand(loggingSumologicCmdRoot.CmdClause, globals, data)
	loggingSyslogCmdRoot := syslog.NewRootCommand(loggingCmdRoot.CmdClause, globals)
	loggingSyslogCreate := syslog.NewCreateCommand(loggingSyslogCmdRoot.CmdClause, globals, data)
	loggingSyslogDelete := syslog.NewDeleteCommand(loggingSyslogCmdRoot.CmdClause, globals, data)
	loggingSyslogDescribe := syslog.NewDescribeCommand(loggingSyslogCmdRoot.CmdClause, globals, data)
	loggingSyslogList := syslog.NewListCommand(loggingSyslogCmdRoot.CmdClause, globals, data)
	loggingSyslogUpdate := syslog.NewUpdateCommand(loggingSyslogCmdRoot.CmdClause, globals, data)
	popCmdRoot := pop.NewRootCommand(app, globals)
	profileCmdRoot := profile.NewRootCommand(app, globals)
	profileCreate := profile.NewCreateCommand(profileCmdRoot.CmdClause, profile.APIClientFactory(opts.APIClient), globals)
	profileDelete := profile.NewDeleteCommand(profileCmdRoot.CmdClause, globals)
	profileList := profile.NewListCommand(profileCmdRoot.CmdClause, globals)
	profileSwitch := profile.NewSwitchCommand(profileCmdRoot.CmdClause, globals)
	profileToken := profile.NewTokenCommand(profileCmdRoot.CmdClause, globals)
	profileUpdate := profile.NewUpdateCommand(profileCmdRoot.CmdClause, profile.APIClientFactory(opts.APIClient), globals)
	purgeCmdRoot := purge.NewRootCommand(app, globals, data)
	serviceCmdRoot := service.NewRootCommand(app, globals)
	serviceCreate := service.NewCreateCommand(serviceCmdRoot.CmdClause, globals)
	serviceDelete := service.NewDeleteCommand(serviceCmdRoot.CmdClause, globals, data)
	serviceDescribe := service.NewDescribeCommand(serviceCmdRoot.CmdClause, globals, data)
	serviceList := service.NewListCommand(serviceCmdRoot.CmdClause, globals)
	serviceSearch := service.NewSearchCommand(serviceCmdRoot.CmdClause, globals, data)
	serviceUpdate := service.NewUpdateCommand(serviceCmdRoot.CmdClause, globals, data)
	serviceVersionCmdRoot := serviceversion.NewRootCommand(app, globals)
	serviceVersionActivate := serviceversion.NewActivateCommand(serviceVersionCmdRoot.CmdClause, globals, data)
	serviceVersionClone := serviceversion.NewCloneCommand(serviceVersionCmdRoot.CmdClause, globals, data)
	serviceVersionDeactivate := serviceversion.NewDeactivateCommand(serviceVersionCmdRoot.CmdClause, globals, data)
	serviceVersionList := serviceversion.NewListCommand(serviceVersionCmdRoot.CmdClause, globals, data)
	serviceVersionLock := serviceversion.NewLockCommand(serviceVersionCmdRoot.CmdClause, globals, data)
	serviceVersionUpdate := serviceversion.NewUpdateCommand(serviceVersionCmdRoot.CmdClause, globals, data)
	statsCmdRoot := stats.NewRootCommand(app, globals)
	statsHistorical := stats.NewHistoricalCommand(statsCmdRoot.CmdClause, globals, data)
	statsRealtime := stats.NewRealtimeCommand(statsCmdRoot.CmdClause, globals, data)
	statsRegions := stats.NewRegionsCommand(statsCmdRoot.CmdClause, globals)
	updateRoot := update.NewRootCommand(app, opts.ConfigPath, opts.Versioners.CLI, globals)
	userCmdRoot := user.NewRootCommand(app, globals)
	userCreate := user.NewCreateCommand(userCmdRoot.CmdClause, globals, data)
	userDelete := user.NewDeleteCommand(userCmdRoot.CmdClause, globals, data)
	userDescribe := user.NewDescribeCommand(userCmdRoot.CmdClause, globals, data)
	userList := user.NewListCommand(userCmdRoot.CmdClause, globals, data)
	userUpdate := user.NewUpdateCommand(userCmdRoot.CmdClause, globals, data)
	vclCmdRoot := vcl.NewRootCommand(app, globals)
	vclCustomCmdRoot := custom.NewRootCommand(vclCmdRoot.CmdClause, globals)
	vclCustomCreate := custom.NewCreateCommand(vclCustomCmdRoot.CmdClause, globals, data)
	vclCustomDelete := custom.NewDeleteCommand(vclCustomCmdRoot.CmdClause, globals, data)
	vclCustomDescribe := custom.NewDescribeCommand(vclCustomCmdRoot.CmdClause, globals, data)
	vclCustomList := custom.NewListCommand(vclCustomCmdRoot.CmdClause, globals, data)
	vclCustomUpdate := custom.NewUpdateCommand(vclCustomCmdRoot.CmdClause, globals, data)
	vclSnippetCmdRoot := snippet.NewRootCommand(vclCmdRoot.CmdClause, globals)
	vclSnippetCreate := snippet.NewCreateCommand(vclSnippetCmdRoot.CmdClause, globals, data)
	vclSnippetDelete := snippet.NewDeleteCommand(vclSnippetCmdRoot.CmdClause, globals, data)
	vclSnippetDescribe := snippet.NewDescribeCommand(vclSnippetCmdRoot.CmdClause, globals, data)
	vclSnippetList := snippet.NewListCommand(vclSnippetCmdRoot.CmdClause, globals, data)
	vclSnippetUpdate := snippet.NewUpdateCommand(vclSnippetCmdRoot.CmdClause, globals, data)
	versionCmdRoot := version.NewRootCommand(app, opts.Versioners.Viceroy)
	whoamiCmdRoot := whoami.NewRootCommand(app, globals)

	return []cmd.Command{
		shellcompleteCmdRoot,
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
		configCmdRoot,
		dictionaryCmdRoot,
		dictionaryCreate,
		dictionaryDelete,
		dictionaryDescribe,
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
		logtailCmdRoot,
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
		popCmdRoot,
		profileCmdRoot,
		profileCreate,
		profileDelete,
		profileList,
		profileSwitch,
		profileToken,
		profileUpdate,
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
}
