package app_test

import (
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
			wantErr: helpDefault + "\nERROR: error parsing arguments: command not specified.\n\n",
		},
		{
			name:    "help flag only",
			args:    []string{"--help"},
			wantErr: helpDefault + "\nERROR: error parsing arguments: command not specified.\n\n",
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

			testutil.AssertString(t, testcase.wantOut, stdout.String())

			wantErrLines := strings.Split(testcase.wantErr, "\n")
			outputErrLines := strings.Split(stderr.String(), "\n")

			for i, line := range outputErrLines {
				testutil.AssertString(t, strings.TrimSpace(wantErrLines[i]), strings.TrimSpace(line))
			}
		})
	}
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
        --version=VERSION        Number of version you wish to clone
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

    -n, --name=NAME                Name of package, defaulting to directory name
                                   of the --path destination
    -d, --description=DESCRIPTION  Description of the package
    -a, --author=AUTHOR            Author of the package
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
    -d, --name=NAME              Name of domain

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
                                   https://www.openssl.org/docs/manmaster/man1/ciphers.html
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
                                   https://www.openssl.org/docs/manmaster/man1/ciphers.html
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

  logging bigquery list --version=VERSION [<flags>]
    List BigQuery endpoints on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version

  logging bigquery describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a BigQuery logging endpoint on a Fastly
    service version

    -s, --service-id=SERVICE-ID  Service ID
        --version=VERSION        Number of service version
    -d, --name=NAME              The name of the BigQuery logging object

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

For help on a specific command, try e.g.

	fastly help configure
	fastly configure --help
`) + "\n\n"
