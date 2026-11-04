package logging

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/service/logging/azureblob"
	"github.com/fastly/cli/pkg/commands/service/logging/bigquery"
	"github.com/fastly/cli/pkg/commands/service/logging/cloudfiles"
	"github.com/fastly/cli/pkg/commands/service/logging/datadog"
	"github.com/fastly/cli/pkg/commands/service/logging/debug"
	"github.com/fastly/cli/pkg/commands/service/logging/digitalocean"
	"github.com/fastly/cli/pkg/commands/service/logging/elasticsearch"
	"github.com/fastly/cli/pkg/commands/service/logging/ftp"
	"github.com/fastly/cli/pkg/commands/service/logging/gcs"
	"github.com/fastly/cli/pkg/commands/service/logging/googlepubsub"
	"github.com/fastly/cli/pkg/commands/service/logging/grafanacloudlogs"
	"github.com/fastly/cli/pkg/commands/service/logging/heroku"
	"github.com/fastly/cli/pkg/commands/service/logging/honeycomb"
	"github.com/fastly/cli/pkg/commands/service/logging/https"
	"github.com/fastly/cli/pkg/commands/service/logging/insights"
	"github.com/fastly/cli/pkg/commands/service/logging/kafka"
	"github.com/fastly/cli/pkg/commands/service/logging/kinesis"
	"github.com/fastly/cli/pkg/commands/service/logging/logexplorer"
	"github.com/fastly/cli/pkg/commands/service/logging/loggly"
	"github.com/fastly/cli/pkg/commands/service/logging/logshuttle"
	"github.com/fastly/cli/pkg/commands/service/logging/newrelic"
	"github.com/fastly/cli/pkg/commands/service/logging/newrelicotlp"
	"github.com/fastly/cli/pkg/commands/service/logging/openstack"
	"github.com/fastly/cli/pkg/commands/service/logging/papertrail"
	"github.com/fastly/cli/pkg/commands/service/logging/s3"
	"github.com/fastly/cli/pkg/commands/service/logging/scalyr"
	"github.com/fastly/cli/pkg/commands/service/logging/sftp"
	"github.com/fastly/cli/pkg/commands/service/logging/splunk"
	"github.com/fastly/cli/pkg/commands/service/logging/sumologic"
	"github.com/fastly/cli/pkg/commands/service/logging/syslog"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Debug            debug.Command                `cmd:"" help:"Stream live logging endpoint errors"`
	Insights         insights.Command             `cmd:"" help:"Retrieve statistics from sampled log records"`
	LogExplorer      logexplorer.Command          `cmd:"" help:"Retrieve sampled log records"`
	Azureblob        azureblob.RootCommand        `cmd:"" help:"Manipulate Fastly service version Azure Blob Storage logging endpoints"`
	Bigquery         bigquery.RootCommand         `cmd:"" help:"Manipulate Fastly service version BigQuery logging endpoints"`
	Cloudfiles       cloudfiles.RootCommand       `cmd:"" help:"Manipulate Fastly service version Cloudfiles logging endpoints"`
	Datadog          datadog.RootCommand          `cmd:"" help:"Manipulate Fastly service version Datadog logging endpoints"`
	Digitalocean     digitalocean.RootCommand     `cmd:"" help:"Manipulate Fastly service version DigitalOcean Spaces logging endpoints"`
	Elasticsearch    elasticsearch.RootCommand    `cmd:"" help:"Manipulate Fastly service version Elasticsearch logging endpoints"`
	FTP              ftp.RootCommand              `cmd:"" help:"Manipulate Fastly service version FTP logging endpoints"`
	GCS              gcs.RootCommand              `cmd:"" help:"Manipulate Fastly service version GCS logging endpoints"`
	Googlepubsub     googlepubsub.RootCommand     `cmd:"" help:"Manipulate Fastly service version Google Cloud Pub/Sub logging endpoints"`
	Grafanacloudlogs grafanacloudlogs.RootCommand `cmd:"" help:"Manipulate Fastly service version Grafana Cloud Logs logging endpoints"`
	Heroku           heroku.RootCommand           `cmd:"" help:"Manipulate Fastly service version Heroku logging endpoints"`
	Honeycomb        honeycomb.RootCommand        `cmd:"" help:"Manipulate Fastly service version Honeycomb logging endpoints"`
	HTTPS            https.RootCommand            `cmd:"" help:"Manipulate Fastly service version HTTPS logging endpoints"`
	Kafka            kafka.RootCommand            `cmd:"" help:"Manipulate Fastly service version Kafka logging endpoints"`
	Kinesis          kinesis.RootCommand          `cmd:"" help:"Manipulate a Kinesis logging endpoint for a specific Fastly service version"`
	Loggly           loggly.RootCommand           `cmd:"" help:"Manipulate Fastly service version Loggly logging endpoints"`
	Logshuttle       logshuttle.RootCommand       `cmd:"" help:"Manipulate Fastly service version Logshuttle logging endpoints"`
	Newrelic         newrelic.RootCommand         `cmd:"" help:"Manipulate a NewRelic logging endpoint for a specific Fastly service version"`
	Newrelicotlp     newrelicotlp.RootCommand     `cmd:"" help:"Manipulate a NewRelic OTLP logging endpoint for a specific Fastly service version"`
	Openstack        openstack.RootCommand        `cmd:"" help:"Manipulate Fastly service version OpenStack logging endpoints"`
	Papertrail       papertrail.RootCommand       `cmd:"" help:"Manipulate Fastly service version Papertrail logging endpoints."`
	S3               s3.RootCommand               `cmd:"" name:"s3" help:"Manipulate Fastly service version S3 logging endpoints"`
	Scalyr           scalyr.RootCommand           `cmd:"" help:"Manipulate Fastly service version Scalyr logging endpoints"`
	SFTP             sftp.RootCommand             `cmd:"" help:"Manipulate Fastly service version SFTP logging endpoints"`
	Splunk           splunk.RootCommand           `cmd:"" help:"Manipulate Fastly service version Splunk logging endpoints"`
	Sumologic        sumologic.RootCommand        `cmd:"" help:"Manipulate Fastly service version Sumologic logging endpoints"`
	Syslog           syslog.RootCommand           `cmd:"" help:"Manipulate Fastly service version Syslog logging endpoints"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "logging"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
