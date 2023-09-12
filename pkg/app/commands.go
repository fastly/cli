package app

import (
	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/acl"
	"github.com/fastly/cli/pkg/commands/aclentry"
	"github.com/fastly/cli/pkg/commands/authenticate"
	"github.com/fastly/cli/pkg/commands/authtoken"
	"github.com/fastly/cli/pkg/commands/backend"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/commands/config"
	"github.com/fastly/cli/pkg/commands/configstore"
	"github.com/fastly/cli/pkg/commands/configstoreentry"
	"github.com/fastly/cli/pkg/commands/dictionary"
	"github.com/fastly/cli/pkg/commands/dictionaryentry"
	"github.com/fastly/cli/pkg/commands/domain"
	"github.com/fastly/cli/pkg/commands/healthcheck"
	"github.com/fastly/cli/pkg/commands/ip"
	"github.com/fastly/cli/pkg/commands/kvstore"
	"github.com/fastly/cli/pkg/commands/kvstoreentry"
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
	"github.com/fastly/cli/pkg/commands/logging/loggly"
	"github.com/fastly/cli/pkg/commands/logging/logshuttle"
	"github.com/fastly/cli/pkg/commands/logging/newrelic"
	"github.com/fastly/cli/pkg/commands/logging/newrelicotlp"
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
	"github.com/fastly/cli/pkg/commands/products"
	"github.com/fastly/cli/pkg/commands/profile"
	"github.com/fastly/cli/pkg/commands/purge"
	"github.com/fastly/cli/pkg/commands/ratelimit"
	"github.com/fastly/cli/pkg/commands/resourcelink"
	"github.com/fastly/cli/pkg/commands/secretstore"
	"github.com/fastly/cli/pkg/commands/secretstoreentry"
	"github.com/fastly/cli/pkg/commands/service"
	"github.com/fastly/cli/pkg/commands/serviceauth"
	"github.com/fastly/cli/pkg/commands/serviceversion"
	"github.com/fastly/cli/pkg/commands/shellcomplete"
	"github.com/fastly/cli/pkg/commands/stats"
	tlsconfig "github.com/fastly/cli/pkg/commands/tls/config"
	tlscustom "github.com/fastly/cli/pkg/commands/tls/custom"
	tlscustomactivation "github.com/fastly/cli/pkg/commands/tls/custom/activation"
	tlscustomcertificate "github.com/fastly/cli/pkg/commands/tls/custom/certificate"
	tlscustomdomain "github.com/fastly/cli/pkg/commands/tls/custom/domain"
	tlscustomprivatekey "github.com/fastly/cli/pkg/commands/tls/custom/privatekey"
	tlsplatform "github.com/fastly/cli/pkg/commands/tls/platform"
	tlssubscription "github.com/fastly/cli/pkg/commands/tls/subscription"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/commands/user"
	"github.com/fastly/cli/pkg/commands/vcl"
	"github.com/fastly/cli/pkg/commands/vcl/condition"
	"github.com/fastly/cli/pkg/commands/vcl/custom"
	"github.com/fastly/cli/pkg/commands/vcl/snippet"
	"github.com/fastly/cli/pkg/commands/version"
	"github.com/fastly/cli/pkg/commands/whoami"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
)

func defineCommands(
	app *kingpin.Application,
	g *global.Data,
	m manifest.Data,
	opts RunOpts,
) []cmd.Command {
	shellcompleteCmdRoot := shellcomplete.NewRootCommand(app, g)
	aclCmdRoot := acl.NewRootCommand(app, g)
	aclCreate := acl.NewCreateCommand(aclCmdRoot.CmdClause, g, m)
	aclDelete := acl.NewDeleteCommand(aclCmdRoot.CmdClause, g, m)
	aclDescribe := acl.NewDescribeCommand(aclCmdRoot.CmdClause, g, m)
	aclList := acl.NewListCommand(aclCmdRoot.CmdClause, g, m)
	aclUpdate := acl.NewUpdateCommand(aclCmdRoot.CmdClause, g, m)
	aclEntryCmdRoot := aclentry.NewRootCommand(app, g)
	aclEntryCreate := aclentry.NewCreateCommand(aclEntryCmdRoot.CmdClause, g, m)
	aclEntryDelete := aclentry.NewDeleteCommand(aclEntryCmdRoot.CmdClause, g, m)
	aclEntryDescribe := aclentry.NewDescribeCommand(aclEntryCmdRoot.CmdClause, g, m)
	aclEntryList := aclentry.NewListCommand(aclEntryCmdRoot.CmdClause, g, m)
	aclEntryUpdate := aclentry.NewUpdateCommand(aclEntryCmdRoot.CmdClause, g, m)
	authenticateCmdRoot := authenticate.NewRootCommand(app, g, opts.Opener)
	authtokenCmdRoot := authtoken.NewRootCommand(app, g)
	authtokenCreate := authtoken.NewCreateCommand(authtokenCmdRoot.CmdClause, g, m)
	authtokenDelete := authtoken.NewDeleteCommand(authtokenCmdRoot.CmdClause, g, m)
	authtokenDescribe := authtoken.NewDescribeCommand(authtokenCmdRoot.CmdClause, g, m)
	authtokenList := authtoken.NewListCommand(authtokenCmdRoot.CmdClause, g, m)
	backendCmdRoot := backend.NewRootCommand(app, g)
	backendCreate := backend.NewCreateCommand(backendCmdRoot.CmdClause, g, m)
	backendDelete := backend.NewDeleteCommand(backendCmdRoot.CmdClause, g, m)
	backendDescribe := backend.NewDescribeCommand(backendCmdRoot.CmdClause, g, m)
	backendList := backend.NewListCommand(backendCmdRoot.CmdClause, g, m)
	backendUpdate := backend.NewUpdateCommand(backendCmdRoot.CmdClause, g, m)
	computeCmdRoot := compute.NewRootCommand(app, g)
	computeBuild := compute.NewBuildCommand(computeCmdRoot.CmdClause, g, opts.Versioners.WasmTools)
	computeDeploy := compute.NewDeployCommand(computeCmdRoot.CmdClause, g)
	computeHashFiles := compute.NewHashFilesCommand(computeCmdRoot.CmdClause, g, computeBuild)
	computeHashsum := compute.NewHashsumCommand(computeCmdRoot.CmdClause, g, computeBuild)
	computeInit := compute.NewInitCommand(computeCmdRoot.CmdClause, g, m)
	computeMetadata := compute.NewMetadataCommand(computeCmdRoot.CmdClause, g)
	computePack := compute.NewPackCommand(computeCmdRoot.CmdClause, g, m)
	computePublish := compute.NewPublishCommand(computeCmdRoot.CmdClause, g, computeBuild, computeDeploy)
	computeServe := compute.NewServeCommand(computeCmdRoot.CmdClause, g, computeBuild, opts.Versioners.Viceroy)
	computeUpdate := compute.NewUpdateCommand(computeCmdRoot.CmdClause, g, m)
	computeValidate := compute.NewValidateCommand(computeCmdRoot.CmdClause, g)
	configCmdRoot := config.NewRootCommand(app, g)
	configstoreCmdRoot := configstore.NewRootCommand(app, g)
	configstoreCreate := configstore.NewCreateCommand(configstoreCmdRoot.CmdClause, g, m)
	configstoreDelete := configstore.NewDeleteCommand(configstoreCmdRoot.CmdClause, g, m)
	configstoreDescribe := configstore.NewDescribeCommand(configstoreCmdRoot.CmdClause, g, m)
	configstoreList := configstore.NewListCommand(configstoreCmdRoot.CmdClause, g, m)
	configstoreListServices := configstore.NewListServicesCommand(configstoreCmdRoot.CmdClause, g, m)
	configstoreUpdate := configstore.NewUpdateCommand(configstoreCmdRoot.CmdClause, g, m)
	configstoreentryCmdRoot := configstoreentry.NewRootCommand(app, g)
	configstoreentryCreate := configstoreentry.NewCreateCommand(configstoreentryCmdRoot.CmdClause, g, m)
	configstoreentryDelete := configstoreentry.NewDeleteCommand(configstoreentryCmdRoot.CmdClause, g, m)
	configstoreentryDescribe := configstoreentry.NewDescribeCommand(configstoreentryCmdRoot.CmdClause, g, m)
	configstoreentryList := configstoreentry.NewListCommand(configstoreentryCmdRoot.CmdClause, g, m)
	configstoreentryUpdate := configstoreentry.NewUpdateCommand(configstoreentryCmdRoot.CmdClause, g, m)
	dictionaryCmdRoot := dictionary.NewRootCommand(app, g)
	dictionaryCreate := dictionary.NewCreateCommand(dictionaryCmdRoot.CmdClause, g, m)
	dictionaryDelete := dictionary.NewDeleteCommand(dictionaryCmdRoot.CmdClause, g, m)
	dictionaryDescribe := dictionary.NewDescribeCommand(dictionaryCmdRoot.CmdClause, g, m)
	dictionaryEntryCmdRoot := dictionaryentry.NewRootCommand(app, g)
	dictionaryEntryCreate := dictionaryentry.NewCreateCommand(dictionaryEntryCmdRoot.CmdClause, g, m)
	dictionaryEntryDelete := dictionaryentry.NewDeleteCommand(dictionaryEntryCmdRoot.CmdClause, g, m)
	dictionaryEntryDescribe := dictionaryentry.NewDescribeCommand(dictionaryEntryCmdRoot.CmdClause, g, m)
	dictionaryEntryList := dictionaryentry.NewListCommand(dictionaryEntryCmdRoot.CmdClause, g, m)
	dictionaryEntryUpdate := dictionaryentry.NewUpdateCommand(dictionaryEntryCmdRoot.CmdClause, g, m)
	dictionaryList := dictionary.NewListCommand(dictionaryCmdRoot.CmdClause, g, m)
	dictionaryUpdate := dictionary.NewUpdateCommand(dictionaryCmdRoot.CmdClause, g, m)
	domainCmdRoot := domain.NewRootCommand(app, g)
	domainCreate := domain.NewCreateCommand(domainCmdRoot.CmdClause, g, m)
	domainDelete := domain.NewDeleteCommand(domainCmdRoot.CmdClause, g, m)
	domainDescribe := domain.NewDescribeCommand(domainCmdRoot.CmdClause, g, m)
	domainList := domain.NewListCommand(domainCmdRoot.CmdClause, g, m)
	domainUpdate := domain.NewUpdateCommand(domainCmdRoot.CmdClause, g, m)
	domainValidate := domain.NewValidateCommand(domainCmdRoot.CmdClause, g, m)
	healthcheckCmdRoot := healthcheck.NewRootCommand(app, g)
	healthcheckCreate := healthcheck.NewCreateCommand(healthcheckCmdRoot.CmdClause, g, m)
	healthcheckDelete := healthcheck.NewDeleteCommand(healthcheckCmdRoot.CmdClause, g, m)
	healthcheckDescribe := healthcheck.NewDescribeCommand(healthcheckCmdRoot.CmdClause, g, m)
	healthcheckList := healthcheck.NewListCommand(healthcheckCmdRoot.CmdClause, g, m)
	healthcheckUpdate := healthcheck.NewUpdateCommand(healthcheckCmdRoot.CmdClause, g, m)
	ipCmdRoot := ip.NewRootCommand(app, g)
	kvstoreCmdRoot := kvstore.NewRootCommand(app, g)
	kvstoreCreate := kvstore.NewCreateCommand(kvstoreCmdRoot.CmdClause, g, m)
	kvstoreDelete := kvstore.NewDeleteCommand(kvstoreCmdRoot.CmdClause, g, m)
	kvstoreDescribe := kvstore.NewDescribeCommand(kvstoreCmdRoot.CmdClause, g, m)
	kvstoreList := kvstore.NewListCommand(kvstoreCmdRoot.CmdClause, g, m)
	kvstoreentryCmdRoot := kvstoreentry.NewRootCommand(app, g)
	kvstoreentryCreate := kvstoreentry.NewCreateCommand(kvstoreentryCmdRoot.CmdClause, g, m)
	kvstoreentryDelete := kvstoreentry.NewDeleteCommand(kvstoreentryCmdRoot.CmdClause, g, m)
	kvstoreentryDescribe := kvstoreentry.NewDescribeCommand(kvstoreentryCmdRoot.CmdClause, g, m)
	kvstoreentryList := kvstoreentry.NewListCommand(kvstoreentryCmdRoot.CmdClause, g, m)
	logtailCmdRoot := logtail.NewRootCommand(app, g, m)
	loggingCmdRoot := logging.NewRootCommand(app, g)
	loggingAzureblobCmdRoot := azureblob.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingAzureblobCreate := azureblob.NewCreateCommand(loggingAzureblobCmdRoot.CmdClause, g, m)
	loggingAzureblobDelete := azureblob.NewDeleteCommand(loggingAzureblobCmdRoot.CmdClause, g, m)
	loggingAzureblobDescribe := azureblob.NewDescribeCommand(loggingAzureblobCmdRoot.CmdClause, g, m)
	loggingAzureblobList := azureblob.NewListCommand(loggingAzureblobCmdRoot.CmdClause, g, m)
	loggingAzureblobUpdate := azureblob.NewUpdateCommand(loggingAzureblobCmdRoot.CmdClause, g, m)
	loggingBigQueryCmdRoot := bigquery.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingBigQueryCreate := bigquery.NewCreateCommand(loggingBigQueryCmdRoot.CmdClause, g, m)
	loggingBigQueryDelete := bigquery.NewDeleteCommand(loggingBigQueryCmdRoot.CmdClause, g, m)
	loggingBigQueryDescribe := bigquery.NewDescribeCommand(loggingBigQueryCmdRoot.CmdClause, g, m)
	loggingBigQueryList := bigquery.NewListCommand(loggingBigQueryCmdRoot.CmdClause, g, m)
	loggingBigQueryUpdate := bigquery.NewUpdateCommand(loggingBigQueryCmdRoot.CmdClause, g, m)
	loggingCloudfilesCmdRoot := cloudfiles.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingCloudfilesCreate := cloudfiles.NewCreateCommand(loggingCloudfilesCmdRoot.CmdClause, g, m)
	loggingCloudfilesDelete := cloudfiles.NewDeleteCommand(loggingCloudfilesCmdRoot.CmdClause, g, m)
	loggingCloudfilesDescribe := cloudfiles.NewDescribeCommand(loggingCloudfilesCmdRoot.CmdClause, g, m)
	loggingCloudfilesList := cloudfiles.NewListCommand(loggingCloudfilesCmdRoot.CmdClause, g, m)
	loggingCloudfilesUpdate := cloudfiles.NewUpdateCommand(loggingCloudfilesCmdRoot.CmdClause, g, m)
	loggingDatadogCmdRoot := datadog.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingDatadogCreate := datadog.NewCreateCommand(loggingDatadogCmdRoot.CmdClause, g, m)
	loggingDatadogDelete := datadog.NewDeleteCommand(loggingDatadogCmdRoot.CmdClause, g, m)
	loggingDatadogDescribe := datadog.NewDescribeCommand(loggingDatadogCmdRoot.CmdClause, g, m)
	loggingDatadogList := datadog.NewListCommand(loggingDatadogCmdRoot.CmdClause, g, m)
	loggingDatadogUpdate := datadog.NewUpdateCommand(loggingDatadogCmdRoot.CmdClause, g, m)
	loggingDigitaloceanCmdRoot := digitalocean.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingDigitaloceanCreate := digitalocean.NewCreateCommand(loggingDigitaloceanCmdRoot.CmdClause, g, m)
	loggingDigitaloceanDelete := digitalocean.NewDeleteCommand(loggingDigitaloceanCmdRoot.CmdClause, g, m)
	loggingDigitaloceanDescribe := digitalocean.NewDescribeCommand(loggingDigitaloceanCmdRoot.CmdClause, g, m)
	loggingDigitaloceanList := digitalocean.NewListCommand(loggingDigitaloceanCmdRoot.CmdClause, g, m)
	loggingDigitaloceanUpdate := digitalocean.NewUpdateCommand(loggingDigitaloceanCmdRoot.CmdClause, g, m)
	loggingElasticsearchCmdRoot := elasticsearch.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingElasticsearchCreate := elasticsearch.NewCreateCommand(loggingElasticsearchCmdRoot.CmdClause, g, m)
	loggingElasticsearchDelete := elasticsearch.NewDeleteCommand(loggingElasticsearchCmdRoot.CmdClause, g, m)
	loggingElasticsearchDescribe := elasticsearch.NewDescribeCommand(loggingElasticsearchCmdRoot.CmdClause, g, m)
	loggingElasticsearchList := elasticsearch.NewListCommand(loggingElasticsearchCmdRoot.CmdClause, g, m)
	loggingElasticsearchUpdate := elasticsearch.NewUpdateCommand(loggingElasticsearchCmdRoot.CmdClause, g, m)
	loggingFtpCmdRoot := ftp.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingFtpCreate := ftp.NewCreateCommand(loggingFtpCmdRoot.CmdClause, g, m)
	loggingFtpDelete := ftp.NewDeleteCommand(loggingFtpCmdRoot.CmdClause, g, m)
	loggingFtpDescribe := ftp.NewDescribeCommand(loggingFtpCmdRoot.CmdClause, g, m)
	loggingFtpList := ftp.NewListCommand(loggingFtpCmdRoot.CmdClause, g, m)
	loggingFtpUpdate := ftp.NewUpdateCommand(loggingFtpCmdRoot.CmdClause, g, m)
	loggingGcsCmdRoot := gcs.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingGcsCreate := gcs.NewCreateCommand(loggingGcsCmdRoot.CmdClause, g, m)
	loggingGcsDelete := gcs.NewDeleteCommand(loggingGcsCmdRoot.CmdClause, g, m)
	loggingGcsDescribe := gcs.NewDescribeCommand(loggingGcsCmdRoot.CmdClause, g, m)
	loggingGcsList := gcs.NewListCommand(loggingGcsCmdRoot.CmdClause, g, m)
	loggingGcsUpdate := gcs.NewUpdateCommand(loggingGcsCmdRoot.CmdClause, g, m)
	loggingGooglepubsubCmdRoot := googlepubsub.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingGooglepubsubCreate := googlepubsub.NewCreateCommand(loggingGooglepubsubCmdRoot.CmdClause, g, m)
	loggingGooglepubsubDelete := googlepubsub.NewDeleteCommand(loggingGooglepubsubCmdRoot.CmdClause, g, m)
	loggingGooglepubsubDescribe := googlepubsub.NewDescribeCommand(loggingGooglepubsubCmdRoot.CmdClause, g, m)
	loggingGooglepubsubList := googlepubsub.NewListCommand(loggingGooglepubsubCmdRoot.CmdClause, g, m)
	loggingGooglepubsubUpdate := googlepubsub.NewUpdateCommand(loggingGooglepubsubCmdRoot.CmdClause, g, m)
	loggingHerokuCmdRoot := heroku.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingHerokuCreate := heroku.NewCreateCommand(loggingHerokuCmdRoot.CmdClause, g, m)
	loggingHerokuDelete := heroku.NewDeleteCommand(loggingHerokuCmdRoot.CmdClause, g, m)
	loggingHerokuDescribe := heroku.NewDescribeCommand(loggingHerokuCmdRoot.CmdClause, g, m)
	loggingHerokuList := heroku.NewListCommand(loggingHerokuCmdRoot.CmdClause, g, m)
	loggingHerokuUpdate := heroku.NewUpdateCommand(loggingHerokuCmdRoot.CmdClause, g, m)
	loggingHoneycombCmdRoot := honeycomb.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingHoneycombCreate := honeycomb.NewCreateCommand(loggingHoneycombCmdRoot.CmdClause, g, m)
	loggingHoneycombDelete := honeycomb.NewDeleteCommand(loggingHoneycombCmdRoot.CmdClause, g, m)
	loggingHoneycombDescribe := honeycomb.NewDescribeCommand(loggingHoneycombCmdRoot.CmdClause, g, m)
	loggingHoneycombList := honeycomb.NewListCommand(loggingHoneycombCmdRoot.CmdClause, g, m)
	loggingHoneycombUpdate := honeycomb.NewUpdateCommand(loggingHoneycombCmdRoot.CmdClause, g, m)
	loggingHTTPSCmdRoot := https.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingHTTPSCreate := https.NewCreateCommand(loggingHTTPSCmdRoot.CmdClause, g, m)
	loggingHTTPSDelete := https.NewDeleteCommand(loggingHTTPSCmdRoot.CmdClause, g, m)
	loggingHTTPSDescribe := https.NewDescribeCommand(loggingHTTPSCmdRoot.CmdClause, g, m)
	loggingHTTPSList := https.NewListCommand(loggingHTTPSCmdRoot.CmdClause, g, m)
	loggingHTTPSUpdate := https.NewUpdateCommand(loggingHTTPSCmdRoot.CmdClause, g, m)
	loggingKafkaCmdRoot := kafka.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingKafkaCreate := kafka.NewCreateCommand(loggingKafkaCmdRoot.CmdClause, g, m)
	loggingKafkaDelete := kafka.NewDeleteCommand(loggingKafkaCmdRoot.CmdClause, g, m)
	loggingKafkaDescribe := kafka.NewDescribeCommand(loggingKafkaCmdRoot.CmdClause, g, m)
	loggingKafkaList := kafka.NewListCommand(loggingKafkaCmdRoot.CmdClause, g, m)
	loggingKafkaUpdate := kafka.NewUpdateCommand(loggingKafkaCmdRoot.CmdClause, g, m)
	loggingKinesisCmdRoot := kinesis.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingKinesisCreate := kinesis.NewCreateCommand(loggingKinesisCmdRoot.CmdClause, g, m)
	loggingKinesisDelete := kinesis.NewDeleteCommand(loggingKinesisCmdRoot.CmdClause, g, m)
	loggingKinesisDescribe := kinesis.NewDescribeCommand(loggingKinesisCmdRoot.CmdClause, g, m)
	loggingKinesisList := kinesis.NewListCommand(loggingKinesisCmdRoot.CmdClause, g, m)
	loggingKinesisUpdate := kinesis.NewUpdateCommand(loggingKinesisCmdRoot.CmdClause, g, m)
	loggingLogglyCmdRoot := loggly.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingLogglyCreate := loggly.NewCreateCommand(loggingLogglyCmdRoot.CmdClause, g, m)
	loggingLogglyDelete := loggly.NewDeleteCommand(loggingLogglyCmdRoot.CmdClause, g, m)
	loggingLogglyDescribe := loggly.NewDescribeCommand(loggingLogglyCmdRoot.CmdClause, g, m)
	loggingLogglyList := loggly.NewListCommand(loggingLogglyCmdRoot.CmdClause, g, m)
	loggingLogglyUpdate := loggly.NewUpdateCommand(loggingLogglyCmdRoot.CmdClause, g, m)
	loggingLogshuttleCmdRoot := logshuttle.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingLogshuttleCreate := logshuttle.NewCreateCommand(loggingLogshuttleCmdRoot.CmdClause, g, m)
	loggingLogshuttleDelete := logshuttle.NewDeleteCommand(loggingLogshuttleCmdRoot.CmdClause, g, m)
	loggingLogshuttleDescribe := logshuttle.NewDescribeCommand(loggingLogshuttleCmdRoot.CmdClause, g, m)
	loggingLogshuttleList := logshuttle.NewListCommand(loggingLogshuttleCmdRoot.CmdClause, g, m)
	loggingLogshuttleUpdate := logshuttle.NewUpdateCommand(loggingLogshuttleCmdRoot.CmdClause, g, m)
	loggingNewRelicCmdRoot := newrelic.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingNewRelicCreate := newrelic.NewCreateCommand(loggingNewRelicCmdRoot.CmdClause, g, m)
	loggingNewRelicDelete := newrelic.NewDeleteCommand(loggingNewRelicCmdRoot.CmdClause, g, m)
	loggingNewRelicDescribe := newrelic.NewDescribeCommand(loggingNewRelicCmdRoot.CmdClause, g, m)
	loggingNewRelicList := newrelic.NewListCommand(loggingNewRelicCmdRoot.CmdClause, g, m)
	loggingNewRelicUpdate := newrelic.NewUpdateCommand(loggingNewRelicCmdRoot.CmdClause, g, m)
	loggingNewRelicOTLPCmdRoot := newrelicotlp.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingNewRelicOTLPCreate := newrelicotlp.NewCreateCommand(loggingNewRelicOTLPCmdRoot.CmdClause, g, m)
	loggingNewRelicOTLPDelete := newrelicotlp.NewDeleteCommand(loggingNewRelicOTLPCmdRoot.CmdClause, g, m)
	loggingNewRelicOTLPDescribe := newrelicotlp.NewDescribeCommand(loggingNewRelicOTLPCmdRoot.CmdClause, g, m)
	loggingNewRelicOTLPList := newrelicotlp.NewListCommand(loggingNewRelicOTLPCmdRoot.CmdClause, g, m)
	loggingNewRelicOTLPUpdate := newrelicotlp.NewUpdateCommand(loggingNewRelicOTLPCmdRoot.CmdClause, g, m)
	loggingOpenstackCmdRoot := openstack.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingOpenstackCreate := openstack.NewCreateCommand(loggingOpenstackCmdRoot.CmdClause, g, m)
	loggingOpenstackDelete := openstack.NewDeleteCommand(loggingOpenstackCmdRoot.CmdClause, g, m)
	loggingOpenstackDescribe := openstack.NewDescribeCommand(loggingOpenstackCmdRoot.CmdClause, g, m)
	loggingOpenstackList := openstack.NewListCommand(loggingOpenstackCmdRoot.CmdClause, g, m)
	loggingOpenstackUpdate := openstack.NewUpdateCommand(loggingOpenstackCmdRoot.CmdClause, g, m)
	loggingPapertrailCmdRoot := papertrail.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingPapertrailCreate := papertrail.NewCreateCommand(loggingPapertrailCmdRoot.CmdClause, g, m)
	loggingPapertrailDelete := papertrail.NewDeleteCommand(loggingPapertrailCmdRoot.CmdClause, g, m)
	loggingPapertrailDescribe := papertrail.NewDescribeCommand(loggingPapertrailCmdRoot.CmdClause, g, m)
	loggingPapertrailList := papertrail.NewListCommand(loggingPapertrailCmdRoot.CmdClause, g, m)
	loggingPapertrailUpdate := papertrail.NewUpdateCommand(loggingPapertrailCmdRoot.CmdClause, g, m)
	loggingS3CmdRoot := s3.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingS3Create := s3.NewCreateCommand(loggingS3CmdRoot.CmdClause, g, m)
	loggingS3Delete := s3.NewDeleteCommand(loggingS3CmdRoot.CmdClause, g, m)
	loggingS3Describe := s3.NewDescribeCommand(loggingS3CmdRoot.CmdClause, g, m)
	loggingS3List := s3.NewListCommand(loggingS3CmdRoot.CmdClause, g, m)
	loggingS3Update := s3.NewUpdateCommand(loggingS3CmdRoot.CmdClause, g, m)
	loggingScalyrCmdRoot := scalyr.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingScalyrCreate := scalyr.NewCreateCommand(loggingScalyrCmdRoot.CmdClause, g, m)
	loggingScalyrDelete := scalyr.NewDeleteCommand(loggingScalyrCmdRoot.CmdClause, g, m)
	loggingScalyrDescribe := scalyr.NewDescribeCommand(loggingScalyrCmdRoot.CmdClause, g, m)
	loggingScalyrList := scalyr.NewListCommand(loggingScalyrCmdRoot.CmdClause, g, m)
	loggingScalyrUpdate := scalyr.NewUpdateCommand(loggingScalyrCmdRoot.CmdClause, g, m)
	loggingSftpCmdRoot := sftp.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingSftpCreate := sftp.NewCreateCommand(loggingSftpCmdRoot.CmdClause, g, m)
	loggingSftpDelete := sftp.NewDeleteCommand(loggingSftpCmdRoot.CmdClause, g, m)
	loggingSftpDescribe := sftp.NewDescribeCommand(loggingSftpCmdRoot.CmdClause, g, m)
	loggingSftpList := sftp.NewListCommand(loggingSftpCmdRoot.CmdClause, g, m)
	loggingSftpUpdate := sftp.NewUpdateCommand(loggingSftpCmdRoot.CmdClause, g, m)
	loggingSplunkCmdRoot := splunk.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingSplunkCreate := splunk.NewCreateCommand(loggingSplunkCmdRoot.CmdClause, g, m)
	loggingSplunkDelete := splunk.NewDeleteCommand(loggingSplunkCmdRoot.CmdClause, g, m)
	loggingSplunkDescribe := splunk.NewDescribeCommand(loggingSplunkCmdRoot.CmdClause, g, m)
	loggingSplunkList := splunk.NewListCommand(loggingSplunkCmdRoot.CmdClause, g, m)
	loggingSplunkUpdate := splunk.NewUpdateCommand(loggingSplunkCmdRoot.CmdClause, g, m)
	loggingSumologicCmdRoot := sumologic.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingSumologicCreate := sumologic.NewCreateCommand(loggingSumologicCmdRoot.CmdClause, g, m)
	loggingSumologicDelete := sumologic.NewDeleteCommand(loggingSumologicCmdRoot.CmdClause, g, m)
	loggingSumologicDescribe := sumologic.NewDescribeCommand(loggingSumologicCmdRoot.CmdClause, g, m)
	loggingSumologicList := sumologic.NewListCommand(loggingSumologicCmdRoot.CmdClause, g, m)
	loggingSumologicUpdate := sumologic.NewUpdateCommand(loggingSumologicCmdRoot.CmdClause, g, m)
	loggingSyslogCmdRoot := syslog.NewRootCommand(loggingCmdRoot.CmdClause, g)
	loggingSyslogCreate := syslog.NewCreateCommand(loggingSyslogCmdRoot.CmdClause, g, m)
	loggingSyslogDelete := syslog.NewDeleteCommand(loggingSyslogCmdRoot.CmdClause, g, m)
	loggingSyslogDescribe := syslog.NewDescribeCommand(loggingSyslogCmdRoot.CmdClause, g, m)
	loggingSyslogList := syslog.NewListCommand(loggingSyslogCmdRoot.CmdClause, g, m)
	loggingSyslogUpdate := syslog.NewUpdateCommand(loggingSyslogCmdRoot.CmdClause, g, m)
	popCmdRoot := pop.NewRootCommand(app, g)
	productsCmdRoot := products.NewRootCommand(app, g, m)
	profileCmdRoot := profile.NewRootCommand(app, g)
	profileCreate := profile.NewCreateCommand(profileCmdRoot.CmdClause, profile.APIClientFactory(opts.APIClient), g, authenticateCmdRoot)
	profileDelete := profile.NewDeleteCommand(profileCmdRoot.CmdClause, g)
	profileList := profile.NewListCommand(profileCmdRoot.CmdClause, g)
	profileSwitch := profile.NewSwitchCommand(profileCmdRoot.CmdClause, g)
	profileToken := profile.NewTokenCommand(profileCmdRoot.CmdClause, g)
	profileUpdate := profile.NewUpdateCommand(profileCmdRoot.CmdClause, profile.APIClientFactory(opts.APIClient), g, authenticateCmdRoot)
	purgeCmdRoot := purge.NewRootCommand(app, g, m)
	rateLimitCmdRoot := ratelimit.NewRootCommand(app, g)
	rateLimitCreate := ratelimit.NewCreateCommand(rateLimitCmdRoot.CmdClause, g, m)
	rateLimitDelete := ratelimit.NewDeleteCommand(rateLimitCmdRoot.CmdClause, g, m)
	rateLimitDescribe := ratelimit.NewDescribeCommand(rateLimitCmdRoot.CmdClause, g, m)
	rateLimitList := ratelimit.NewListCommand(rateLimitCmdRoot.CmdClause, g, m)
	rateLimitUpdate := ratelimit.NewUpdateCommand(rateLimitCmdRoot.CmdClause, g, m)
	resourcelinkCmdRoot := resourcelink.NewRootCommand(app, g)
	resourcelinkCreate := resourcelink.NewCreateCommand(resourcelinkCmdRoot.CmdClause, g, m)
	resourcelinkDelete := resourcelink.NewDeleteCommand(resourcelinkCmdRoot.CmdClause, g, m)
	resourcelinkDescribe := resourcelink.NewDescribeCommand(resourcelinkCmdRoot.CmdClause, g, m)
	resourcelinkList := resourcelink.NewListCommand(resourcelinkCmdRoot.CmdClause, g, m)
	resourcelinkUpdate := resourcelink.NewUpdateCommand(resourcelinkCmdRoot.CmdClause, g, m)
	secretstoreCmdRoot := secretstore.NewRootCommand(app, g)
	secretstoreCreate := secretstore.NewCreateCommand(secretstoreCmdRoot.CmdClause, g, m)
	secretstoreDescribe := secretstore.NewDescribeCommand(secretstoreCmdRoot.CmdClause, g, m)
	secretstoreDelete := secretstore.NewDeleteCommand(secretstoreCmdRoot.CmdClause, g, m)
	secretstoreList := secretstore.NewListCommand(secretstoreCmdRoot.CmdClause, g, m)
	secretstoreentryCmdRoot := secretstoreentry.NewRootCommand(app, g)
	secretstoreentryCreate := secretstoreentry.NewCreateCommand(secretstoreentryCmdRoot.CmdClause, g, m)
	secretstoreentryDescribe := secretstoreentry.NewDescribeCommand(secretstoreentryCmdRoot.CmdClause, g, m)
	secretstoreentryDelete := secretstoreentry.NewDeleteCommand(secretstoreentryCmdRoot.CmdClause, g, m)
	secretstoreentryList := secretstoreentry.NewListCommand(secretstoreentryCmdRoot.CmdClause, g, m)
	serviceCmdRoot := service.NewRootCommand(app, g)
	serviceCreate := service.NewCreateCommand(serviceCmdRoot.CmdClause, g)
	serviceDelete := service.NewDeleteCommand(serviceCmdRoot.CmdClause, g, m)
	serviceDescribe := service.NewDescribeCommand(serviceCmdRoot.CmdClause, g, m)
	serviceList := service.NewListCommand(serviceCmdRoot.CmdClause, g)
	serviceSearch := service.NewSearchCommand(serviceCmdRoot.CmdClause, g, m)
	serviceUpdate := service.NewUpdateCommand(serviceCmdRoot.CmdClause, g, m)
	serviceauthCmdRoot := serviceauth.NewRootCommand(app, g)
	serviceauthCreate := serviceauth.NewCreateCommand(serviceauthCmdRoot.CmdClause, g, m)
	serviceauthDelete := serviceauth.NewDeleteCommand(serviceauthCmdRoot.CmdClause, g, m)
	serviceauthDescribe := serviceauth.NewDescribeCommand(serviceauthCmdRoot.CmdClause, g, m)
	serviceauthList := serviceauth.NewListCommand(serviceauthCmdRoot.CmdClause, g)
	serviceauthUpdate := serviceauth.NewUpdateCommand(serviceauthCmdRoot.CmdClause, g, m)
	serviceVersionCmdRoot := serviceversion.NewRootCommand(app, g)
	serviceVersionActivate := serviceversion.NewActivateCommand(serviceVersionCmdRoot.CmdClause, g, m)
	serviceVersionClone := serviceversion.NewCloneCommand(serviceVersionCmdRoot.CmdClause, g, m)
	serviceVersionDeactivate := serviceversion.NewDeactivateCommand(serviceVersionCmdRoot.CmdClause, g, m)
	serviceVersionList := serviceversion.NewListCommand(serviceVersionCmdRoot.CmdClause, g, m)
	serviceVersionLock := serviceversion.NewLockCommand(serviceVersionCmdRoot.CmdClause, g, m)
	serviceVersionUpdate := serviceversion.NewUpdateCommand(serviceVersionCmdRoot.CmdClause, g, m)
	statsCmdRoot := stats.NewRootCommand(app, g)
	statsHistorical := stats.NewHistoricalCommand(statsCmdRoot.CmdClause, g, m)
	statsRealtime := stats.NewRealtimeCommand(statsCmdRoot.CmdClause, g, m)
	statsRegions := stats.NewRegionsCommand(statsCmdRoot.CmdClause, g)
	tlsConfigCmdRoot := tlsconfig.NewRootCommand(app, g)
	tlsConfigDescribe := tlsconfig.NewDescribeCommand(tlsConfigCmdRoot.CmdClause, g, m)
	tlsConfigList := tlsconfig.NewListCommand(tlsConfigCmdRoot.CmdClause, g, m)
	tlsConfigUpdate := tlsconfig.NewUpdateCommand(tlsConfigCmdRoot.CmdClause, g, m)
	tlsCustomCmdRoot := tlscustom.NewRootCommand(app, g)
	tlsCustomActivationCmdRoot := tlscustomactivation.NewRootCommand(tlsCustomCmdRoot.CmdClause, g)
	tlsCustomActivationCreate := tlscustomactivation.NewCreateCommand(tlsCustomActivationCmdRoot.CmdClause, g, m)
	tlsCustomActivationDelete := tlscustomactivation.NewDeleteCommand(tlsCustomActivationCmdRoot.CmdClause, g, m)
	tlsCustomActivationDescribe := tlscustomactivation.NewDescribeCommand(tlsCustomActivationCmdRoot.CmdClause, g, m)
	tlsCustomActivationList := tlscustomactivation.NewListCommand(tlsCustomActivationCmdRoot.CmdClause, g, m)
	tlsCustomActivationUpdate := tlscustomactivation.NewUpdateCommand(tlsCustomActivationCmdRoot.CmdClause, g, m)
	tlsCustomCertificateCmdRoot := tlscustomcertificate.NewRootCommand(tlsCustomCmdRoot.CmdClause, g)
	tlsCustomCertificateCreate := tlscustomcertificate.NewCreateCommand(tlsCustomCertificateCmdRoot.CmdClause, g, m)
	tlsCustomCertificateDelete := tlscustomcertificate.NewDeleteCommand(tlsCustomCertificateCmdRoot.CmdClause, g, m)
	tlsCustomCertificateDescribe := tlscustomcertificate.NewDescribeCommand(tlsCustomCertificateCmdRoot.CmdClause, g, m)
	tlsCustomCertificateList := tlscustomcertificate.NewListCommand(tlsCustomCertificateCmdRoot.CmdClause, g, m)
	tlsCustomCertificateUpdate := tlscustomcertificate.NewUpdateCommand(tlsCustomCertificateCmdRoot.CmdClause, g, m)
	tlsCustomDomainCmdRoot := tlscustomdomain.NewRootCommand(tlsCustomCmdRoot.CmdClause, g)
	tlsCustomDomainList := tlscustomdomain.NewListCommand(tlsCustomDomainCmdRoot.CmdClause, g, m)
	tlsCustomPrivateKeyCmdRoot := tlscustomprivatekey.NewRootCommand(tlsCustomCmdRoot.CmdClause, g)
	tlsCustomPrivateKeyCreate := tlscustomprivatekey.NewCreateCommand(tlsCustomPrivateKeyCmdRoot.CmdClause, g, m)
	tlsCustomPrivateKeyDelete := tlscustomprivatekey.NewDeleteCommand(tlsCustomPrivateKeyCmdRoot.CmdClause, g, m)
	tlsCustomPrivateKeyDescribe := tlscustomprivatekey.NewDescribeCommand(tlsCustomPrivateKeyCmdRoot.CmdClause, g, m)
	tlsCustomPrivateKeyList := tlscustomprivatekey.NewListCommand(tlsCustomPrivateKeyCmdRoot.CmdClause, g, m)
	tlsPlatformCmdRoot := tlsplatform.NewRootCommand(app, g)
	tlsPlatformCreate := tlsplatform.NewCreateCommand(tlsPlatformCmdRoot.CmdClause, g, m)
	tlsPlatformDelete := tlsplatform.NewDeleteCommand(tlsPlatformCmdRoot.CmdClause, g, m)
	tlsPlatformDescribe := tlsplatform.NewDescribeCommand(tlsPlatformCmdRoot.CmdClause, g, m)
	tlsPlatformList := tlsplatform.NewListCommand(tlsPlatformCmdRoot.CmdClause, g, m)
	tlsPlatformUpdate := tlsplatform.NewUpdateCommand(tlsPlatformCmdRoot.CmdClause, g, m)
	tlsSubscriptionCmdRoot := tlssubscription.NewRootCommand(app, g)
	tlsSubscriptionCreate := tlssubscription.NewCreateCommand(tlsSubscriptionCmdRoot.CmdClause, g, m)
	tlsSubscriptionDelete := tlssubscription.NewDeleteCommand(tlsSubscriptionCmdRoot.CmdClause, g, m)
	tlsSubscriptionDescribe := tlssubscription.NewDescribeCommand(tlsSubscriptionCmdRoot.CmdClause, g, m)
	tlsSubscriptionList := tlssubscription.NewListCommand(tlsSubscriptionCmdRoot.CmdClause, g, m)
	tlsSubscriptionUpdate := tlssubscription.NewUpdateCommand(tlsSubscriptionCmdRoot.CmdClause, g, m)
	updateRoot := update.NewRootCommand(app, opts.ConfigPath, opts.Versioners.CLI, g)
	userCmdRoot := user.NewRootCommand(app, g)
	userCreate := user.NewCreateCommand(userCmdRoot.CmdClause, g, m)
	userDelete := user.NewDeleteCommand(userCmdRoot.CmdClause, g, m)
	userDescribe := user.NewDescribeCommand(userCmdRoot.CmdClause, g, m)
	userList := user.NewListCommand(userCmdRoot.CmdClause, g, m)
	userUpdate := user.NewUpdateCommand(userCmdRoot.CmdClause, g, m)
	vclCmdRoot := vcl.NewRootCommand(app, g)
	vclConditionCmdRoot := condition.NewRootCommand(vclCmdRoot.CmdClause, g)
	vclConditionCreate := condition.NewCreateCommand(vclConditionCmdRoot.CmdClause, g, m)
	vclConditionDelete := condition.NewDeleteCommand(vclConditionCmdRoot.CmdClause, g, m)
	vclConditionDescribe := condition.NewDescribeCommand(vclConditionCmdRoot.CmdClause, g, m)
	vclConditionList := condition.NewListCommand(vclConditionCmdRoot.CmdClause, g, m)
	vclConditionUpdate := condition.NewUpdateCommand(vclConditionCmdRoot.CmdClause, g, m)
	vclCustomCmdRoot := custom.NewRootCommand(vclCmdRoot.CmdClause, g)
	vclCustomCreate := custom.NewCreateCommand(vclCustomCmdRoot.CmdClause, g, m)
	vclCustomDelete := custom.NewDeleteCommand(vclCustomCmdRoot.CmdClause, g, m)
	vclCustomDescribe := custom.NewDescribeCommand(vclCustomCmdRoot.CmdClause, g, m)
	vclCustomList := custom.NewListCommand(vclCustomCmdRoot.CmdClause, g, m)
	vclCustomUpdate := custom.NewUpdateCommand(vclCustomCmdRoot.CmdClause, g, m)
	vclSnippetCmdRoot := snippet.NewRootCommand(vclCmdRoot.CmdClause, g)
	vclSnippetCreate := snippet.NewCreateCommand(vclSnippetCmdRoot.CmdClause, g, m)
	vclSnippetDelete := snippet.NewDeleteCommand(vclSnippetCmdRoot.CmdClause, g, m)
	vclSnippetDescribe := snippet.NewDescribeCommand(vclSnippetCmdRoot.CmdClause, g, m)
	vclSnippetList := snippet.NewListCommand(vclSnippetCmdRoot.CmdClause, g, m)
	vclSnippetUpdate := snippet.NewUpdateCommand(vclSnippetCmdRoot.CmdClause, g, m)
	versionCmdRoot := version.NewRootCommand(app, opts.Versioners.Viceroy)
	whoamiCmdRoot := whoami.NewRootCommand(app, g)

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
		authenticateCmdRoot,
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
		computeHashFiles,
		computeHashsum,
		computeInit,
		computeMetadata,
		computePack,
		computePublish,
		computeServe,
		computeUpdate,
		computeValidate,
		configCmdRoot,
		configstoreCmdRoot,
		configstoreCreate,
		configstoreDelete,
		configstoreDescribe,
		configstoreList,
		configstoreListServices,
		configstoreUpdate,
		configstoreentryCmdRoot,
		configstoreentryCreate,
		configstoreentryDelete,
		configstoreentryDescribe,
		configstoreentryList,
		configstoreentryUpdate,
		dictionaryCmdRoot,
		dictionaryCreate,
		dictionaryDelete,
		dictionaryDescribe,
		dictionaryEntryCmdRoot,
		dictionaryEntryCreate,
		dictionaryEntryDelete,
		dictionaryEntryDescribe,
		dictionaryEntryList,
		dictionaryEntryUpdate,
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
		kvstoreCreate,
		kvstoreDelete,
		kvstoreDescribe,
		kvstoreList,
		kvstoreentryCreate,
		kvstoreentryDelete,
		kvstoreentryDescribe,
		kvstoreentryList,
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
		loggingNewRelicOTLPCmdRoot,
		loggingNewRelicOTLPCreate,
		loggingNewRelicOTLPDelete,
		loggingNewRelicOTLPDescribe,
		loggingNewRelicOTLPList,
		loggingNewRelicOTLPUpdate,
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
		productsCmdRoot,
		profileCmdRoot,
		profileCreate,
		profileDelete,
		profileList,
		profileSwitch,
		profileToken,
		profileUpdate,
		purgeCmdRoot,
		rateLimitCmdRoot,
		rateLimitCreate,
		rateLimitDelete,
		rateLimitDescribe,
		rateLimitList,
		rateLimitUpdate,
		resourcelinkCmdRoot,
		resourcelinkCreate,
		resourcelinkDelete,
		resourcelinkDescribe,
		resourcelinkList,
		resourcelinkUpdate,
		secretstoreCreate,
		secretstoreDescribe,
		secretstoreDelete,
		secretstoreList,
		secretstoreentryCreate,
		secretstoreentryDescribe,
		secretstoreentryDelete,
		secretstoreentryList,
		serviceCmdRoot,
		serviceCreate,
		serviceDelete,
		serviceDescribe,
		serviceList,
		serviceSearch,
		serviceUpdate,
		serviceauthCmdRoot,
		serviceauthCreate,
		serviceauthDelete,
		serviceauthDescribe,
		serviceauthList,
		serviceauthUpdate,
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
		tlsConfigCmdRoot,
		tlsConfigDescribe,
		tlsConfigList,
		tlsConfigUpdate,
		tlsCustomCmdRoot,
		tlsCustomActivationCmdRoot,
		tlsCustomActivationCreate,
		tlsCustomActivationDelete,
		tlsCustomActivationDescribe,
		tlsCustomActivationList,
		tlsCustomActivationUpdate,
		tlsCustomCertificateCmdRoot,
		tlsCustomCertificateCreate,
		tlsCustomCertificateDelete,
		tlsCustomCertificateDescribe,
		tlsCustomCertificateList,
		tlsCustomCertificateUpdate,
		tlsCustomDomainCmdRoot,
		tlsCustomDomainList,
		tlsCustomPrivateKeyCmdRoot,
		tlsCustomPrivateKeyCreate,
		tlsCustomPrivateKeyDelete,
		tlsCustomPrivateKeyDescribe,
		tlsCustomPrivateKeyList,
		tlsPlatformCmdRoot,
		tlsPlatformCreate,
		tlsPlatformDelete,
		tlsPlatformDescribe,
		tlsPlatformList,
		tlsPlatformUpdate,
		tlsSubscriptionCmdRoot,
		tlsSubscriptionCreate,
		tlsSubscriptionDelete,
		tlsSubscriptionDescribe,
		tlsSubscriptionList,
		tlsSubscriptionUpdate,
		updateRoot,
		userCmdRoot,
		userCreate,
		userDelete,
		userDescribe,
		userList,
		userUpdate,
		vclCmdRoot,
		vclConditionCmdRoot,
		vclConditionCreate,
		vclConditionDelete,
		vclConditionDescribe,
		vclConditionList,
		vclConditionUpdate,
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
