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
	aliaslogging "github.com/fastly/cli/pkg/commands/alias/logging"
	aliasazureblob "github.com/fastly/cli/pkg/commands/alias/logging/azureblob"
	aliasbigquery "github.com/fastly/cli/pkg/commands/alias/logging/bigquery"
	aliascloudfiles "github.com/fastly/cli/pkg/commands/alias/logging/cloudfiles"
	aliasdatadog "github.com/fastly/cli/pkg/commands/alias/logging/datadog"
	aliasdigitalocean "github.com/fastly/cli/pkg/commands/alias/logging/digitalocean"
	aliaselasticsearch "github.com/fastly/cli/pkg/commands/alias/logging/elasticsearch"
	aliasftp "github.com/fastly/cli/pkg/commands/alias/logging/ftp"
	aliasgcs "github.com/fastly/cli/pkg/commands/alias/logging/gcs"
	aliasgooglepubsub "github.com/fastly/cli/pkg/commands/alias/logging/googlepubsub"
	aliasgrafanacloudlogs "github.com/fastly/cli/pkg/commands/alias/logging/grafanacloudlogs"
	aliasheroku "github.com/fastly/cli/pkg/commands/alias/logging/heroku"
	aliashoneycomb "github.com/fastly/cli/pkg/commands/alias/logging/honeycomb"
	aliashttps "github.com/fastly/cli/pkg/commands/alias/logging/https"
	aliaskafka "github.com/fastly/cli/pkg/commands/alias/logging/kafka"
	aliaskinesis "github.com/fastly/cli/pkg/commands/alias/logging/kinesis"
	aliasloggly "github.com/fastly/cli/pkg/commands/alias/logging/loggly"
	aliaslogshuttle "github.com/fastly/cli/pkg/commands/alias/logging/logshuttle"
	aliasnewrelic "github.com/fastly/cli/pkg/commands/alias/logging/newrelic"
	aliasnewrelicotlp "github.com/fastly/cli/pkg/commands/alias/logging/newrelicotlp"
	aliasopenstack "github.com/fastly/cli/pkg/commands/alias/logging/openstack"
	aliaspapertrail "github.com/fastly/cli/pkg/commands/alias/logging/papertrail"
	aliass3 "github.com/fastly/cli/pkg/commands/alias/logging/s3"
	aliasscalyr "github.com/fastly/cli/pkg/commands/alias/logging/scalyr"
	aliassftp "github.com/fastly/cli/pkg/commands/alias/logging/sftp"
	aliassplunk "github.com/fastly/cli/pkg/commands/alias/logging/splunk"
	aliassumologic "github.com/fastly/cli/pkg/commands/alias/logging/sumologic"
	aliassyslog "github.com/fastly/cli/pkg/commands/alias/logging/syslog"
	aliaspurge "github.com/fastly/cli/pkg/commands/alias/purge"
	aliasratelimit "github.com/fastly/cli/pkg/commands/alias/ratelimit"
	aliasresourcelink "github.com/fastly/cli/pkg/commands/alias/resourcelink"
	aliasserviceauth "github.com/fastly/cli/pkg/commands/alias/serviceauth"
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
	serviceauth "github.com/fastly/cli/pkg/commands/service/auth"
	servicebackend "github.com/fastly/cli/pkg/commands/service/backend"
	servicedictionary "github.com/fastly/cli/pkg/commands/service/dictionary"
	servicedictionaryentry "github.com/fastly/cli/pkg/commands/service/dictionaryentry"
	servicedomain "github.com/fastly/cli/pkg/commands/service/domain"
	servicehealthcheck "github.com/fastly/cli/pkg/commands/service/healthcheck"
	serviceimageoptimizerdefaults "github.com/fastly/cli/pkg/commands/service/imageoptimizerdefaults"
	servicelogging "github.com/fastly/cli/pkg/commands/service/logging"
	serviceloggingazureblob "github.com/fastly/cli/pkg/commands/service/logging/azureblob"
	serviceloggingbigquery "github.com/fastly/cli/pkg/commands/service/logging/bigquery"
	serviceloggingcloudfiles "github.com/fastly/cli/pkg/commands/service/logging/cloudfiles"
	serviceloggingdatadog "github.com/fastly/cli/pkg/commands/service/logging/datadog"
	serviceloggingdigitalocean "github.com/fastly/cli/pkg/commands/service/logging/digitalocean"
	serviceloggingelasticsearch "github.com/fastly/cli/pkg/commands/service/logging/elasticsearch"
	serviceloggingftp "github.com/fastly/cli/pkg/commands/service/logging/ftp"
	servicelogginggcs "github.com/fastly/cli/pkg/commands/service/logging/gcs"
	servicelogginggooglepubsub "github.com/fastly/cli/pkg/commands/service/logging/googlepubsub"
	servicelogginggrafanacloudlogs "github.com/fastly/cli/pkg/commands/service/logging/grafanacloudlogs"
	serviceloggingheroku "github.com/fastly/cli/pkg/commands/service/logging/heroku"
	servicelogginghoneycomb "github.com/fastly/cli/pkg/commands/service/logging/honeycomb"
	servicelogginghttps "github.com/fastly/cli/pkg/commands/service/logging/https"
	serviceloggingkafka "github.com/fastly/cli/pkg/commands/service/logging/kafka"
	serviceloggingkinesis "github.com/fastly/cli/pkg/commands/service/logging/kinesis"
	serviceloggingloggly "github.com/fastly/cli/pkg/commands/service/logging/loggly"
	servicelogginglogshuttle "github.com/fastly/cli/pkg/commands/service/logging/logshuttle"
	serviceloggingnewrelic "github.com/fastly/cli/pkg/commands/service/logging/newrelic"
	serviceloggingnewrelicotlp "github.com/fastly/cli/pkg/commands/service/logging/newrelicotlp"
	serviceloggingopenstack "github.com/fastly/cli/pkg/commands/service/logging/openstack"
	serviceloggingpapertrail "github.com/fastly/cli/pkg/commands/service/logging/papertrail"
	serviceloggings3 "github.com/fastly/cli/pkg/commands/service/logging/s3"
	serviceloggingscalyr "github.com/fastly/cli/pkg/commands/service/logging/scalyr"
	serviceloggingsftp "github.com/fastly/cli/pkg/commands/service/logging/sftp"
	serviceloggingsplunk "github.com/fastly/cli/pkg/commands/service/logging/splunk"
	serviceloggingsumologic "github.com/fastly/cli/pkg/commands/service/logging/sumologic"
	serviceloggingsyslog "github.com/fastly/cli/pkg/commands/service/logging/syslog"
	servicepurge "github.com/fastly/cli/pkg/commands/service/purge"
	serviceratelimit "github.com/fastly/cli/pkg/commands/service/ratelimit"
	serviceresourcelink "github.com/fastly/cli/pkg/commands/service/resourcelink"
	servicevcl "github.com/fastly/cli/pkg/commands/service/vcl"
	servicevclcondition "github.com/fastly/cli/pkg/commands/service/vcl/condition"
	servicevclcustom "github.com/fastly/cli/pkg/commands/service/vcl/custom"
	servicevclsnippet "github.com/fastly/cli/pkg/commands/service/vcl/snippet"
	serviceversion "github.com/fastly/cli/pkg/commands/service/version"
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
	serviceauthCmdRoot := serviceauth.NewRootCommand(serviceCmdRoot.CmdClause, data)
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
	serviceloggingCmdRoot := servicelogging.NewRootCommand(serviceCmdRoot.CmdClause, data)
	serviceloggingAzureblobCmdRoot := serviceloggingazureblob.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingAzureblobCreate := serviceloggingazureblob.NewCreateCommand(serviceloggingAzureblobCmdRoot.CmdClause, data)
	serviceloggingAzureblobDelete := serviceloggingazureblob.NewDeleteCommand(serviceloggingAzureblobCmdRoot.CmdClause, data)
	serviceloggingAzureblobDescribe := serviceloggingazureblob.NewDescribeCommand(serviceloggingAzureblobCmdRoot.CmdClause, data)
	serviceloggingAzureblobList := serviceloggingazureblob.NewListCommand(serviceloggingAzureblobCmdRoot.CmdClause, data)
	serviceloggingAzureblobUpdate := serviceloggingazureblob.NewUpdateCommand(serviceloggingAzureblobCmdRoot.CmdClause, data)
	serviceloggingBigQueryCmdRoot := serviceloggingbigquery.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingBigQueryCreate := serviceloggingbigquery.NewCreateCommand(serviceloggingBigQueryCmdRoot.CmdClause, data)
	serviceloggingBigQueryDelete := serviceloggingbigquery.NewDeleteCommand(serviceloggingBigQueryCmdRoot.CmdClause, data)
	serviceloggingBigQueryDescribe := serviceloggingbigquery.NewDescribeCommand(serviceloggingBigQueryCmdRoot.CmdClause, data)
	serviceloggingBigQueryList := serviceloggingbigquery.NewListCommand(serviceloggingBigQueryCmdRoot.CmdClause, data)
	serviceloggingBigQueryUpdate := serviceloggingbigquery.NewUpdateCommand(serviceloggingBigQueryCmdRoot.CmdClause, data)
	serviceloggingCloudfilesCmdRoot := serviceloggingcloudfiles.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingCloudfilesCreate := serviceloggingcloudfiles.NewCreateCommand(serviceloggingCloudfilesCmdRoot.CmdClause, data)
	serviceloggingCloudfilesDelete := serviceloggingcloudfiles.NewDeleteCommand(serviceloggingCloudfilesCmdRoot.CmdClause, data)
	serviceloggingCloudfilesDescribe := serviceloggingcloudfiles.NewDescribeCommand(serviceloggingCloudfilesCmdRoot.CmdClause, data)
	serviceloggingCloudfilesList := serviceloggingcloudfiles.NewListCommand(serviceloggingCloudfilesCmdRoot.CmdClause, data)
	serviceloggingCloudfilesUpdate := serviceloggingcloudfiles.NewUpdateCommand(serviceloggingCloudfilesCmdRoot.CmdClause, data)
	serviceloggingDatadogCmdRoot := serviceloggingdatadog.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingDatadogCreate := serviceloggingdatadog.NewCreateCommand(serviceloggingDatadogCmdRoot.CmdClause, data)
	serviceloggingDatadogDelete := serviceloggingdatadog.NewDeleteCommand(serviceloggingDatadogCmdRoot.CmdClause, data)
	serviceloggingDatadogDescribe := serviceloggingdatadog.NewDescribeCommand(serviceloggingDatadogCmdRoot.CmdClause, data)
	serviceloggingDatadogList := serviceloggingdatadog.NewListCommand(serviceloggingDatadogCmdRoot.CmdClause, data)
	serviceloggingDatadogUpdate := serviceloggingdatadog.NewUpdateCommand(serviceloggingDatadogCmdRoot.CmdClause, data)
	serviceloggingDigitaloceanCmdRoot := serviceloggingdigitalocean.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingDigitaloceanCreate := serviceloggingdigitalocean.NewCreateCommand(serviceloggingDigitaloceanCmdRoot.CmdClause, data)
	serviceloggingDigitaloceanDelete := serviceloggingdigitalocean.NewDeleteCommand(serviceloggingDigitaloceanCmdRoot.CmdClause, data)
	serviceloggingDigitaloceanDescribe := serviceloggingdigitalocean.NewDescribeCommand(serviceloggingDigitaloceanCmdRoot.CmdClause, data)
	serviceloggingDigitaloceanList := serviceloggingdigitalocean.NewListCommand(serviceloggingDigitaloceanCmdRoot.CmdClause, data)
	serviceloggingDigitaloceanUpdate := serviceloggingdigitalocean.NewUpdateCommand(serviceloggingDigitaloceanCmdRoot.CmdClause, data)
	serviceloggingElasticsearchCmdRoot := serviceloggingelasticsearch.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingElasticsearchCreate := serviceloggingelasticsearch.NewCreateCommand(serviceloggingElasticsearchCmdRoot.CmdClause, data)
	serviceloggingElasticsearchDelete := serviceloggingelasticsearch.NewDeleteCommand(serviceloggingElasticsearchCmdRoot.CmdClause, data)
	serviceloggingElasticsearchDescribe := serviceloggingelasticsearch.NewDescribeCommand(serviceloggingElasticsearchCmdRoot.CmdClause, data)
	serviceloggingElasticsearchList := serviceloggingelasticsearch.NewListCommand(serviceloggingElasticsearchCmdRoot.CmdClause, data)
	serviceloggingElasticsearchUpdate := serviceloggingelasticsearch.NewUpdateCommand(serviceloggingElasticsearchCmdRoot.CmdClause, data)
	serviceloggingFtpCmdRoot := serviceloggingftp.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingFtpCreate := serviceloggingftp.NewCreateCommand(serviceloggingFtpCmdRoot.CmdClause, data)
	serviceloggingFtpDelete := serviceloggingftp.NewDeleteCommand(serviceloggingFtpCmdRoot.CmdClause, data)
	serviceloggingFtpDescribe := serviceloggingftp.NewDescribeCommand(serviceloggingFtpCmdRoot.CmdClause, data)
	serviceloggingFtpList := serviceloggingftp.NewListCommand(serviceloggingFtpCmdRoot.CmdClause, data)
	serviceloggingFtpUpdate := serviceloggingftp.NewUpdateCommand(serviceloggingFtpCmdRoot.CmdClause, data)
	serviceloggingGcsCmdRoot := servicelogginggcs.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingGcsCreate := servicelogginggcs.NewCreateCommand(serviceloggingGcsCmdRoot.CmdClause, data)
	serviceloggingGcsDelete := servicelogginggcs.NewDeleteCommand(serviceloggingGcsCmdRoot.CmdClause, data)
	serviceloggingGcsDescribe := servicelogginggcs.NewDescribeCommand(serviceloggingGcsCmdRoot.CmdClause, data)
	serviceloggingGcsList := servicelogginggcs.NewListCommand(serviceloggingGcsCmdRoot.CmdClause, data)
	serviceloggingGcsUpdate := servicelogginggcs.NewUpdateCommand(serviceloggingGcsCmdRoot.CmdClause, data)
	serviceloggingGooglepubsubCmdRoot := servicelogginggooglepubsub.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingGooglepubsubCreate := servicelogginggooglepubsub.NewCreateCommand(serviceloggingGooglepubsubCmdRoot.CmdClause, data)
	serviceloggingGooglepubsubDelete := servicelogginggooglepubsub.NewDeleteCommand(serviceloggingGooglepubsubCmdRoot.CmdClause, data)
	serviceloggingGooglepubsubDescribe := servicelogginggooglepubsub.NewDescribeCommand(serviceloggingGooglepubsubCmdRoot.CmdClause, data)
	serviceloggingGooglepubsubList := servicelogginggooglepubsub.NewListCommand(serviceloggingGooglepubsubCmdRoot.CmdClause, data)
	serviceloggingGooglepubsubUpdate := servicelogginggooglepubsub.NewUpdateCommand(serviceloggingGooglepubsubCmdRoot.CmdClause, data)
	serviceloggingGrafanacloudlogsCmdRoot := servicelogginggrafanacloudlogs.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingGrafanacloudlogsCreate := servicelogginggrafanacloudlogs.NewCreateCommand(serviceloggingGrafanacloudlogsCmdRoot.CmdClause, data)
	serviceloggingGrafanacloudlogsDelete := servicelogginggrafanacloudlogs.NewDeleteCommand(serviceloggingGrafanacloudlogsCmdRoot.CmdClause, data)
	serviceloggingGrafanacloudlogsDescribe := servicelogginggrafanacloudlogs.NewDescribeCommand(serviceloggingGrafanacloudlogsCmdRoot.CmdClause, data)
	serviceloggingGrafanacloudlogsList := servicelogginggrafanacloudlogs.NewListCommand(serviceloggingGrafanacloudlogsCmdRoot.CmdClause, data)
	serviceloggingGrafanacloudlogsUpdate := servicelogginggrafanacloudlogs.NewUpdateCommand(serviceloggingGrafanacloudlogsCmdRoot.CmdClause, data)
	serviceloggingHerokuCmdRoot := serviceloggingheroku.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingHerokuCreate := serviceloggingheroku.NewCreateCommand(serviceloggingHerokuCmdRoot.CmdClause, data)
	serviceloggingHerokuDelete := serviceloggingheroku.NewDeleteCommand(serviceloggingHerokuCmdRoot.CmdClause, data)
	serviceloggingHerokuDescribe := serviceloggingheroku.NewDescribeCommand(serviceloggingHerokuCmdRoot.CmdClause, data)
	serviceloggingHerokuList := serviceloggingheroku.NewListCommand(serviceloggingHerokuCmdRoot.CmdClause, data)
	serviceloggingHerokuUpdate := serviceloggingheroku.NewUpdateCommand(serviceloggingHerokuCmdRoot.CmdClause, data)
	serviceloggingHoneycombCmdRoot := servicelogginghoneycomb.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingHoneycombCreate := servicelogginghoneycomb.NewCreateCommand(serviceloggingHoneycombCmdRoot.CmdClause, data)
	serviceloggingHoneycombDelete := servicelogginghoneycomb.NewDeleteCommand(serviceloggingHoneycombCmdRoot.CmdClause, data)
	serviceloggingHoneycombDescribe := servicelogginghoneycomb.NewDescribeCommand(serviceloggingHoneycombCmdRoot.CmdClause, data)
	serviceloggingHoneycombList := servicelogginghoneycomb.NewListCommand(serviceloggingHoneycombCmdRoot.CmdClause, data)
	serviceloggingHoneycombUpdate := servicelogginghoneycomb.NewUpdateCommand(serviceloggingHoneycombCmdRoot.CmdClause, data)
	serviceloggingHTTPSCmdRoot := servicelogginghttps.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingHTTPSCreate := servicelogginghttps.NewCreateCommand(serviceloggingHTTPSCmdRoot.CmdClause, data)
	serviceloggingHTTPSDelete := servicelogginghttps.NewDeleteCommand(serviceloggingHTTPSCmdRoot.CmdClause, data)
	serviceloggingHTTPSDescribe := servicelogginghttps.NewDescribeCommand(serviceloggingHTTPSCmdRoot.CmdClause, data)
	serviceloggingHTTPSList := servicelogginghttps.NewListCommand(serviceloggingHTTPSCmdRoot.CmdClause, data)
	serviceloggingHTTPSUpdate := servicelogginghttps.NewUpdateCommand(serviceloggingHTTPSCmdRoot.CmdClause, data)
	serviceloggingKafkaCmdRoot := serviceloggingkafka.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingKafkaCreate := serviceloggingkafka.NewCreateCommand(serviceloggingKafkaCmdRoot.CmdClause, data)
	serviceloggingKafkaDelete := serviceloggingkafka.NewDeleteCommand(serviceloggingKafkaCmdRoot.CmdClause, data)
	serviceloggingKafkaDescribe := serviceloggingkafka.NewDescribeCommand(serviceloggingKafkaCmdRoot.CmdClause, data)
	serviceloggingKafkaList := serviceloggingkafka.NewListCommand(serviceloggingKafkaCmdRoot.CmdClause, data)
	serviceloggingKafkaUpdate := serviceloggingkafka.NewUpdateCommand(serviceloggingKafkaCmdRoot.CmdClause, data)
	serviceloggingKinesisCmdRoot := serviceloggingkinesis.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingKinesisCreate := serviceloggingkinesis.NewCreateCommand(serviceloggingKinesisCmdRoot.CmdClause, data)
	serviceloggingKinesisDelete := serviceloggingkinesis.NewDeleteCommand(serviceloggingKinesisCmdRoot.CmdClause, data)
	serviceloggingKinesisDescribe := serviceloggingkinesis.NewDescribeCommand(serviceloggingKinesisCmdRoot.CmdClause, data)
	serviceloggingKinesisList := serviceloggingkinesis.NewListCommand(serviceloggingKinesisCmdRoot.CmdClause, data)
	serviceloggingKinesisUpdate := serviceloggingkinesis.NewUpdateCommand(serviceloggingKinesisCmdRoot.CmdClause, data)
	serviceloggingLogglyCmdRoot := serviceloggingloggly.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingLogglyCreate := serviceloggingloggly.NewCreateCommand(serviceloggingLogglyCmdRoot.CmdClause, data)
	serviceloggingLogglyDelete := serviceloggingloggly.NewDeleteCommand(serviceloggingLogglyCmdRoot.CmdClause, data)
	serviceloggingLogglyDescribe := serviceloggingloggly.NewDescribeCommand(serviceloggingLogglyCmdRoot.CmdClause, data)
	serviceloggingLogglyList := serviceloggingloggly.NewListCommand(serviceloggingLogglyCmdRoot.CmdClause, data)
	serviceloggingLogglyUpdate := serviceloggingloggly.NewUpdateCommand(serviceloggingLogglyCmdRoot.CmdClause, data)
	serviceloggingLogshuttleCmdRoot := servicelogginglogshuttle.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingLogshuttleCreate := servicelogginglogshuttle.NewCreateCommand(serviceloggingLogshuttleCmdRoot.CmdClause, data)
	serviceloggingLogshuttleDelete := servicelogginglogshuttle.NewDeleteCommand(serviceloggingLogshuttleCmdRoot.CmdClause, data)
	serviceloggingLogshuttleDescribe := servicelogginglogshuttle.NewDescribeCommand(serviceloggingLogshuttleCmdRoot.CmdClause, data)
	serviceloggingLogshuttleList := servicelogginglogshuttle.NewListCommand(serviceloggingLogshuttleCmdRoot.CmdClause, data)
	serviceloggingLogshuttleUpdate := servicelogginglogshuttle.NewUpdateCommand(serviceloggingLogshuttleCmdRoot.CmdClause, data)
	serviceloggingNewRelicCmdRoot := serviceloggingnewrelic.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingNewRelicCreate := serviceloggingnewrelic.NewCreateCommand(serviceloggingNewRelicCmdRoot.CmdClause, data)
	serviceloggingNewRelicDelete := serviceloggingnewrelic.NewDeleteCommand(serviceloggingNewRelicCmdRoot.CmdClause, data)
	serviceloggingNewRelicDescribe := serviceloggingnewrelic.NewDescribeCommand(serviceloggingNewRelicCmdRoot.CmdClause, data)
	serviceloggingNewRelicList := serviceloggingnewrelic.NewListCommand(serviceloggingNewRelicCmdRoot.CmdClause, data)
	serviceloggingNewRelicUpdate := serviceloggingnewrelic.NewUpdateCommand(serviceloggingNewRelicCmdRoot.CmdClause, data)
	serviceloggingNewRelicOTLPCmdRoot := serviceloggingnewrelicotlp.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingNewRelicOTLPCreate := serviceloggingnewrelicotlp.NewCreateCommand(serviceloggingNewRelicOTLPCmdRoot.CmdClause, data)
	serviceloggingNewRelicOTLPDelete := serviceloggingnewrelicotlp.NewDeleteCommand(serviceloggingNewRelicOTLPCmdRoot.CmdClause, data)
	serviceloggingNewRelicOTLPDescribe := serviceloggingnewrelicotlp.NewDescribeCommand(serviceloggingNewRelicOTLPCmdRoot.CmdClause, data)
	serviceloggingNewRelicOTLPList := serviceloggingnewrelicotlp.NewListCommand(serviceloggingNewRelicOTLPCmdRoot.CmdClause, data)
	serviceloggingNewRelicOTLPUpdate := serviceloggingnewrelicotlp.NewUpdateCommand(serviceloggingNewRelicOTLPCmdRoot.CmdClause, data)
	serviceloggingOpenstackCmdRoot := serviceloggingopenstack.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingOpenstackCreate := serviceloggingopenstack.NewCreateCommand(serviceloggingOpenstackCmdRoot.CmdClause, data)
	serviceloggingOpenstackDelete := serviceloggingopenstack.NewDeleteCommand(serviceloggingOpenstackCmdRoot.CmdClause, data)
	serviceloggingOpenstackDescribe := serviceloggingopenstack.NewDescribeCommand(serviceloggingOpenstackCmdRoot.CmdClause, data)
	serviceloggingOpenstackList := serviceloggingopenstack.NewListCommand(serviceloggingOpenstackCmdRoot.CmdClause, data)
	serviceloggingOpenstackUpdate := serviceloggingopenstack.NewUpdateCommand(serviceloggingOpenstackCmdRoot.CmdClause, data)
	serviceloggingPapertrailCmdRoot := serviceloggingpapertrail.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingPapertrailCreate := serviceloggingpapertrail.NewCreateCommand(serviceloggingPapertrailCmdRoot.CmdClause, data)
	serviceloggingPapertrailDelete := serviceloggingpapertrail.NewDeleteCommand(serviceloggingPapertrailCmdRoot.CmdClause, data)
	serviceloggingPapertrailDescribe := serviceloggingpapertrail.NewDescribeCommand(serviceloggingPapertrailCmdRoot.CmdClause, data)
	serviceloggingPapertrailList := serviceloggingpapertrail.NewListCommand(serviceloggingPapertrailCmdRoot.CmdClause, data)
	serviceloggingPapertrailUpdate := serviceloggingpapertrail.NewUpdateCommand(serviceloggingPapertrailCmdRoot.CmdClause, data)
	serviceloggingS3CmdRoot := serviceloggings3.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingS3Create := serviceloggings3.NewCreateCommand(serviceloggingS3CmdRoot.CmdClause, data)
	serviceloggingS3Delete := serviceloggings3.NewDeleteCommand(serviceloggingS3CmdRoot.CmdClause, data)
	serviceloggingS3Describe := serviceloggings3.NewDescribeCommand(serviceloggingS3CmdRoot.CmdClause, data)
	serviceloggingS3List := serviceloggings3.NewListCommand(serviceloggingS3CmdRoot.CmdClause, data)
	serviceloggingS3Update := serviceloggings3.NewUpdateCommand(serviceloggingS3CmdRoot.CmdClause, data)
	serviceloggingScalyrCmdRoot := serviceloggingscalyr.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingScalyrCreate := serviceloggingscalyr.NewCreateCommand(serviceloggingScalyrCmdRoot.CmdClause, data)
	serviceloggingScalyrDelete := serviceloggingscalyr.NewDeleteCommand(serviceloggingScalyrCmdRoot.CmdClause, data)
	serviceloggingScalyrDescribe := serviceloggingscalyr.NewDescribeCommand(serviceloggingScalyrCmdRoot.CmdClause, data)
	serviceloggingScalyrList := serviceloggingscalyr.NewListCommand(serviceloggingScalyrCmdRoot.CmdClause, data)
	serviceloggingScalyrUpdate := serviceloggingscalyr.NewUpdateCommand(serviceloggingScalyrCmdRoot.CmdClause, data)
	serviceloggingSftpCmdRoot := serviceloggingsftp.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingSftpCreate := serviceloggingsftp.NewCreateCommand(serviceloggingSftpCmdRoot.CmdClause, data)
	serviceloggingSftpDelete := serviceloggingsftp.NewDeleteCommand(serviceloggingSftpCmdRoot.CmdClause, data)
	serviceloggingSftpDescribe := serviceloggingsftp.NewDescribeCommand(serviceloggingSftpCmdRoot.CmdClause, data)
	serviceloggingSftpList := serviceloggingsftp.NewListCommand(serviceloggingSftpCmdRoot.CmdClause, data)
	serviceloggingSftpUpdate := serviceloggingsftp.NewUpdateCommand(serviceloggingSftpCmdRoot.CmdClause, data)
	serviceloggingSplunkCmdRoot := serviceloggingsplunk.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingSplunkCreate := serviceloggingsplunk.NewCreateCommand(serviceloggingSplunkCmdRoot.CmdClause, data)
	serviceloggingSplunkDelete := serviceloggingsplunk.NewDeleteCommand(serviceloggingSplunkCmdRoot.CmdClause, data)
	serviceloggingSplunkDescribe := serviceloggingsplunk.NewDescribeCommand(serviceloggingSplunkCmdRoot.CmdClause, data)
	serviceloggingSplunkList := serviceloggingsplunk.NewListCommand(serviceloggingSplunkCmdRoot.CmdClause, data)
	serviceloggingSplunkUpdate := serviceloggingsplunk.NewUpdateCommand(serviceloggingSplunkCmdRoot.CmdClause, data)
	serviceloggingSumologicCmdRoot := serviceloggingsumologic.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingSumologicCreate := serviceloggingsumologic.NewCreateCommand(serviceloggingSumologicCmdRoot.CmdClause, data)
	serviceloggingSumologicDelete := serviceloggingsumologic.NewDeleteCommand(serviceloggingSumologicCmdRoot.CmdClause, data)
	serviceloggingSumologicDescribe := serviceloggingsumologic.NewDescribeCommand(serviceloggingSumologicCmdRoot.CmdClause, data)
	serviceloggingSumologicList := serviceloggingsumologic.NewListCommand(serviceloggingSumologicCmdRoot.CmdClause, data)
	serviceloggingSumologicUpdate := serviceloggingsumologic.NewUpdateCommand(serviceloggingSumologicCmdRoot.CmdClause, data)
	serviceloggingSyslogCmdRoot := serviceloggingsyslog.NewRootCommand(serviceloggingCmdRoot.CmdClause, data)
	serviceloggingSyslogCreate := serviceloggingsyslog.NewCreateCommand(serviceloggingSyslogCmdRoot.CmdClause, data)
	serviceloggingSyslogDelete := serviceloggingsyslog.NewDeleteCommand(serviceloggingSyslogCmdRoot.CmdClause, data)
	serviceloggingSyslogDescribe := serviceloggingsyslog.NewDescribeCommand(serviceloggingSyslogCmdRoot.CmdClause, data)
	serviceloggingSyslogList := serviceloggingsyslog.NewListCommand(serviceloggingSyslogCmdRoot.CmdClause, data)
	serviceloggingSyslogUpdate := serviceloggingsyslog.NewUpdateCommand(serviceloggingSyslogCmdRoot.CmdClause, data)
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
	aliasServiceAuthRoot := aliasserviceauth.NewRootCommand(app, data)
	aliasServiceAuthCreate := aliasserviceauth.NewCreateCommand(aliasServiceAuthRoot.CmdClause, data)
	aliasServiceAuthDelete := aliasserviceauth.NewDeleteCommand(aliasServiceAuthRoot.CmdClause, data)
	aliasServiceAuthDescribe := aliasserviceauth.NewDescribeCommand(aliasServiceAuthRoot.CmdClause, data)
	aliasServiceAuthList := aliasserviceauth.NewListCommand(aliasServiceAuthRoot.CmdClause, data)
	aliasServiceAuthUpdate := aliasserviceauth.NewUpdateCommand(aliasServiceAuthRoot.CmdClause, data)
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
	aliasLoggingRoot := aliaslogging.NewRootCommand(app, data)
	aliasAzureblobRoot := aliasazureblob.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasAzureblobCreate := aliasazureblob.NewCreateCommand(aliasAzureblobRoot.CmdClause, data)
	aliasAzureblobDelete := aliasazureblob.NewDeleteCommand(aliasAzureblobRoot.CmdClause, data)
	aliasAzureblobDescribe := aliasazureblob.NewDescribeCommand(aliasAzureblobRoot.CmdClause, data)
	aliasAzureblobList := aliasazureblob.NewListCommand(aliasAzureblobRoot.CmdClause, data)
	aliasAzureblobUpdate := aliasazureblob.NewUpdateCommand(aliasAzureblobRoot.CmdClause, data)
	aliasBigqueryRoot := aliasbigquery.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasBigqueryCreate := aliasbigquery.NewCreateCommand(aliasBigqueryRoot.CmdClause, data)
	aliasBigqueryDelete := aliasbigquery.NewDeleteCommand(aliasBigqueryRoot.CmdClause, data)
	aliasBigqueryDescribe := aliasbigquery.NewDescribeCommand(aliasBigqueryRoot.CmdClause, data)
	aliasBigqueryList := aliasbigquery.NewListCommand(aliasBigqueryRoot.CmdClause, data)
	aliasBigqueryUpdate := aliasbigquery.NewUpdateCommand(aliasBigqueryRoot.CmdClause, data)
	aliasCloudfilesRoot := aliascloudfiles.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasCloudfilesCreate := aliascloudfiles.NewCreateCommand(aliasCloudfilesRoot.CmdClause, data)
	aliasCloudfilesDelete := aliascloudfiles.NewDeleteCommand(aliasCloudfilesRoot.CmdClause, data)
	aliasCloudfilesDescribe := aliascloudfiles.NewDescribeCommand(aliasCloudfilesRoot.CmdClause, data)
	aliasCloudfilesList := aliascloudfiles.NewListCommand(aliasCloudfilesRoot.CmdClause, data)
	aliasCloudfilesUpdate := aliascloudfiles.NewUpdateCommand(aliasCloudfilesRoot.CmdClause, data)
	aliasDatadogRoot := aliasdatadog.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasDatadogCreate := aliasdatadog.NewCreateCommand(aliasDatadogRoot.CmdClause, data)
	aliasDatadogDelete := aliasdatadog.NewDeleteCommand(aliasDatadogRoot.CmdClause, data)
	aliasDatadogDescribe := aliasdatadog.NewDescribeCommand(aliasDatadogRoot.CmdClause, data)
	aliasDatadogList := aliasdatadog.NewListCommand(aliasDatadogRoot.CmdClause, data)
	aliasDatadogUpdate := aliasdatadog.NewUpdateCommand(aliasDatadogRoot.CmdClause, data)
	aliasDigitaloceanRoot := aliasdigitalocean.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasDigitaloceanCreate := aliasdigitalocean.NewCreateCommand(aliasDigitaloceanRoot.CmdClause, data)
	aliasDigitaloceanDelete := aliasdigitalocean.NewDeleteCommand(aliasDigitaloceanRoot.CmdClause, data)
	aliasDigitaloceanDescribe := aliasdigitalocean.NewDescribeCommand(aliasDigitaloceanRoot.CmdClause, data)
	aliasDigitaloceanList := aliasdigitalocean.NewListCommand(aliasDigitaloceanRoot.CmdClause, data)
	aliasDigitaloceanUpdate := aliasdigitalocean.NewUpdateCommand(aliasDigitaloceanRoot.CmdClause, data)
	aliasElasticsearchRoot := aliaselasticsearch.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasElasticsearchCreate := aliaselasticsearch.NewCreateCommand(aliasElasticsearchRoot.CmdClause, data)
	aliasElasticsearchDelete := aliaselasticsearch.NewDeleteCommand(aliasElasticsearchRoot.CmdClause, data)
	aliasElasticsearchDescribe := aliaselasticsearch.NewDescribeCommand(aliasElasticsearchRoot.CmdClause, data)
	aliasElasticsearchList := aliaselasticsearch.NewListCommand(aliasElasticsearchRoot.CmdClause, data)
	aliasElasticsearchUpdate := aliaselasticsearch.NewUpdateCommand(aliasElasticsearchRoot.CmdClause, data)
	aliasFtpRoot := aliasftp.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasFtpCreate := aliasftp.NewCreateCommand(aliasFtpRoot.CmdClause, data)
	aliasFtpDelete := aliasftp.NewDeleteCommand(aliasFtpRoot.CmdClause, data)
	aliasFtpDescribe := aliasftp.NewDescribeCommand(aliasFtpRoot.CmdClause, data)
	aliasFtpList := aliasftp.NewListCommand(aliasFtpRoot.CmdClause, data)
	aliasFtpUpdate := aliasftp.NewUpdateCommand(aliasFtpRoot.CmdClause, data)
	aliasGcsRoot := aliasgcs.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasGcsCreate := aliasgcs.NewCreateCommand(aliasGcsRoot.CmdClause, data)
	aliasGcsDelete := aliasgcs.NewDeleteCommand(aliasGcsRoot.CmdClause, data)
	aliasGcsDescribe := aliasgcs.NewDescribeCommand(aliasGcsRoot.CmdClause, data)
	aliasGcsList := aliasgcs.NewListCommand(aliasGcsRoot.CmdClause, data)
	aliasGcsUpdate := aliasgcs.NewUpdateCommand(aliasGcsRoot.CmdClause, data)
	aliasGooglepubsubRoot := aliasgooglepubsub.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasGooglepubsubCreate := aliasgooglepubsub.NewCreateCommand(aliasGooglepubsubRoot.CmdClause, data)
	aliasGooglepubsubDelete := aliasgooglepubsub.NewDeleteCommand(aliasGooglepubsubRoot.CmdClause, data)
	aliasGooglepubsubDescribe := aliasgooglepubsub.NewDescribeCommand(aliasGooglepubsubRoot.CmdClause, data)
	aliasGooglepubsubList := aliasgooglepubsub.NewListCommand(aliasGooglepubsubRoot.CmdClause, data)
	aliasGooglepubsubUpdate := aliasgooglepubsub.NewUpdateCommand(aliasGooglepubsubRoot.CmdClause, data)
	aliasGrafanacloudlogsRoot := aliasgrafanacloudlogs.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasGrafanacloudlogsCreate := aliasgrafanacloudlogs.NewCreateCommand(aliasGrafanacloudlogsRoot.CmdClause, data)
	aliasGrafanacloudlogsDelete := aliasgrafanacloudlogs.NewDeleteCommand(aliasGrafanacloudlogsRoot.CmdClause, data)
	aliasGrafanacloudlogsDescribe := aliasgrafanacloudlogs.NewDescribeCommand(aliasGrafanacloudlogsRoot.CmdClause, data)
	aliasGrafanacloudlogsList := aliasgrafanacloudlogs.NewListCommand(aliasGrafanacloudlogsRoot.CmdClause, data)
	aliasGrafanacloudlogsUpdate := aliasgrafanacloudlogs.NewUpdateCommand(aliasGrafanacloudlogsRoot.CmdClause, data)
	aliasHerokuRoot := aliasheroku.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasHerokuCreate := aliasheroku.NewCreateCommand(aliasHerokuRoot.CmdClause, data)
	aliasHerokuDelete := aliasheroku.NewDeleteCommand(aliasHerokuRoot.CmdClause, data)
	aliasHerokuDescribe := aliasheroku.NewDescribeCommand(aliasHerokuRoot.CmdClause, data)
	aliasHerokuList := aliasheroku.NewListCommand(aliasHerokuRoot.CmdClause, data)
	aliasHerokuUpdate := aliasheroku.NewUpdateCommand(aliasHerokuRoot.CmdClause, data)
	aliasHoneycombRoot := aliashoneycomb.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasHoneycombCreate := aliashoneycomb.NewCreateCommand(aliasHoneycombRoot.CmdClause, data)
	aliasHoneycombDelete := aliashoneycomb.NewDeleteCommand(aliasHoneycombRoot.CmdClause, data)
	aliasHoneycombDescribe := aliashoneycomb.NewDescribeCommand(aliasHoneycombRoot.CmdClause, data)
	aliasHoneycombList := aliashoneycomb.NewListCommand(aliasHoneycombRoot.CmdClause, data)
	aliasHoneycombUpdate := aliashoneycomb.NewUpdateCommand(aliasHoneycombRoot.CmdClause, data)
	aliasHttpsRoot := aliashttps.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasHttpsCreate := aliashttps.NewCreateCommand(aliasHttpsRoot.CmdClause, data)
	aliasHttpsDelete := aliashttps.NewDeleteCommand(aliasHttpsRoot.CmdClause, data)
	aliasHttpsDescribe := aliashttps.NewDescribeCommand(aliasHttpsRoot.CmdClause, data)
	aliasHttpsList := aliashttps.NewListCommand(aliasHttpsRoot.CmdClause, data)
	aliasHttpsUpdate := aliashttps.NewUpdateCommand(aliasHttpsRoot.CmdClause, data)
	aliasKafkaRoot := aliaskafka.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasKafkaCreate := aliaskafka.NewCreateCommand(aliasKafkaRoot.CmdClause, data)
	aliasKafkaDelete := aliaskafka.NewDeleteCommand(aliasKafkaRoot.CmdClause, data)
	aliasKafkaDescribe := aliaskafka.NewDescribeCommand(aliasKafkaRoot.CmdClause, data)
	aliasKafkaList := aliaskafka.NewListCommand(aliasKafkaRoot.CmdClause, data)
	aliasKafkaUpdate := aliaskafka.NewUpdateCommand(aliasKafkaRoot.CmdClause, data)
	aliasKinesisRoot := aliaskinesis.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasKinesisCreate := aliaskinesis.NewCreateCommand(aliasKinesisRoot.CmdClause, data)
	aliasKinesisDelete := aliaskinesis.NewDeleteCommand(aliasKinesisRoot.CmdClause, data)
	aliasKinesisDescribe := aliaskinesis.NewDescribeCommand(aliasKinesisRoot.CmdClause, data)
	aliasKinesisList := aliaskinesis.NewListCommand(aliasKinesisRoot.CmdClause, data)
	aliasKinesisUpdate := aliaskinesis.NewUpdateCommand(aliasKinesisRoot.CmdClause, data)
	aliasLogglyRoot := aliasloggly.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasLogglyCreate := aliasloggly.NewCreateCommand(aliasLogglyRoot.CmdClause, data)
	aliasLogglyDelete := aliasloggly.NewDeleteCommand(aliasLogglyRoot.CmdClause, data)
	aliasLogglyDescribe := aliasloggly.NewDescribeCommand(aliasLogglyRoot.CmdClause, data)
	aliasLogglyList := aliasloggly.NewListCommand(aliasLogglyRoot.CmdClause, data)
	aliasLogglyUpdate := aliasloggly.NewUpdateCommand(aliasLogglyRoot.CmdClause, data)
	aliasLogshuttleRoot := aliaslogshuttle.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasLogshuttleCreate := aliaslogshuttle.NewCreateCommand(aliasLogshuttleRoot.CmdClause, data)
	aliasLogshuttleDelete := aliaslogshuttle.NewDeleteCommand(aliasLogshuttleRoot.CmdClause, data)
	aliasLogshuttleDescribe := aliaslogshuttle.NewDescribeCommand(aliasLogshuttleRoot.CmdClause, data)
	aliasLogshuttleList := aliaslogshuttle.NewListCommand(aliasLogshuttleRoot.CmdClause, data)
	aliasLogshuttleUpdate := aliaslogshuttle.NewUpdateCommand(aliasLogshuttleRoot.CmdClause, data)
	aliasNewrelicRoot := aliasnewrelic.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasNewrelicCreate := aliasnewrelic.NewCreateCommand(aliasNewrelicRoot.CmdClause, data)
	aliasNewrelicDelete := aliasnewrelic.NewDeleteCommand(aliasNewrelicRoot.CmdClause, data)
	aliasNewrelicDescribe := aliasnewrelic.NewDescribeCommand(aliasNewrelicRoot.CmdClause, data)
	aliasNewrelicList := aliasnewrelic.NewListCommand(aliasNewrelicRoot.CmdClause, data)
	aliasNewrelicUpdate := aliasnewrelic.NewUpdateCommand(aliasNewrelicRoot.CmdClause, data)
	aliasNewrelicotlpRoot := aliasnewrelicotlp.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasNewrelicotlpCreate := aliasnewrelicotlp.NewCreateCommand(aliasNewrelicotlpRoot.CmdClause, data)
	aliasNewrelicotlpDelete := aliasnewrelicotlp.NewDeleteCommand(aliasNewrelicotlpRoot.CmdClause, data)
	aliasNewrelicotlpDescribe := aliasnewrelicotlp.NewDescribeCommand(aliasNewrelicotlpRoot.CmdClause, data)
	aliasNewrelicotlpList := aliasnewrelicotlp.NewListCommand(aliasNewrelicotlpRoot.CmdClause, data)
	aliasNewrelicotlpUpdate := aliasnewrelicotlp.NewUpdateCommand(aliasNewrelicotlpRoot.CmdClause, data)
	aliasOpenstackRoot := aliasopenstack.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasOpenstackCreate := aliasopenstack.NewCreateCommand(aliasOpenstackRoot.CmdClause, data)
	aliasOpenstackDelete := aliasopenstack.NewDeleteCommand(aliasOpenstackRoot.CmdClause, data)
	aliasOpenstackDescribe := aliasopenstack.NewDescribeCommand(aliasOpenstackRoot.CmdClause, data)
	aliasOpenstackList := aliasopenstack.NewListCommand(aliasOpenstackRoot.CmdClause, data)
	aliasOpenstackUpdate := aliasopenstack.NewUpdateCommand(aliasOpenstackRoot.CmdClause, data)
	aliasPapertrailRoot := aliaspapertrail.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasPapertrailCreate := aliaspapertrail.NewCreateCommand(aliasPapertrailRoot.CmdClause, data)
	aliasPapertrailDelete := aliaspapertrail.NewDeleteCommand(aliasPapertrailRoot.CmdClause, data)
	aliasPapertrailDescribe := aliaspapertrail.NewDescribeCommand(aliasPapertrailRoot.CmdClause, data)
	aliasPapertrailList := aliaspapertrail.NewListCommand(aliasPapertrailRoot.CmdClause, data)
	aliasPapertrailUpdate := aliaspapertrail.NewUpdateCommand(aliasPapertrailRoot.CmdClause, data)
	aliasS3Root := aliass3.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasS3Create := aliass3.NewCreateCommand(aliasS3Root.CmdClause, data)
	aliasS3Delete := aliass3.NewDeleteCommand(aliasS3Root.CmdClause, data)
	aliasS3Describe := aliass3.NewDescribeCommand(aliasS3Root.CmdClause, data)
	aliasS3List := aliass3.NewListCommand(aliasS3Root.CmdClause, data)
	aliasS3Update := aliass3.NewUpdateCommand(aliasS3Root.CmdClause, data)
	aliasScalyrRoot := aliasscalyr.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasScalyrCreate := aliasscalyr.NewCreateCommand(aliasScalyrRoot.CmdClause, data)
	aliasScalyrDelete := aliasscalyr.NewDeleteCommand(aliasScalyrRoot.CmdClause, data)
	aliasScalyrDescribe := aliasscalyr.NewDescribeCommand(aliasScalyrRoot.CmdClause, data)
	aliasScalyrList := aliasscalyr.NewListCommand(aliasScalyrRoot.CmdClause, data)
	aliasScalyrUpdate := aliasscalyr.NewUpdateCommand(aliasScalyrRoot.CmdClause, data)
	aliasSftpRoot := aliassftp.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasSftpCreate := aliassftp.NewCreateCommand(aliasSftpRoot.CmdClause, data)
	aliasSftpDelete := aliassftp.NewDeleteCommand(aliasSftpRoot.CmdClause, data)
	aliasSftpDescribe := aliassftp.NewDescribeCommand(aliasSftpRoot.CmdClause, data)
	aliasSftpList := aliassftp.NewListCommand(aliasSftpRoot.CmdClause, data)
	aliasSftpUpdate := aliassftp.NewUpdateCommand(aliasSftpRoot.CmdClause, data)
	aliasSplunkRoot := aliassplunk.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasSplunkCreate := aliassplunk.NewCreateCommand(aliasSplunkRoot.CmdClause, data)
	aliasSplunkDelete := aliassplunk.NewDeleteCommand(aliasSplunkRoot.CmdClause, data)
	aliasSplunkDescribe := aliassplunk.NewDescribeCommand(aliasSplunkRoot.CmdClause, data)
	aliasSplunkList := aliassplunk.NewListCommand(aliasSplunkRoot.CmdClause, data)
	aliasSplunkUpdate := aliassplunk.NewUpdateCommand(aliasSplunkRoot.CmdClause, data)
	aliasSumologicRoot := aliassumologic.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasSumologicCreate := aliassumologic.NewCreateCommand(aliasSumologicRoot.CmdClause, data)
	aliasSumologicDelete := aliassumologic.NewDeleteCommand(aliasSumologicRoot.CmdClause, data)
	aliasSumologicDescribe := aliassumologic.NewDescribeCommand(aliasSumologicRoot.CmdClause, data)
	aliasSumologicList := aliassumologic.NewListCommand(aliasSumologicRoot.CmdClause, data)
	aliasSumologicUpdate := aliassumologic.NewUpdateCommand(aliasSumologicRoot.CmdClause, data)
	aliasSyslogRoot := aliassyslog.NewRootCommand(aliasLoggingRoot.CmdClause, data)
	aliasSyslogCreate := aliassyslog.NewCreateCommand(aliasSyslogRoot.CmdClause, data)
	aliasSyslogDelete := aliassyslog.NewDeleteCommand(aliasSyslogRoot.CmdClause, data)
	aliasSyslogDescribe := aliassyslog.NewDescribeCommand(aliasSyslogRoot.CmdClause, data)
	aliasSyslogList := aliassyslog.NewListCommand(aliasSyslogRoot.CmdClause, data)
	aliasSyslogUpdate := aliassyslog.NewUpdateCommand(aliasSyslogRoot.CmdClause, data)

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
		serviceloggingAzureblobCmdRoot,
		serviceloggingAzureblobCreate,
		serviceloggingAzureblobDelete,
		serviceloggingAzureblobDescribe,
		serviceloggingAzureblobList,
		serviceloggingAzureblobUpdate,
		serviceloggingBigQueryCmdRoot,
		serviceloggingBigQueryCreate,
		serviceloggingBigQueryDelete,
		serviceloggingBigQueryDescribe,
		serviceloggingBigQueryList,
		serviceloggingBigQueryUpdate,
		serviceloggingCloudfilesCmdRoot,
		serviceloggingCloudfilesCreate,
		serviceloggingCloudfilesDelete,
		serviceloggingCloudfilesDescribe,
		serviceloggingCloudfilesList,
		serviceloggingCloudfilesUpdate,
		serviceloggingCmdRoot,
		serviceloggingDatadogCmdRoot,
		serviceloggingDatadogCreate,
		serviceloggingDatadogDelete,
		serviceloggingDatadogDescribe,
		serviceloggingDatadogList,
		serviceloggingDatadogUpdate,
		serviceloggingDigitaloceanCmdRoot,
		serviceloggingDigitaloceanCreate,
		serviceloggingDigitaloceanDelete,
		serviceloggingDigitaloceanDescribe,
		serviceloggingDigitaloceanList,
		serviceloggingDigitaloceanUpdate,
		serviceloggingElasticsearchCmdRoot,
		serviceloggingElasticsearchCreate,
		serviceloggingElasticsearchDelete,
		serviceloggingElasticsearchDescribe,
		serviceloggingElasticsearchList,
		serviceloggingElasticsearchUpdate,
		serviceloggingFtpCmdRoot,
		serviceloggingFtpCreate,
		serviceloggingFtpDelete,
		serviceloggingFtpDescribe,
		serviceloggingFtpList,
		serviceloggingFtpUpdate,
		serviceloggingGcsCmdRoot,
		serviceloggingGcsCreate,
		serviceloggingGcsDelete,
		serviceloggingGcsDescribe,
		serviceloggingGcsList,
		serviceloggingGcsUpdate,
		serviceloggingGooglepubsubCmdRoot,
		serviceloggingGooglepubsubCreate,
		serviceloggingGooglepubsubDelete,
		serviceloggingGooglepubsubDescribe,
		serviceloggingGooglepubsubList,
		serviceloggingGooglepubsubUpdate,
		serviceloggingGrafanacloudlogsCmdRoot,
		serviceloggingGrafanacloudlogsCreate,
		serviceloggingGrafanacloudlogsDelete,
		serviceloggingGrafanacloudlogsDescribe,
		serviceloggingGrafanacloudlogsList,
		serviceloggingGrafanacloudlogsUpdate,
		serviceloggingHerokuCmdRoot,
		serviceloggingHerokuCreate,
		serviceloggingHerokuDelete,
		serviceloggingHerokuDescribe,
		serviceloggingHerokuList,
		serviceloggingHerokuUpdate,
		serviceloggingHoneycombCmdRoot,
		serviceloggingHoneycombCreate,
		serviceloggingHoneycombDelete,
		serviceloggingHoneycombDescribe,
		serviceloggingHoneycombList,
		serviceloggingHoneycombUpdate,
		serviceloggingHTTPSCmdRoot,
		serviceloggingHTTPSCreate,
		serviceloggingHTTPSDelete,
		serviceloggingHTTPSDescribe,
		serviceloggingHTTPSList,
		serviceloggingHTTPSUpdate,
		serviceloggingKafkaCmdRoot,
		serviceloggingKafkaCreate,
		serviceloggingKafkaDelete,
		serviceloggingKafkaDescribe,
		serviceloggingKafkaList,
		serviceloggingKafkaUpdate,
		serviceloggingKinesisCmdRoot,
		serviceloggingKinesisCreate,
		serviceloggingKinesisDelete,
		serviceloggingKinesisDescribe,
		serviceloggingKinesisList,
		serviceloggingKinesisUpdate,
		serviceloggingLogglyCmdRoot,
		serviceloggingLogglyCreate,
		serviceloggingLogglyDelete,
		serviceloggingLogglyDescribe,
		serviceloggingLogglyList,
		serviceloggingLogglyUpdate,
		serviceloggingLogshuttleCmdRoot,
		serviceloggingLogshuttleCreate,
		serviceloggingLogshuttleDelete,
		serviceloggingLogshuttleDescribe,
		serviceloggingLogshuttleList,
		serviceloggingLogshuttleUpdate,
		serviceloggingNewRelicCmdRoot,
		serviceloggingNewRelicCreate,
		serviceloggingNewRelicDelete,
		serviceloggingNewRelicDescribe,
		serviceloggingNewRelicList,
		serviceloggingNewRelicUpdate,
		serviceloggingNewRelicOTLPCmdRoot,
		serviceloggingNewRelicOTLPCreate,
		serviceloggingNewRelicOTLPDelete,
		serviceloggingNewRelicOTLPDescribe,
		serviceloggingNewRelicOTLPList,
		serviceloggingNewRelicOTLPUpdate,
		serviceloggingOpenstackCmdRoot,
		serviceloggingOpenstackCreate,
		serviceloggingOpenstackDelete,
		serviceloggingOpenstackDescribe,
		serviceloggingOpenstackList,
		serviceloggingOpenstackUpdate,
		serviceloggingPapertrailCmdRoot,
		serviceloggingPapertrailCreate,
		serviceloggingPapertrailDelete,
		serviceloggingPapertrailDescribe,
		serviceloggingPapertrailList,
		serviceloggingPapertrailUpdate,
		serviceloggingS3CmdRoot,
		serviceloggingS3Create,
		serviceloggingS3Delete,
		serviceloggingS3Describe,
		serviceloggingS3List,
		serviceloggingS3Update,
		serviceloggingScalyrCmdRoot,
		serviceloggingScalyrCreate,
		serviceloggingScalyrDelete,
		serviceloggingScalyrDescribe,
		serviceloggingScalyrList,
		serviceloggingScalyrUpdate,
		serviceloggingSftpCmdRoot,
		serviceloggingSftpCreate,
		serviceloggingSftpDelete,
		serviceloggingSftpDescribe,
		serviceloggingSftpList,
		serviceloggingSftpUpdate,
		serviceloggingSplunkCmdRoot,
		serviceloggingSplunkCreate,
		serviceloggingSplunkDelete,
		serviceloggingSplunkDescribe,
		serviceloggingSplunkList,
		serviceloggingSplunkUpdate,
		serviceloggingSumologicCmdRoot,
		serviceloggingSumologicCreate,
		serviceloggingSumologicDelete,
		serviceloggingSumologicDescribe,
		serviceloggingSumologicList,
		serviceloggingSumologicUpdate,
		serviceloggingSyslogCmdRoot,
		serviceloggingSyslogCreate,
		serviceloggingSyslogDelete,
		serviceloggingSyslogDescribe,
		serviceloggingSyslogList,
		serviceloggingSyslogUpdate,
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
		aliasServiceAuthCreate,
		aliasServiceAuthDelete,
		aliasServiceAuthDescribe,
		aliasServiceAuthList,
		aliasServiceAuthUpdate,
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
		aliasLoggingRoot,
		aliasAzureblobRoot,
		aliasAzureblobCreate,
		aliasAzureblobDelete,
		aliasAzureblobDescribe,
		aliasAzureblobList,
		aliasAzureblobUpdate,
		aliasBigqueryRoot,
		aliasBigqueryCreate,
		aliasBigqueryDelete,
		aliasBigqueryDescribe,
		aliasBigqueryList,
		aliasBigqueryUpdate,
		aliasCloudfilesRoot,
		aliasCloudfilesCreate,
		aliasCloudfilesDelete,
		aliasCloudfilesDescribe,
		aliasCloudfilesList,
		aliasCloudfilesUpdate,
		aliasDatadogRoot,
		aliasDatadogCreate,
		aliasDatadogDelete,
		aliasDatadogDescribe,
		aliasDatadogList,
		aliasDatadogUpdate,
		aliasDigitaloceanRoot,
		aliasDigitaloceanCreate,
		aliasDigitaloceanDelete,
		aliasDigitaloceanDescribe,
		aliasDigitaloceanList,
		aliasDigitaloceanUpdate,
		aliasElasticsearchRoot,
		aliasElasticsearchCreate,
		aliasElasticsearchDelete,
		aliasElasticsearchDescribe,
		aliasElasticsearchList,
		aliasElasticsearchUpdate,
		aliasFtpRoot,
		aliasFtpCreate,
		aliasFtpDelete,
		aliasFtpDescribe,
		aliasFtpList,
		aliasFtpUpdate,
		aliasGcsRoot,
		aliasGcsCreate,
		aliasGcsDelete,
		aliasGcsDescribe,
		aliasGcsList,
		aliasGcsUpdate,
		aliasGooglepubsubRoot,
		aliasGooglepubsubCreate,
		aliasGooglepubsubDelete,
		aliasGooglepubsubDescribe,
		aliasGooglepubsubList,
		aliasGooglepubsubUpdate,
		aliasGrafanacloudlogsRoot,
		aliasGrafanacloudlogsCreate,
		aliasGrafanacloudlogsDelete,
		aliasGrafanacloudlogsDescribe,
		aliasGrafanacloudlogsList,
		aliasGrafanacloudlogsUpdate,
		aliasHerokuRoot,
		aliasHerokuCreate,
		aliasHerokuDelete,
		aliasHerokuDescribe,
		aliasHerokuList,
		aliasHerokuUpdate,
		aliasHoneycombRoot,
		aliasHoneycombCreate,
		aliasHoneycombDelete,
		aliasHoneycombDescribe,
		aliasHoneycombList,
		aliasHoneycombUpdate,
		aliasHttpsRoot,
		aliasHttpsCreate,
		aliasHttpsDelete,
		aliasHttpsDescribe,
		aliasHttpsList,
		aliasHttpsUpdate,
		aliasKafkaRoot,
		aliasKafkaCreate,
		aliasKafkaDelete,
		aliasKafkaDescribe,
		aliasKafkaList,
		aliasKafkaUpdate,
		aliasKinesisRoot,
		aliasKinesisCreate,
		aliasKinesisDelete,
		aliasKinesisDescribe,
		aliasKinesisList,
		aliasKinesisUpdate,
		aliasLogglyRoot,
		aliasLogglyCreate,
		aliasLogglyDelete,
		aliasLogglyDescribe,
		aliasLogglyList,
		aliasLogglyUpdate,
		aliasLogshuttleRoot,
		aliasLogshuttleCreate,
		aliasLogshuttleDelete,
		aliasLogshuttleDescribe,
		aliasLogshuttleList,
		aliasLogshuttleUpdate,
		aliasNewrelicRoot,
		aliasNewrelicCreate,
		aliasNewrelicDelete,
		aliasNewrelicDescribe,
		aliasNewrelicList,
		aliasNewrelicUpdate,
		aliasNewrelicotlpRoot,
		aliasNewrelicotlpCreate,
		aliasNewrelicotlpDelete,
		aliasNewrelicotlpDescribe,
		aliasNewrelicotlpList,
		aliasNewrelicotlpUpdate,
		aliasOpenstackRoot,
		aliasOpenstackCreate,
		aliasOpenstackDelete,
		aliasOpenstackDescribe,
		aliasOpenstackList,
		aliasOpenstackUpdate,
		aliasPapertrailRoot,
		aliasPapertrailCreate,
		aliasPapertrailDelete,
		aliasPapertrailDescribe,
		aliasPapertrailList,
		aliasPapertrailUpdate,
		aliasS3Root,
		aliasS3Create,
		aliasS3Delete,
		aliasS3Describe,
		aliasS3List,
		aliasS3Update,
		aliasScalyrRoot,
		aliasScalyrCreate,
		aliasScalyrDelete,
		aliasScalyrDescribe,
		aliasScalyrList,
		aliasScalyrUpdate,
		aliasSftpRoot,
		aliasSftpCreate,
		aliasSftpDelete,
		aliasSftpDescribe,
		aliasSftpList,
		aliasSftpUpdate,
		aliasSplunkRoot,
		aliasSplunkCreate,
		aliasSplunkDelete,
		aliasSplunkDescribe,
		aliasSplunkList,
		aliasSplunkUpdate,
		aliasSumologicRoot,
		aliasSumologicCreate,
		aliasSumologicDelete,
		aliasSumologicDescribe,
		aliasSumologicList,
		aliasSumologicUpdate,
		aliasSyslogRoot,
		aliasSyslogCreate,
		aliasSyslogDelete,
		aliasSyslogDescribe,
		aliasSyslogList,
		aliasSyslogUpdate,
	}
}
