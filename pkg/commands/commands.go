package commands

import (
	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/argparser"
	aliasacl "github.com/fastly/cli/pkg/commands/alias/acl"
	aliasaclentry "github.com/fastly/cli/pkg/commands/alias/aclentry"
	aliasalerts "github.com/fastly/cli/pkg/commands/alias/alerts"
	aliasbackend "github.com/fastly/cli/pkg/commands/alias/backend"
	aliasdictionary "github.com/fastly/cli/pkg/commands/alias/dictionary"
	aliasdictionaryentry "github.com/fastly/cli/pkg/commands/alias/dictionaryentry"
	aliashealthcheck "github.com/fastly/cli/pkg/commands/alias/healthcheck"
	aliasimageoptimizerdefaults "github.com/fastly/cli/pkg/commands/alias/imageoptimizerdefaults"
	aliaspurge "github.com/fastly/cli/pkg/commands/alias/purge"
	aliasratelimit "github.com/fastly/cli/pkg/commands/alias/ratelimit"
	aliasresourcelink "github.com/fastly/cli/pkg/commands/alias/resourcelink"
	aliasserviceversion "github.com/fastly/cli/pkg/commands/alias/serviceversion"
	aliasvcl "github.com/fastly/cli/pkg/commands/alias/vcl"
	aliasvclcondition "github.com/fastly/cli/pkg/commands/alias/vcl/condition"
	aliasvclcustom "github.com/fastly/cli/pkg/commands/alias/vcl/custom"
	aliasvclsnippet "github.com/fastly/cli/pkg/commands/alias/vcl/snippet"
	"github.com/fastly/cli/pkg/commands/authtoken"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/commands/compute/computeacl"
	"github.com/fastly/cli/pkg/commands/config"
	"github.com/fastly/cli/pkg/commands/configstore"
	"github.com/fastly/cli/pkg/commands/configstoreentry"
	"github.com/fastly/cli/pkg/commands/dashboard"
	dashboardItem "github.com/fastly/cli/pkg/commands/dashboard/item"
	"github.com/fastly/cli/pkg/commands/domain"
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
	"github.com/fastly/cli/pkg/commands/ngwaf"
	"github.com/fastly/cli/pkg/commands/ngwaf/countrylist"
	"github.com/fastly/cli/pkg/commands/ngwaf/customsignal"
	"github.com/fastly/cli/pkg/commands/ngwaf/iplist"
	"github.com/fastly/cli/pkg/commands/ngwaf/rule"
	"github.com/fastly/cli/pkg/commands/ngwaf/signallist"
	"github.com/fastly/cli/pkg/commands/ngwaf/stringlist"
	"github.com/fastly/cli/pkg/commands/ngwaf/wildcardlist"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert"
	workspaceAlertDatadog "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/datadog"
	workspaceAlertJira "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/jira"
	workspaceAlertMailinglist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/mailinglist"
	workspaceAlertMicrosoftteams "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/microsoftteams"
	workspaceAlertOpsgenie "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/opsgenie"
	workspaceAlertPagerduty "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/pagerduty"
	workspaceAlertSlack "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/slack"
	workspaceAlertWebhook "github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/webhook"
	wscountrylist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/countrylist"
	wscustomsignal "github.com/fastly/cli/pkg/commands/ngwaf/workspace/customsignal"
	wsiplist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/iplist"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/redaction"
	workspaceRule "github.com/fastly/cli/pkg/commands/ngwaf/workspace/rule"
	wssignallistlist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/signallist"
	wsstringlistlist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/stringlist"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/threshold"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/virtualpatch"
	wswildcardlistlist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/wildcardlist"
	"github.com/fastly/cli/pkg/commands/objectstorage"
	"github.com/fastly/cli/pkg/commands/objectstorage/accesskeys"
	"github.com/fastly/cli/pkg/commands/pop"
	"github.com/fastly/cli/pkg/commands/products"
	"github.com/fastly/cli/pkg/commands/profile"
	"github.com/fastly/cli/pkg/commands/secretstore"
	"github.com/fastly/cli/pkg/commands/secretstoreentry"
	"github.com/fastly/cli/pkg/commands/service"
	serviceacl "github.com/fastly/cli/pkg/commands/service/acl"
	serviceaclentry "github.com/fastly/cli/pkg/commands/service/aclentry"
	servicealert "github.com/fastly/cli/pkg/commands/service/alert"
	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"
	servicedictionary "github.com/fastly/cli/pkg/commands/service/dictionary"
	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"
	servicedomain "github.com/fastly/cli/pkg/commands/service/domain"
	servicehealthcheck "github.com/fastly/cli/pkg/commands/service/healthcheck"
	serviceimageoptimizerdefaults "github.com/fastly/cli/pkg/commands/service/imageoptimizerdefaults"
	servicepurge "github.com/fastly/cli/pkg/commands/service/purge"
	serviceratelimit "github.com/fastly/cli/pkg/commands/service/ratelimit"
	serviceresourcelink "github.com/fastly/cli/pkg/commands/service/resourcelink"
	servicevcl "github.com/fastly/cli/pkg/commands/service/vcl"
	servicevclcondition "github.com/fastly/cli/pkg/commands/service/vcl/condition"
	servicevclcustom "github.com/fastly/cli/pkg/commands/service/vcl/custom"
	servicevclsnippet "github.com/fastly/cli/pkg/commands/service/vcl/snippet"
	serviceversion "github.com/fastly/cli/pkg/commands/service/version"
	"github.com/fastly/cli/pkg/commands/serviceauth"
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
	"github.com/fastly/cli/pkg/commands/tools"
	domainTools "github.com/fastly/cli/pkg/commands/tools/domain"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/commands/user"
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

	authtokenCmdRoot := authtoken.NewRootCommand(app, data)
	authtokenCreate := authtoken.NewCreateCommand(authtokenCmdRoot.CmdClause, data)
	authtokenDelete := authtoken.NewDeleteCommand(authtokenCmdRoot.CmdClause, data)
	authtokenDescribe := authtoken.NewDescribeCommand(authtokenCmdRoot.CmdClause, data)
	authtokenList := authtoken.NewListCommand(authtokenCmdRoot.CmdClause, data)
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
	domainCmdRoot := domain.NewRootCommand(app, data)
	domainCreate := domain.NewCreateCommand(domainCmdRoot.CmdClause, data)
	domainDelete := domain.NewDeleteCommand(domainCmdRoot.CmdClause, data)
	domainDescribe := domain.NewDescribeCommand(domainCmdRoot.CmdClause, data)
	domainList := domain.NewListCommand(domainCmdRoot.CmdClause, data)
	domainUpdate := domain.NewUpdateCommand(domainCmdRoot.CmdClause, data)
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
	kvstoreentryGet := kvstoreentry.NewGetCommand(kvstoreentryCmdRoot.CmdClause, data)
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
	ngwafRoot := ngwaf.NewRootCommand(app, data)
	ngwafWorkspaceRoot := workspace.NewRootCommand(ngwafRoot.CmdClause, data)
	ngwafWorkspaceCreate := workspace.NewCreateCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceDelete := workspace.NewDeleteCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceGet := workspace.NewGetCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceList := workspace.NewListCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceUpdate := workspace.NewUpdateCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafRedactionRoot := redaction.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafRedactionCreate := redaction.NewCreateCommand(ngwafRedactionRoot.CmdClause, data)
	ngwafRedactionDelete := redaction.NewDeleteCommand(ngwafRedactionRoot.CmdClause, data)
	ngwafRedactionList := redaction.NewListCommand(ngwafRedactionRoot.CmdClause, data)
	ngwafRedactionRetrieve := redaction.NewRetrieveCommand(ngwafRedactionRoot.CmdClause, data)
	ngwafRedactionUpdate := redaction.NewUpdateCommand(ngwafRedactionRoot.CmdClause, data)
	ngwafCountryListRoot := countrylist.NewRootCommand(ngwafRoot.CmdClause, data)
	ngwafCountryListCreate := countrylist.NewCreateCommand(ngwafCountryListRoot.CmdClause, data)
	ngwafCountryListDelete := countrylist.NewDeleteCommand(ngwafCountryListRoot.CmdClause, data)
	ngwafCountryListGet := countrylist.NewGetCommand(ngwafCountryListRoot.CmdClause, data)
	ngwafCountryListList := countrylist.NewListCommand(ngwafCountryListRoot.CmdClause, data)
	ngwafCountryListUpdate := countrylist.NewUpdateCommand(ngwafCountryListRoot.CmdClause, data)
	ngwafCustomSignalRoot := customsignal.NewRootCommand(ngwafRoot.CmdClause, data)
	ngwafCustomSignalCreate := customsignal.NewCreateCommand(ngwafCustomSignalRoot.CmdClause, data)
	ngwafCustomSignalDelete := customsignal.NewDeleteCommand(ngwafCustomSignalRoot.CmdClause, data)
	ngwafCustomSignalGet := customsignal.NewGetCommand(ngwafCustomSignalRoot.CmdClause, data)
	ngwafCustomSignalList := customsignal.NewListCommand(ngwafCustomSignalRoot.CmdClause, data)
	ngwafCustomSignalUpdate := customsignal.NewUpdateCommand(ngwafCustomSignalRoot.CmdClause, data)
	ngwafIPListRoot := iplist.NewRootCommand(ngwafRoot.CmdClause, data)
	ngwafIPListCreate := iplist.NewCreateCommand(ngwafIPListRoot.CmdClause, data)
	ngwafIPListDelete := iplist.NewDeleteCommand(ngwafIPListRoot.CmdClause, data)
	ngwafIPListGet := iplist.NewGetCommand(ngwafIPListRoot.CmdClause, data)
	ngwafIPListList := iplist.NewListCommand(ngwafIPListRoot.CmdClause, data)
	ngwafIPListUpdate := iplist.NewUpdateCommand(ngwafIPListRoot.CmdClause, data)
	ngwafRuleRoot := rule.NewRootCommand(ngwafRoot.CmdClause, data)
	ngwafRuleCreate := rule.NewCreateCommand(ngwafRuleRoot.CmdClause, data)
	ngwafRuleDelete := rule.NewDeleteCommand(ngwafRuleRoot.CmdClause, data)
	ngwafRuleGet := rule.NewGetCommand(ngwafRuleRoot.CmdClause, data)
	ngwafRuleList := rule.NewListCommand(ngwafRuleRoot.CmdClause, data)
	ngwafRuleUpdate := rule.NewUpdateCommand(ngwafRuleRoot.CmdClause, data)
	ngwafSignalListRoot := signallist.NewRootCommand(ngwafRoot.CmdClause, data)
	ngwafSignalListCreate := signallist.NewCreateCommand(ngwafSignalListRoot.CmdClause, data)
	ngwafSignalListDelete := signallist.NewDeleteCommand(ngwafSignalListRoot.CmdClause, data)
	ngwafSignalListGet := signallist.NewGetCommand(ngwafSignalListRoot.CmdClause, data)
	ngwafSignalListList := signallist.NewListCommand(ngwafSignalListRoot.CmdClause, data)
	ngwafSignalListUpdate := signallist.NewUpdateCommand(ngwafSignalListRoot.CmdClause, data)
	ngwafStringListRoot := stringlist.NewRootCommand(ngwafRoot.CmdClause, data)
	ngwafStringListCreate := stringlist.NewCreateCommand(ngwafStringListRoot.CmdClause, data)
	ngwafStringListDelete := stringlist.NewDeleteCommand(ngwafStringListRoot.CmdClause, data)
	ngwafStringListGet := stringlist.NewGetCommand(ngwafStringListRoot.CmdClause, data)
	ngwafStringListList := stringlist.NewListCommand(ngwafStringListRoot.CmdClause, data)
	ngwafStringListUpdate := stringlist.NewUpdateCommand(ngwafStringListRoot.CmdClause, data)
	ngwafWildcardListRoot := wildcardlist.NewRootCommand(ngwafRoot.CmdClause, data)
	ngwafWildcardListCreate := wildcardlist.NewCreateCommand(ngwafWildcardListRoot.CmdClause, data)
	ngwafWildcardListDelete := wildcardlist.NewDeleteCommand(ngwafWildcardListRoot.CmdClause, data)
	ngwafWildcardListGet := wildcardlist.NewGetCommand(ngwafWildcardListRoot.CmdClause, data)
	ngwafWildcardListList := wildcardlist.NewListCommand(ngwafWildcardListRoot.CmdClause, data)
	ngwafWildcardListUpdate := wildcardlist.NewUpdateCommand(ngwafWildcardListRoot.CmdClause, data)
	ngwafWorkspaceCountryListRoot := wscountrylist.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceCountryListCreate := wscountrylist.NewCreateCommand(ngwafWorkspaceCountryListRoot.CmdClause, data)
	ngwafWorkspaceCountryListDelete := wscountrylist.NewDeleteCommand(ngwafWorkspaceCountryListRoot.CmdClause, data)
	ngwafWorkspaceCountryListGet := wscountrylist.NewGetCommand(ngwafWorkspaceCountryListRoot.CmdClause, data)
	ngwafWorkspaceCountryListList := wscountrylist.NewListCommand(ngwafWorkspaceCountryListRoot.CmdClause, data)
	ngwafWorkspaceCountryListUpdate := wscountrylist.NewUpdateCommand(ngwafWorkspaceCountryListRoot.CmdClause, data)
	ngwafWorkspaceCustomSignalRoot := wscustomsignal.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceCustomSignalCreate := wscustomsignal.NewCreateCommand(ngwafWorkspaceCustomSignalRoot.CmdClause, data)
	ngwafWorkspaceCustomSignalDelete := wscustomsignal.NewDeleteCommand(ngwafWorkspaceCustomSignalRoot.CmdClause, data)
	ngwafWorkspaceCustomSignalGet := wscustomsignal.NewGetCommand(ngwafWorkspaceCustomSignalRoot.CmdClause, data)
	ngwafWorkspaceCustomSignalList := wscustomsignal.NewListCommand(ngwafWorkspaceCustomSignalRoot.CmdClause, data)
	ngwafWorkspaceCustomSignalUpdate := wscustomsignal.NewUpdateCommand(ngwafWorkspaceCustomSignalRoot.CmdClause, data)
	ngwafWorkspaceIPListRoot := wsiplist.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceIPListCreate := wsiplist.NewCreateCommand(ngwafWorkspaceIPListRoot.CmdClause, data)
	ngwafWorkspaceIPListDelete := wsiplist.NewDeleteCommand(ngwafWorkspaceIPListRoot.CmdClause, data)
	ngwafWorkspaceIPListGet := wsiplist.NewGetCommand(ngwafWorkspaceIPListRoot.CmdClause, data)
	ngwafWorkspaceIPListList := wsiplist.NewListCommand(ngwafWorkspaceIPListRoot.CmdClause, data)
	ngwafWorkspaceIPListUpdate := wsiplist.NewUpdateCommand(ngwafWorkspaceIPListRoot.CmdClause, data)
	ngwafWorkspaceRuleRoot := workspaceRule.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceRuleCreate := workspaceRule.NewCreateCommand(ngwafWorkspaceRuleRoot.CmdClause, data)
	ngwafWorkspaceRuleDelete := workspaceRule.NewDeleteCommand(ngwafWorkspaceRuleRoot.CmdClause, data)
	ngwafWorkspaceRuleGet := workspaceRule.NewGetCommand(ngwafWorkspaceRuleRoot.CmdClause, data)
	ngwafWorkspaceRuleList := workspaceRule.NewListCommand(ngwafWorkspaceRuleRoot.CmdClause, data)
	ngwafWorkspaceRuleUpdate := workspaceRule.NewUpdateCommand(ngwafWorkspaceRuleRoot.CmdClause, data)
	ngwafWorkspaceSignalListRoot := wssignallistlist.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceSignalListCreate := wssignallistlist.NewCreateCommand(ngwafWorkspaceSignalListRoot.CmdClause, data)
	ngwafWorkspaceSignalListDelete := wssignallistlist.NewDeleteCommand(ngwafWorkspaceSignalListRoot.CmdClause, data)
	ngwafWorkspaceSignalListGet := wssignallistlist.NewGetCommand(ngwafWorkspaceSignalListRoot.CmdClause, data)
	ngwafWorkspaceSignalListList := wssignallistlist.NewListCommand(ngwafWorkspaceSignalListRoot.CmdClause, data)
	ngwafWorkspaceSignalListUpdate := wssignallistlist.NewUpdateCommand(ngwafWorkspaceSignalListRoot.CmdClause, data)
	ngwafWorkspaceStringListRoot := wsstringlistlist.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceStringListCreate := wsstringlistlist.NewCreateCommand(ngwafWorkspaceStringListRoot.CmdClause, data)
	ngwafWorkspaceStringListDelete := wsstringlistlist.NewDeleteCommand(ngwafWorkspaceStringListRoot.CmdClause, data)
	ngwafWorkspaceStringListGet := wsstringlistlist.NewGetCommand(ngwafWorkspaceStringListRoot.CmdClause, data)
	ngwafWorkspaceStringListList := wsstringlistlist.NewListCommand(ngwafWorkspaceStringListRoot.CmdClause, data)
	ngwafWorkspaceStringListUpdate := wsstringlistlist.NewUpdateCommand(ngwafWorkspaceStringListRoot.CmdClause, data)
	ngwafWorkspaceThresholdRoot := threshold.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceThresholdCreate := threshold.NewCreateCommand(ngwafWorkspaceThresholdRoot.CmdClause, data)
	ngwafWorkspaceThresholdDelete := threshold.NewDeleteCommand(ngwafWorkspaceThresholdRoot.CmdClause, data)
	ngwafWorkspaceThresholdGet := threshold.NewGetCommand(ngwafWorkspaceThresholdRoot.CmdClause, data)
	ngwafWorkspaceThresholdList := threshold.NewListCommand(ngwafWorkspaceThresholdRoot.CmdClause, data)
	ngwafWorkspaceThresholdUpdate := threshold.NewUpdateCommand(ngwafWorkspaceThresholdRoot.CmdClause, data)
	ngwafWorkspaceWildcardListRoot := wildcardlist.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceWildcardListCreate := wswildcardlistlist.NewCreateCommand(ngwafWorkspaceWildcardListRoot.CmdClause, data)
	ngwafWorkspaceWildcardListDelete := wswildcardlistlist.NewDeleteCommand(ngwafWorkspaceWildcardListRoot.CmdClause, data)
	ngwafWorkspaceWildcardListGet := wswildcardlistlist.NewGetCommand(ngwafWorkspaceWildcardListRoot.CmdClause, data)
	ngwafWorkspaceWildcardListList := wswildcardlistlist.NewListCommand(ngwafWorkspaceWildcardListRoot.CmdClause, data)
	ngwafWorkspaceWildcardListUpdate := wswildcardlistlist.NewUpdateCommand(ngwafWorkspaceWildcardListRoot.CmdClause, data)
	ngwafVirtualpatchRoot := virtualpatch.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafVirtualpatchList := virtualpatch.NewListCommand(ngwafVirtualpatchRoot.CmdClause, data)
	ngwafVirtualpatchUpdate := virtualpatch.NewUpdateCommand(ngwafVirtualpatchRoot.CmdClause, data)
	ngwafVirtualpatchRetrieve := virtualpatch.NewRetrieveCommand(ngwafVirtualpatchRoot.CmdClause, data)
	ngwafWorkspaceAlertRoot := alert.NewRootCommand(ngwafWorkspaceRoot.CmdClause, data)
	ngwafWorkspaceAlertDatadogRoot := workspaceAlertDatadog.NewRootCommand(ngwafWorkspaceAlertRoot.CmdClause, data)
	ngwafWorkspaceAlertDatadogCreate := workspaceAlertDatadog.NewCreateCommand(ngwafWorkspaceAlertDatadogRoot.CmdClause, data)
	ngwafWorkspaceAlertDatadogDelete := workspaceAlertDatadog.NewDeleteCommand(ngwafWorkspaceAlertDatadogRoot.CmdClause, data)
	ngwafWorkspaceAlertDatadogGet := workspaceAlertDatadog.NewGetCommand(ngwafWorkspaceAlertDatadogRoot.CmdClause, data)
	ngwafWorkspaceAlertDatadogList := workspaceAlertDatadog.NewListCommand(ngwafWorkspaceAlertDatadogRoot.CmdClause, data)
	ngwafWorkspaceAlertDatadogUpdate := workspaceAlertDatadog.NewUpdateCommand(ngwafWorkspaceAlertDatadogRoot.CmdClause, data)
	ngwafWorkspaceAlertJiraRoot := workspaceAlertJira.NewRootCommand(ngwafWorkspaceAlertRoot.CmdClause, data)
	ngwafWorkspaceAlertJiraCreate := workspaceAlertJira.NewCreateCommand(ngwafWorkspaceAlertJiraRoot.CmdClause, data)
	ngwafWorkspaceAlertJiraDelete := workspaceAlertJira.NewDeleteCommand(ngwafWorkspaceAlertJiraRoot.CmdClause, data)
	ngwafWorkspaceAlertJiraGet := workspaceAlertJira.NewGetCommand(ngwafWorkspaceAlertJiraRoot.CmdClause, data)
	ngwafWorkspaceAlertJiraList := workspaceAlertJira.NewListCommand(ngwafWorkspaceAlertJiraRoot.CmdClause, data)
	ngwafWorkspaceAlertJiraUpdate := workspaceAlertJira.NewUpdateCommand(ngwafWorkspaceAlertJiraRoot.CmdClause, data)
	ngwafWorkspaceAlertMailinglistRoot := workspaceAlertMailinglist.NewRootCommand(ngwafWorkspaceAlertRoot.CmdClause, data)
	ngwafWorkspaceAlertMailinglistCreate := workspaceAlertMailinglist.NewCreateCommand(ngwafWorkspaceAlertMailinglistRoot.CmdClause, data)
	ngwafWorkspaceAlertMailinglistDelete := workspaceAlertMailinglist.NewDeleteCommand(ngwafWorkspaceAlertMailinglistRoot.CmdClause, data)
	ngwafWorkspaceAlertMailinglistGet := workspaceAlertMailinglist.NewGetCommand(ngwafWorkspaceAlertMailinglistRoot.CmdClause, data)
	ngwafWorkspaceAlertMailinglistList := workspaceAlertMailinglist.NewListCommand(ngwafWorkspaceAlertMailinglistRoot.CmdClause, data)
	ngwafWorkspaceAlertMailinglistUpdate := workspaceAlertMailinglist.NewUpdateCommand(ngwafWorkspaceAlertMailinglistRoot.CmdClause, data)
	ngwafWorkspaceAlertMicrosoftteamsRoot := workspaceAlertMicrosoftteams.NewRootCommand(ngwafWorkspaceAlertRoot.CmdClause, data)
	ngwafWorkspaceAlertMicrosoftteamsCreate := workspaceAlertMicrosoftteams.NewCreateCommand(ngwafWorkspaceAlertMicrosoftteamsRoot.CmdClause, data)
	ngwafWorkspaceAlertMicrosoftteamsDelete := workspaceAlertMicrosoftteams.NewDeleteCommand(ngwafWorkspaceAlertMicrosoftteamsRoot.CmdClause, data)
	ngwafWorkspaceAlertMicrosoftteamsGet := workspaceAlertMicrosoftteams.NewGetCommand(ngwafWorkspaceAlertMicrosoftteamsRoot.CmdClause, data)
	ngwafWorkspaceAlertMicrosoftteamsList := workspaceAlertMicrosoftteams.NewListCommand(ngwafWorkspaceAlertMicrosoftteamsRoot.CmdClause, data)
	ngwafWorkspaceAlertMicrosoftteamsUpdate := workspaceAlertMicrosoftteams.NewUpdateCommand(ngwafWorkspaceAlertMicrosoftteamsRoot.CmdClause, data)
	ngwafWorkspaceAlertOpsgenieRoot := workspaceAlertOpsgenie.NewRootCommand(ngwafWorkspaceAlertRoot.CmdClause, data)
	ngwafWorkspaceAlertOpsgenieCreate := workspaceAlertOpsgenie.NewCreateCommand(ngwafWorkspaceAlertOpsgenieRoot.CmdClause, data)
	ngwafWorkspaceAlertOpsgenieDelete := workspaceAlertOpsgenie.NewDeleteCommand(ngwafWorkspaceAlertOpsgenieRoot.CmdClause, data)
	ngwafWorkspaceAlertOpsgenieGet := workspaceAlertOpsgenie.NewGetCommand(ngwafWorkspaceAlertOpsgenieRoot.CmdClause, data)
	ngwafWorkspaceAlertOpsgenieList := workspaceAlertOpsgenie.NewListCommand(ngwafWorkspaceAlertOpsgenieRoot.CmdClause, data)
	ngwafWorkspaceAlertOpsgenieUpdate := workspaceAlertOpsgenie.NewUpdateCommand(ngwafWorkspaceAlertOpsgenieRoot.CmdClause, data)
	ngwafWorkspaceAlertPagerdutyRoot := workspaceAlertPagerduty.NewRootCommand(ngwafWorkspaceAlertRoot.CmdClause, data)
	ngwafWorkspaceAlertPagerdutyCreate := workspaceAlertPagerduty.NewCreateCommand(ngwafWorkspaceAlertPagerdutyRoot.CmdClause, data)
	ngwafWorkspaceAlertPagerdutyDelete := workspaceAlertPagerduty.NewDeleteCommand(ngwafWorkspaceAlertPagerdutyRoot.CmdClause, data)
	ngwafWorkspaceAlertPagerdutyGet := workspaceAlertPagerduty.NewGetCommand(ngwafWorkspaceAlertPagerdutyRoot.CmdClause, data)
	ngwafWorkspaceAlertPagerdutyList := workspaceAlertPagerduty.NewListCommand(ngwafWorkspaceAlertPagerdutyRoot.CmdClause, data)
	ngwafWorkspaceAlertPagerdutyUpdate := workspaceAlertPagerduty.NewUpdateCommand(ngwafWorkspaceAlertPagerdutyRoot.CmdClause, data)
	ngwafWorkspaceAlertSlackRoot := workspaceAlertSlack.NewRootCommand(ngwafWorkspaceAlertRoot.CmdClause, data)
	ngwafWorkspaceAlertSlackCreate := workspaceAlertSlack.NewCreateCommand(ngwafWorkspaceAlertSlackRoot.CmdClause, data)
	ngwafWorkspaceAlertSlackDelete := workspaceAlertSlack.NewDeleteCommand(ngwafWorkspaceAlertSlackRoot.CmdClause, data)
	ngwafWorkspaceAlertSlackGet := workspaceAlertSlack.NewGetCommand(ngwafWorkspaceAlertSlackRoot.CmdClause, data)
	ngwafWorkspaceAlertSlackList := workspaceAlertSlack.NewListCommand(ngwafWorkspaceAlertSlackRoot.CmdClause, data)
	ngwafWorkspaceAlertSlackUpdate := workspaceAlertSlack.NewUpdateCommand(ngwafWorkspaceAlertSlackRoot.CmdClause, data)
	ngwafWorkspaceAlertWebhookRoot := workspaceAlertWebhook.NewRootCommand(ngwafWorkspaceAlertRoot.CmdClause, data)
	ngwafWorkspaceAlertWebhookCreate := workspaceAlertWebhook.NewCreateCommand(ngwafWorkspaceAlertWebhookRoot.CmdClause, data)
	ngwafWorkspaceAlertWebhookDelete := workspaceAlertWebhook.NewDeleteCommand(ngwafWorkspaceAlertWebhookRoot.CmdClause, data)
	ngwafWorkspaceAlertWebhookGet := workspaceAlertWebhook.NewGetCommand(ngwafWorkspaceAlertWebhookRoot.CmdClause, data)
	ngwafWorkspaceAlertWebhookGetSigningKey := workspaceAlertWebhook.NewGetSigningKeyCommand(ngwafWorkspaceAlertWebhookRoot.CmdClause, data)
	ngwafWorkspaceAlertWebhookList := workspaceAlertWebhook.NewListCommand(ngwafWorkspaceAlertWebhookRoot.CmdClause, data)
	ngwafWorkspaceAlertWebhookRotateSigningKey := workspaceAlertWebhook.NewRotateSigningKeyCommand(ngwafWorkspaceAlertWebhookRoot.CmdClause, data)
	ngwafWorkspaceAlertWebhookUpdate := workspaceAlertWebhook.NewUpdateCommand(ngwafWorkspaceAlertWebhookRoot.CmdClause, data)
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
	servicePurge := servicepurge.NewPurgeCommand(serviceCmdRoot.CmdClause, data)
	servicealertCmdRoot := servicealert.NewRootCommand(serviceCmdRoot.CmdClause, data)
	servicealertCreate := servicealert.NewCreateCommand(servicealertCmdRoot.CmdClause, data)
	servicealertDelete := servicealert.NewDeleteCommand(servicealertCmdRoot.CmdClause, data)
	servicealertDescribe := servicealert.NewDescribeCommand(servicealertCmdRoot.CmdClause, data)
	servicealertList := servicealert.NewListCommand(servicealertCmdRoot.CmdClause, data)
	servicealertListHistory := servicealert.NewListHistoryCommand(servicealertCmdRoot.CmdClause, data)
	servicealertUpdate := servicealert.NewUpdateCommand(servicealertCmdRoot.CmdClause, data)
	serviceaclCmdRoot := serviceacl.NewRootCommand(serviceCmdRoot.CmdClause, data)
	serviceaclCreate := serviceacl.NewCreateCommand(serviceaclCmdRoot.CmdClause, data)
	serviceaclDelete := serviceacl.NewDeleteCommand(serviceaclCmdRoot.CmdClause, data)
	serviceaclDescribe := serviceacl.NewDescribeCommand(serviceaclCmdRoot.CmdClause, data)
	serviceaclList := serviceacl.NewListCommand(serviceaclCmdRoot.CmdClause, data)
	serviceaclUpdate := serviceacl.NewUpdateCommand(serviceaclCmdRoot.CmdClause, data)
	serviceaclentryCmdRoot := serviceaclentry.NewRootCommand(serviceCmdRoot.CmdClause, data)
	serviceaclentryCreate := serviceaclentry.NewCreateCommand(serviceaclentryCmdRoot.CmdClause, data)
	serviceaclentryDelete := serviceaclentry.NewDeleteCommand(serviceaclentryCmdRoot.CmdClause, data)
	serviceaclentryDescribe := serviceaclentry.NewDescribeCommand(serviceaclentryCmdRoot.CmdClause, data)
	serviceaclentryList := serviceaclentry.NewListCommand(serviceaclentryCmdRoot.CmdClause, data)
	serviceaclentryUpdate := serviceaclentry.NewUpdateCommand(serviceaclentryCmdRoot.CmdClause, data)
	serviceauthCmdRoot := serviceauth.NewRootCommand(app, data)
	serviceauthCreate := serviceauth.NewCreateCommand(serviceauthCmdRoot.CmdClause, data)
	serviceauthDelete := serviceauth.NewDeleteCommand(serviceauthCmdRoot.CmdClause, data)
	serviceauthDescribe := serviceauth.NewDescribeCommand(serviceauthCmdRoot.CmdClause, data)
	serviceauthList := serviceauth.NewListCommand(serviceauthCmdRoot.CmdClause, data)
	serviceauthUpdate := serviceauth.NewUpdateCommand(serviceauthCmdRoot.CmdClause, data)
	servicedictionaryCmdRoot := servicedictionary.NewRootCommand(serviceCmdRoot.CmdClause, data)
	servicedictionaryCreate := servicedictionary.NewCreateCommand(servicedictionaryCmdRoot.CmdClause, data)
	servicedictionaryDelete := servicedictionary.NewDeleteCommand(servicedictionaryCmdRoot.CmdClause, data)
	servicedictionaryDescribe := servicedictionary.NewDescribeCommand(servicedictionaryCmdRoot.CmdClause, data)
	servicedictionaryList := servicedictionary.NewListCommand(servicedictionaryCmdRoot.CmdClause, data)
	servicedictionaryUpdate := servicedictionary.NewUpdateCommand(servicedictionaryCmdRoot.CmdClause, data)
	servicevclCmdRoot := servicevcl.NewRootCommand(serviceCmdRoot.CmdClause, data)
	servicevclDescribe := servicevcl.NewDescribeCommand(servicevclCmdRoot.CmdClause, data)
	servicevclConditionCmdRoot := servicevclcondition.NewRootCommand(servicevclCmdRoot.CmdClause, data)
	servicevclConditionCreate := servicevclcondition.NewCreateCommand(servicevclConditionCmdRoot.CmdClause, data)
	servicevclConditionDelete := servicevclcondition.NewDeleteCommand(servicevclConditionCmdRoot.CmdClause, data)
	servicevclConditionDescribe := servicevclcondition.NewDescribeCommand(servicevclConditionCmdRoot.CmdClause, data)
	servicevclConditionList := servicevclcondition.NewListCommand(servicevclConditionCmdRoot.CmdClause, data)
	servicevclConditionUpdate := servicevclcondition.NewUpdateCommand(servicevclConditionCmdRoot.CmdClause, data)
	servicevclCustomCmdRoot := servicevclcustom.NewRootCommand(servicevclCmdRoot.CmdClause, data)
	servicevclCustomCreate := servicevclcustom.NewCreateCommand(servicevclCustomCmdRoot.CmdClause, data)
	servicevclCustomDelete := servicevclcustom.NewDeleteCommand(servicevclCustomCmdRoot.CmdClause, data)
	servicevclCustomDescribe := servicevclcustom.NewDescribeCommand(servicevclCustomCmdRoot.CmdClause, data)
	servicevclCustomList := servicevclcustom.NewListCommand(servicevclCustomCmdRoot.CmdClause, data)
	servicevclCustomUpdate := servicevclcustom.NewUpdateCommand(servicevclCustomCmdRoot.CmdClause, data)
	servicevclSnippetCmdRoot := servicevclsnippet.NewRootCommand(servicevclCmdRoot.CmdClause, data)
	servicevclSnippetCreate := servicevclsnippet.NewCreateCommand(servicevclSnippetCmdRoot.CmdClause, data)
	servicevclSnippetDelete := servicevclsnippet.NewDeleteCommand(servicevclSnippetCmdRoot.CmdClause, data)
	servicevclSnippetDescribe := servicevclsnippet.NewDescribeCommand(servicevclSnippetCmdRoot.CmdClause, data)
	servicevclSnippetList := servicevclsnippet.NewListCommand(servicevclSnippetCmdRoot.CmdClause, data)
	servicevclSnippetUpdate := servicevclsnippet.NewUpdateCommand(servicevclSnippetCmdRoot.CmdClause, data)
	serviceVersionCmdRoot := serviceversion.NewRootCommand(serviceCmdRoot.CmdClause, data)
	serviceVersionActivate := serviceversion.NewActivateCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionClone := serviceversion.NewCloneCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionDeactivate := serviceversion.NewDeactivateCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionList := serviceversion.NewListCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionLock := serviceversion.NewLockCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionStage := serviceversion.NewStageCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionUnstage := serviceversion.NewUnstageCommand(serviceVersionCmdRoot.CmdClause, data)
	serviceVersionUpdate := serviceversion.NewUpdateCommand(serviceVersionCmdRoot.CmdClause, data)
	servicedomainCmdRoot := servicedomain.NewRootCommand(serviceCmdRoot.CmdClause, data)
	servicedomainCreate := servicedomain.NewCreateCommand(servicedomainCmdRoot.CmdClause, data)
	servicedomainDelete := servicedomain.NewDeleteCommand(servicedomainCmdRoot.CmdClause, data)
	servicedomainDescribe := servicedomain.NewDescribeCommand(servicedomainCmdRoot.CmdClause, data)
	servicedomainList := servicedomain.NewListCommand(servicedomainCmdRoot.CmdClause, data)
	servicedomainUpdate := servicedomain.NewUpdateCommand(servicedomainCmdRoot.CmdClause, data)
	servicedomainValidate := servicedomain.NewValidateCommand(servicedomainCmdRoot.CmdClause, data)
	servicedictionaryentryCmdRoot := servicedictionaryentry.NewRootCommand(serviceCmdRoot.CmdClause, data)
	servicedictionaryentryCreate := servicedictionaryentry.NewCreateCommand(servicedictionaryentryCmdRoot.CmdClause, data)
	servicedictionaryentryDelete := servicedictionaryentry.NewDeleteCommand(servicedictionaryentryCmdRoot.CmdClause, data)
	servicedictionaryentryDescribe := servicedictionaryentry.NewDescribeCommand(servicedictionaryentryCmdRoot.CmdClause, data)
	servicedictionaryentryList := servicedictionaryentry.NewListCommand(servicedictionaryentryCmdRoot.CmdClause, data)
	servicedictionaryentryUpdate := servicedictionaryentry.NewUpdateCommand(servicedictionaryentryCmdRoot.CmdClause, data)
	servicebackendCmdRoot := servicebackend.NewRootCommand(serviceCmdRoot.CmdClause, data)
	servicebackendCreate := servicebackend.NewCreateCommand(servicebackendCmdRoot.CmdClause, data)
	servicebackendDelete := servicebackend.NewDeleteCommand(servicebackendCmdRoot.CmdClause, data)
	servicebackendDescribe := servicebackend.NewDescribeCommand(servicebackendCmdRoot.CmdClause, data)
	servicebackendList := servicebackend.NewListCommand(servicebackendCmdRoot.CmdClause, data)
	servicebackendUpdate := servicebackend.NewUpdateCommand(servicebackendCmdRoot.CmdClause, data)
	servicehealthcheckCmdRoot := servicehealthcheck.NewRootCommand(serviceCmdRoot.CmdClause, data)
	servicehealthcheckCreate := servicehealthcheck.NewCreateCommand(servicehealthcheckCmdRoot.CmdClause, data)
	servicehealthcheckDelete := servicehealthcheck.NewDeleteCommand(servicehealthcheckCmdRoot.CmdClause, data)
	servicehealthcheckDescribe := servicehealthcheck.NewDescribeCommand(servicehealthcheckCmdRoot.CmdClause, data)
	servicehealthcheckList := servicehealthcheck.NewListCommand(servicehealthcheckCmdRoot.CmdClause, data)
	servicehealthcheckUpdate := servicehealthcheck.NewUpdateCommand(servicehealthcheckCmdRoot.CmdClause, data)
	serviceimageoptimizerdefaultsCmdRoot := serviceimageoptimizerdefaults.NewRootCommand(serviceCmdRoot.CmdClause, data)
	serviceimageoptimizerdefaultsGet := serviceimageoptimizerdefaults.NewGetCommand(serviceimageoptimizerdefaultsCmdRoot.CmdClause, data)
	serviceimageoptimizerdefaultsUpdate := serviceimageoptimizerdefaults.NewUpdateCommand(serviceimageoptimizerdefaultsCmdRoot.CmdClause, data)
	serviceratelimitCmdRoot := serviceratelimit.NewRootCommand(serviceCmdRoot.CmdClause, data)
	serviceratelimitCreate := serviceratelimit.NewCreateCommand(serviceratelimitCmdRoot.CmdClause, data)
	serviceratelimitDelete := serviceratelimit.NewDeleteCommand(serviceratelimitCmdRoot.CmdClause, data)
	serviceratelimitDescribe := serviceratelimit.NewDescribeCommand(serviceratelimitCmdRoot.CmdClause, data)
	serviceratelimitList := serviceratelimit.NewListCommand(serviceratelimitCmdRoot.CmdClause, data)
	serviceratelimitUpdate := serviceratelimit.NewUpdateCommand(serviceratelimitCmdRoot.CmdClause, data)
	serviceresourcelinkCmdRoot := serviceresourcelink.NewRootCommand(serviceCmdRoot.CmdClause, data)
	serviceresourcelinkCreate := serviceresourcelink.NewCreateCommand(serviceresourcelinkCmdRoot.CmdClause, data)
	serviceresourcelinkDelete := serviceresourcelink.NewDeleteCommand(serviceresourcelinkCmdRoot.CmdClause, data)
	serviceresourcelinkDescribe := serviceresourcelink.NewDescribeCommand(serviceresourcelinkCmdRoot.CmdClause, data)
	serviceresourcelinkList := serviceresourcelink.NewListCommand(serviceresourcelinkCmdRoot.CmdClause, data)
	serviceresourcelinkUpdate := serviceresourcelink.NewUpdateCommand(serviceresourcelinkCmdRoot.CmdClause, data)
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
	toolsCmdRoot := tools.NewRootCommand(app, data)
	toolsDomainCmdRoot := domainTools.NewRootCommand(toolsCmdRoot.CmdClause, data)
	toolsDomainStatus := domainTools.NewDomainStatusCommand(toolsDomainCmdRoot.CmdClause, data)
	toolsDomainSuggestions := domainTools.NewDomainSuggestionsCommand(toolsDomainCmdRoot.CmdClause, data)
	updateRoot := update.NewRootCommand(app, data)
	userCmdRoot := user.NewRootCommand(app, data)
	userCreate := user.NewCreateCommand(userCmdRoot.CmdClause, data)
	userDelete := user.NewDeleteCommand(userCmdRoot.CmdClause, data)
	userDescribe := user.NewDescribeCommand(userCmdRoot.CmdClause, data)
	userList := user.NewListCommand(userCmdRoot.CmdClause, data)
	userUpdate := user.NewUpdateCommand(userCmdRoot.CmdClause, data)
	versionCmdRoot := version.NewRootCommand(app, data)
	whoamiCmdRoot := whoami.NewRootCommand(app, data)

	// Aliases for deprecated commands
	aliasBackendRoot := aliasbackend.NewRootCommand(app, data)
	aliasBackendCreate := aliasbackend.NewCreateCommand(aliasBackendRoot.CmdClause, data)
	aliasBackendDelete := aliasbackend.NewDeleteCommand(aliasBackendRoot.CmdClause, data)
	aliasBackendDescribe := aliasbackend.NewDescribeCommand(aliasBackendRoot.CmdClause, data)
	aliasBackendList := aliasbackend.NewListCommand(aliasBackendRoot.CmdClause, data)
	aliasBackendUpdate := aliasbackend.NewUpdateCommand(aliasBackendRoot.CmdClause, data)
	aliasDictionaryEntryRoot := aliasdictionaryentry.NewRootCommand(app, data)
	aliasDictionaryEntryCreate := aliasdictionaryentry.NewCreateCommand(aliasDictionaryEntryRoot.CmdClause, data)
	aliasDictionaryEntryDelete := aliasdictionaryentry.NewDeleteCommand(aliasDictionaryEntryRoot.CmdClause, data)
	aliasDictionaryEntryDescribe := aliasdictionaryentry.NewDescribeCommand(aliasDictionaryEntryRoot.CmdClause, data)
	aliasDictionaryEntryList := aliasdictionaryentry.NewListCommand(aliasDictionaryEntryRoot.CmdClause, data)
	aliasDictionaryEntryUpdate := aliasdictionaryentry.NewUpdateCommand(aliasDictionaryEntryRoot.CmdClause, data)
	aliasDictionaryRoot := aliasdictionary.NewRootCommand(app, data)
	aliasDictionaryCreate := aliasdictionary.NewCreateCommand(aliasDictionaryRoot.CmdClause, data)
	aliasDictionaryDelete := aliasdictionary.NewDeleteCommand(aliasDictionaryRoot.CmdClause, data)
	aliasDictionaryDescribe := aliasdictionary.NewDescribeCommand(aliasDictionaryRoot.CmdClause, data)
	aliasDictionaryList := aliasdictionary.NewListCommand(aliasDictionaryRoot.CmdClause, data)
	aliasDictionaryUpdate := aliasdictionary.NewUpdateCommand(aliasDictionaryRoot.CmdClause, data)
	aliasHealthcheckRoot := aliashealthcheck.NewRootCommand(app, data)
	aliasHealthcheckCreate := aliashealthcheck.NewCreateCommand(aliasHealthcheckRoot.CmdClause, data)
	aliasHealthcheckDelete := aliashealthcheck.NewDeleteCommand(aliasHealthcheckRoot.CmdClause, data)
	aliasHealthcheckDescribe := aliashealthcheck.NewDescribeCommand(aliasHealthcheckRoot.CmdClause, data)
	aliasHealthcheckList := aliashealthcheck.NewListCommand(aliasHealthcheckRoot.CmdClause, data)
	aliasHealthcheckUpdate := aliashealthcheck.NewUpdateCommand(aliasHealthcheckRoot.CmdClause, data)
	aliasimageoptimizerdefaultsRoot := aliasimageoptimizerdefaults.NewRootCommand(app, data)
	aliasimageoptimizerdefaultsGet := aliasimageoptimizerdefaults.NewGetCommand(aliasimageoptimizerdefaultsRoot.CmdClause, data)
	aliasimageoptimizerdefaultsUpdate := aliasimageoptimizerdefaults.NewUpdateCommand(aliasimageoptimizerdefaultsRoot.CmdClause, data)
	aliasPurge := aliaspurge.NewCommand(app, data)
	aliasAlertRoot := aliasalerts.NewRootCommand(app, data)
	aliasAlertCreate := aliasalerts.NewCreateCommand(aliasAlertRoot.CmdClause, data)
	aliasAlertDelete := aliasalerts.NewDeleteCommand(aliasAlertRoot.CmdClause, data)
	aliasAlertDescribe := aliasalerts.NewDescribeCommand(aliasAlertRoot.CmdClause, data)
	aliasAlertList := aliasalerts.NewListCommand(aliasAlertRoot.CmdClause, data)
	aliasAlertListHistory := aliasalerts.NewListHistoryCommand(aliasAlertRoot.CmdClause, data)
	aliasAlertUpdate := aliasalerts.NewUpdateCommand(aliasAlertRoot.CmdClause, data)
	aliasACLRoot := aliasacl.NewRootCommand(app, data)
	aliasACLCreate := aliasacl.NewCreateCommand(aliasACLRoot.CmdClause, data)
	aliasACLDelete := aliasacl.NewDeleteCommand(aliasACLRoot.CmdClause, data)
	aliasACLDescribe := aliasacl.NewDescribeCommand(aliasACLRoot.CmdClause, data)
	aliasACLList := aliasacl.NewListCommand(aliasACLRoot.CmdClause, data)
	aliasACLUpdate := aliasacl.NewUpdateCommand(aliasACLRoot.CmdClause, data)
	aliasACLEntryRoot := aliasaclentry.NewRootCommand(app, data)
	aliasACLEntryCreate := aliasaclentry.NewCreateCommand(aliasACLEntryRoot.CmdClause, data)
	aliasACLEntryDelete := aliasaclentry.NewDeleteCommand(aliasACLEntryRoot.CmdClause, data)
	aliasACLEntryDescribe := aliasaclentry.NewDescribeCommand(aliasACLEntryRoot.CmdClause, data)
	aliasACLEntryList := aliasaclentry.NewListCommand(aliasACLEntryRoot.CmdClause, data)
	aliasACLEntryUpdate := aliasaclentry.NewUpdateCommand(aliasACLEntryRoot.CmdClause, data)
	aliasRateLimitRoot := aliasratelimit.NewRootCommand(app, data)
	aliasRateLimitCreate := aliasratelimit.NewCreateCommand(aliasRateLimitRoot.CmdClause, data)
	aliasRateLimitDelete := aliasratelimit.NewDeleteCommand(aliasRateLimitRoot.CmdClause, data)
	aliasRateLimitDescribe := aliasratelimit.NewDescribeCommand(aliasRateLimitRoot.CmdClause, data)
	aliasRateLimitList := aliasratelimit.NewListCommand(aliasRateLimitRoot.CmdClause, data)
	aliasRateLimitUpdate := aliasratelimit.NewUpdateCommand(aliasRateLimitRoot.CmdClause, data)
	aliasResourceLinkRoot := aliasresourcelink.NewRootCommand(app, data)
	aliasResourceLinkCreate := aliasresourcelink.NewCreateCommand(aliasResourceLinkRoot.CmdClause, data)
	aliasResourceLinkDelete := aliasresourcelink.NewDeleteCommand(aliasResourceLinkRoot.CmdClause, data)
	aliasResourceLinkDescribe := aliasresourcelink.NewDescribeCommand(aliasResourceLinkRoot.CmdClause, data)
	aliasResourceLinkList := aliasresourcelink.NewListCommand(aliasResourceLinkRoot.CmdClause, data)
	aliasResourceLinkUpdate := aliasresourcelink.NewUpdateCommand(aliasResourceLinkRoot.CmdClause, data)
	aliasVclRoot := aliasvcl.NewRootCommand(app, data)
	aliasVclDescribe := aliasvcl.NewDescribeCommand(aliasVclRoot.CmdClause, data)
	aliasVclConditionRoot := aliasvclcondition.NewRootCommand(aliasVclRoot.CmdClause, data)
	aliasVclConditionCreate := aliasvclcondition.NewCreateCommand(aliasVclConditionRoot.CmdClause, data)
	aliasVclConditionDelete := aliasvclcondition.NewDeleteCommand(aliasVclConditionRoot.CmdClause, data)
	aliasVclConditionDescribe := aliasvclcondition.NewDescribeCommand(aliasVclConditionRoot.CmdClause, data)
	aliasVclConditionList := aliasvclcondition.NewListCommand(aliasVclConditionRoot.CmdClause, data)
	aliasVclConditionUpdate := aliasvclcondition.NewUpdateCommand(aliasVclConditionRoot.CmdClause, data)
	aliasVclCustomRoot := aliasvclcustom.NewRootCommand(aliasVclRoot.CmdClause, data)
	aliasVclCustomCreate := aliasvclcustom.NewCreateCommand(aliasVclCustomRoot.CmdClause, data)
	aliasVclCustomDelete := aliasvclcustom.NewDeleteCommand(aliasVclCustomRoot.CmdClause, data)
	aliasVclCustomDescribe := aliasvclcustom.NewDescribeCommand(aliasVclCustomRoot.CmdClause, data)
	aliasVclCustomList := aliasvclcustom.NewListCommand(aliasVclCustomRoot.CmdClause, data)
	aliasVclCustomUpdate := aliasvclcustom.NewUpdateCommand(aliasVclCustomRoot.CmdClause, data)
	aliasVclSnippetRoot := aliasvclsnippet.NewRootCommand(aliasVclRoot.CmdClause, data)
	aliasVclSnippetCreate := aliasvclsnippet.NewCreateCommand(aliasVclSnippetRoot.CmdClause, data)
	aliasVclSnippetDelete := aliasvclsnippet.NewDeleteCommand(aliasVclSnippetRoot.CmdClause, data)
	aliasVclSnippetDescribe := aliasvclsnippet.NewDescribeCommand(aliasVclSnippetRoot.CmdClause, data)
	aliasVclSnippetList := aliasvclsnippet.NewListCommand(aliasVclSnippetRoot.CmdClause, data)
	aliasVclSnippetUpdate := aliasvclsnippet.NewUpdateCommand(aliasVclSnippetRoot.CmdClause, data)
	aliasServiceVersionRoot := aliasserviceversion.NewRootCommand(app, data)
	aliasServiceVersionActivate := aliasserviceversion.NewActivateCommand(aliasServiceVersionRoot.CmdClause, data)
	aliasServiceVersionClone := aliasserviceversion.NewCloneCommand(aliasServiceVersionRoot.CmdClause, data)
	aliasServiceVersionDeactivate := aliasserviceversion.NewDeactivateCommand(aliasServiceVersionRoot.CmdClause, data)
	aliasServiceVersionList := aliasserviceversion.NewListCommand(aliasServiceVersionRoot.CmdClause, data)
	aliasServiceVersionLock := aliasserviceversion.NewLockCommand(aliasServiceVersionRoot.CmdClause, data)
	aliasServiceVersionStage := aliasserviceversion.NewStageCommand(aliasServiceVersionRoot.CmdClause, data)
	aliasServiceVersionUnstage := aliasserviceversion.NewUnstageCommand(aliasServiceVersionRoot.CmdClause, data)
	aliasServiceVersionUpdate := aliasserviceversion.NewUpdateCommand(aliasServiceVersionRoot.CmdClause, data)

	return []argparser.Command{
		shellcompleteCmdRoot,
		authtokenCmdRoot,
		authtokenCreate,
		authtokenDelete,
		authtokenDescribe,
		authtokenList,
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
		domainCmdRoot,
		domainCreate,
		domainDelete,
		domainDescribe,
		domainList,
		domainUpdate,
		installRoot,
		ipCmdRoot,
		kvstoreCreate,
		kvstoreDelete,
		kvstoreDescribe,
		kvstoreList,
		kvstoreentryCreate,
		kvstoreentryDelete,
		kvstoreentryGet,
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
		ngwafRoot,
		ngwafRedactionCreate,
		ngwafRedactionDelete,
		ngwafRedactionList,
		ngwafRedactionRetrieve,
		ngwafRedactionUpdate,
		ngwafRedactionRoot,
		ngwafCountryListRoot,
		ngwafCountryListCreate,
		ngwafCountryListDelete,
		ngwafCountryListGet,
		ngwafCountryListList,
		ngwafCountryListUpdate,
		ngwafCustomSignalRoot,
		ngwafCustomSignalCreate,
		ngwafCustomSignalDelete,
		ngwafCustomSignalGet,
		ngwafCustomSignalList,
		ngwafCustomSignalUpdate,
		ngwafIPListRoot,
		ngwafIPListCreate,
		ngwafIPListDelete,
		ngwafIPListGet,
		ngwafIPListList,
		ngwafIPListUpdate,
		ngwafRuleRoot,
		ngwafRuleCreate,
		ngwafRuleDelete,
		ngwafRuleGet,
		ngwafRuleList,
		ngwafRuleUpdate,
		ngwafSignalListRoot,
		ngwafSignalListCreate,
		ngwafSignalListDelete,
		ngwafSignalListGet,
		ngwafSignalListList,
		ngwafSignalListUpdate,
		ngwafStringListRoot,
		ngwafStringListCreate,
		ngwafStringListDelete,
		ngwafStringListGet,
		ngwafStringListList,
		ngwafStringListUpdate,
		ngwafWildcardListCreate,
		ngwafWildcardListDelete,
		ngwafWildcardListGet,
		ngwafWildcardListList,
		ngwafWildcardListUpdate,
		ngwafWorkspaceCountryListRoot,
		ngwafWorkspaceCountryListCreate,
		ngwafWorkspaceCountryListDelete,
		ngwafWorkspaceCountryListGet,
		ngwafWorkspaceCountryListList,
		ngwafWorkspaceCountryListUpdate,
		ngwafWorkspaceCustomSignalRoot,
		ngwafWorkspaceCustomSignalCreate,
		ngwafWorkspaceCustomSignalDelete,
		ngwafWorkspaceCustomSignalGet,
		ngwafWorkspaceCustomSignalList,
		ngwafWorkspaceCustomSignalUpdate,
		ngwafWorkspaceIPListRoot,
		ngwafWorkspaceIPListCreate,
		ngwafWorkspaceIPListDelete,
		ngwafWorkspaceIPListGet,
		ngwafWorkspaceIPListList,
		ngwafWorkspaceIPListUpdate,
		ngwafWorkspaceRuleRoot,
		ngwafWorkspaceRuleCreate,
		ngwafWorkspaceRuleDelete,
		ngwafWorkspaceRuleGet,
		ngwafWorkspaceRuleList,
		ngwafWorkspaceRuleUpdate,
		ngwafWorkspaceSignalListRoot,
		ngwafWorkspaceSignalListCreate,
		ngwafWorkspaceSignalListDelete,
		ngwafWorkspaceSignalListGet,
		ngwafWorkspaceSignalListList,
		ngwafWorkspaceSignalListUpdate,
		ngwafWorkspaceStringListRoot,
		ngwafWorkspaceStringListCreate,
		ngwafWorkspaceStringListDelete,
		ngwafWorkspaceStringListGet,
		ngwafWorkspaceStringListList,
		ngwafWorkspaceStringListUpdate,
		ngwafWorkspaceThresholdRoot,
		ngwafWorkspaceThresholdCreate,
		ngwafWorkspaceThresholdDelete,
		ngwafWorkspaceThresholdGet,
		ngwafWorkspaceThresholdList,
		ngwafWorkspaceThresholdUpdate,
		ngwafWorkspaceWildcardListCreate,
		ngwafWorkspaceWildcardListDelete,
		ngwafWorkspaceWildcardListGet,
		ngwafWorkspaceWildcardListList,
		ngwafWorkspaceWildcardListUpdate,
		ngwafVirtualpatchList,
		ngwafVirtualpatchRetrieve,
		ngwafVirtualpatchRoot,
		ngwafVirtualpatchUpdate,
		ngwafWorkspaceAlertRoot,
		ngwafWorkspaceAlertDatadogRoot,
		ngwafWorkspaceAlertDatadogCreate,
		ngwafWorkspaceAlertDatadogDelete,
		ngwafWorkspaceAlertDatadogGet,
		ngwafWorkspaceAlertDatadogList,
		ngwafWorkspaceAlertDatadogUpdate,
		ngwafWorkspaceAlertJiraRoot,
		ngwafWorkspaceAlertJiraCreate,
		ngwafWorkspaceAlertJiraDelete,
		ngwafWorkspaceAlertJiraGet,
		ngwafWorkspaceAlertJiraList,
		ngwafWorkspaceAlertJiraUpdate,
		ngwafWorkspaceAlertMailinglistRoot,
		ngwafWorkspaceAlertMailinglistCreate,
		ngwafWorkspaceAlertMailinglistDelete,
		ngwafWorkspaceAlertMailinglistGet,
		ngwafWorkspaceAlertMailinglistList,
		ngwafWorkspaceAlertMailinglistUpdate,
		ngwafWorkspaceAlertMicrosoftteamsRoot,
		ngwafWorkspaceAlertMicrosoftteamsCreate,
		ngwafWorkspaceAlertMicrosoftteamsDelete,
		ngwafWorkspaceAlertMicrosoftteamsGet,
		ngwafWorkspaceAlertMicrosoftteamsList,
		ngwafWorkspaceAlertMicrosoftteamsUpdate,
		ngwafWorkspaceAlertOpsgenieRoot,
		ngwafWorkspaceAlertOpsgenieCreate,
		ngwafWorkspaceAlertOpsgenieDelete,
		ngwafWorkspaceAlertOpsgenieGet,
		ngwafWorkspaceAlertOpsgenieList,
		ngwafWorkspaceAlertOpsgenieUpdate,
		ngwafWorkspaceAlertPagerdutyRoot,
		ngwafWorkspaceAlertPagerdutyCreate,
		ngwafWorkspaceAlertPagerdutyDelete,
		ngwafWorkspaceAlertPagerdutyGet,
		ngwafWorkspaceAlertPagerdutyList,
		ngwafWorkspaceAlertPagerdutyUpdate,
		ngwafWorkspaceAlertSlackRoot,
		ngwafWorkspaceAlertSlackCreate,
		ngwafWorkspaceAlertSlackDelete,
		ngwafWorkspaceAlertSlackGet,
		ngwafWorkspaceAlertSlackList,
		ngwafWorkspaceAlertSlackUpdate,
		ngwafWorkspaceAlertWebhookRoot,
		ngwafWorkspaceAlertWebhookCreate,
		ngwafWorkspaceAlertWebhookDelete,
		ngwafWorkspaceAlertWebhookGet,
		ngwafWorkspaceAlertWebhookGetSigningKey,
		ngwafWorkspaceAlertWebhookList,
		ngwafWorkspaceAlertWebhookRotateSigningKey,
		ngwafWorkspaceAlertWebhookUpdate,
		ngwafWorkspaceRoot,
		ngwafWorkspaceCreate,
		ngwafWorkspaceDelete,
		ngwafWorkspaceGet,
		ngwafWorkspaceList,
		ngwafWorkspaceUpdate,
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
		servicePurge,
		servicealertCreate,
		servicealertDelete,
		servicealertDescribe,
		servicealertList,
		servicealertListHistory,
		servicealertUpdate,
		serviceaclCmdRoot,
		serviceaclCreate,
		serviceaclDelete,
		serviceaclDescribe,
		serviceaclList,
		serviceaclUpdate,
		serviceaclentryCmdRoot,
		serviceaclentryCreate,
		serviceaclentryDelete,
		serviceaclentryDescribe,
		serviceaclentryList,
		serviceaclentryUpdate,
		serviceauthCmdRoot,
		serviceauthCreate,
		serviceauthDelete,
		serviceauthDescribe,
		serviceauthList,
		serviceauthUpdate,
		servicedictionaryCmdRoot,
		servicedictionaryCreate,
		servicedictionaryDelete,
		servicedictionaryDescribe,
		servicedictionaryList,
		servicedictionaryUpdate,
		servicevclCmdRoot,
		servicevclDescribe,
		servicevclConditionCmdRoot,
		servicevclConditionCreate,
		servicevclConditionDelete,
		servicevclConditionDescribe,
		servicevclConditionList,
		servicevclConditionUpdate,
		servicevclCustomCmdRoot,
		servicevclCustomCreate,
		servicevclCustomDelete,
		servicevclCustomDescribe,
		servicevclCustomList,
		servicevclCustomUpdate,
		servicevclSnippetCmdRoot,
		servicevclSnippetCreate,
		servicevclSnippetDelete,
		servicevclSnippetDescribe,
		servicevclSnippetList,
		servicevclSnippetUpdate,
		servicedomainCmdRoot,
		servicedomainCreate,
		servicedomainDelete,
		servicedomainDescribe,
		servicedomainList,
		servicedomainUpdate,
		servicedomainValidate,
		servicedictionaryentryCmdRoot,
		servicedictionaryentryCreate,
		servicedictionaryentryDelete,
		servicedictionaryentryDescribe,
		servicedictionaryentryList,
		servicedictionaryentryUpdate,
		servicebackendCmdRoot,
		servicebackendCreate,
		servicebackendDelete,
		servicebackendDescribe,
		servicebackendList,
		servicebackendUpdate,
		servicehealthcheckCmdRoot,
		servicehealthcheckCreate,
		servicehealthcheckDelete,
		servicehealthcheckDescribe,
		servicehealthcheckList,
		servicehealthcheckUpdate,
		serviceimageoptimizerdefaultsCmdRoot,
		serviceimageoptimizerdefaultsGet,
		serviceimageoptimizerdefaultsUpdate,
		serviceratelimitCmdRoot,
		serviceratelimitCreate,
		serviceratelimitDelete,
		serviceratelimitDescribe,
		serviceratelimitList,
		serviceratelimitUpdate,
		serviceresourcelinkCmdRoot,
		serviceresourcelinkCreate,
		serviceresourcelinkDelete,
		serviceresourcelinkDescribe,
		serviceresourcelinkList,
		serviceresourcelinkUpdate,
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
		toolsCmdRoot,
		toolsDomainCmdRoot,
		toolsDomainStatus,
		toolsDomainSuggestions,
		updateRoot,
		userCmdRoot,
		userCreate,
		userDelete,
		userDescribe,
		userList,
		userUpdate,
		versionCmdRoot,
		whoamiCmdRoot,
		aliasBackendCreate,
		aliasBackendDelete,
		aliasBackendDescribe,
		aliasBackendList,
		aliasBackendUpdate,
		aliasDictionaryEntryCreate,
		aliasDictionaryEntryDelete,
		aliasDictionaryEntryDescribe,
		aliasDictionaryEntryList,
		aliasDictionaryEntryUpdate,
		aliasDictionaryCreate,
		aliasDictionaryDelete,
		aliasDictionaryDescribe,
		aliasDictionaryList,
		aliasDictionaryUpdate,
		aliasHealthcheckCreate,
		aliasHealthcheckDelete,
		aliasHealthcheckDescribe,
		aliasHealthcheckList,
		aliasHealthcheckUpdate,
		aliasimageoptimizerdefaultsGet,
		aliasimageoptimizerdefaultsUpdate,
		aliasPurge,
		aliasAlertRoot,
		aliasAlertCreate,
		aliasAlertDelete,
		aliasAlertDescribe,
		aliasAlertList,
		aliasAlertListHistory,
		aliasAlertUpdate,
		aliasACLCreate,
		aliasACLDelete,
		aliasACLDescribe,
		aliasACLList,
		aliasACLUpdate,
		aliasACLEntryCreate,
		aliasACLEntryDelete,
		aliasACLEntryDescribe,
		aliasACLEntryList,
		aliasACLEntryUpdate,
		aliasRateLimitCreate,
		aliasRateLimitDelete,
		aliasRateLimitDescribe,
		aliasRateLimitList,
		aliasRateLimitUpdate,
		aliasResourceLinkCreate,
		aliasResourceLinkDelete,
		aliasResourceLinkDescribe,
		aliasResourceLinkList,
		aliasResourceLinkUpdate,
		aliasVclDescribe,
		aliasVclConditionCreate,
		aliasVclConditionDelete,
		aliasVclConditionDescribe,
		aliasVclConditionList,
		aliasVclConditionUpdate,
		aliasVclCustomCreate,
		aliasVclCustomDelete,
		aliasVclCustomDescribe,
		aliasVclCustomList,
		aliasVclCustomUpdate,
		aliasVclSnippetCreate,
		aliasVclSnippetDelete,
		aliasVclSnippetDescribe,
		aliasVclSnippetList,
		aliasVclSnippetUpdate,
		aliasServiceVersionActivate,
		aliasServiceVersionClone,
		aliasServiceVersionDeactivate,
		aliasServiceVersionList,
		aliasServiceVersionLock,
		aliasServiceVersionStage,
		aliasServiceVersionUnstage,
		aliasServiceVersionUpdate,
	}
}
