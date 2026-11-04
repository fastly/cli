package commands

import (
	"github.com/alecthomas/kong"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/commands/airuntimecontrol"
	"github.com/fastly/cli/pkg/commands/alias/acl"
	"github.com/fastly/cli/pkg/commands/alias/aclentry"
	"github.com/fastly/cli/pkg/commands/alias/alerts"
	"github.com/fastly/cli/pkg/commands/alias/backend"
	"github.com/fastly/cli/pkg/commands/alias/dictionary"
	"github.com/fastly/cli/pkg/commands/alias/dictionaryentry"
	"github.com/fastly/cli/pkg/commands/alias/healthcheck"
	"github.com/fastly/cli/pkg/commands/alias/imageoptimizerdefaults"
	"github.com/fastly/cli/pkg/commands/alias/logging"
	"github.com/fastly/cli/pkg/commands/alias/purge"
	"github.com/fastly/cli/pkg/commands/alias/ratelimit"
	"github.com/fastly/cli/pkg/commands/alias/resourcelink"
	"github.com/fastly/cli/pkg/commands/alias/serviceauth"
	"github.com/fastly/cli/pkg/commands/alias/serviceversion"
	"github.com/fastly/cli/pkg/commands/alias/vcl"
	"github.com/fastly/cli/pkg/commands/apisecurity"
	"github.com/fastly/cli/pkg/commands/auditlog"
	"github.com/fastly/cli/pkg/commands/auth"
	"github.com/fastly/cli/pkg/commands/authtoken"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/commands/config"
	"github.com/fastly/cli/pkg/commands/configstore"
	"github.com/fastly/cli/pkg/commands/configstoreentry"
	"github.com/fastly/cli/pkg/commands/dashboard"
	"github.com/fastly/cli/pkg/commands/dns"
	"github.com/fastly/cli/pkg/commands/domain"
	"github.com/fastly/cli/pkg/commands/install"
	"github.com/fastly/cli/pkg/commands/integration"
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
	"github.com/fastly/cli/pkg/commands/ngwaf/timeseries"
	ngwafwildcardlist "github.com/fastly/cli/pkg/commands/ngwaf/wildcardlist"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/datadog"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/jira"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/mailinglist"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/microsoftteams"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/opsgenie"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/pagerduty"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/slack"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/alert/webhook"
	wscountrylist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/countrylist"
	wscustomsignal "github.com/fastly/cli/pkg/commands/ngwaf/workspace/customsignal"
	wsiplist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/iplist"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/redaction"
	workspaceRule "github.com/fastly/cli/pkg/commands/ngwaf/workspace/rule"
	wssignallistlist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/signallist"
	wsstringlistlist "github.com/fastly/cli/pkg/commands/ngwaf/workspace/stringlist"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/threshold"
	wstimeseries "github.com/fastly/cli/pkg/commands/ngwaf/workspace/timeseries"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/virtualpatch"
	"github.com/fastly/cli/pkg/commands/ngwaf/workspace/wildcardlist"
	"github.com/fastly/cli/pkg/commands/objectstorage"
	"github.com/fastly/cli/pkg/commands/pop"
	"github.com/fastly/cli/pkg/commands/products"
	"github.com/fastly/cli/pkg/commands/profile"
	"github.com/fastly/cli/pkg/commands/routingconfig"
	"github.com/fastly/cli/pkg/commands/secretstore"
	"github.com/fastly/cli/pkg/commands/secretstoreentry"
	"github.com/fastly/cli/pkg/commands/service"
	"github.com/fastly/cli/pkg/commands/shellcomplete"
	"github.com/fastly/cli/pkg/commands/sso"
	"github.com/fastly/cli/pkg/commands/stats"
	tlsconfig "github.com/fastly/cli/pkg/commands/tls/config"
	"github.com/fastly/cli/pkg/commands/tls/custom"
	"github.com/fastly/cli/pkg/commands/tls/platform"
	"github.com/fastly/cli/pkg/commands/tls/subscription"
	"github.com/fastly/cli/pkg/commands/tools"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/commands/user"
	"github.com/fastly/cli/pkg/commands/version"
	"github.com/fastly/cli/pkg/commands/whoami"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/global"
)

// CLI is the fastly command line, declared as a kong grammar.
type CLI struct {
	// Flags holds the global flags. Point it at the program's own Flags before
	// parsing, so kong writes them where the rest of the code reads them.
	*global.Flags
	Help kongcompat.HelpCommand `cmd:"" help:"Show help."`

	Shellcomplete          shellcomplete.RootCommand          `cmd:"" hidden:"" help:"Hidden command used to prevent help output when using --completion-script-<T>"`
	SSO                    sso.RootCommand                    `cmd:"" hidden:"" help:"Single Sign-On authentication (deprecated: use 'fastly auth login --sso --token <name>' instead)"`
	Auth                   auth.RootCommand                   `cmd:"" help:"Manage stored Fastly API tokens and token policies"`
	AuthToken              authtoken.RootCommand              `cmd:"" hidden:"" help:"Manage API tokens (deprecated: use 'fastly auth' subcommands instead)"`
	AiRuntimeControl       airuntimecontrol.RootCommand       `cmd:"" help:"Manage AI Runtime Control"`
	Apisecurity            apisecurity.RootCommand            `cmd:"" help:"Manipulate Fastly API security operations"`
	AuditLog               auditlog.RootCommand               `cmd:"" help:"Manage Fastly audit log configuration"`
	Compute                compute.RootCommand                `cmd:"" help:"Manage Compute packages"`
	Config                 config.RootCommand                 `cmd:"" help:"Display the Fastly CLI configuration"`
	ConfigStore            configstore.RootCommand            `cmd:"" help:"Manipulate Fastly Config Stores"`
	ConfigStoreEntry       configstoreentry.RootCommand       `cmd:"" help:"Manipulate Fastly Config Store items"`
	Dashboard              dashboard.RootCommand              `cmd:"" help:"Manipulate Fastly Custom Dashboards"`
	DNS                    dns.RootCommand                    `cmd:"" help:"Manipulate Fastly DNS Zones and TSIG Keys"`
	Domain                 domain.RootCommand                 `cmd:"" help:"Manipulate Fastly domains"`
	Install                install.RootCommand                `cmd:"" help:"Install the specified version of the CLI"`
	Integration            integration.RootCommand            `cmd:"" help:"Manage Fastly notification integrations"`
	IPList                 ip.RootCommand                     `cmd:"" help:"List Fastly's public IPs"`
	KVStore                kvstore.RootCommand                `cmd:"" aliases:"object-store" help:"Manipulate Fastly KV Stores"`
	KVStoreEntry           kvstoreentry.RootCommand           `cmd:"" help:"Manipulate Fastly KV Store keys"`
	LogTail                logtail.RootCommand                `cmd:"" help:"Tail Compute logs"`
	NGWAF                  ngwafCommands                      `cmd:"" help:"Manage NGWAF"`
	ObjectStorage          objectstorage.RootCommand          `cmd:"" help:"Manage object storage"`
	Pops                   pop.RootCommand                    `cmd:"" help:"List Fastly datacenters"`
	Products               products.RootCommand               `cmd:"" help:"Enable, disable, and check the enablement status of products"`
	Profile                profile.RootCommand                `cmd:"" hidden:"" help:"Manage user profiles (deprecated: use 'fastly auth' instead)"`
	RoutingConfig          routingconfig.RootCommand          `cmd:"" help:"Manipulate Fastly routing configs"`
	SecretStore            secretstore.RootCommand            `cmd:"" help:"Manipulate Fastly Secret Stores"`
	SecretStoreEntry       secretstoreentry.RootCommand       `cmd:"" help:"Manipulate Fastly Secret Store secrets"`
	Service                service.RootCommand                `cmd:"" help:"Manipulate Fastly services"`
	Stats                  stats.RootCommand                  `cmd:"" help:"View statistics for Fastly services: historical, realtime, aggregate, usage, domain and origin inspection"`
	TLSConfig              tlsconfig.RootCommand              `cmd:"" help:"Apply configuration options for each TLS enabled domain"`
	TLSCustom              custom.RootCommand                 `cmd:"" help:"Manage custom keys and certs used to enable TLS"`
	TLSPlatform            platform.RootCommand               `cmd:"" help:"Manage large numbers of TLS certificates"`
	TLSSubscription        subscription.RootCommand           `cmd:"" help:"Generate TLS certificates procured and renewed by Fastly"`
	Tools                  tools.RootCommand                  `cmd:"" help:"Tools for working with the Fastly platform"`
	Update                 update.RootCommand                 `cmd:"" help:"Update the CLI to the latest version"`
	User                   user.RootCommand                   `cmd:"" help:"Manipulate users of the Fastly API and web interface"`
	Version                version.RootCommand                `cmd:"" help:"Display version information for the Fastly CLI"`
	Whoami                 whoami.RootCommand                 `cmd:"" help:"Get information about the currently authenticated account"`
	Backend                backend.RootCommand                `cmd:"" hidden:"" help:"Manipulate Fastly service version backends"`
	DictionaryEntry        dictionaryentry.RootCommand        `cmd:"" hidden:"" help:"Manipulate Fastly edge dictionary items"`
	Dictionary             dictionary.RootCommand             `cmd:"" hidden:"" help:"Manipulate FastlyService Dictionaries"`
	Healthcheck            healthcheck.RootCommand            `cmd:"" hidden:"" help:"Manipulate Fastly service version healthchecks"`
	Imageoptimizerdefaults imageoptimizerdefaults.RootCommand `cmd:"" hidden:"" help:"Manipulate Fastly Image Optimizer Defaults"`
	Purge                  purge.Command                      `cmd:"" hidden:"" help:"Invalidate objects in the Fastly cache"`
	Alerts                 alerts.RootCommand                 `cmd:"" hidden:"" help:"Manipulate Fastly Alerts"`
	ACL                    acl.RootCommand                    `cmd:"" hidden:"" help:"Manipulate Fastly ACLs"`
	Aclentry               aclentry.RootCommand               `cmd:"" hidden:"" help:"Manipulate Fastly ACLs"`
	RateLimit              ratelimit.RootCommand              `cmd:"" hidden:"" help:"Manipulate rate-limiters of the Fastly API and web interface"`
	ResourceLink           resourcelink.RootCommand           `cmd:"" hidden:"" help:"Manipulate Fastly service resource links"`
	ServiceAuth            serviceauth.RootCommand            `cmd:"" hidden:"" help:"Manipulate Fastly Service Authentication"`
	VCL                    vcl.RootCommand                    `cmd:"" hidden:"" help:"Manipulate Fastly service version VCL"`
	ServiceVersion         serviceversion.RootCommand         `cmd:"" hidden:"" help:"Manipulate Fastly service versions"`
	Logging                logging.RootCommand                `cmd:"" hidden:"" help:"Manipulate Fastly service version logging endpoints"`
}

// ngwafCommands is the "ngwaf" command and its subcommands.
type ngwafCommands struct {
	ngwaf.RootCommand

	Workspace    ngwafWorkspaceCommands    `cmd:"" help:"Manage NGWAF Workspaces"`
	CountryList  countrylist.RootCommand   `cmd:"" help:"Manage NGWAF Account Country Lists"`
	Customsignal customsignal.RootCommand  `cmd:"" help:"Manage NGWAF Account-Level Custom Signals"`
	IPList       iplist.RootCommand        `cmd:"" help:"Manage NGWAF Account IP Lists"`
	Rule         rule.RootCommand          `cmd:"" help:"Manage NGWAF Account-Level Rules"`
	SignalList   signallist.RootCommand    `cmd:"" help:"Manage NGWAF Account Signal Lists"`
	StringList   stringlist.RootCommand    `cmd:"" help:"Manage NGWAF Account String Lists"`
	TimeSeries   timeseries.RootCommand    `cmd:"" help:"Manage NGWAF"`
	WildcardList ngwafWildcardListCommands `cmd:"" help:"Manage NGWAF Account Wildcard Lists"`
}

// ngwafWorkspaceCommands is the "ngwaf workspace" command and its subcommands.
type ngwafWorkspaceCommands struct {
	workspace.RootCommand

	Create       workspace.CreateCommand            `cmd:"" aliases:"add" help:"Create a workspace"`
	Delete       workspace.DeleteCommand            `cmd:"" help:"Delete a workspace"`
	Get          workspace.GetCommand               `cmd:"" help:"Get a workspace"`
	List         workspace.ListCommand              `cmd:"" help:"List all workspaces"`
	Update       workspace.UpdateCommand            `cmd:"" help:"Update a workspace"`
	Redaction    redaction.RootCommand              `cmd:"" help:"Manage NGWAF Redactions"`
	CountryList  wscountrylist.RootCommand          `cmd:"" help:"Manage NGWAF Workspace Country Lists"`
	TimeSeries   wstimeseries.RootCommand           `cmd:"" help:"Manage workspace time series metrics"`
	Customsignal wscustomsignal.RootCommand         `cmd:"" help:"Manage NGWAF Workspace-Level Custom Signals"`
	IPList       wsiplist.RootCommand               `cmd:"" help:"Manage NGWAF Workspace IP Lists"`
	Rule         workspaceRule.RootCommand          `cmd:"" help:"Manage NGWAF Account-Level Rules"`
	SignalList   wssignallistlist.RootCommand       `cmd:"" help:"Manage NGWAF Workspace Signal Lists"`
	StringList   wsstringlistlist.RootCommand       `cmd:"" help:"Manage NGWAF Workspace String Lists"`
	Threshold    threshold.RootCommand              `cmd:"" help:"Manage NGWAF Workspace Thresholds"`
	WildcardList ngwafWorkspaceWildcardListCommands `cmd:"" help:"Manage NGWAF Account Wildcard Lists"`
	Virtualpatch virtualpatch.RootCommand           `cmd:"" help:"Manage NGWAF Virtual Patches"`
	Alert        ngwafWorkspaceAlertCommands        `cmd:"" help:"Manage workspace alerts"`
}

// ngwafWorkspaceWildcardListCommands is the "ngwaf workspace wildcard-list" command and its subcommands.
type ngwafWorkspaceWildcardListCommands struct {
	ngwafwildcardlist.RootCommand

	Create wildcardlist.CreateCommand `cmd:"" aliases:"add" help:"Create a workspace-level wildcard list"`
	Delete wildcardlist.DeleteCommand `cmd:"" help:"Delete a workspace wildcard list"`
	Get    wildcardlist.GetCommand    `cmd:"" help:"Get a workspace-level wildcard list"`
	List   wildcardlist.ListCommand   `cmd:"" help:"List all wildcard lists for your workspace"`
	Update wildcardlist.UpdateCommand `cmd:"" help:"Update a workspace-level wildcard list"`
}

// ngwafWorkspaceAlertCommands is the "ngwaf workspace alert" command and its subcommands.
type ngwafWorkspaceAlertCommands struct {
	alert.RootCommand

	Datadog        datadog.RootCommand        `cmd:"" help:"Manage Datadog workspace alerts"`
	Jira           jira.RootCommand           `cmd:"" help:"Manage Jira workspace alerts"`
	Mailinglist    mailinglist.RootCommand    `cmd:"" help:"Manage Mailing List workspace alerts"`
	Microsoftteams microsoftteams.RootCommand `cmd:"" help:"Manage Microsoft Teams workspace alerts"`
	Opsgenie       opsgenie.RootCommand       `cmd:"" help:"Manage Opsgenie workspace alerts"`
	Pagerduty      pagerduty.RootCommand      `cmd:"" help:"Manage PagerDuty workspace alerts"`
	Slack          slack.RootCommand          `cmd:"" help:"Manage Slack workspace alerts"`
	Webhook        webhook.RootCommand        `cmd:"" help:"Manage Webhook workspace alerts"`
}

// ngwafWildcardListCommands is the "ngwaf wildcard-list" command and its subcommands.
type ngwafWildcardListCommands struct {
	ngwafwildcardlist.RootCommand

	Create ngwafwildcardlist.CreateCommand `cmd:"" aliases:"add" help:"Create an account-level wildcard list"`
	Delete ngwafwildcardlist.DeleteCommand `cmd:"" help:"Delete an account wildcard list"`
	Get    ngwafwildcardlist.GetCommand    `cmd:"" help:"Get an account-level wildcard list"`
	List   ngwafwildcardlist.ListCommand   `cmd:"" help:"List all wildcard lists for your account"`
	Update ngwafwildcardlist.UpdateCommand `cmd:"" help:"Update an account-level wildcard list"`
}

// HintActions returns the functions completing the flags whose values
// are only known at run time, as kingpin's HintAction did.
func (c *CLI) HintActions() map[string]func() []string {
	return map[string]func() []string{
		"token": env.Vars,
	}
}

// Conditional returns the kong option that leaves out the commands and flags
// the command line only offers under some conditions.
func Conditional() kong.Option {
	return kong.OptionFunc(func(k *kong.Kong) error {
		if env.AuthCommandDisabled() {
			return kong.IgnoreFields(`^CLI\.(SSO|Auth|AuthToken|Profile|Whoami|Token)$`).Apply(k)
		}
		return nil
	})
}
