package commands

import (
	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/acl"
	"github.com/fastly/cli/pkg/commands/aclentry"
	"github.com/fastly/cli/pkg/commands/alerts"
	"github.com/fastly/cli/pkg/commands/authtoken"
	"github.com/fastly/cli/pkg/commands/backend"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/commands/compute/computeacl"
	"github.com/fastly/cli/pkg/commands/config"
	"github.com/fastly/cli/pkg/commands/configstore"
	"github.com/fastly/cli/pkg/commands/configstoreentry"
	"github.com/fastly/cli/pkg/commands/dashboard"
	dashboardItem "github.com/fastly/cli/pkg/commands/dashboard/item"
	"github.com/fastly/cli/pkg/commands/dictionary"
	"github.com/fastly/cli/pkg/commands/dictionaryentry"
	"github.com/fastly/cli/pkg/commands/domain"
	"github.com/fastly/cli/pkg/commands/domainv1"
	"github.com/fastly/cli/pkg/commands/healthcheck"
	"github.com/fastly/cli/pkg/commands/install"
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
	"github.com/fastly/cli/pkg/commands/logging/grafanacloudlogs"
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
	"github.com/fastly/cli/pkg/commands/objectstorage"
	"github.com/fastly/cli/pkg/commands/objectstorage/accesskeys"
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
	"github.com/fastly/cli/pkg/commands/sso"
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
)

// Define constructs all the commands exposed by the CLI.
func Define( // nolint:revive // function-length
	app *kingpin.Application,
	data *global.Data,
) []argparser.Command {
	shellcompleteCmdRoot := shellcomplete.NewRootCommand(app, data)

	// NOTE: The order commands are created are the order they appear in 'help'.
	// But because we need to pass the SSO command into the profile commands, it
	// means the SSO command must be created _before_ the profile commands. This
	// messes up the order of the commands in the `--help` output. So to make the
	// placement of the `sso` subcommand not look too odd we place it at the
	// beginning of the list of commands.
	ssoCmdRoot := sso.NewRootCommand(app, data)

	aclCmdRoot := acl.NewRootCommand(app, data)
	aclCreate := acl.NewCreateCommand(aclCmdRoot.CmdClause, data)
	aclDelete := acl.NewDeleteCommand(aclCmdRoot.CmdClause, data)
	aclDescribe := acl.NewDescribeCommand(aclCmdRoot.CmdClause, data)
	aclList := acl.NewListCommand(aclCmdRoot.CmdClause, data)
	aclUpdate := acl.NewUpdateCommand(aclCmdRoot.CmdClause, data)
	aclEntryCmdRoot := aclentry.NewRootCommand(app, data)
	aclEntryCreate := aclentry.NewCreateCommand(aclEntryCmdRoot.CmdClause, data)
	aclEntryDelete := aclentry.NewDeleteCommand(aclEntryCmdRoot.CmdClause, data)
	aclEntryDescribe := aclentry.NewDescribeCommand(aclEntryCmdRoot.CmdClause, data)
	aclEntryList := aclentry.NewListCommand(aclEntryCmdRoot.CmdClause, data)
	aclEntryUpdate := aclentry.NewUpdateCommand(aclEntryCmdRoot.CmdClause, data)
	alertsCmdRoot := alerts.NewRootCommand(app, data)
	alertsCreate := alerts.NewCreateCommand(alertsCmdRoot.CmdClause, data)
	alertsDelete := alerts.NewDeleteCommand(alertsCmdRoot.CmdClause, data)
	alertsDescribe := alerts.NewDescribeCommand(alertsCmdRoot.CmdClause, data)
	alertsList := alerts.NewListCommand(alertsCmdRoot.CmdClause, data)
	alertsListHistory := alerts.NewListHistoryCommand(alertsCmdRoot.CmdClause, data)
	alertsUpdate := alerts.NewUpdateCommand(alertsCmdRoot.CmdClause, data)
	authtokenCmdRoot := authtoken.NewRootCommand(app, data)
	authtokenCreate := authtoken.NewCreateCommand(authtokenCmdRoot.CmdClause, data)
	authtokenDelete := authtoken.NewDeleteCommand(authtokenCmdRoot.CmdClause, data)
	authtokenDescribe := authtoken.NewDescribeCommand(authtokenCmdRoot.CmdClause, data)
	authtokenList := authtoken.NewListCommand(authtokenCmdRoot.CmdClause, data)
	backendCmdRoot := backend.NewRootCommand(app, data)
	backendCreate := backend.NewCreateCommand(backendCmdRoot.CmdClause, data)
	backendDelete := backend.NewDeleteCommand(backendCmdRoot.CmdClause, data)
	backendDescribe := backend.NewDescribeCommand(backendCmdRoot.CmdClause, data)
	backendList := backend.NewListCommand(backendCmdRoot.CmdClause, data)
	backendUpdate := backend.NewUpdateCommand(backendCmdRoot.CmdClause, data)
	computeCmdRoot := compute.NewRootCommand(app, data)
	computeACLCmdRoot := computeacl.NewRootCommand(computeCmdRoot.CmdClause, data)
	computeACLCreate := computeacl.NewCreateCommand(computeACLCmdRoot.CmdClause, data)
	computeACLList := computeacl.NewListCommand(computeACLCmdRoot.CmdClause, data)
	computeACLDescribe := computeacl.NewDescribeCommand(computeACLCmdRoot.CmdClause, data)
	computeACLUpdate := computeacl.NewUpdateCommand(computeACLCmdRoot.CmdClause, data)
	computeACLLookup := computeacl.NewLookupCommand(computeACLCmdRoot.CmdClause, data)
	computeACLDelete := computeacl.NewDeleteCommand(computeACLCmdRoot.CmdClause, data)
	computeACLEntriesList := computeacl.NewListEntriesCommand(computeACLCmdRoot.CmdClause, data)
	computeBuild := compute.NewBuildCommand(computeCmdRoot.CmdClause, data)
	computeDeploy := compute.NewDeployCommand(computeCmdRoot.CmdClause, data)
	computeHashFiles := compute.NewHashFilesCommand(computeCmdRoot.CmdClause, data, computeBuild)
	computeHashsum := compute.NewHashsumCommand(computeCmdRoot.CmdClause, data, computeBuild)
	computeInit := compute.NewInitCommand(computeCmdRoot.CmdClause, data)
	computeMetadata := compute.NewMetadataCommand(computeCmdRoot.CmdClause, data)
	computePack := compute.NewPackCommand(computeCmdRoot.CmdClause, data)
	computePublish := compute.NewPublishCommand(computeCmdRoot.CmdClause, data, computeBuild, computeDeploy)
	computeServe := compute.NewServeCommand(computeCmdRoot.CmdClause, data, computeBuild)
	computeUpdate := compute.NewUpdateCommand(computeCmdRoot.CmdClause, data)
	computeValidate := compute.NewValidateCommand(computeCmdRoot.CmdClause, data)
	configCmdRoot := config.NewRootCommand(app, data)
	configstoreCmdRoot := configstore.NewRootCommand(app, data)
	configstoreCreate := configstore.NewCreateCommand(configstoreCmdRoot.CmdClause, data)
	configstoreDelete := configstore.NewDeleteCommand(configstoreCmdRoot.CmdClause, data)
	configstoreDescribe := configstore.NewDescribeCommand(configstoreCmdRoot.CmdClause, data)
	configstoreList := configstore.NewListCommand(configstoreCmdRoot.CmdClause, data)
	configstoreListServices := configstore.NewListServicesCommand(configstoreCmdRoot.CmdClause, data)
	configstoreUpdate := configstore.NewUpdateCommand(configstoreCmdRoot.CmdClause, data)
	configstoreentryCmdRoot := configstoreentry.NewRootCommand(app, data)
	configstoreentryCreate := configstoreentry.NewCreateCommand(configstoreentryCmdRoot.CmdClause, data)
	configstoreentryDelete := configstoreentry.NewDeleteCommand(configstoreentryCmdRoot.CmdClause, data)
	configstoreentryDescribe := configstoreentry.NewDescribeCommand(configstoreentryCmdRoot.CmdClause, data)
	configstoreentryList := configstoreentry.NewListCommand(configstoreentryCmdRoot.CmdClause, data)
	configstoreentryUpdate := configstoreentry.NewUpdateCommand(configstoreentryCmdRoot.CmdClause, data)
	dashboardCmdRoot := dashboard.NewRootCommand(app, data)
	dashboardList := dashboard.NewListCommand(dashboardCmdRoot.CmdClause, data)
	dashboardCreate := dashboard.NewCreateCommand(dashboardCmdRoot.CmdClause, data)
	dashboardDescribe := dashboard.NewDescribeCommand(dashboardCmdRoot.CmdClause, data)
	dashboardUpdate := dashboard.NewUpdateCommand(dashboardCmdRoot.CmdClause, data)
	dashboardDelete := dashboard.NewDeleteCommand(dashboardCmdRoot.CmdClause, data)
	dashboardItemCmdRoot := dashboardItem.NewRootCommand(dashboardCmdRoot.CmdClause, data)
	dashboardItemCreate := dashboardItem.NewCreateCommand(dashboardItemCmdRoot.CmdClause, data)
	dashboardItemDescribe := dashboardItem.NewDescribeCommand(dashboardItemCmdRoot.CmdClause, data)
	dashboardItemUpdate := dashboardItem.NewUpdateCommand(dashboardItemCmdRoot.CmdClause, data)
	dashboardItemDelete := dashboardItem.NewDeleteCommand(dashboardItemCmdRoot.CmdClause, data)
	dictionaryCmdRoot := dictionary.NewRootCommand(app, data)
	dictionaryCreate := dictionary.NewCreateCommand(dictionaryCmdRoot.CmdClause, data)
	dictionaryDelete := dictionary.NewDeleteCommand(dictionaryCmdRoot.CmdClause, data)
	dictionaryDescribe := dictionary.NewDescribeCommand(dictionaryCmdRoot.CmdClause, data)
	dictionaryEntryCmdRoot := dictionaryentry.NewRootCommand(app, data)
	dictionaryEntryCreate := dictionaryentry.NewCreateCommand(dictionaryEntryCmdRoot.CmdClause, data)
	dictionaryEntryDelete := dictionaryentry.NewDeleteCommand(dictionaryEntryCmdRoot.CmdClause, data)
	dictionaryEntryDescribe := dictionaryentry.NewDescribeCommand(dictionaryEntryCmdRoot.CmdClause, data)
	dictionaryEntryList := dictionaryentry.NewListCommand(dictionaryEntryCmdRoot.CmdClause, data)
	dictionaryEntryUpdate := dictionaryentry.NewUpdateCommand(dictionaryEntryCmdRoot.CmdClause, data)
	dictionaryList := dictionary.NewListCommand(dictionaryCmdRoot.CmdClause, data)
	dictionaryUpdate := dictionary.NewUpdateCommand(dictionaryCmdRoot.CmdClause, data)
	domainCmdRoot := domain.NewRootCommand(app, data)
	domainCreate := domain.NewCreateCommand(domainCmdRoot.CmdClause, data)
	domainDelete := domain.NewDeleteCommand(domainCmdRoot.CmdClause, data)
	domainDescribe := domain.NewDescribeCommand(domainCmdRoot.CmdClause, data)
	domainList := domain.NewListCommand(domainCmdRoot.CmdClause, data)
	domainUpdate := domain.NewUpdateCommand(domainCmdRoot.CmdClause, data)
	domainValidate := domain.NewValidateCommand(domainCmdRoot.CmdClause, data)
	domainv1CmdRoot := domainv1.NewRootCommand(app, data)
	domainv1Create := domainv1.NewCreateCommand(domainv1CmdRoot.CmdClause, data)
	domainv1Delete := domainv1.NewDeleteCommand(domainv1CmdRoot.CmdClause, data)
	domainv1Describe := domainv1.NewDescribeCommand(domainv1CmdRoot.CmdClause, data)
	domainv1List := domainv1.NewListCommand(domainv1CmdRoot.CmdClause, data)
	domainv1Update := domainv1.NewUpdateCommand(domainv1CmdRoot.CmdClause, data)
	healthcheckCmdRoot := healthcheck.NewRootCommand(app, data)
	healthcheckCreate := healthcheck.NewCreateCommand(healthcheckCmdRoot.CmdClause, data)
	healthcheckDelete := healthcheck.NewDeleteCommand(healthcheckCmdRoot.CmdClause, data)
	healthcheckDescribe := healthcheck.NewDescribeCommand(healthcheckCmdRoot.CmdClause, data)
	healthcheckList := healthcheck.NewListCommand(healthcheckCmdRoot.CmdClause, data)
	healthcheckUpdate := healthcheck.NewUpdateCommand(healthcheckCmdRoot.CmdClause, data)
	installRoot := install.NewRootCommand(app, data)
	ipCmdRoot := ip.NewRootCommand(app, data)
	kvstoreCmdRoot := kvstore.NewRootCommand(app, data)
	kvstoreCreate := kvstore.NewCreateCommand(kvstoreCmdRoot.CmdClause, data)
	kvstoreDelete := kvstore.NewDeleteCommand(kvstoreCmdRoot.CmdClause, data)
	kvstoreDescribe := kvstore.NewDescribeCommand(kvstoreCmdRoot.CmdClause, data)
	kvstoreList := kvstore.NewListCommand(kvstoreCmdRoot.CmdClause, data)
	kvstoreentryCmdRoot := kvstoreentry.NewRootCommand(app, data)
	kvstoreentryCreate := kvstoreentry.NewCreateCommand(kvstoreentryCmdRoot.CmdClause, data)
	kvstoreentryDelete := kvstoreentry.NewDeleteCommand(kvstoreentryCmdRoot.CmdClause, data)
	kvstoreentryDescribe := kvstoreentry.NewDescribeCommand(kvstoreentryCmdRoot.CmdClause, data)
	kvstoreentryList := kvstoreentry.NewListCommand(kvstoreentryCmdRoot.CmdClause, data)
	logtailCmdRoot := logtail.NewRootCommand(app, data)
	loggingCmdRoot := logging.NewRootCommand(app, data)
	loggingAzureblobCmdRoot := azureblob.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingAzureblobCreate := azureblob.NewCreateCommand(loggingAzureblobCmdRoot.CmdClause, data)
	loggingAzureblobDelete := azureblob.NewDeleteCommand(loggingAzureblobCmdRoot.CmdClause, data)
	loggingAzureblobDescribe := azureblob.NewDescribeCommand(loggingAzureblobCmdRoot.CmdClause, data)
	loggingAzureblobList := azureblob.NewListCommand(loggingAzureblobCmdRoot.CmdClause, data)
	loggingAzureblobUpdate := azureblob.NewUpdateCommand(loggingAzureblobCmdRoot.CmdClause, data)
	loggingBigQueryCmdRoot := bigquery.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingBigQueryCreate := bigquery.NewCreateCommand(loggingBigQueryCmdRoot.CmdClause, data)
	loggingBigQueryDelete := bigquery.NewDeleteCommand(loggingBigQueryCmdRoot.CmdClause, data)
	loggingBigQueryDescribe := bigquery.NewDescribeCommand(loggingBigQueryCmdRoot.CmdClause, data)
	loggingBigQueryList := bigquery.NewListCommand(loggingBigQueryCmdRoot.CmdClause, data)
	loggingBigQueryUpdate := bigquery.NewUpdateCommand(loggingBigQueryCmdRoot.CmdClause, data)
	loggingCloudfilesCmdRoot := cloudfiles.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingCloudfilesCreate := cloudfiles.NewCreateCommand(loggingCloudfilesCmdRoot.CmdClause, data)
	loggingCloudfilesDelete := cloudfiles.NewDeleteCommand(loggingCloudfilesCmdRoot.CmdClause, data)
	loggingCloudfilesDescribe := cloudfiles.NewDescribeCommand(loggingCloudfilesCmdRoot.CmdClause, data)
	loggingCloudfilesList := cloudfiles.NewListCommand(loggingCloudfilesCmdRoot.CmdClause, data)
	loggingCloudfilesUpdate := cloudfiles.NewUpdateCommand(loggingCloudfilesCmdRoot.CmdClause, data)
	loggingDatadogCmdRoot := datadog.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingDatadogCreate := datadog.NewCreateCommand(loggingDatadogCmdRoot.CmdClause, data)
	loggingDatadogDelete := datadog.NewDeleteCommand(loggingDatadogCmdRoot.CmdClause, data)
	loggingDatadogDescribe := datadog.NewDescribeCommand(loggingDatadogCmdRoot.CmdClause, data)
	loggingDatadogList := datadog.NewListCommand(loggingDatadogCmdRoot.CmdClause, data)
	loggingDatadogUpdate := datadog.NewUpdateCommand(loggingDatadogCmdRoot.CmdClause, data)
	loggingDigitaloceanCmdRoot := digitalocean.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingDigitaloceanCreate := digitalocean.NewCreateCommand(loggingDigitaloceanCmdRoot.CmdClause, data)
	loggingDigitaloceanDelete := digitalocean.NewDeleteCommand(loggingDigitaloceanCmdRoot.CmdClause, data)
	loggingDigitaloceanDescribe := digitalocean.NewDescribeCommand(loggingDigitaloceanCmdRoot.CmdClause, data)
	loggingDigitaloceanList := digitalocean.NewListCommand(loggingDigitaloceanCmdRoot.CmdClause, data)
	loggingDigitaloceanUpdate := digitalocean.NewUpdateCommand(loggingDigitaloceanCmdRoot.CmdClause, data)
	loggingElasticsearchCmdRoot := elasticsearch.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingElasticsearchCreate := elasticsearch.NewCreateCommand(loggingElasticsearchCmdRoot.CmdClause, data)
	loggingElasticsearchDelete := elasticsearch.NewDeleteCommand(loggingElasticsearchCmdRoot.CmdClause, data)
	loggingElasticsearchDescribe := elasticsearch.NewDescribeCommand(loggingElasticsearchCmdRoot.CmdClause, data)
	loggingElasticsearchList := elasticsearch.NewListCommand(loggingElasticsearchCmdRoot.CmdClause, data)
	loggingElasticsearchUpdate := elasticsearch.NewUpdateCommand(loggingElasticsearchCmdRoot.CmdClause, data)
	loggingFtpCmdRoot := ftp.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingFtpCreate := ftp.NewCreateCommand(loggingFtpCmdRoot.CmdClause, data)
	loggingFtpDelete := ftp.NewDeleteCommand(loggingFtpCmdRoot.CmdClause, data)
	loggingFtpDescribe := ftp.NewDescribeCommand(loggingFtpCmdRoot.CmdClause, data)
	loggingFtpList := ftp.NewListCommand(loggingFtpCmdRoot.CmdClause, data)
	loggingFtpUpdate := ftp.NewUpdateCommand(loggingFtpCmdRoot.CmdClause, data)
	loggingGcsCmdRoot := gcs.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingGcsCreate := gcs.NewCreateCommand(loggingGcsCmdRoot.CmdClause, data)
	loggingGcsDelete := gcs.NewDeleteCommand(loggingGcsCmdRoot.CmdClause, data)
	loggingGcsDescribe := gcs.NewDescribeCommand(loggingGcsCmdRoot.CmdClause, data)
	loggingGcsList := gcs.NewListCommand(loggingGcsCmdRoot.CmdClause, data)
	loggingGcsUpdate := gcs.NewUpdateCommand(loggingGcsCmdRoot.CmdClause, data)
	loggingGooglepubsubCmdRoot := googlepubsub.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingGooglepubsubCreate := googlepubsub.NewCreateCommand(loggingGooglepubsubCmdRoot.CmdClause, data)
	loggingGooglepubsubDelete := googlepubsub.NewDeleteCommand(loggingGooglepubsubCmdRoot.CmdClause, data)
	loggingGooglepubsubDescribe := googlepubsub.NewDescribeCommand(loggingGooglepubsubCmdRoot.CmdClause, data)
	loggingGooglepubsubList := googlepubsub.NewListCommand(loggingGooglepubsubCmdRoot.CmdClause, data)
	loggingGooglepubsubUpdate := googlepubsub.NewUpdateCommand(loggingGooglepubsubCmdRoot.CmdClause, data)
	loggingGrafanacloudlogsCmdRoot := grafanacloudlogs.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingGrafanacloudlogsCreate := grafanacloudlogs.NewCreateCommand(loggingGrafanacloudlogsCmdRoot.CmdClause, data)
	loggingGrafanacloudlogsDelete := grafanacloudlogs.NewDeleteCommand(loggingGrafanacloudlogsCmdRoot.CmdClause, data)
	loggingGrafanacloudlogsDescribe := grafanacloudlogs.NewDescribeCommand(loggingGrafanacloudlogsCmdRoot.CmdClause, data)
	loggingGrafanacloudlogsList := grafanacloudlogs.NewListCommand(loggingGrafanacloudlogsCmdRoot.CmdClause, data)
	loggingGrafanacloudlogsUpdate := grafanacloudlogs.NewUpdateCommand(loggingGrafanacloudlogsCmdRoot.CmdClause, data)
	loggingHerokuCmdRoot := heroku.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingHerokuCreate := heroku.NewCreateCommand(loggingHerokuCmdRoot.CmdClause, data)
	loggingHerokuDelete := heroku.NewDeleteCommand(loggingHerokuCmdRoot.CmdClause, data)
	loggingHerokuDescribe := heroku.NewDescribeCommand(loggingHerokuCmdRoot.CmdClause, data)
	loggingHerokuList := heroku.NewListCommand(loggingHerokuCmdRoot.CmdClause, data)
	loggingHerokuUpdate := heroku.NewUpdateCommand(loggingHerokuCmdRoot.CmdClause, data)
	loggingHoneycombCmdRoot := honeycomb.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingHoneycombCreate := honeycomb.NewCreateCommand(loggingHoneycombCmdRoot.CmdClause, data)
	loggingHoneycombDelete := honeycomb.NewDeleteCommand(loggingHoneycombCmdRoot.CmdClause, data)
	loggingHoneycombDescribe := honeycomb.NewDescribeCommand(loggingHoneycombCmdRoot.CmdClause, data)
	loggingHoneycombList := honeycomb.NewListCommand(loggingHoneycombCmdRoot.CmdClause, data)
	loggingHoneycombUpdate := honeycomb.NewUpdateCommand(loggingHoneycombCmdRoot.CmdClause, data)
	loggingHTTPSCmdRoot := https.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingHTTPSCreate := https.NewCreateCommand(loggingHTTPSCmdRoot.CmdClause, data)
	loggingHTTPSDelete := https.NewDeleteCommand(loggingHTTPSCmdRoot.CmdClause, data)
	loggingHTTPSDescribe := https.NewDescribeCommand(loggingHTTPSCmdRoot.CmdClause, data)
	loggingHTTPSList := https.NewListCommand(loggingHTTPSCmdRoot.CmdClause, data)
	loggingHTTPSUpdate := https.NewUpdateCommand(loggingHTTPSCmdRoot.CmdClause, data)
	loggingKafkaCmdRoot := kafka.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingKafkaCreate := kafka.NewCreateCommand(loggingKafkaCmdRoot.CmdClause, data)
	loggingKafkaDelete := kafka.NewDeleteCommand(loggingKafkaCmdRoot.CmdClause, data)
	loggingKafkaDescribe := kafka.NewDescribeCommand(loggingKafkaCmdRoot.CmdClause, data)
	loggingKafkaList := kafka.NewListCommand(loggingKafkaCmdRoot.CmdClause, data)
	loggingKafkaUpdate := kafka.NewUpdateCommand(loggingKafkaCmdRoot.CmdClause, data)
	loggingKinesisCmdRoot := kinesis.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingKinesisCreate := kinesis.NewCreateCommand(loggingKinesisCmdRoot.CmdClause, data)
	loggingKinesisDelete := kinesis.NewDeleteCommand(loggingKinesisCmdRoot.CmdClause, data)
	loggingKinesisDescribe := kinesis.NewDescribeCommand(loggingKinesisCmdRoot.CmdClause, data)
	loggingKinesisList := kinesis.NewListCommand(loggingKinesisCmdRoot.CmdClause, data)
	loggingKinesisUpdate := kinesis.NewUpdateCommand(loggingKinesisCmdRoot.CmdClause, data)
	loggingLogglyCmdRoot := loggly.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingLogglyCreate := loggly.NewCreateCommand(loggingLogglyCmdRoot.CmdClause, data)
	loggingLogglyDelete := loggly.NewDeleteCommand(loggingLogglyCmdRoot.CmdClause, data)
	loggingLogglyDescribe := loggly.NewDescribeCommand(loggingLogglyCmdRoot.CmdClause, data)
	loggingLogglyList := loggly.NewListCommand(loggingLogglyCmdRoot.CmdClause, data)
	loggingLogglyUpdate := loggly.NewUpdateCommand(loggingLogglyCmdRoot.CmdClause, data)
	loggingLogshuttleCmdRoot := logshuttle.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingLogshuttleCreate := logshuttle.NewCreateCommand(loggingLogshuttleCmdRoot.CmdClause, data)
	loggingLogshuttleDelete := logshuttle.NewDeleteCommand(loggingLogshuttleCmdRoot.CmdClause, data)
	loggingLogshuttleDescribe := logshuttle.NewDescribeCommand(loggingLogshuttleCmdRoot.CmdClause, data)
	loggingLogshuttleList := logshuttle.NewListCommand(loggingLogshuttleCmdRoot.CmdClause, data)
	loggingLogshuttleUpdate := logshuttle.NewUpdateCommand(loggingLogshuttleCmdRoot.CmdClause, data)
	loggingNewRelicCmdRoot := newrelic.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingNewRelicCreate := newrelic.NewCreateCommand(loggingNewRelicCmdRoot.CmdClause, data)
	loggingNewRelicDelete := newrelic.NewDeleteCommand(loggingNewRelicCmdRoot.CmdClause, data)
	loggingNewRelicDescribe := newrelic.NewDescribeCommand(loggingNewRelicCmdRoot.CmdClause, data)
	loggingNewRelicList := newrelic.NewListCommand(loggingNewRelicCmdRoot.CmdClause, data)
	loggingNewRelicUpdate := newrelic.NewUpdateCommand(loggingNewRelicCmdRoot.CmdClause, data)
	loggingNewRelicOTLPCmdRoot := newrelicotlp.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingNewRelicOTLPCreate := newrelicotlp.NewCreateCommand(loggingNewRelicOTLPCmdRoot.CmdClause, data)
	loggingNewRelicOTLPDelete := newrelicotlp.NewDeleteCommand(loggingNewRelicOTLPCmdRoot.CmdClause, data)
	loggingNewRelicOTLPDescribe := newrelicotlp.NewDescribeCommand(loggingNewRelicOTLPCmdRoot.CmdClause, data)
	loggingNewRelicOTLPList := newrelicotlp.NewListCommand(loggingNewRelicOTLPCmdRoot.CmdClause, data)
	loggingNewRelicOTLPUpdate := newrelicotlp.NewUpdateCommand(loggingNewRelicOTLPCmdRoot.CmdClause, data)
	loggingOpenstackCmdRoot := openstack.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingOpenstackCreate := openstack.NewCreateCommand(loggingOpenstackCmdRoot.CmdClause, data)
	loggingOpenstackDelete := openstack.NewDeleteCommand(loggingOpenstackCmdRoot.CmdClause, data)
	loggingOpenstackDescribe := openstack.NewDescribeCommand(loggingOpenstackCmdRoot.CmdClause, data)
	loggingOpenstackList := openstack.NewListCommand(loggingOpenstackCmdRoot.CmdClause, data)
	loggingOpenstackUpdate := openstack.NewUpdateCommand(loggingOpenstackCmdRoot.CmdClause, data)
	loggingPapertrailCmdRoot := papertrail.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingPapertrailCreate := papertrail.NewCreateCommand(loggingPapertrailCmdRoot.CmdClause, data)
	loggingPapertrailDelete := papertrail.NewDeleteCommand(loggingPapertrailCmdRoot.CmdClause, data)
	loggingPapertrailDescribe := papertrail.NewDescribeCommand(loggingPapertrailCmdRoot.CmdClause, data)
	loggingPapertrailList := papertrail.NewListCommand(loggingPapertrailCmdRoot.CmdClause, data)
	loggingPapertrailUpdate := papertrail.NewUpdateCommand(loggingPapertrailCmdRoot.CmdClause, data)
	loggingS3CmdRoot := s3.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingS3Create := s3.NewCreateCommand(loggingS3CmdRoot.CmdClause, data)
	loggingS3Delete := s3.NewDeleteCommand(loggingS3CmdRoot.CmdClause, data)
	loggingS3Describe := s3.NewDescribeCommand(loggingS3CmdRoot.CmdClause, data)
	loggingS3List := s3.NewListCommand(loggingS3CmdRoot.CmdClause, data)
	loggingS3Update := s3.NewUpdateCommand(loggingS3CmdRoot.CmdClause, data)
	loggingScalyrCmdRoot := scalyr.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingScalyrCreate := scalyr.NewCreateCommand(loggingScalyrCmdRoot.CmdClause, data)
	loggingScalyrDelete := scalyr.NewDeleteCommand(loggingScalyrCmdRoot.CmdClause, data)
	loggingScalyrDescribe := scalyr.NewDescribeCommand(loggingScalyrCmdRoot.CmdClause, data)
	loggingScalyrList := scalyr.NewListCommand(loggingScalyrCmdRoot.CmdClause, data)
	loggingScalyrUpdate := scalyr.NewUpdateCommand(loggingScalyrCmdRoot.CmdClause, data)
	loggingSftpCmdRoot := sftp.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingSftpCreate := sftp.NewCreateCommand(loggingSftpCmdRoot.CmdClause, data)
	loggingSftpDelete := sftp.NewDeleteCommand(loggingSftpCmdRoot.CmdClause, data)
	loggingSftpDescribe := sftp.NewDescribeCommand(loggingSftpCmdRoot.CmdClause, data)
	loggingSftpList := sftp.NewListCommand(loggingSftpCmdRoot.CmdClause, data)
	loggingSftpUpdate := sftp.NewUpdateCommand(loggingSftpCmdRoot.CmdClause, data)
	loggingSplunkCmdRoot := splunk.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingSplunkCreate := splunk.NewCreateCommand(loggingSplunkCmdRoot.CmdClause, data)
	loggingSplunkDelete := splunk.NewDeleteCommand(loggingSplunkCmdRoot.CmdClause, data)
	loggingSplunkDescribe := splunk.NewDescribeCommand(loggingSplunkCmdRoot.CmdClause, data)
	loggingSplunkList := splunk.NewListCommand(loggingSplunkCmdRoot.CmdClause, data)
	loggingSplunkUpdate := splunk.NewUpdateCommand(loggingSplunkCmdRoot.CmdClause, data)
	loggingSumologicCmdRoot := sumologic.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingSumologicCreate := sumologic.NewCreateCommand(loggingSumologicCmdRoot.CmdClause, data)
	loggingSumologicDelete := sumologic.NewDeleteCommand(loggingSumologicCmdRoot.CmdClause, data)
	loggingSumologicDescribe := sumologic.NewDescribeCommand(loggingSumologicCmdRoot.CmdClause, data)
	loggingSumologicList := sumologic.NewListCommand(loggingSumologicCmdRoot.CmdClause, data)
	loggingSumologicUpdate := sumologic.NewUpdateCommand(loggingSumologicCmdRoot.CmdClause, data)
	loggingSyslogCmdRoot := syslog.NewRootCommand(loggingCmdRoot.CmdClause, data)
	loggingSyslogCreate := syslog.NewCreateCommand(loggingSyslogCmdRoot.CmdClause, data)
	loggingSyslogDelete := syslog.NewDeleteCommand(loggingSyslogCmdRoot.CmdClause, data)
	loggingSyslogDescribe := syslog.NewDescribeCommand(loggingSyslogCmdRoot.CmdClause, data)
	loggingSyslogList := syslog.NewListCommand(loggingSyslogCmdRoot.CmdClause, data)
	loggingSyslogUpdate := syslog.NewUpdateCommand(loggingSyslogCmdRoot.CmdClause, data)
	objectStorageRoot := objectstorage.NewRootCommand(app, data)
	objectStorageAccesskeysRoot := accesskeys.NewRootCommand(objectStorageRoot.CmdClause, data)
	objectStorageAccesskeysCreate := accesskeys.NewCreateCommand(objectStorageAccesskeysRoot.CmdClause, data)
	objectStorageAccesskeysDelete := accesskeys.NewDeleteCommand(objectStorageAccesskeysRoot.CmdClause, data)
	objectStorageAccesskeysGet := accesskeys.NewGetCommand(objectStorageAccesskeysRoot.CmdClause, data)
	objectStorageAccesskeysList := accesskeys.NewListCommand(objectStorageAccesskeysRoot.CmdClause, data)
	popCmdRoot := pop.NewRootCommand(app, data)
	productsCmdRoot := products.NewRootCommand(app, data)
	profileCmdRoot := profile.NewRootCommand(app, data)
	profileCreate := profile.NewCreateCommand(profileCmdRoot.CmdClause, data, ssoCmdRoot)
	profileDelete := profile.NewDeleteCommand(profileCmdRoot.CmdClause, data)
	profileList := profile.NewListCommand(profileCmdRoot.CmdClause, data)
	profileSwitch := profile.NewSwitchCommand(profileCmdRoot.CmdClause, data, ssoCmdRoot)
	profileToken := profile.NewTokenCommand(profileCmdRoot.CmdClause, data)
	profileUpdate := profile.NewUpdateCommand(profileCmdRoot.CmdClause, data, ssoCmdRoot)
	purgeCmdRoot := purge.NewRootCommand(app, data)
	rateLimitCmdRoot := ratelimit.NewRootCommand(app, data)
	rateLimitCreate := ratelimit.NewCreateCommand(rateLimitCmdRoot.CmdClause, data)
	rateLimitDelete := ratelimit.NewDeleteCommand(rateLimitCmdRoot.CmdClause, data)
	rateLimitDescribe := ratelimit.NewDescribeCommand(rateLimitCmdRoot.CmdClause, data)
	rateLimitList := ratelimit.NewListCommand(rateLimitCmdRoot.CmdClause, data)
	rateLimitUpdate := ratelimit.NewUpdateCommand(rateLimitCmdRoot.CmdClause, data)
	resourcelinkCmdRoot := resourcelink.NewRootCommand(app, data)
	resourcelinkCreate := resourcelink.NewCreateCommand(resourcelinkCmdRoot.CmdClause, data)
	resourcelinkDelete := resourcelink.NewDeleteCommand(resourcelinkCmdRoot.CmdClause, data)
	resourcelinkDescribe := resourcelink.NewDescribeCommand(resourcelinkCmdRoot.CmdClause, data)
	resourcelinkList := resourcelink.NewListCommand(resourcelinkCmdRoot.CmdClause, data)
	resourcelinkUpdate := resourcelink.NewUpdateCommand(resourcelinkCmdRoot.CmdClause, data)
	secretstoreCmdRoot := secretstore.NewRootCommand(app, data)
	secretstoreCreate := secretstore.NewCreateCommand(secretstoreCmdRoot.CmdClause, data)
	secretstoreDescribe := secretstore.NewDescribeCommand(secretstoreCmdRoot.CmdClause, data)
	secretstoreDelete := secretstore.NewDeleteCommand(secretstoreCmdRoot.CmdClause, data)
	secretstoreList := secretstore.NewListCommand(secretstoreCmdRoot.CmdClause, data)
	secretstoreentryCmdRoot := secretstoreentry.NewRootCommand(app, data)
	secretstoreentryCreate := secretstoreentry.NewCreateCommand(secretstoreentryCmdRoot.CmdClause, data)
	secretstoreentryDescribe := secretstoreentry.NewDescribeCommand(secretstoreentryCmdRoot.CmdClause, data)
	secretstoreentryDelete := secretstoreentry.NewDeleteCommand(secretstoreentryCmdRoot.CmdClause, data)
	secretstoreentryList := secretstoreentry.NewListCommand(secretstoreentryCmdRoot.CmdClause, data)
	serviceCmdRoot := service.NewRootCommand(app, data)
	serviceCreate := service.NewCreateCommand(serviceCmdRoot.CmdClause, data)
	serviceDelete := service.NewDeleteCommand(serviceCmdRoot.CmdClause, data)
	serviceDescribe := service.NewDescribeCommand(serviceCmdRoot.CmdClause, data)
	serviceList := service.NewListCommand(serviceCmdRoot.CmdClause, data)
	serviceSearch := service.NewSearchCommand(serviceCmdRoot.CmdClause, data)
	serviceUpdate := service.NewUpdateCommand(serviceCmdRoot.CmdClause, data)
	serviceauthCmdRoot := serviceauth.NewRootCommand(app, data)
	serviceauthCreate := serviceauth.NewCreateCommand(serviceauthCmdRoot.CmdClause, data)
	serviceauthDelete := serviceauth.NewDeleteCommand(serviceauthCmdRoot.CmdClause, data)
	serviceauthDescribe := serviceauth.NewDescribeCommand(serviceauthCmdRoot.CmdClause, data)
	serviceauthList := serviceauth.NewListCommand(serviceauthCmdRoot.CmdClause, data)
	serviceauthUpdate := serviceauth.NewUpdateCommand(serviceauthCmdRoot.CmdClause, data)
	serviceVersionCmdRoot := serviceversion.NewRootCommand(app, data)
	serviceVersionActivate := serviceversion.NewActivateCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionClone := serviceversion.NewCloneCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionDeactivate := serviceversion.NewDeactivateCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionList := serviceversion.NewListCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionLock := serviceversion.NewLockCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionStage := serviceversion.NewStageCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionUnstage := serviceversion.NewUnstageCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionUpdate := serviceversion.NewUpdateCommand(serviceVersionCmdRoot.CmdClause, data)
	statsCmdRoot := stats.NewRootCommand(app, data)
	statsHistorical := stats.NewHistoricalCommand(statsCmdRoot.CmdClause, data)
	statsRealtime := stats.NewRealtimeCommand(statsCmdRoot.CmdClause, data)
	statsRegions := stats.NewRegionsCommand(statsCmdRoot.CmdClause, data)
	tlsConfigCmdRoot := tlsconfig.NewRootCommand(app, data)
	tlsConfigDescribe := tlsconfig.NewDescribeCommand(tlsConfigCmdRoot.CmdClause, data)
	tlsConfigList := tlsconfig.NewListCommand(tlsConfigCmdRoot.CmdClause, data)
	tlsConfigUpdate := tlsconfig.NewUpdateCommand(tlsConfigCmdRoot.CmdClause, data)
	tlsCustomCmdRoot := tlscustom.NewRootCommand(app, data)
	tlsCustomActivationCmdRoot := tlscustomactivation.NewRootCommand(tlsCustomCmdRoot.CmdClause, data)
	tlsCustomActivationCreate := tlscustomactivation.NewCreateCommand(tlsCustomActivationCmdRoot.CmdClause, data)
	tlsCustomActivationDelete := tlscustomactivation.NewDeleteCommand(tlsCustomActivationCmdRoot.CmdClause, data)
	tlsCustomActivationDescribe := tlscustomactivation.NewDescribeCommand(tlsCustomActivationCmdRoot.CmdClause, data)
	tlsCustomActivationList := tlscustomactivation.NewListCommand(tlsCustomActivationCmdRoot.CmdClause, data)
	tlsCustomActivationUpdate := tlscustomactivation.NewUpdateCommand(tlsCustomActivationCmdRoot.CmdClause, data)
	tlsCustomCertificateCmdRoot := tlscustomcertificate.NewRootCommand(tlsCustomCmdRoot.CmdClause, data)
	tlsCustomCertificateCreate := tlscustomcertificate.NewCreateCommand(tlsCustomCertificateCmdRoot.CmdClause, data)
	tlsCustomCertificateDelete := tlscustomcertificate.NewDeleteCommand(tlsCustomCertificateCmdRoot.CmdClause, data)
	tlsCustomCertificateDescribe := tlscustomcertificate.NewDescribeCommand(tlsCustomCertificateCmdRoot.CmdClause, data)
	tlsCustomCertificateList := tlscustomcertificate.NewListCommand(tlsCustomCertificateCmdRoot.CmdClause, data)
	tlsCustomCertificateUpdate := tlscustomcertificate.NewUpdateCommand(tlsCustomCertificateCmdRoot.CmdClause, data)
	tlsCustomDomainCmdRoot := tlscustomdomain.NewRootCommand(tlsCustomCmdRoot.CmdClause, data)
	tlsCustomDomainList := tlscustomdomain.NewListCommand(tlsCustomDomainCmdRoot.CmdClause, data)
	tlsCustomPrivateKeyCmdRoot := tlscustomprivatekey.NewRootCommand(tlsCustomCmdRoot.CmdClause, data)
	tlsCustomPrivateKeyCreate := tlscustomprivatekey.NewCreateCommand(tlsCustomPrivateKeyCmdRoot.CmdClause, data)
	tlsCustomPrivateKeyDelete := tlscustomprivatekey.NewDeleteCommand(tlsCustomPrivateKeyCmdRoot.CmdClause, data)
	tlsCustomPrivateKeyDescribe := tlscustomprivatekey.NewDescribeCommand(tlsCustomPrivateKeyCmdRoot.CmdClause, data)
	tlsCustomPrivateKeyList := tlscustomprivatekey.NewListCommand(tlsCustomPrivateKeyCmdRoot.CmdClause, data)
	tlsPlatformCmdRoot := tlsplatform.NewRootCommand(app, data)
	tlsPlatformCreate := tlsplatform.NewCreateCommand(tlsPlatformCmdRoot.CmdClause, data)
	tlsPlatformDelete := tlsplatform.NewDeleteCommand(tlsPlatformCmdRoot.CmdClause, data)
	tlsPlatformDescribe := tlsplatform.NewDescribeCommand(tlsPlatformCmdRoot.CmdClause, data)
	tlsPlatformList := tlsplatform.NewListCommand(tlsPlatformCmdRoot.CmdClause, data)
	tlsPlatformUpdate := tlsplatform.NewUpdateCommand(tlsPlatformCmdRoot.CmdClause, data)
	tlsSubscriptionCmdRoot := tlssubscription.NewRootCommand(app, data)
	tlsSubscriptionCreate := tlssubscription.NewCreateCommand(tlsSubscriptionCmdRoot.CmdClause, data)
	tlsSubscriptionDelete := tlssubscription.NewDeleteCommand(tlsSubscriptionCmdRoot.CmdClause, data)
	tlsSubscriptionDescribe := tlssubscription.NewDescribeCommand(tlsSubscriptionCmdRoot.CmdClause, data)
	tlsSubscriptionList := tlssubscription.NewListCommand(tlsSubscriptionCmdRoot.CmdClause, data)
	tlsSubscriptionUpdate := tlssubscription.NewUpdateCommand(tlsSubscriptionCmdRoot.CmdClause, data)
	updateRoot := update.NewRootCommand(app, data)
	userCmdRoot := user.NewRootCommand(app, data)
	userCreate := user.NewCreateCommand(userCmdRoot.CmdClause, data)
	userDelete := user.NewDeleteCommand(userCmdRoot.CmdClause, data)
	userDescribe := user.NewDescribeCommand(userCmdRoot.CmdClause, data)
	userList := user.NewListCommand(userCmdRoot.CmdClause, data)
	userUpdate := user.NewUpdateCommand(userCmdRoot.CmdClause, data)
	vclCmdRoot := vcl.NewRootCommand(app, data)
	vclConditionCmdRoot := condition.NewRootCommand(vclCmdRoot.CmdClause, data)
	vclConditionCreate := condition.NewCreateCommand(vclConditionCmdRoot.CmdClause, data)
	vclConditionDelete := condition.NewDeleteCommand(vclConditionCmdRoot.CmdClause, data)
	vclConditionDescribe := condition.NewDescribeCommand(vclConditionCmdRoot.CmdClause, data)
	vclConditionList := condition.NewListCommand(vclConditionCmdRoot.CmdClause, data)
	vclConditionUpdate := condition.NewUpdateCommand(vclConditionCmdRoot.CmdClause, data)
	vclCustomCmdRoot := custom.NewRootCommand(vclCmdRoot.CmdClause, data)
	vclCustomCreate := custom.NewCreateCommand(vclCustomCmdRoot.CmdClause, data)
	vclCustomDelete := custom.NewDeleteCommand(vclCustomCmdRoot.CmdClause, data)
	vclCustomDescribe := custom.NewDescribeCommand(vclCustomCmdRoot.CmdClause, data)
	vclCustomList := custom.NewListCommand(vclCustomCmdRoot.CmdClause, data)
	vclCustomUpdate := custom.NewUpdateCommand(vclCustomCmdRoot.CmdClause, data)
	vclSnippetCmdRoot := snippet.NewRootCommand(vclCmdRoot.CmdClause, data)
	vclSnippetCreate := snippet.NewCreateCommand(vclSnippetCmdRoot.CmdClause, data)
	vclSnippetDelete := snippet.NewDeleteCommand(vclSnippetCmdRoot.CmdClause, data)
	vclSnippetDescribe := snippet.NewDescribeCommand(vclSnippetCmdRoot.CmdClause, data)
	vclSnippetList := snippet.NewListCommand(vclSnippetCmdRoot.CmdClause, data)
	vclSnippetUpdate := snippet.NewUpdateCommand(vclSnippetCmdRoot.CmdClause, data)
	versionCmdRoot := version.NewRootCommand(app, data)
	whoamiCmdRoot := whoami.NewRootCommand(app, data)

	return []argparser.Command{
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
		alertsCreate,
		alertsDelete,
		alertsDescribe,
		alertsList,
		alertsListHistory,
		alertsUpdate,
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
		computeCmdRoot,
		computeACLCmdRoot,
		computeACLCreate,
		computeACLList,
		computeACLDescribe,
		computeACLDelete,
		computeACLUpdate,
		computeACLLookup,
		computeACLEntriesList,
		computeBuild,
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
		dashboardCmdRoot,
		dashboardList,
		dashboardCreate,
		dashboardDescribe,
		dashboardUpdate,
		dashboardDelete,
		dashboardItemCmdRoot,
		dashboardItemCreate,
		dashboardItemDescribe,
		dashboardItemUpdate,
		dashboardItemDelete,
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
		domainv1CmdRoot,
		domainv1Create,
		domainv1Delete,
		domainv1Describe,
		domainv1List,
		domainv1Update,
		healthcheckCmdRoot,
		healthcheckCreate,
		healthcheckDelete,
		healthcheckDescribe,
		healthcheckList,
		healthcheckUpdate,
		installRoot,
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
		loggingGrafanacloudlogsCmdRoot,
		loggingGrafanacloudlogsCreate,
		loggingGrafanacloudlogsDelete,
		loggingGrafanacloudlogsDescribe,
		loggingGrafanacloudlogsList,
		loggingGrafanacloudlogsUpdate,
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
		objectStorageRoot,
		objectStorageAccesskeysRoot,
		objectStorageAccesskeysCreate,
		objectStorageAccesskeysDelete,
		objectStorageAccesskeysGet,
		objectStorageAccesskeysList,
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
		serviceVersionStage,
		serviceVersionUnstage,
		serviceVersionUpdate,
		ssoCmdRoot,
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
