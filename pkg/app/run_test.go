package app_test

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/update"
)

func TestApplication(t *testing.T) {
	// These tests should only verify the app.Run helper wires things up
	// correctly, and check behaviors that can't be associated with a specific
	// command or subcommand. Commands should be tested in their packages,
	// leveraging the app.Run helper as appropriate.
	for _, testcase := range []struct {
		name    string
		args    []string
		wantOut string
		wantErr string
	}{
		{
			name:    "no args",
			args:    nil,
			wantErr: helpDefault + "\nERROR: error parsing arguments: command not specified.\n",
		},
		{
			name:    "help flag only",
			args:    []string{"--help"},
			wantErr: helpDefault + "\nERROR: error parsing arguments: command not specified.\n",
		},
		{
			name:    "help argument only",
			args:    []string{"help"},
			wantErr: fullFatHelpDefault,
		},
		{
			name:    "help service",
			args:    []string{"help", "service"},
			wantErr: helpService,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var (
				args                            = testcase.args
				env                             = config.Environment{}
				file                            = config.File{}
				configFilePath                  = "/dev/null"
				clientFactory                   = mock.APIClient(mock.API{})
				httpClient     api.HTTPClient   = nil
				versioner      update.Versioner = nil
				stdin          io.Reader        = nil
				stdout         bytes.Buffer
				stderr         bytes.Buffer
			)
			err := app.Run(args, env, file, configFilePath, clientFactory, httpClient, versioner, stdin, &stdout)
			if err != nil {
				errors.Deduce(err).Print(&stderr)
			}

			// Our flag package creates trailing space on
			// some lines. Strip what we get so we don't
			// need to maintain invisible spaces in
			// wantOut/wantErr below.
			testutil.AssertString(t, testcase.wantOut, stripTrailingSpace(stdout.String()))
			testutil.AssertString(t, testcase.wantErr, stripTrailingSpace(stderr.String()))
		})
	}
}

// stripTrailingSpace removes any trailing spaces from the multiline str.
func stripTrailingSpace(str string) string {
	buf := bytes.NewBuffer(nil)

	scan := bufio.NewScanner(strings.NewReader(str))
	for scan.Scan() {
		buf.WriteString(strings.TrimRight(scan.Text(), " \t\r\n"))
		buf.WriteString("\n")
	}
	return buf.String()
}

var helpDefault = strings.TrimSpace(`
USAGE
  fastly [<flags>] <command> [<args> ...]

A tool to interact with the Fastly API

GLOBAL FLAGS
      --help         Show context-sensitive help.
  -t, --token=TOKEN  Fastly API token (or via FASTLY_API_TOKEN)
  -v, --verbose      Verbose logging

COMMANDS
  help             Show help.
  configure        Configure the Fastly CLI
  whoami           Get information about the currently authenticated account
  version          Display version information for the Fastly CLI
  update           Update the CLI to the latest version
  service          Manipulate Fastly services
  service-version  Manipulate Fastly service versions
  compute          Manage Compute@Edge packages
  domain           Manipulate Fastly service version domains
  backend          Manipulate Fastly service version backends
  healthcheck      Manipulate Fastly service version healthchecks
  logging          Manipulate Fastly service version logging endpoints
  stats            View statistics (historical and realtime) for a Fastly
                   service
`) + "\n\n"

var helpService = strings.TrimSpace(`
USAGE
  fastly [<flags>] service

GLOBAL FLAGS
      --help         Show context-sensitive help.
  -t, --token=TOKEN  Fastly API token (or via FASTLY_API_TOKEN)
  -v, --verbose      Verbose logging

SUBCOMMANDS

  service create --name=NAME [<flags>]
    Create a Fastly service

    -n, --name=NAME        Service name
        --type=wasm        Service type. Can be one of "wasm" or "vcl", defaults
                           to "wasm".
        --comment=COMMENT  Human-readable comment

  service list
    List Fastly services


  service describe [<flags>]
    Show detailed information about a Fastly service

    -s, --service-id=SERVICE-ID  Service ID

  service update [<flags>]
    Update a Fastly service

    -s, --service-id=SERVICE-ID  Service ID
    -n, --name=NAME              Service name
        --comment=COMMENT        Human-readable comment

  service delete [<flags>]
    Delete a Fastly service

    -s, --service-id=SERVICE-ID  Service ID

  service search [<flags>]
    Search for a Fastly service by name

    -n, --name=NAME  Service name

`) + "\n\n"

var fullFatHelpDefault = strings.TrimSpace(`
USAGE
  fastly [<flags>] <command>

A tool to interact with the Fastly API

GLOBAL FLAGS
      --help         Show context-sensitive help.
  -t, --token=TOKEN  Fastly API token (or via FASTLY_API_TOKEN)
  -v, --verbose      Verbose logging

COMMANDS
  help [<command> ...]
    Show help.


  configure
    Configure the Fastly CLI


  whoami
    Get information about the currently authenticated account


  version
    Display version information for the Fastly CLI


  update
    Update the CLI to the latest version


  service create --name=NAME [<flags>]
    Create a Fastly service

    -n, --name=NAME        Service name
        --type=wasm        Service type. Can be one of "wasm" or "vcl", defaults
                           to "wasm".
        --comment=COMMENT  Human-readable comment

  service list
    List Fastly services


  service describe [<flags>]
    Show detailed information about a Fastly service

    -s, --service-id=SERVICE-ID  Service ID

  service update [<flags>]
    Update a Fastly service

    -s, --service-id=SERVICE-ID  Service ID
    -n, --name=NAME              Service name
        --comment=COMMENT        Human-readable comment

  service delete [<flags>]
    Delete a Fastly service

    -s, --service-id=SERVICE-ID  Service ID

  service search [<flags>]
    Search for a Fastly service by name

    -n, --name=NAME  Service name

  service-version clone --version=VERSION [<flags>]
    Clone a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of version you wish to clone

  service-version list [<flags>]
    List Fastly service versions

    -s, --service-id=SERVICE-ID  Service ID

  service-version update --version=VERSION --comment=COMMENT [<flags>]
    Update a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of version you wish to update
        --comment=COMMENT        Human-readable comment

  service-version activate --version=VERSION [<flags>]
    Activate a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of version you wish to activate

  service-version deactivate --version=VERSION [<flags>]
    Deactivate a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of version you wish to deactivate

  service-version lock --version=VERSION [<flags>]
    Lock a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of version you wish to lock

  compute init [<flags>]
    Initialize a new Compute@Edge package locally

    -s, --service-id=SERVICE-ID    Existing service ID to use. By default, this
                                   command creates a new service
    -n, --name=NAME                Name of package, defaulting to directory name
                                   of the --path destination
    -d, --description=DESCRIPTION  Description of the package
    -a, --author=AUTHOR ...        Author(s) of the package
    -l, --language=LANGUAGE        Language of the package
    -f, --from=FROM                Git repository containing package template
    -p, --path=PATH                Destination to write the new package,
                                   defaulting to the current directory
        --domain=DOMAIN            The name of the domain associated to the
                                   package
        --backend=BACKEND          A hostname, IPv4, or IPv6 address for the
                                   package backend

  compute build [<flags>]
    Build a Compute@Edge package locally

    --name=NAME          Package name
    --language=LANGUAGE  Language type
    --include-source     Include source code in built package
    --force              Skip verification steps and force build

  compute deploy [<flags>]
    Deploy a package to a Fastly Compute@Edge service

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of version to activate
    -p, --path=PATH              Path to package

  compute update --service-id=SERVICE-ID --version=VERSION --path=PATH
    Update a package on a Fastly Compute@Edge service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -p, --path=PATH              Path to package

  compute validate --path=PATH
    Validate a Compute@Edge package

    -p, --path=PATH  Path to package

  domain create --name=NAME --version=VERSION [<flags>]
    Create a domain on a Fastly service version

    -n, --name=NAME              Domain name
        --comment=COMMENT        A descriptive note
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  domain list --version=VERSION [<flags>]
    List domains on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  domain describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a domain on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              Name of domain

  domain update --version=VERSION --name=NAME [<flags>]
    Update a domain on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              Domain name
        --new-name=NEW-NAME      New domain name
        --comment=COMMENT        A descriptive note

  domain delete --name=NAME --version=VERSION [<flags>]
    Delete a domain on a Fastly service version

    -n, --name=NAME              Domain name
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  backend create --service-id=SERVICE-ID --version=VERSION --name=NAME --address=ADDRESS [<flags>]
    Create a backend on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
    -n, --name=NAME                Backend name
        --address=ADDRESS          A hostname, IPv4, or IPv6 address for the
                                   backend
        --comment=COMMENT          A descriptive note
        --port=PORT                Port number of the address
        --override-host=OVERRIDE-HOST
                                   The hostname to override the Host header
        --connect-timeout=CONNECT-TIMEOUT
                                   How long to wait for a timeout in
                                   milliseconds
        --max-conn=MAX-CONN        Maximum number of connections
        --first-byte-timeout=FIRST-BYTE-TIMEOUT
                                   How long to wait for the first bytes in
                                   milliseconds
        --between-bytes-timeout=BETWEEN-BYTES-TIMEOUT
                                   How long to wait between bytes in
                                   milliseconds
        --auto-loadbalance         Whether or not this backend should be
                                   automatically load balanced
        --weight=WEIGHT            Weight used to load balance this backend
                                   against others
        --request-condition=REQUEST-CONDITION
                                   Condition, which if met, will select this
                                   backend during a request
        --healthcheck=HEALTHCHECK  The name of the healthcheck to use with this
                                   backend
        --shield=SHIELD            The shield POP designated to reduce inbound
                                   load on this origin by serving the cached
                                   data to the rest of the network
        --use-ssl                  Whether or not to use SSL to reach the
                                   backend
        --ssl-check-cert           Be strict on checking SSL certs
        --ssl-ca-cert=SSL-CA-CERT  CA certificate attached to origin
        --ssl-client-cert=SSL-CLIENT-CERT
                                   Client certificate attached to origin
        --ssl-client-key=SSL-CLIENT-KEY
                                   Client key attached to origin
        --ssl-cert-hostname=SSL-CERT-HOSTNAME
                                   Overrides ssl_hostname, but only for cert
                                   verification. Does not affect SNI at all.
        --ssl-sni-hostname=SSL-SNI-HOSTNAME
                                   Overrides ssl_hostname, but only for SNI in
                                   the handshake. Does not affect cert
                                   validation at all.
        --min-tls-version=MIN-TLS-VERSION
                                   Minimum allowed TLS version on SSL
                                   connections to this backend
        --max-tls-version=MAX-TLS-VERSION
                                   Maximum allowed TLS version on SSL
                                   connections to this backend
        --ssl-ciphers=SSL-CIPHERS ...
                                   List of OpenSSL ciphers (see
                                   https://www.openssl.org/docs/man1.0.2/man1/ciphers
                                   for details)

  backend list --service-id=SERVICE-ID --version=VERSION
    List backends on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  backend describe --service-id=SERVICE-ID --version=VERSION --name=NAME
    Show detailed information about a backend on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              Name of backend

  backend update --service-id=SERVICE-ID --version=VERSION --name=NAME [<flags>]
    Update a backend on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
    -n, --name=NAME                backend name
        --new-name=NEW-NAME        New backend name
        --comment=COMMENT          A descriptive note
        --address=ADDRESS          A hostname, IPv4, or IPv6 address for the
                                   backend
        --port=PORT                Port number of the address
        --override-host=OVERRIDE-HOST
                                   The hostname to override the Host header
        --connect-timeout=CONNECT-TIMEOUT
                                   How long to wait for a timeout in
                                   milliseconds
        --max-conn=MAX-CONN        Maximum number of connections
        --first-byte-timeout=FIRST-BYTE-TIMEOUT
                                   How long to wait for the first bytes in
                                   milliseconds
        --between-bytes-timeout=BETWEEN-BYTES-TIMEOUT
                                   How long to wait between bytes in
                                   milliseconds
        --auto-loadbalance         Whether or not this backend should be
                                   automatically load balanced
        --weight=WEIGHT            Weight used to load balance this backend
                                   against others
        --request-condition=REQUEST-CONDITION
                                   condition, which if met, will select this
                                   backend during a request
        --healthcheck=HEALTHCHECK  The name of the healthcheck to use with this
                                   backend
        --shield=SHIELD            The shield POP designated to reduce inbound
                                   load on this origin by serving the cached
                                   data to the rest of the network
        --use-ssl                  Whether or not to use SSL to reach the
                                   backend
        --ssl-check-cert           Be strict on checking SSL certs
        --ssl-ca-cert=SSL-CA-CERT  CA certificate attached to origin
        --ssl-client-cert=SSL-CLIENT-CERT
                                   Client certificate attached to origin
        --ssl-client-key=SSL-CLIENT-KEY
                                   Client key attached to origin
        --ssl-cert-hostname=SSL-CERT-HOSTNAME
                                   Overrides ssl_hostname, but only for cert
                                   verification. Does not affect SNI at all.
        --ssl-sni-hostname=SSL-SNI-HOSTNAME
                                   Overrides ssl_hostname, but only for SNI in
                                   the handshake. Does not affect cert
                                   validation at all.
        --min-tls-version=MIN-TLS-VERSION
                                   Minimum allowed TLS version on SSL
                                   connections to this backend
        --max-tls-version=MAX-TLS-VERSION
                                   Maximum allowed TLS version on SSL
                                   connections to this backend
        --ssl-ciphers=SSL-CIPHERS ...
                                   List of OpenSSL ciphers (see
                                   https://www.openssl.org/docs/man1.0.2/man1/ciphers
                                   for details)

  backend delete --service-id=SERVICE-ID --version=VERSION --name=NAME
    Delete a backend on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              Backend name

  healthcheck create --version=VERSION --name=NAME [<flags>]
    Create a healthcheck on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              Healthcheck name
        --comment=COMMENT        A descriptive note
        --method=METHOD          Which HTTP method to use
        --host=HOST              Which host to check
        --path=PATH              The path to check
        --http-version=HTTP-VERSION
                                 Whether to use version 1.0 or 1.1 HTTP
        --timeout=TIMEOUT        Timeout in milliseconds
        --check-interval=CHECK-INTERVAL
                                 How often to run the healthcheck in
                                 milliseconds
        --expected-response=EXPECTED-RESPONSE
                                 The status code expected from the host
        --window=WINDOW          The number of most recent healthcheck queries
                                 to keep for this healthcheck
        --threshold=THRESHOLD    How many healthchecks must succeed to be
                                 considered healthy
        --initial=INITIAL        When loading a config, the initial number of
                                 probes to be seen as OK

  healthcheck list --version=VERSION [<flags>]
    List healthchecks on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  healthcheck describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a healthcheck on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              Name of healthcheck

  healthcheck update --version=VERSION --name=NAME [<flags>]
    Update a healthcheck on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              Healthcheck name
        --new-name=NEW-NAME      Healthcheck name
        --comment=COMMENT        A descriptive note
        --method=METHOD          Which HTTP method to use
        --host=HOST              Which host to check
        --path=PATH              The path to check
        --http-version=HTTP-VERSION
                                 Whether to use version 1.0 or 1.1 HTTP
        --timeout=TIMEOUT        Timeout in milliseconds
        --check-interval=CHECK-INTERVAL
                                 How often to run the healthcheck in
                                 milliseconds
        --expected-response=EXPECTED-RESPONSE
                                 The status code expected from the host
        --window=WINDOW          The number of most recent healthcheck queries
                                 to keep for this healthcheck
        --threshold=THRESHOLD    How many healthchecks must succeed to be
                                 considered healthy
        --initial=INITIAL        When loading a config, the initial number of
                                 probes to be seen as OK

  healthcheck delete --version=VERSION --name=NAME [<flags>]
    Delete a healthcheck on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              Healthcheck name

  logging bigquery create --name=NAME --version=VERSION --project-id=PROJECT-ID --dataset=DATASET --table=TABLE --user=USER --secret-key=SECRET-KEY [<flags>]
    Create a BigQuery logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the BigQuery logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --project-id=PROJECT-ID  Your Google Cloud Platform project ID
        --dataset=DATASET        Your BigQuery dataset
        --table=TABLE            Your BigQuery table
        --user=USER              Your Google Cloud Platform service account
                                 email address. The client_email field in your
                                 service account authentication JSON.
        --secret-key=SECRET-KEY  Your Google Cloud Platform account secret key.
                                 The private_key field in your service account
                                 authentication JSON.
        --template-suffix=TEMPLATE-SUFFIX
                                 BigQuery table name suffix template
        --format=FORMAT          Apache style log formatting. Must produce JSON
                                 that matches the schema of your BigQuery table
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute

  logging bigquery list --version=VERSION [<flags>]
    List BigQuery endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging bigquery describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a BigQuery logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the BigQuery logging object

  logging bigquery update --version=VERSION --name=NAME [<flags>]
    Update a BigQuery logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the BigQuery logging object
        --new-name=NEW-NAME      New name of the BigQuery logging object
        --project-id=PROJECT-ID  Your Google Cloud Platform project ID
        --dataset=DATASET        Your BigQuery dataset
        --table=TABLE            Your BigQuery table
        --user=USER              Your Google Cloud Platform service account
                                 email address. The client_email field in your
                                 service account authentication JSON.
        --secret-key=SECRET-KEY  Your Google Cloud Platform account secret key.
                                 The private_key field in your service account
                                 authentication JSON.
        --template-suffix=TEMPLATE-SUFFIX
                                 BigQuery table name suffix template
        --format=FORMAT          Apache style log formatting. Must produce JSON
                                 that matches the schema of your BigQuery table
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute

  logging bigquery delete --version=VERSION --name=NAME [<flags>]
    Delete a BigQuery logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the BigQuery logging object

  logging s3 create --name=NAME --version=VERSION --bucket=BUCKET --access-key=ACCESS-KEY --secret-key=SECRET-KEY [<flags>]
    Create an Amazon S3 logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the S3 logging object. Used as a
                                 primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --bucket=BUCKET          Your S3 bucket name
        --access-key=ACCESS-KEY  Your S3 account access key
        --secret-key=SECRET-KEY  Your S3 account secret key
        --domain=DOMAIN          The domain of the S3 endpoint
        --path=PATH              The path to upload logs to
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --redundancy=REDUNDANCY  The S3 redundancy level. Can be either standard
                                 or reduced_redundancy
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk
        --server-side-encryption=SERVER-SIDE-ENCRYPTION
                                 Set to enable S3 Server Side Encryption. Can be
                                 either AES256 or aws:kms
        --server-side-encryption-kms-key-id=SERVER-SIDE-ENCRYPTION-KMS-KEY-ID
                                 Server-side KMS Key ID. Must be set if
                                 server-side-encryption is set to aws:kms

  logging s3 list --version=VERSION [<flags>]
    List S3 endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging s3 describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a S3 logging endpoint on a Fastly service
    version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the S3 logging object

  logging s3 update --version=VERSION --name=NAME [<flags>]
    Update a S3 logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the S3 logging object
        --new-name=NEW-NAME      New name of the S3 logging object
        --bucket=BUCKET          Your S3 bucket name
        --access-key=ACCESS-KEY  Your S3 account access key
        --secret-key=SECRET-KEY  Your S3 account secret key
        --domain=DOMAIN          The domain of the S3 endpoint
        --path=PATH              The path to upload logs to
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --redundancy=REDUNDANCY  The S3 redundancy level. Can be either standard
                                 or reduced_redundancy
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk
        --server-side-encryption=SERVER-SIDE-ENCRYPTION
                                 Set to enable S3 Server Side Encryption. Can be
                                 either AES256 or aws:kms
        --server-side-encryption-kms-key-id=SERVER-SIDE-ENCRYPTION-KMS-KEY-ID
                                 Server-side KMS Key ID. Must be set if
                                 server-side-encryption is set to aws:kms

  logging s3 delete --version=VERSION --name=NAME [<flags>]
    Delete a S3 logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the S3 logging object

  logging syslog create --name=NAME --version=VERSION --address=ADDRESS [<flags>]
    Create a Syslog logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Syslog logging object. Used
                                   as a primary key for API access
    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
        --address=ADDRESS          A hostname or IPv4 address
        --port=PORT                The port number
        --use-tls                  Whether to use TLS for secure logging. Can be
                                   either true or false
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   Used during the TLS handshake to validate the
                                   certificate
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
        --auth-token=AUTH-TOKEN    Whether to prepend each message with a
                                   specific token
        --format=FORMAT            Apache style log formatting
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --message-type=MESSAGE-TYPE
                                   How the message should be formatted. One of:
                                   classic (default), loggly, logplex or blank
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug

  logging syslog list --version=VERSION [<flags>]
    List Syslog endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging syslog describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Syslog logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Syslog logging object

  logging syslog update --version=VERSION --name=NAME [<flags>]
    Update a Syslog logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
    -n, --name=NAME                The name of the Syslog logging object
        --new-name=NEW-NAME        New name of the Syslog logging object
        --address=ADDRESS          A hostname or IPv4 address
        --port=PORT                The port number
        --use-tls                  Whether to use TLS for secure logging. Can be
                                   either true or false
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   Used during the TLS handshake to validate the
                                   certificate
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
        --auth-token=AUTH-TOKEN    Whether to prepend each message with a
                                   specific token
        --format=FORMAT            Apache style log formatting
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --message-type=MESSAGE-TYPE
                                   How the message should be formatted. One of:
                                   classic (default), loggly, logplex or blank
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug

  logging syslog delete --version=VERSION --name=NAME [<flags>]
    Delete a Syslog logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Syslog logging object

  logging logentries create --name=NAME --version=VERSION [<flags>]
    Create a Logentries logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Logentries logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --port=PORT              The port number
        --use-tls                Whether to use TLS for secure logging. Can be
                                 either true or false
        --auth-token=AUTH-TOKEN  Use token based authentication
                                 (https://logentries.com/doc/input-token/)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value

  logging logentries list --version=VERSION [<flags>]
    List Logentries endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging logentries describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Logentries logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Logentries logging object

  logging logentries update --version=VERSION --name=NAME [<flags>]
    Update a Logentries logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Logentries logging object
        --new-name=NEW-NAME      New name of the Logentries logging object
        --port=PORT              The port number
        --use-tls                Whether to use TLS for secure logging. Can be
                                 either true or false
        --auth-token=AUTH-TOKEN  Use token based authentication
                                 (https://logentries.com/doc/input-token/)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value

  logging logentries delete --version=VERSION --name=NAME [<flags>]
    Delete a Logentries logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Logentries logging object

  logging papertrail create --name=NAME --version=VERSION --address=ADDRESS [<flags>]
    Create a Papertrail logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Papertrail logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --address=ADDRESS        A hostname or IPv4 address
        --port=PORT              The port number
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --format=FORMAT          Apache style log formatting
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value

  logging papertrail list --version=VERSION [<flags>]
    List Papertrail endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging papertrail describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Papertrail logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Papertrail logging object

  logging papertrail update --version=VERSION --name=NAME [<flags>]
    Update a Papertrail logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Papertrail logging object
        --new-name=NEW-NAME      New name of the Papertrail logging object
        --address=ADDRESS        A hostname or IPv4 address
        --port=PORT              The port number
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --format=FORMAT          Apache style log formatting
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value

  logging papertrail delete --version=VERSION --name=NAME [<flags>]
    Delete a Papertrail logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Papertrail logging object

  logging sumologic create --name=NAME --version=VERSION --url=URL [<flags>]
    Create a Sumologic logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Sumologic logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --url=URL                The URL to POST to
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value

  logging sumologic list --version=VERSION [<flags>]
    List Sumologic endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging sumologic describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Sumologic logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Sumologic logging object

  logging sumologic update --version=VERSION --name=NAME [<flags>]
    Update a Sumologic logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Sumologic logging object
        --new-name=NEW-NAME      New name of the Sumologic logging object
        --url=URL                The URL to POST to
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value

  logging sumologic delete --version=VERSION --name=NAME [<flags>]
    Delete a Sumologic logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Sumologic logging object

  logging gcs create --name=NAME --version=VERSION --user=USER --bucket=BUCKET --secret-key=SECRET-KEY [<flags>]
    Create a GCS logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the GCS logging object. Used as a
                                 primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --user=USER              Your GCS service account email address. The
                                 client_email field in your service account
                                 authentication JSON
        --bucket=BUCKET          The bucket of the GCS bucket
        --secret-key=SECRET-KEY  Your GCS account secret key. The private_key
                                 field in your service account authentication
                                 JSON
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --path=PATH              The path to upload logs to (default '/')
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging gcs list --version=VERSION [<flags>]
    List GCS endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging gcs describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a GCS logging endpoint on a Fastly service
    version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the GCS logging object

  logging gcs update --version=VERSION --name=NAME [<flags>]
    Update a GCS logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the GCS logging object
        --new-name=NEW-NAME      New name of the GCS logging object
        --bucket=BUCKET          The bucket of the GCS bucket
        --user=USER              Your GCS service account email address. The
                                 client_email field in your service account
                                 authentication JSON
        --secret-key=SECRET-KEY  Your GCS account secret key. The private_key
                                 field in your service account authentication
                                 JSON
        --path=PATH              The path to upload logs to (default '/')
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging gcs delete --version=VERSION --name=NAME [<flags>]
    Delete a GCS logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the GCS logging object

  logging ftp create --name=NAME --version=VERSION --address=ADDRESS --user=USER --password=PASSWORD [<flags>]
    Create an FTP logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the FTP logging object. Used as a
                                 primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --address=ADDRESS        An hostname or IPv4 address
        --user=USER              The username for the server (can be anonymous)
        --password=PASSWORD      The password for the server (for anonymous use
                                 an email address)
        --port=PORT              The port number
        --path=PATH              The path to upload log files to. If the path
                                 ends in / then it is treated as a directory
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging ftp list --version=VERSION [<flags>]
    List FTP endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging ftp describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an FTP logging endpoint on a Fastly service
    version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the FTP logging object

  logging ftp update --version=VERSION --name=NAME [<flags>]
    Update an FTP logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the FTP logging object
        --new-name=NEW-NAME      New name of the FTP logging object
        --address=ADDRESS        An hostname or IPv4 address
        --port=PORT              The port number
        --username=USERNAME      The username for the server (can be anonymous)
        --password=PASSWORD      The password for the server (for anonymous use
                                 an email address)
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk
        --path=PATH              The path to upload log files to. If the path
                                 ends in / then it is treated as a directory
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (the default, version 2 log format) or 1 (the
                                 version 1 log format). The logging call gets
                                 placed by default in vcl_log if format_version
                                 is set to 2 and in vcl_deliver if
                                 format_version is set to 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging ftp delete --version=VERSION --name=NAME [<flags>]
    Delete an FTP logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the FTP logging object

  logging splunk create --name=NAME --version=VERSION --url=URL [<flags>]
    Create a Splunk logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Splunk logging object. Used
                                   as a primary key for API access
    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
        --url=URL                  The URL to POST to
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --format=FORMAT            Apache style log formatting
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug
        --auth-token=AUTH-TOKEN    A Splunk token for use in posting logs over
                                   HTTP to your collector

  logging splunk list --version=VERSION [<flags>]
    List Splunk endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging splunk describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Splunk logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Splunk logging object

  logging splunk update --version=VERSION --name=NAME [<flags>]
    Update a Splunk logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
    -n, --name=NAME                The name of the Splunk logging object
        --new-name=NEW-NAME        New name of the Splunk logging object
        --url=URL                  The URL to POST to.
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --format=FORMAT            Apache style log formatting
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug. This field is not required and has
                                   no default value
        --auth-token=AUTH-TOKEN

  logging splunk delete --version=VERSION --name=NAME [<flags>]
    Delete a Splunk logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Splunk logging object

  logging scalyr create --name=NAME --version=VERSION --auth-token=AUTH-TOKEN [<flags>]
    Create a Scalyr logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Scalyr logging object. Used as
                                 a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://www.scalyr.com/keys)
        --region=REGION          The region that log data will be sent to. One
                                 of US or EU. Defaults to US if undefined
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging scalyr list --version=VERSION [<flags>]
    List Scalyr endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging scalyr describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Scalyr logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Scalyr logging object

  logging scalyr update --version=VERSION --name=NAME [<flags>]
    Update a Scalyr logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Scalyr logging object
        --new-name=NEW-NAME      New name of the Scalyr logging object
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://www.scalyr.com/keys)
        --region=REGION          The region that log data will be sent to. One
                                 of US or EU. Defaults to US if undefined
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging scalyr delete --version=VERSION --name=NAME [<flags>]
    Delete a Scalyr logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Scalyr logging object

  logging loggly create --name=NAME --version=VERSION --auth-token=AUTH-TOKEN [<flags>]
    Create a Loggly logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Loggly logging object. Used as
                                 a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://www.loggly.com/docs/customer-token-authentication-token/)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging loggly list --version=VERSION [<flags>]
    List Loggly endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging loggly describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Loggly logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Loggly logging object

  logging loggly update --version=VERSION --name=NAME [<flags>]
    Update a Loggly logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Loggly logging object
        --new-name=NEW-NAME      New name of the Loggly logging object
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://www.loggly.com/docs/customer-token-authentication-token/)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging loggly delete --version=VERSION --name=NAME [<flags>]
    Delete a Loggly logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Loggly logging object

  logging honeycomb create --name=NAME --version=VERSION --dataset=DATASET --auth-token=AUTH-TOKEN [<flags>]
    Create a Honeycomb logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Honeycomb logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --dataset=DATASET        The Honeycomb Dataset you want to log to
        --auth-token=AUTH-TOKEN  The Write Key from the Account page of your
                                 Honeycomb account
        --format=FORMAT          Apache style log formatting. Your log must
                                 produce valid JSON that Honeycomb can ingest
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging honeycomb list --version=VERSION [<flags>]
    List Honeycomb endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging honeycomb describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Honeycomb logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Honeycomb logging object

  logging honeycomb update --version=VERSION --name=NAME [<flags>]
    Update a Honeycomb logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Honeycomb logging object
        --new-name=NEW-NAME      New name of the Honeycomb logging object
        --format=FORMAT          Apache style log formatting. Your log must
                                 produce valid JSON that Honeycomb can ingest
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --dataset=DATASET        The Honeycomb Dataset you want to log to
        --auth-token=AUTH-TOKEN  The Write Key from the Account page of your
                                 Honeycomb account
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging honeycomb delete --version=VERSION --name=NAME [<flags>]
    Delete a Honeycomb logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Honeycomb logging object

  logging heroku create --name=NAME --version=VERSION --url=URL --auth-token=AUTH-TOKEN [<flags>]
    Create a Heroku logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Heroku logging object. Used as
                                 a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --url=URL                The url to stream logs to
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://devcenter.heroku.com/articles/add-on-partner-log-integration)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging heroku list --version=VERSION [<flags>]
    List Heroku endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging heroku describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Heroku logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Heroku logging object

  logging heroku update --version=VERSION --name=NAME [<flags>]
    Update a Heroku logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Heroku logging object
        --new-name=NEW-NAME      New name of the Heroku logging object
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --url=URL                The url to stream logs to
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://devcenter.heroku.com/articles/add-on-partner-log-integration)
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging heroku delete --version=VERSION --name=NAME [<flags>]
    Delete a Heroku logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Heroku logging object

  logging sftp create --name=NAME --version=VERSION --address=ADDRESS --user=USER --ssh-known-hosts=SSH-KNOWN-HOSTS [<flags>]
    Create an SFTP logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the SFTP logging object. Used as a
                                 primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --address=ADDRESS        The hostname or IPv4 addres
        --user=USER              The username for the server
        --ssh-known-hosts=SSH-KNOWN-HOSTS
                                 A list of host keys for all hosts we can
                                 connect to over SFTP
        --port=PORT              The port number
        --password=PASSWORD      The password for the server. If both password
                                 and secret_key are passed, secret_key will be
                                 used in preference
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk
        --secret-key=SECRET-KEY  The SSH private key for the server. If both
                                 password and secret_key are passed, secret_key
                                 will be used in preference
        --path=PATH              The path to upload logs to. The directory must
                                 exist on the SFTP server before logs can be
                                 saved to it
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging sftp list --version=VERSION [<flags>]
    List SFTP endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging sftp describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an SFTP logging endpoint on a Fastly service
    version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the SFTP logging object

  logging sftp update --version=VERSION --name=NAME [<flags>]
    Update an SFTP logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the SFTP logging object
        --new-name=NEW-NAME      New name of the SFTP logging object
        --address=ADDRESS        The hostname or IPv4 address
        --port=PORT              The port number
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk
        --secret-key=SECRET-KEY  The SSH private key for the server. If both
                                 password and secret_key are passed, secret_key
                                 will be used in preference
        --ssh-known-hosts=SSH-KNOWN-HOSTS
                                 A list of host keys for all hosts we can
                                 connect to over SFTP
        --user=USER              The username for the server
        --password=PASSWORD      The password for the server. If both password
                                 and secret_key are passed, secret_key will be
                                 used in preference
        --path=PATH              The path to upload logs to. The directory must
                                 exist on the SFTP server before logs can be
                                 saved to it
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging sftp delete --version=VERSION --name=NAME [<flags>]
    Delete an SFTP logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the SFTP logging object

  logging logshuttle create --name=NAME --version=VERSION --url=URL --auth-token=AUTH-TOKEN [<flags>]
    Create a Logshuttle logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Logshuttle logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --url=URL                Your Log Shuttle endpoint url
        --auth-token=AUTH-TOKEN  The data authentication token associated with
                                 this endpoint
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging logshuttle list --version=VERSION [<flags>]
    List Logshuttle endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging logshuttle describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Logshuttle logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Logshuttle logging object

  logging logshuttle update --version=VERSION --name=NAME [<flags>]
    Update a Logshuttle logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Logshuttle logging object
        --new-name=NEW-NAME      New name of the Logshuttle logging object
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --url=URL                Your Log Shuttle endpoint url
        --auth-token=AUTH-TOKEN  The data authentication token associated with
                                 this endpoint
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging logshuttle delete --version=VERSION --name=NAME [<flags>]
    Delete a Logshuttle logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Logshuttle logging object

  logging cloudfiles create --name=NAME --version=VERSION --user=USER --access-key=ACCESS-KEY --bucket=BUCKET [<flags>]
    Create a Cloudfiles logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Cloudfiles logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --user=USER              The username for your Cloudfile account
        --access-key=ACCESS-KEY  Your Cloudfile account access key
        --bucket=BUCKET          The name of your Cloudfiles container
        --path=PATH              The path to upload logs to
        --region=REGION          The region to stream logs to. One of:
                                 DFW-Dallas, ORD-Chicago, IAD-Northern Virginia,
                                 LON-London, SYD-Sydney, HKG-Hong Kong
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk

  logging cloudfiles list --version=VERSION [<flags>]
    List Cloudfiles endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging cloudfiles describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Cloudfiles logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Cloudfiles logging object

  logging cloudfiles update --version=VERSION --name=NAME [<flags>]
    Update a Cloudfiles logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Cloudfiles logging object
        --new-name=NEW-NAME      New name of the Cloudfiles logging object
        --user=USER              The username for your Cloudfile account
        --access-key=ACCESS-KEY  Your Cloudfile account access key
        --bucket=BUCKET          The name of your Cloudfiles container
        --path=PATH              The path to upload logs to
        --region=REGION          The region to stream logs to. One of:
                                 DFW-Dallas, ORD-Chicago, IAD-Northern Virginia,
                                 LON-London, SYD-Sydney, HKG-Hong Kong
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk

  logging cloudfiles delete --version=VERSION --name=NAME [<flags>]
    Delete a Cloudfiles logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Cloudfiles logging object

  logging digitalocean create --name=NAME --version=VERSION --bucket=BUCKET --access-key=ACCESS-KEY --secret-key=SECRET-KEY [<flags>]
    Create a DigitalOcean Spaces logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the DigitalOcean Spaces logging
                                 object. Used as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --bucket=BUCKET          The name of the DigitalOcean Space
        --access-key=ACCESS-KEY  Your DigitalOcean Spaces account access key
        --secret-key=SECRET-KEY  Your DigitalOcean Spaces account secret key
        --domain=DOMAIN          The domain of the DigitalOcean Spaces endpoint
                                 (default 'nyc3.digitaloceanspaces.com')
        --path=PATH              The path to upload logs to
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk

  logging digitalocean list --version=VERSION [<flags>]
    List DigitalOcean Spaces logging endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging digitalocean describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a DigitalOcean Spaces logging endpoint on a
    Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the DigitalOcean Spaces logging
                                 object

  logging digitalocean update --version=VERSION --name=NAME [<flags>]
    Update a DigitalOcean Spaces logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the DigitalOcean Spaces logging
                                 object
        --new-name=NEW-NAME      New name of the DigitalOcean Spaces logging
                                 object
        --bucket=BUCKET          The name of the DigitalOcean Space
        --domain=DOMAIN          The domain of the DigitalOcean Spaces endpoint
                                 (default 'nyc3.digitaloceanspaces.com')
        --access-key=ACCESS-KEY  Your DigitalOcean Spaces account access key
        --secret-key=SECRET-KEY  Your DigitalOcean Spaces account secret key
        --path=PATH              The path to upload logs to
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk

  logging digitalocean delete --version=VERSION --name=NAME [<flags>]
    Delete a DigitalOcean Spaces logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the DigitalOcean Spaces logging
                                 object

  logging elasticsearch create --name=NAME --version=VERSION --index=INDEX --url=URL [<flags>]
    Create an Elasticsearch logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Elasticsearch logging object.
                                   Used as a primary key for API access
    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
        --index=INDEX              The name of the Elasticsearch index to send
                                   documents (logs) to. The index must follow
                                   the Elasticsearch index format rules
                                   (https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html).
                                   We support strftime
                                   (http://man7.org/linux/man-pages/man3/strftime.3.html)
                                   interpolated variables inside braces prefixed
                                   with a pound symbol. For example, #{%F} will
                                   interpolate as YYYY-MM-DD with today's date
        --url=URL                  The URL to stream logs to. Must use HTTPS.
        --pipeline=PIPELINE        The ID of the Elasticsearch ingest pipeline
                                   to apply pre-process transformations to
                                   before indexing. For example my_pipeline_id.
                                   Learn more about creating a pipeline in the
                                   Elasticsearch docs
                                   (https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest.html)
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --format=FORMAT            Apache style log formatting. Your log must
                                   produce valid JSON that Elasticsearch can
                                   ingest
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute
        --request-max-entries=REQUEST-MAX-ENTRIES
                                   Maximum number of logs to append to a batch,
                                   if non-zero. Defaults to 0 for unbounded
        --request-max-bytes=REQUEST-MAX-BYTES
                                   Maximum size of log batch, if non-zero.
                                   Defaults to 0 for unbounded

  logging elasticsearch list --version=VERSION [<flags>]
    List Elasticsearch endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging elasticsearch describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an Elasticsearch logging endpoint on a
    Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Elasticsearch logging object

  logging elasticsearch update --version=VERSION --name=NAME [<flags>]
    Update an Elasticsearch logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
    -n, --name=NAME                The name of the Elasticsearch logging object
        --new-name=NEW-NAME        New name of the Elasticsearch logging object
        --index=INDEX              The name of the Elasticsearch index to send
                                   documents (logs) to. The index must follow
                                   the Elasticsearch index format rules
                                   (https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html).
                                   We support strftime
                                   (http://man7.org/linux/man-pages/man3/strftime.3.html)
                                   interpolated variables inside braces prefixed
                                   with a pound symbol. For example, #{%F} will
                                   interpolate as YYYY-MM-DD with today's date
        --url=URL                  The URL to stream logs to. Must use HTTPS.
        --pipeline=PIPELINE        The ID of the Elasticsearch ingest pipeline
                                   to apply pre-process transformations to
                                   before indexing. For example my_pipeline_id.
                                   Learn more about creating a pipeline in the
                                   Elasticsearch docs
                                   (https://www.elastic.co/guide/en/elasticsearch/reference/current/ingest.html)
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --format=FORMAT            Apache style log formatting. Your log must
                                   produce valid JSON that Elasticsearch can
                                   ingest
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute
        --request-max-entries=REQUEST-MAX-ENTRIES
                                   Maximum number of logs to append to a batch,
                                   if non-zero. Defaults to 0 for unbounded
        --request-max-bytes=REQUEST-MAX-BYTES
                                   Maximum size of log batch, if non-zero.
                                   Defaults to 0 for unbounded

  logging elasticsearch delete --version=VERSION --name=NAME [<flags>]
    Delete an Elasticsearch logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Elasticsearch logging object

  logging azureblob create --name=NAME --version=VERSION --container=CONTAINER --account-name=ACCOUNT-NAME --sas-token=SAS-TOKEN [<flags>]
    Create an Azure Blob Storage logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Azure Blob Storage logging
                                 object. Used as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --container=CONTAINER    The name of the Azure Blob Storage container in
                                 which to store logs
        --account-name=ACCOUNT-NAME
                                 The unique Azure Blob Storage namespace in
                                 which your data objects are stored
        --sas-token=SAS-TOKEN    The Azure shared access signature providing
                                 write access to the blob service objects. Be
                                 sure to update your token before it expires or
                                 the logging functionality will not work
        --path=PATH              The path to upload logs to
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk

  logging azureblob list --version=VERSION [<flags>]
    List Azure Blob Storage logging endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging azureblob describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an Azure Blob Storage logging endpoint on a
    Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Azure Blob Storage logging
                                 object

  logging azureblob update --version=VERSION --name=NAME [<flags>]
    Update an Azure Blob Storage logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Azure Blob Storage logging
                                 object
        --new-name=NEW-NAME      New name of the Azure Blob Storage logging
                                 object
        --container=CONTAINER    The name of the Azure Blob Storage container in
                                 which to store logs
        --account-name=ACCOUNT-NAME
                                 The unique Azure Blob Storage namespace in
                                 which your data objects are stored
        --sas-token=SAS-TOKEN    The Azure shared access signature providing
                                 write access to the blob service objects. Be
                                 sure to update your token before it expires or
                                 the logging functionality will not work
        --path=PATH              The path to upload logs to
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk

  logging azureblob delete --version=VERSION --name=NAME [<flags>]
    Delete an Azure Blob Storage logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Azure Blob Storage logging
                                 object

  logging datadog create --name=NAME --version=VERSION --auth-token=AUTH-TOKEN [<flags>]
    Create a Datadog logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Datadog logging object. Used as
                                 a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --auth-token=AUTH-TOKEN  The API key from your Datadog account
        --region=REGION          The region that log data will be sent to. One
                                 of US or EU. Defaults to US if undefined
        --format=FORMAT          Apache style log formatting. For details on the
                                 default value refer to the documentation
                                 (https://developer.fastly.com/reference/api/logging/datadog/)
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging datadog list --version=VERSION [<flags>]
    List Datadog endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging datadog describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Datadog logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Datadog logging object

  logging datadog update --version=VERSION --name=NAME [<flags>]
    Update a Datadog logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Datadog logging object
        --new-name=NEW-NAME      New name of the Datadog logging object
        --auth-token=AUTH-TOKEN  The API key from your Datadog account
        --region=REGION          The region that log data will be sent to. One
                                 of US or EU. Defaults to US if undefined
        --format=FORMAT          Apache style log formatting. For details on the
                                 default value refer to the documentation
                                 (https://developer.fastly.com/reference/api/logging/datadog/)
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging datadog delete --version=VERSION --name=NAME [<flags>]
    Delete a Datadog logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Datadog logging object

  logging https create --name=NAME --version=VERSION --url=URL [<flags>]
    Create an HTTPS logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the HTTPS logging object. Used as
                                   a primary key for API access
    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
        --url=URL                  URL that log data will be sent to. Must use
                                   the https protocol
        --content-type=CONTENT-TYPE
                                   Content type of the header sent with the
                                   request
        --header-name=HEADER-NAME  Name of the custom header sent with the
                                   request
        --header-value=HEADER-VALUE
                                   Value of the custom header sent with the
                                   request
        --method=METHOD            HTTP method used for request. Can be POST or
                                   PUT. Defaults to POST if not specified
        --json-format=JSON-FORMAT  Enforces valid JSON formatting for log
                                   entries. Can be disabled 0, array of json
                                   (wraps JSON log batches in an array) 1, or
                                   newline delimited json (places each JSON log
                                   entry onto a new line in a batch) 2
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --message-type=MESSAGE-TYPE
                                   How the message should be formatted. One of:
                                   classic (default), loggly, logplex or blank
        --format=FORMAT            Apache style log formatting. Your log must
                                   produce valid JSON that HTTPS can ingest
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute
        --request-max-entries=REQUEST-MAX-ENTRIES
                                   Maximum number of logs to append to a batch,
                                   if non-zero. Defaults to 0 for unbounded
        --request-max-bytes=REQUEST-MAX-BYTES
                                   Maximum size of log batch, if non-zero.
                                   Defaults to 0 for unbounded

  logging https list --version=VERSION [<flags>]
    List HTTPS endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging https describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an HTTPS logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the HTTPS logging object

  logging https update --version=VERSION --name=NAME [<flags>]
    Update an HTTPS logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
    -n, --name=NAME                The name of the HTTPS logging object
        --new-name=NEW-NAME        New name of the HTTPS logging object
        --url=URL                  URL that log data will be sent to. Must use
                                   the https protocol
        --content-type=CONTENT-TYPE
                                   Content type of the header sent with the
                                   request
        --header-name=HEADER-NAME  Name of the custom header sent with the
                                   request
        --header-value=HEADER-VALUE
                                   Value of the custom header sent with the
                                   request
        --method=METHOD            HTTP method used for request. Can be POST or
                                   PUT. Defaults to POST if not specified
        --json-format=JSON-FORMAT  Enforces valid JSON formatting for log
                                   entries. Can be disabled 0, array of json
                                   (wraps JSON log batches in an array) 1, or
                                   newline delimited json (places each JSON log
                                   entry onto a new line in a batch) 2
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --message-type=MESSAGE-TYPE
                                   How the message should be formatted. One of:
                                   classic (default), loggly, logplex or blank
        --format=FORMAT            Apache style log formatting. Your log must
                                   produce valid JSON that HTTPS can ingest
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute
        --request-max-entries=REQUEST-MAX-ENTRIES
                                   Maximum number of logs to append to a batch,
                                   if non-zero. Defaults to 0 for unbounded
        --request-max-bytes=REQUEST-MAX-BYTES
                                   Maximum size of log batch, if non-zero.
                                   Defaults to 0 for unbounded

  logging https delete --version=VERSION --name=NAME [<flags>]
    Delete an HTTPS logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the HTTPS logging object

  logging kafka create --name=NAME --version=VERSION --topic=TOPIC --brokers=BROKERS [<flags>]
    Create a Kafka logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Kafka logging object. Used as
                                   a primary key for API access
    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
        --topic=TOPIC              The Kafka topic to send logs to
        --brokers=BROKERS          A comma-separated list of IP addresses or
                                   hostnames of Kafka brokers
        --compression-codec=COMPRESSION-CODEC
                                   The codec used for compression of your logs.
                                   One of: gzip, snappy, lz4
        --required-acks=REQUIRED-ACKS
                                   The Number of acknowledgements a leader must
                                   receive before a write is considered
                                   successful. One of: 1 (default) One server
                                   needs to respond. 0 No servers need to
                                   respond. -1 Wait for all in-sync replicas to
                                   respond
        --use-tls                  Whether to use TLS for secure logging. Can be
                                   either true or false
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --format=FORMAT            Apache style log formatting. Your log must
                                   produce valid JSON that Kafka can ingest
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute

  logging kafka list --version=VERSION [<flags>]
    List Kafka endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging kafka describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Kafka logging endpoint on a Fastly service
    version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Kafka logging object

  logging kafka update --version=VERSION --name=NAME [<flags>]
    Update a Kafka logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID
        --version=VERSION          Number of service version
    -n, --name=NAME                The name of the Kafka logging object
        --new-name=NEW-NAME        New name of the Kafka logging object
        --topic=TOPIC              The Kafka topic to send logs to
        --brokers=BROKERS          A comma-separated list of IP addresses or
                                   hostnames of Kafka brokers
        --compression-codec=COMPRESSION-CODEC
                                   The codec used for compression of your logs.
                                   One of: gzip, snappy, lz4
        --required-acks=REQUIRED-ACKS
                                   The Number of acknowledgements a leader must
                                   receive before a write is considered
                                   successful. One of: 1 (default) One server
                                   needs to respond. 0 No servers need to
                                   respond. -1 Wait for all in-sync replicas to
                                   respond
        --use-tls                  Whether to use TLS for secure logging. Can be
                                   either true or false
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --format=FORMAT            Apache style log formatting. Your log must
                                   produce valid JSON that Kafka can ingest
        --format-version=FORMAT-VERSION
                                   The version of the custom logging format used
                                   for the configured endpoint. Can be either 2
                                   (default) or 1
        --placement=PLACEMENT      Where in the generated VCL the logging call
                                   should be placed, overriding any
                                   format_version default. Can be none or
                                   waf_debug
        --response-condition=RESPONSE-CONDITION
                                   The name of an existing condition in the
                                   configured endpoint, or leave blank to always
                                   execute

  logging kafka delete --version=VERSION --name=NAME [<flags>]
    Delete a Kafka logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Kafka logging object

  logging googlepubsub create --name=NAME --version=VERSION --user=USER --secret-key=SECRET-KEY --topic=TOPIC --project-id=PROJECT-ID [<flags>]
    Create a Google Cloud Pub/Sub logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Google Cloud Pub/Sub logging
                                 object. Used as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --user=USER              Your Google Cloud Platform service account
                                 email address. The client_email field in your
                                 service account authentication JSON
        --secret-key=SECRET-KEY  Your Google Cloud Platform account secret key.
                                 The private_key field in your service account
                                 authentication JSON
        --topic=TOPIC            The Google Cloud Pub/Sub topic to which logs
                                 will be published
        --project-id=PROJECT-ID  The ID of your Google Cloud Platform project
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute

  logging googlepubsub list --version=VERSION [<flags>]
    List Google Cloud Pub/Sub endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging googlepubsub describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Google Cloud Pub/Sub logging endpoint on a
    Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Google Cloud Pub/Sub logging
                                 object

  logging googlepubsub update --version=VERSION --name=NAME [<flags>]
    Update a Google Cloud Pub/Sub logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Google Cloud Pub/Sub logging
                                 object
        --new-name=NEW-NAME      New name of the Google Cloud Pub/Sub logging
                                 object
        --user=USER              Your Google Cloud Platform service account
                                 email address. The client_email field in your
                                 service account authentication JSON
        --secret-key=SECRET-KEY  Your Google Cloud Platform account secret key.
                                 The private_key field in your service account
                                 authentication JSON
        --topic=TOPIC            The Google Cloud Pub/Sub topic to which logs
                                 will be published
        --project-id=PROJECT-ID  The ID of your Google Cloud Platform project
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug. This field
                                 is not required and has no default value
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute

  logging googlepubsub delete --version=VERSION --name=NAME [<flags>]
    Delete a Google Cloud Pub/Sub logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the Google Cloud Pub/Sub logging
                                 object

  logging openstack create --name=NAME --version=VERSION --bucket=BUCKET --access-key=ACCESS-KEY --user=USER --url=URL [<flags>]
    Create an OpenStack logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the OpenStack logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
        --bucket=BUCKET          The name of your OpenStack container
        --access-key=ACCESS-KEY  Your OpenStack account access key
        --user=USER              The username for your OpenStack account
        --url=URL                Your OpenStack auth url
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk
        --path=PATH              The path to upload logs to
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug

  logging openstack list --version=VERSION [<flags>]
    List OpenStack logging endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging openstack describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an OpenStack logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the OpenStack logging object

  logging openstack update --version=VERSION --name=NAME [<flags>]
    Update an OpenStack logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the OpenStack logging object
        --new-name=NEW-NAME      New name of the OpenStack logging object
        --bucket=BUCKET          The name of the Openstack Space
        --access-key=ACCESS-KEY  Your OpenStack account access key
        --user=USER              The username for your OpenStack account.
        --url=URL                Your OpenStack auth url.
        --path=PATH              The path to upload logs to
        --period=PERIOD          How frequently log files are finalized so they
                                 can be available for reading (in seconds,
                                 default 3600)
        --gzip-level=GZIP-LEVEL  What level of GZIP encoding to have when
                                 dumping logs (default 0, no compression)
        --format=FORMAT          Apache style log formatting
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint. Can be either 2
                                 (default) or 1
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint, or leave blank to always
                                 execute
        --message-type=MESSAGE-TYPE
                                 How the message should be formatted. One of:
                                 classic (default), loggly, logplex or blank
        --timestamp-format=TIMESTAMP-FORMAT
                                 strftime specified timestamp formatting
                                 (default "%Y-%m-%dT%H:%M:%S.000")
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed, overriding any format_version
                                 default. Can be none or waf_debug
        --public-key=PUBLIC-KEY  A PGP public key that Fastly will use to
                                 encrypt your log files before writing them to
                                 disk

  logging openstack delete --version=VERSION --name=NAME [<flags>]
    Delete an OpenStack logging endpoint on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -n, --name=NAME              The name of the OpenStack logging object

  stats regions
    List stats regions


  stats historical --service-id=SERVICE-ID [<flags>]
    View historical stats for a Fastly service

    -s, --service-id=SERVICE-ID  Service ID
        --from=FROM              From time, accepted formats at
                                 https://docs.fastly.com/api/stats#Range
        --to=TO                  To time
        --by=BY                  Aggregation period (minute/hour/day)
        --region=REGION          Filter by region ('stats regions' to list)
        --format=FORMAT          Output format (json)

  stats realtime --service-id=SERVICE-ID [<flags>]
    View realtime stats for a Fastly service

    -s, --service-id=SERVICE-ID  Service ID
        --format=FORMAT          Output format (json)

For help on a specific command, try e.g.

	fastly help configure
	fastly configure --help


`) + "\n\n"
