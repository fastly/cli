package logging

import (
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/alias/logging/azureblob"
	"github.com/fastly/cli/pkg/commands/alias/logging/bigquery"
	"github.com/fastly/cli/pkg/commands/alias/logging/cloudfiles"
	"github.com/fastly/cli/pkg/commands/alias/logging/datadog"
	"github.com/fastly/cli/pkg/commands/alias/logging/digitalocean"
	"github.com/fastly/cli/pkg/commands/alias/logging/elasticsearch"
	"github.com/fastly/cli/pkg/commands/alias/logging/ftp"
	"github.com/fastly/cli/pkg/commands/alias/logging/gcs"
	"github.com/fastly/cli/pkg/commands/alias/logging/googlepubsub"
	"github.com/fastly/cli/pkg/commands/alias/logging/grafanacloudlogs"
	"github.com/fastly/cli/pkg/commands/alias/logging/heroku"
	"github.com/fastly/cli/pkg/commands/alias/logging/honeycomb"
	"github.com/fastly/cli/pkg/commands/alias/logging/https"
	"github.com/fastly/cli/pkg/commands/alias/logging/kafka"
	"github.com/fastly/cli/pkg/commands/alias/logging/kinesis"
	"github.com/fastly/cli/pkg/commands/alias/logging/loggly"
	"github.com/fastly/cli/pkg/commands/alias/logging/logshuttle"
	"github.com/fastly/cli/pkg/commands/alias/logging/newrelic"
	"github.com/fastly/cli/pkg/commands/alias/logging/newrelicotlp"
	"github.com/fastly/cli/pkg/commands/alias/logging/openstack"
	"github.com/fastly/cli/pkg/commands/alias/logging/papertrail"
	"github.com/fastly/cli/pkg/commands/alias/logging/s3"
	"github.com/fastly/cli/pkg/commands/alias/logging/scalyr"
	"github.com/fastly/cli/pkg/commands/alias/logging/sftp"
	"github.com/fastly/cli/pkg/commands/alias/logging/splunk"
	"github.com/fastly/cli/pkg/commands/alias/logging/sumologic"
	"github.com/fastly/cli/pkg/commands/alias/logging/syslog"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags

	Azureblob        azureblob.RootCommand        `cmd:"" hidden:"" help:"Manipulate Fastly service version Azure Blob Storage logging endpoints"`
	Bigquery         bigquery.RootCommand         `cmd:"" hidden:"" help:"Manipulate Fastly service version Google BigQuery logging endpoints"`
	Cloudfiles       cloudfiles.RootCommand       `cmd:"" hidden:"" help:"Manipulate Fastly service version Rackspace Cloud Files logging endpoints"`
	Datadog          datadog.RootCommand          `cmd:"" hidden:"" help:"Manipulate Fastly service version Datadog logging endpoints"`
	Digitalocean     digitalocean.RootCommand     `cmd:"" hidden:"" help:"Manipulate Fastly service version DigitalOcean Spaces logging endpoints"`
	Elasticsearch    elasticsearch.RootCommand    `cmd:"" hidden:"" help:"Manipulate Fastly service version Elasticsearch logging endpoints"`
	FTP              ftp.RootCommand              `cmd:"" hidden:"" help:"Manipulate Fastly service version FTP logging endpoints"`
	GCS              gcs.RootCommand              `cmd:"" hidden:"" help:"Manipulate Fastly service version Google Cloud Storage logging endpoints"`
	Googlepubsub     googlepubsub.RootCommand     `cmd:"" hidden:"" help:"Manipulate Fastly service version Google Cloud Pub/Sub logging endpoints"`
	Grafanacloudlogs grafanacloudlogs.RootCommand `cmd:"" hidden:"" help:"Manipulate Fastly service version Grafana Cloud Logs logging endpoints"`
	Heroku           heroku.RootCommand           `cmd:"" hidden:"" help:"Manipulate Fastly service version Heroku logging endpoints"`
	Honeycomb        honeycomb.RootCommand        `cmd:"" hidden:"" help:"Manipulate Fastly service version Honeycomb logging endpoints"`
	HTTPS            https.RootCommand            `cmd:"" hidden:"" help:"Manipulate Fastly service version HTTPS logging endpoints"`
	Kafka            kafka.RootCommand            `cmd:"" hidden:"" help:"Manipulate Fastly service version Kafka logging endpoints"`
	Kinesis          kinesis.RootCommand          `cmd:"" hidden:"" help:"Manipulate Fastly service version Amazon Kinesis logging endpoints"`
	Loggly           loggly.RootCommand           `cmd:"" hidden:"" help:"Manipulate Fastly service version Loggly logging endpoints"`
	Logshuttle       logshuttle.RootCommand       `cmd:"" hidden:"" help:"Manipulate Fastly service version Log Shuttle logging endpoints"`
	Newrelic         newrelic.RootCommand         `cmd:"" hidden:"" help:"Manipulate Fastly service version New Relic logging endpoints"`
	Newrelicotlp     newrelicotlp.RootCommand     `cmd:"" hidden:"" help:"Manipulate Fastly service version New Relic OTLP logging endpoints"`
	Openstack        openstack.RootCommand        `cmd:"" hidden:"" help:"Manipulate Fastly service version OpenStack Swift logging endpoints"`
	Papertrail       papertrail.RootCommand       `cmd:"" hidden:"" help:"Manipulate Fastly service version Papertrail logging endpoints"`
	S3               s3.RootCommand               `cmd:"" name:"s3" hidden:"" help:"Manipulate Fastly service version Amazon S3 logging endpoints"`
	Scalyr           scalyr.RootCommand           `cmd:"" hidden:"" help:"Manipulate Fastly service version Scalyr logging endpoints"`
	SFTP             sftp.RootCommand             `cmd:"" hidden:"" help:"Manipulate Fastly service version SFTP logging endpoints"`
	Splunk           splunk.RootCommand           `cmd:"" hidden:"" help:"Manipulate Fastly service version Splunk logging endpoints"`
	Sumologic        sumologic.RootCommand        `cmd:"" hidden:"" help:"Manipulate Fastly service version Sumo Logic logging endpoints"`
	Syslog           syslog.RootCommand           `cmd:"" hidden:"" help:"Manipulate Fastly service version Syslog logging endpoints"`
}

// CommandName is the string to be used to invoke this command.
const CommandName = "logging"

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}
