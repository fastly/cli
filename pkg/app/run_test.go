package app_test

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/testutil"
)

func TestApplication(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name:      "no args",
			Args:      nil,
			WantError: helpDefault + "\nERROR: error parsing arguments: command not specified.\n",
		},
		{
			Name:      "help flag only",
			Args:      args("--help"),
			WantError: helpDefault + "\nERROR: error parsing arguments: command not specified.\n",
		},
		{
			Name:      "help argument only",
			Args:      args("help"),
			WantError: fullFatHelpDefault,
		},
		{
			Name:      "help service",
			Args:      args("help service"),
			WantError: helpService,
		},
	}
	// These tests should only verify the app.Run helper wires things up
	// correctly, and check behaviors that can't be associated with a specific
	// command or subcommand. Commands should be tested in their packages,
	// leveraging the app.Run helper as appropriate.
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			err := app.Run(opts)
			if err != nil {
				errors.Deduce(err).Print(&stderr)
			}

			testutil.AssertString(t, testcase.WantError, stripTrailingSpace(stderr.String()))
		})
	}
}

func TestShellCompletion(t *testing.T) {
	args := testutil.Args
	scenarios := []testutil.TestScenario{
		{
			Name: "bash shell complete",
			Args: args("--completion-script-bash"),
			WantOutput: `
_fastly_bash_autocomplete() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[0]} --completion-bash ${COMP_WORDS[@]:1:$COMP_CWORD} )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}
complete -F _fastly_bash_autocomplete fastly

`,
		},
		{
			Name: "zsh shell complete",
			Args: args("--completion-script-zsh"),
			WantOutput: `
#compdef fastly
autoload -U compinit && compinit
autoload -U bashcompinit && bashcompinit

_fastly_bash_autocomplete() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[0]} --completion-bash ${COMP_WORDS[@]:1:$COMP_CWORD} )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    [[ $COMPREPLY ]] && return
    compgen -f
    return 0
}
complete -F _fastly_bash_autocomplete fastly
`,
		},
		{
			Name: "shell evaluate completion options",
			Args: args("--completion-bash"),
			WantOutput: `help
acl
acl-entry
auth-token
backend
compute
config
dictionary
dictionary-item
domain
healthcheck
ip-list
log-tail
logging
pops
profile
purge
service
service-version
stats
update
user
vcl
version
whoami
`,
		},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.Name, func(t *testing.T) {
			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			// NOTE: The Kingpin dependency internally overrides our stdout
			// variable when doing shell completion to the os.Stdout variable and so
			// in order for us to verify it contains the shell completion output, we
			// need an os.Pipe so we can copy off anything written to os.Stdout.
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			outC := make(chan string)

			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, r)
				outC <- buf.String()
			}()

			opts := testutil.NewRunOpts(testcase.Args, &stdout)
			err := app.Run(opts)
			if err != nil {
				errors.Deduce(err).Print(&stderr)
			}

			w.Close()
			os.Stdout = old
			out := <-outC

			testutil.AssertString(t, testcase.WantOutput, stripTrailingSpace(out))
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
      --help             Show context-sensitive help.
  -d, --accept-defaults  Accept default options for all interactive prompts
                         apart from Yes/No confirmations
  -y, --auto-yes         Answer yes automatically to all Yes/No confirmations.
                         This may suppress security warnings
  -i, --non-interactive  Do not prompt for user input - suitable for CI
                         processes. Equivalent to --accept-defaults and
                         --auto-yes
  -o, --profile=PROFILE  Switch account profile for single command execution
                         (see also: 'fastly profile switch')
  -t, --token=TOKEN      Fastly API token (or via FASTLY_API_TOKEN)
  -v, --verbose          Verbose logging

COMMANDS
  help             Show help.
  acl              Manipulate Fastly ACLs (Access Control Lists)
  acl-entry        Manipulate Fastly ACL (Access Control List) entries
  auth-token       Manage API tokens for Fastly service users
  backend          Manipulate Fastly service version backends
  compute          Manage Compute@Edge packages
  config           Display the Fastly CLI configuration
  dictionary       Manipulate Fastly edge dictionaries
  dictionary-item  Manipulate Fastly edge dictionary items
  domain           Manipulate Fastly service version domains
  healthcheck      Manipulate Fastly service version healthchecks
  ip-list          List Fastly's public IPs
  log-tail         Tail Compute@Edge logs
  logging          Manipulate Fastly service version logging endpoints
  pops             List Fastly datacenters
  profile          Manage user profiles
  purge            Invalidate objects in the Fastly cache
  service          Manipulate Fastly services
  service-version  Manipulate Fastly service versions
  stats            View historical and realtime statistics for a Fastly service
  update           Update the CLI to the latest version
  user             Manipulate users of the Fastly API and web interface
  vcl              Manipulate Fastly service version VCL
  version          Display version information for the Fastly CLI
  whoami           Get information about the currently authenticated account

SEE ALSO
  https://developer.fastly.com/reference/cli/
`) + "\n\n"

var helpService = strings.TrimSpace(`
USAGE
  fastly [<flags>] service

GLOBAL FLAGS
      --help             Show context-sensitive help.
  -d, --accept-defaults  Accept default options for all interactive prompts
                         apart from Yes/No confirmations
  -y, --auto-yes         Answer yes automatically to all Yes/No confirmations.
                         This may suppress security warnings
  -i, --non-interactive  Do not prompt for user input - suitable for CI
                         processes. Equivalent to --accept-defaults and
                         --auto-yes
  -o, --profile=PROFILE  Switch account profile for single command execution
                         (see also: 'fastly profile switch')
  -t, --token=TOKEN      Fastly API token (or via FASTLY_API_TOKEN)
  -v, --verbose          Verbose logging

SUBCOMMANDS

  service create --name=NAME [<flags>]
    Create a Fastly service

    -n, --name=NAME        Service name
        --type=vcl         Service type. Can be one of "wasm" or "vcl", defaults
                           to "vcl".
        --comment=COMMENT  Human-readable comment

  service delete [<flags>]
    Delete a Fastly service

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
    -f, --force                  Force deletion of an active service

  service describe [<flags>]
    Show detailed information about a Fastly service

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  service list [<flags>]
    List Fastly services

        --direction=ascend   Direction in which to sort results
    -j, --json               Render output as JSON
        --page=PAGE          Page number of data set to fetch
        --per-page=PER-PAGE  Number of records per page
        --sort="created"     Field on which to sort

  service search --name=NAME
    Search for a Fastly service by name

    -n, --name=NAME  Service name

  service update [<flags>]
    Update a Fastly service

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
    -n, --name=NAME              Service name
        --comment=COMMENT        Human-readable comment

SEE ALSO
  https://developer.fastly.com/reference/cli/service/

`) + "\n\n"

var fullFatHelpDefault = strings.TrimSpace(`
USAGE
  fastly [<flags>] <command>

A tool to interact with the Fastly API

GLOBAL FLAGS
      --help             Show context-sensitive help.
  -d, --accept-defaults  Accept default options for all interactive prompts
                         apart from Yes/No confirmations
  -y, --auto-yes         Answer yes automatically to all Yes/No confirmations.
                         This may suppress security warnings
  -i, --non-interactive  Do not prompt for user input - suitable for CI
                         processes. Equivalent to --accept-defaults and
                         --auto-yes
  -o, --profile=PROFILE  Switch account profile for single command execution
                         (see also: 'fastly profile switch')
  -t, --token=TOKEN      Fastly API token (or via FASTLY_API_TOKEN)
  -v, --verbose          Verbose logging

COMMANDS
  help [<command> ...]
    Show help.


  acl create --name=NAME --version=VERSION [<flags>]
    Create a new ACL attached to the specified service version

        --name=NAME              Name for the ACL. Must start with an
                                 alphanumeric character and contain only
                                 alphanumeric characters, underscores, and
                                 whitespace
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  acl delete --name=NAME --version=VERSION [<flags>]
    Delete an ACL from the specified service version

        --name=NAME              The name of the ACL to delete
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  acl describe --name=NAME --version=VERSION [<flags>]
    Retrieve a single ACL by name for the version and service

        --name=NAME              The name of the ACL
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  acl list --version=VERSION [<flags>]
    List ACLs

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  acl update --name=NAME --new-name=NEW-NAME --version=VERSION [<flags>]
    Update an ACL for a particular service and version

        --name=NAME              The name of the ACL to update
        --new-name=NEW-NAME      The new name of the ACL
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  acl-entry create --acl-id=ACL-ID --ip=IP [<flags>]
    Add an ACL entry to an ACL

        --acl-id=ACL-ID          Alphanumeric string identifying a ACL
        --ip=IP                  An IP address
        --comment=COMMENT        A freeform descriptive note
        --negated                Whether to negate the match
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --subnet=SUBNET          Number of bits for the subnet mask applied to
                                 the IP address

  acl-entry delete --acl-id=ACL-ID --id=ID [<flags>]
    Delete an ACL entry from a specified ACL

        --acl-id=ACL-ID          Alphanumeric string identifying a ACL
        --id=ID                  Alphanumeric string identifying an ACL Entry
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  acl-entry describe --acl-id=ACL-ID --id=ID [<flags>]
    Retrieve a single ACL entry

        --acl-id=ACL-ID          Alphanumeric string identifying a ACL
        --id=ID                  Alphanumeric string identifying an ACL Entry
    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  acl-entry list --acl-id=ACL-ID [<flags>]
    List ACLs

        --acl-id=ACL-ID          Alphanumeric string identifying a ACL
        --direction=ascend       Direction in which to sort results
    -j, --json                   Render output as JSON
        --page=PAGE              Page number of data set to fetch
        --per-page=PER-PAGE      Number of records per page
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --sort="created"         Field on which to sort

  acl-entry update --acl-id=ACL-ID [<flags>]
    Update an ACL entry for a specified ACL

        --acl-id=ACL-ID          Alphanumeric string identifying a ACL
        --comment=COMMENT        A freeform descriptive note
        --file=FILE              Batch update json passed as file path or
                                 content, e.g. $(< batch.json)
        --id=ID                  Alphanumeric string identifying an ACL Entry
        --ip=IP                  An IP address
        --negated                Whether to negate the match
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --subnet=SUBNET          Number of bits for the subnet mask applied to
                                 the IP address

  auth-token create --password=PASSWORD [<flags>]
    Create an API token

    --password=PASSWORD      User password corresponding with --token or
                             $FASTLY_API_TOKEN
    --expires=EXPIRES        Time-stamp (UTC) of when the token will expire
    --name=NAME              Name of the token
    --scope=SCOPE ...        Authorization scope (repeat flag per scope)
    --services=SERVICES ...  A comma-separated list of alphanumeric strings
                             identifying services (default: access to all
                             services)

  auth-token delete [<flags>]
    Revoke an API token

    --current    Revoke the token used to authenticate the request
    --file=FILE  Revoke tokens in bulk from a newline delimited list of tokens
    --id=ID      Alphanumeric string identifying a token

  auth-token describe [<flags>]
    Get the current API token

    -j, --json  Render output as JSON

  auth-token list [<flags>]
    List API tokens

        --customer-id=CUSTOMER-ID  Alphanumeric string identifying the customer
                                   (falls back to FASTLY_CUSTOMER_ID)
    -j, --json                     Render output as JSON

  backend create --version=VERSION --name=NAME --address=ADDRESS [<flags>]
    Create a backend on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
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
        --ssl-ciphers=SSL-CIPHERS  Colon delimited list of OpenSSL ciphers (see
                                   https://www.openssl.org/docs/man1.0.2/man1/ciphers
                                   for details)

  backend delete --version=VERSION --name=NAME [<flags>]
    Delete a backend on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              Backend name

  backend describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a backend on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              Name of backend

  backend list --version=VERSION [<flags>]
    List backends on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  backend update --version=VERSION --name=NAME [<flags>]
    Update a backend on a Fastly service version

    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
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
        --ssl-ciphers=SSL-CIPHERS  Colon delimited list of OpenSSL ciphers (see
                                   https://www.openssl.org/docs/man1.0.2/man1/ciphers
                                   for details)

  compute build [<flags>]
    Build a Compute@Edge package locally

    --include-source     Include source code in built package
    --language=LANGUAGE  Language type
    --name=NAME          Package name
    --skip-verification  Skip verification steps and force build
    --timeout=TIMEOUT    Timeout, in seconds, for the build compilation step

  compute deploy [<flags>]
    Deploy a package to a Fastly Compute@Edge service

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --comment=COMMENT        Human-readable comment
        --domain=DOMAIN          The name of the domain associated to the
                                 package
        --name=NAME              Package name
    -p, --package=PACKAGE        Path to a package tar.gz

  compute init [<flags>]
    Initialize a new Compute@Edge package locally

    -n, --name=NAME                Name of package, falls back to --directory
        --description=DESCRIPTION  Description of the package
    -p, --directory=DIRECTORY      Destination to write the new package,
                                   defaulting to the current directory
    -a, --author=AUTHOR ...        Author(s) of the package
    -l, --language=LANGUAGE        Language of the package
    -f, --from=FROM                Local project directory, or Git repository
                                   URL, or URL referencing a .zip/.tar.gz file,
                                   containing a package template
        --force                    Skip non-empty directory verification step
                                   and force new project creation

  compute pack --wasm-binary=WASM-BINARY
    Package a pre-compiled Wasm binary for a Fastly Compute@Edge service

    -w, --wasm-binary=WASM-BINARY  Path to a pre-compiled Wasm binary

  compute publish [<flags>]
    Build and deploy a Compute@Edge package to a Fastly service

        --comment=COMMENT        Human-readable comment
        --domain=DOMAIN          The name of the domain associated to the
                                 package
        --include-source         Include source code in built package
        --language=LANGUAGE      Language type
        --name=NAME              Package name
    -p, --package=PACKAGE        Path to a package tar.gz
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --skip-verification      Skip verification steps and force build
        --timeout=TIMEOUT        Timeout, in seconds, for the build compilation
                                 step

  compute serve [<flags>]
    Build and run a Compute@Edge package locally

    --addr="127.0.0.1:7676"  The IPv4 address and port to listen on
    --env=ENV                The environment configuration to use (e.g. stage)
    --file="bin/main.wasm"   The Wasm file to run
    --include-source         Include source code in built package
    --language=LANGUAGE      Language type
    --name=NAME              Package name
    --skip-build             Skip the build step
    --skip-verification      Skip verification steps and force build
    --timeout=TIMEOUT        Timeout, in seconds, for the build compilation step
    --watch                  Watch for file changes, then rebuild project and
                             restart local server

  compute update --version=VERSION --package=PACKAGE [<flags>]
    Update a package on a Fastly Compute@Edge service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -p, --package=PACKAGE        Path to a package tar.gz

  compute validate --package=PACKAGE
    Validate a Compute@Edge package

    -p, --package=PACKAGE  Path to a package tar.gz

  config [<flags>]
    Display the Fastly CLI configuration

    -l, --location  Print the location of the CLI configuration file

  dictionary create --version=VERSION --name=NAME [<flags>]
    Create a Fastly edge dictionary on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              Name of Dictionary
        --write-only=WRITE-ONLY  Whether to mark this dictionary as write-only.
                                 Can be true or false (defaults to false)

  dictionary delete --version=VERSION --name=NAME [<flags>]
    Delete a Fastly edge dictionary from a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              Name of Dictionary

  dictionary describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Fastly edge dictionary

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              Name of Dictionary

  dictionary list --version=VERSION [<flags>]
    List all dictionaries on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  dictionary update --version=VERSION --name=NAME [<flags>]
    Update name of dictionary on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              Old name of Dictionary
        --new-name=NEW-NAME      New name of Dictionary
        --write-only=WRITE-ONLY  Whether to mark this dictionary as write-only.
                                 Can be true or false (defaults to false)

  dictionary-item create --dictionary-id=DICTIONARY-ID --key=KEY --value=VALUE [<flags>]
    Create a new item on a Fastly edge dictionary

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --dictionary-id=DICTIONARY-ID
                                 Dictionary ID
        --key=KEY                Dictionary item key
        --value=VALUE            Dictionary item value

  dictionary-item delete --dictionary-id=DICTIONARY-ID --key=KEY [<flags>]
    Delete an item from a Fastly edge dictionary

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --dictionary-id=DICTIONARY-ID
                                 Dictionary ID
        --key=KEY                Dictionary item key

  dictionary-item describe --dictionary-id=DICTIONARY-ID --key=KEY [<flags>]
    Show detailed information about a Fastly edge dictionary item

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --dictionary-id=DICTIONARY-ID
                                 Dictionary ID
        --key=KEY                Dictionary item key

  dictionary-item list --dictionary-id=DICTIONARY-ID [<flags>]
    List items in a Fastly edge dictionary

        --dictionary-id=DICTIONARY-ID
                                 Dictionary ID
        --direction=ascend       Direction in which to sort results
    -j, --json                   Render output as JSON
        --page=PAGE              Page number of data set to fetch
        --per-page=PER-PAGE      Number of records per page
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --sort="created"         Field on which to sort

  dictionary-item update --dictionary-id=DICTIONARY-ID [<flags>]
    Update or insert an item on a Fastly edge dictionary

        --dictionary-id=DICTIONARY-ID
                                 Dictionary ID
        --file=FILE              Batch update json file
        --key=KEY                Dictionary item key
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --value=VALUE            Dictionary item value

  domain create --name=NAME --version=VERSION [<flags>]
    Create a domain on a Fastly service version

    -n, --name=NAME              Domain name
        --comment=COMMENT        A descriptive note
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.

  domain delete --name=NAME --version=VERSION [<flags>]
    Delete a domain on a Fastly service version

    -n, --name=NAME              Domain name
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.

  domain describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a domain on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              Name of domain

  domain list --version=VERSION [<flags>]
    List domains on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  domain update --version=VERSION --name=NAME [<flags>]
    Update a domain on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              Domain name
        --new-name=NEW-NAME      New domain name
        --comment=COMMENT        A descriptive note

  domain validate --version=VERSION [<flags>]
    Checks the status of a specific domain's DNS record for a Service Version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -a, --all                    Checks the status of all domains' DNS records
                                 for a Service Version
    -n, --name=NAME              The name of the domain associated with this
                                 service
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  healthcheck create --version=VERSION --name=NAME [<flags>]
    Create a healthcheck on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
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

  healthcheck delete --version=VERSION --name=NAME [<flags>]
    Delete a healthcheck on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              Healthcheck name

  healthcheck describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a healthcheck on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              Name of healthcheck

  healthcheck list --version=VERSION [<flags>]
    List healthchecks on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  healthcheck update --version=VERSION --name=NAME [<flags>]
    Update a healthcheck on a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
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

  ip-list
    List Fastly's public IPs


  log-tail [<flags>]
    Tail Compute@Edge logs

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --from=FROM              From time, in Unix seconds
        --to=TO                  To time, in Unix seconds
        --sort-buffer=1s         Duration of sort buffer for received logs
        --search-padding=2s      Time beyond from/to to consider in searches
        --stream=STREAM          Output: stdout, stderr, both (default)

  logging azureblob create --name=NAME --version=VERSION --container=CONTAINER --account-name=ACCOUNT-NAME --sas-token=SAS-TOKEN [<flags>]
    Create an Azure Blob Storage logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Azure Blob Storage logging
                                 object. Used as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --container=CONTAINER    The name of the Azure Blob Storage container in
                                 which to store logs
        --account-name=ACCOUNT-NAME
                                 The unique Azure Blob Storage namespace in
                                 which your data objects are stored
        --sas-token=SAS-TOKEN    The Azure shared access signature providing
                                 write access to the blob service objects. Be
                                 sure to update your token before it expires or
                                 the logging functionality will not work
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --file-max-bytes=FILE-MAX-BYTES
                                 The maximum size of a log file in bytes
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging azureblob delete --version=VERSION --name=NAME [<flags>]
    Delete an Azure Blob Storage logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Azure Blob Storage logging
                                 object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging azureblob describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an Azure Blob Storage logging endpoint on a
    Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Azure Blob Storage logging
                                 object

  logging azureblob list --version=VERSION [<flags>]
    List Azure Blob Storage logging endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging azureblob update --version=VERSION --name=NAME [<flags>]
    Update an Azure Blob Storage logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Azure Blob Storage logging
                                 object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --file-max-bytes=FILE-MAX-BYTES
                                 The maximum size of a log file in bytes
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging bigquery create --name=NAME --version=VERSION --project-id=PROJECT-ID --dataset=DATASET --table=TABLE --user=USER --secret-key=SECRET-KEY [<flags>]
    Create a BigQuery logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the BigQuery logging object. Used
                                 as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --project-id=PROJECT-ID  Your Google Cloud Platform project ID
        --dataset=DATASET        Your BigQuery dataset
        --table=TABLE            Your BigQuery table
        --user=USER              Your Google Cloud Platform service account
                                 email address. The client_email field in your
                                 service account authentication JSON.
        --secret-key=SECRET-KEY  Your Google Cloud Platform account secret key.
                                 The private_key field in your service account
                                 authentication JSON.
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the BigQuery logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging bigquery describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a BigQuery logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the BigQuery logging object

  logging bigquery list --version=VERSION [<flags>]
    List BigQuery endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging bigquery update --version=VERSION --name=NAME [<flags>]
    Update a BigQuery logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the BigQuery logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging cloudfiles create --name=NAME --version=VERSION --user=USER --access-key=ACCESS-KEY --bucket=BUCKET [<flags>]
    Create a Cloudfiles logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Cloudfiles logging object. Used
                                 as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --user=USER              The username for your Cloudfile account
        --access-key=ACCESS-KEY  Your Cloudfile account access key
        --bucket=BUCKET          The name of your Cloudfiles container
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging cloudfiles delete --version=VERSION --name=NAME [<flags>]
    Delete a Cloudfiles logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Cloudfiles logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging cloudfiles describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Cloudfiles logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Cloudfiles logging object

  logging cloudfiles list --version=VERSION [<flags>]
    List Cloudfiles endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging cloudfiles update --version=VERSION --name=NAME [<flags>]
    Update a Cloudfiles logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Cloudfiles logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging datadog create --name=NAME --version=VERSION --auth-token=AUTH-TOKEN [<flags>]
    Create a Datadog logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Datadog logging object. Used as
                                 a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --auth-token=AUTH-TOKEN  The API key from your Datadog account
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Datadog logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging datadog describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Datadog logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Datadog logging object

  logging datadog list --version=VERSION [<flags>]
    List Datadog endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging datadog update --version=VERSION --name=NAME [<flags>]
    Update a Datadog logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Datadog logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging digitalocean create --name=NAME --version=VERSION --bucket=BUCKET --access-key=ACCESS-KEY --secret-key=SECRET-KEY [<flags>]
    Create a DigitalOcean Spaces logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the DigitalOcean Spaces logging
                                 object. Used as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --bucket=BUCKET          The name of the DigitalOcean Space
        --access-key=ACCESS-KEY  Your DigitalOcean Spaces account access key
        --secret-key=SECRET-KEY  Your DigitalOcean Spaces account secret key
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging digitalocean delete --version=VERSION --name=NAME [<flags>]
    Delete a DigitalOcean Spaces logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the DigitalOcean Spaces logging
                                 object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging digitalocean describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a DigitalOcean Spaces logging endpoint on a
    Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the DigitalOcean Spaces logging
                                 object

  logging digitalocean list --version=VERSION [<flags>]
    List DigitalOcean Spaces logging endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging digitalocean update --version=VERSION --name=NAME [<flags>]
    Update a DigitalOcean Spaces logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the DigitalOcean Spaces logging
                                 object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging elasticsearch create --name=NAME --version=VERSION --index=INDEX --url=URL [<flags>]
    Create an Elasticsearch logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Elasticsearch logging object.
                                   Used as a primary key for API access
        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
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
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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
                                   if non-zero. Defaults to 10k
        --request-max-bytes=REQUEST-MAX-BYTES
                                   Maximum size of log batch, if non-zero.
                                   Defaults to 100MB

  logging elasticsearch delete --version=VERSION --name=NAME [<flags>]
    Delete an Elasticsearch logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Elasticsearch logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging elasticsearch describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an Elasticsearch logging endpoint on a
    Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Elasticsearch logging object

  logging elasticsearch list --version=VERSION [<flags>]
    List Elasticsearch endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging elasticsearch update --version=VERSION --name=NAME [<flags>]
    Update an Elasticsearch logging endpoint on a Fastly service version

        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
    -n, --name=NAME                The name of the Elasticsearch logging object
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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
                                   if non-zero. Defaults to 10k
        --request-max-bytes=REQUEST-MAX-BYTES
                                   Maximum size of log batch, if non-zero.
                                   Defaults to 100MB

  logging ftp create --name=NAME --version=VERSION --address=ADDRESS --user=USER --password=PASSWORD [<flags>]
    Create an FTP logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the FTP logging object. Used as a
                                 primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --address=ADDRESS        An hostname or IPv4 address
        --user=USER              The username for the server (can be anonymous)
        --password=PASSWORD      The password for the server (for anonymous use
                                 an email address)
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging ftp delete --version=VERSION --name=NAME [<flags>]
    Delete an FTP logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the FTP logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging ftp describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an FTP logging endpoint on a Fastly service
    version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the FTP logging object

  logging ftp list --version=VERSION [<flags>]
    List FTP endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging ftp update --version=VERSION --name=NAME [<flags>]
    Update an FTP logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the FTP logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging gcs create --name=NAME --version=VERSION --user=USER --bucket=BUCKET --secret-key=SECRET-KEY [<flags>]
    Create a GCS logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the GCS logging object. Used as a
                                 primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --user=USER              Your GCS service account email address. The
                                 client_email field in your service account
                                 authentication JSON
        --bucket=BUCKET          The bucket of the GCS bucket
        --secret-key=SECRET-KEY  Your GCS account secret key. The private_key
                                 field in your service account authentication
                                 JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging gcs delete --version=VERSION --name=NAME [<flags>]
    Delete a GCS logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the GCS logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging gcs describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a GCS logging endpoint on a Fastly service
    version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the GCS logging object

  logging gcs list --version=VERSION [<flags>]
    List GCS endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging gcs update --version=VERSION --name=NAME [<flags>]
    Update a GCS logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the GCS logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging googlepubsub create --name=NAME --version=VERSION --user=USER --secret-key=SECRET-KEY --topic=TOPIC --project-id=PROJECT-ID [<flags>]
    Create a Google Cloud Pub/Sub logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Google Cloud Pub/Sub logging
                                 object. Used as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --user=USER              Your Google Cloud Platform service account
                                 email address. The client_email field in your
                                 service account authentication JSON
        --secret-key=SECRET-KEY  Your Google Cloud Platform account secret key.
                                 The private_key field in your service account
                                 authentication JSON
        --topic=TOPIC            The Google Cloud Pub/Sub topic to which logs
                                 will be published
        --project-id=PROJECT-ID  The ID of your Google Cloud Platform project
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Google Cloud Pub/Sub logging
                                 object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging googlepubsub describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Google Cloud Pub/Sub logging endpoint on a
    Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Google Cloud Pub/Sub logging
                                 object

  logging googlepubsub list --version=VERSION [<flags>]
    List Google Cloud Pub/Sub endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging googlepubsub update --version=VERSION --name=NAME [<flags>]
    Update a Google Cloud Pub/Sub logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Google Cloud Pub/Sub logging
                                 object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging heroku create --name=NAME --version=VERSION --url=URL --auth-token=AUTH-TOKEN [<flags>]
    Create a Heroku logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Heroku logging object. Used as
                                 a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --url=URL                The url to stream logs to
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://devcenter.heroku.com/articles/add-on-partner-log-integration)
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging heroku delete --version=VERSION --name=NAME [<flags>]
    Delete a Heroku logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Heroku logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging heroku describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Heroku logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Heroku logging object

  logging heroku list --version=VERSION [<flags>]
    List Heroku endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging heroku update --version=VERSION --name=NAME [<flags>]
    Update a Heroku logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Heroku logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging honeycomb create --name=NAME --version=VERSION --dataset=DATASET --auth-token=AUTH-TOKEN [<flags>]
    Create a Honeycomb logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Honeycomb logging object. Used
                                 as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --dataset=DATASET        The Honeycomb Dataset you want to log to
        --auth-token=AUTH-TOKEN  The Write Key from the Account page of your
                                 Honeycomb account
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging honeycomb delete --version=VERSION --name=NAME [<flags>]
    Delete a Honeycomb logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Honeycomb logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging honeycomb describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Honeycomb logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Honeycomb logging object

  logging honeycomb list --version=VERSION [<flags>]
    List Honeycomb endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging honeycomb update --version=VERSION --name=NAME [<flags>]
    Update a Honeycomb logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Honeycomb logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging https create --name=NAME --version=VERSION --url=URL [<flags>]
    Create an HTTPS logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the HTTPS logging object. Used as
                                   a primary key for API access
        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
        --url=URL                  URL that log data will be sent to. Must use
                                   the https protocol
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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
                                   if non-zero. Defaults to 10k
        --request-max-bytes=REQUEST-MAX-BYTES
                                   Maximum size of log batch, if non-zero.
                                   Defaults to 100MB

  logging https delete --version=VERSION --name=NAME [<flags>]
    Delete an HTTPS logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the HTTPS logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging https describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an HTTPS logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the HTTPS logging object

  logging https list --version=VERSION [<flags>]
    List HTTPS endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging https update --version=VERSION --name=NAME [<flags>]
    Update an HTTPS logging endpoint on a Fastly service version

        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
    -n, --name=NAME                The name of the HTTPS logging object
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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
                                   if non-zero. Defaults to 10k
        --request-max-bytes=REQUEST-MAX-BYTES
                                   Maximum size of log batch, if non-zero.
                                   Defaults to 100MB

  logging kafka create --name=NAME --version=VERSION --topic=TOPIC --brokers=BROKERS [<flags>]
    Create a Kafka logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Kafka logging object. Used as
                                   a primary key for API access
        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
        --topic=TOPIC              The Kafka topic to send logs to
        --brokers=BROKERS          A comma-separated list of IP addresses or
                                   hostnames of Kafka brokers
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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
        --parse-log-keyvals        Parse key-value pairs within the log format
        --max-batch-size=MAX-BATCH-SIZE
                                   The maximum size of the log batch in bytes
        --use-sasl                 Enable SASL authentication. Requires
                                   --auth-method, --username, and --password to
                                   be specified
        --auth-method=AUTH-METHOD  SASL authentication method. Valid values are:
                                   plain, scram-sha-256, scram-sha-512
        --username=USERNAME        SASL authentication username. Required if
                                   --auth-method is specified
        --password=PASSWORD        SASL authentication password. Required if
                                   --auth-method is specified

  logging kafka delete --version=VERSION --name=NAME [<flags>]
    Delete a Kafka logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Kafka logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging kafka describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Kafka logging endpoint on a Fastly service
    version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Kafka logging object

  logging kafka list --version=VERSION [<flags>]
    List Kafka endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging kafka update --version=VERSION --name=NAME [<flags>]
    Update a Kafka logging endpoint on a Fastly service version

        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
    -n, --name=NAME                The name of the Kafka logging object
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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
        --[no-]parse-log-keyvals   Parse key-value pairs within the log format
        --max-batch-size=MAX-BATCH-SIZE
                                   The maximum size of the log batch in bytes
        --use-sasl                 Enable SASL authentication. Requires
                                   --auth-method, --username, and --password to
                                   be specified
        --auth-method=AUTH-METHOD  SASL authentication method. Valid values are:
                                   plain, scram-sha-256, scram-sha-512
        --username=USERNAME        SASL authentication username. Required if
                                   --auth-method is specified
        --password=PASSWORD        SASL authentication password. Required if
                                   --auth-method is specified

  logging kinesis create --name=NAME --version=VERSION --stream-name=STREAM-NAME --region=REGION [<flags>]
    Create an Amazon Kinesis logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Kinesis logging object. Used
                                   as a primary key for API access
        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --stream-name=STREAM-NAME  The Amazon Kinesis stream to send logs to
        --region=REGION            The AWS region where the Kinesis stream
                                   exists
        --access-key=ACCESS-KEY    The access key associated with the target
                                   Amazon Kinesis stream
        --secret-key=SECRET-KEY    The secret key associated with the target
                                   Amazon Kinesis stream
        --iam-role=IAM-ROLE        The IAM role ARN for logging
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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

  logging kinesis delete --version=VERSION --name=NAME [<flags>]
    Delete a Kinesis logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Kinesis logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging kinesis describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Kinesis logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Kinesis logging object

  logging kinesis list --version=VERSION [<flags>]
    List Kinesis endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging kinesis update --version=VERSION --name=NAME [<flags>]
    Update a Kinesis logging endpoint on a Fastly service version

        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
    -n, --name=NAME                The name of the Kinesis logging object
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
        --new-name=NEW-NAME        New name of the Kinesis logging object
        --stream-name=STREAM-NAME  Your Kinesis stream name
        --access-key=ACCESS-KEY    Your Kinesis account access key
        --secret-key=SECRET-KEY    Your Kinesis account secret key
        --iam-role=IAM-ROLE        The IAM role ARN for logging
        --region=REGION            The AWS region where the Kinesis stream
                                   exists
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

  logging logentries create --name=NAME --version=VERSION [<flags>]
    Create a Logentries logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Logentries logging object. Used
                                 as a primary key for API access
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --region=REGION          The region to which to stream logs
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.

  logging logentries delete --version=VERSION --name=NAME [<flags>]
    Delete a Logentries logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Logentries logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging logentries describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Logentries logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Logentries logging object

  logging logentries list --version=VERSION [<flags>]
    List Logentries endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging logentries update --version=VERSION --name=NAME [<flags>]
    Update a Logentries logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Logentries logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --region=REGION          The region to which to stream logs

  logging loggly create --name=NAME --version=VERSION --auth-token=AUTH-TOKEN [<flags>]
    Create a Loggly logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Loggly logging object. Used as
                                 a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://www.loggly.com/docs/customer-token-authentication-token/)
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Loggly logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging loggly describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Loggly logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Loggly logging object

  logging loggly list --version=VERSION [<flags>]
    List Loggly endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging loggly update --version=VERSION --name=NAME [<flags>]
    Update a Loggly logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Loggly logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging logshuttle create --name=NAME --version=VERSION --url=URL --auth-token=AUTH-TOKEN [<flags>]
    Create a Logshuttle logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Logshuttle logging object. Used
                                 as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --url=URL                Your Log Shuttle endpoint url
        --auth-token=AUTH-TOKEN  The data authentication token associated with
                                 this endpoint
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging logshuttle delete --version=VERSION --name=NAME [<flags>]
    Delete a Logshuttle logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Logshuttle logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging logshuttle describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Logshuttle logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Logshuttle logging object

  logging logshuttle list --version=VERSION [<flags>]
    List Logshuttle endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging logshuttle update --version=VERSION --name=NAME [<flags>]
    Update a Logshuttle logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Logshuttle logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging newrelic create --key=KEY --name=NAME --version=VERSION [<flags>]
    Create an New Relic logging endpoint attached to the specified service
    version

        --key=KEY                The Insert API key from the Account page of
                                 your New Relic account
        --name=NAME              The name for the real-time logging
                                 configuration
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --format=FORMAT          A Fastly log format string. Must produce valid
                                 JSON that New Relic Logs can ingest
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed
        --region=REGION          The region to which to stream logs
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging newrelic delete --name=NAME --version=VERSION [<flags>]
    Delete the New Relic Logs logging object for a particular service and
    version

        --name=NAME              The name for the real-time logging
                                 configuration to delete
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging newrelic describe --name=NAME --version=VERSION [<flags>]
    Get the details of a New Relic Logs logging object for a particular service
    and version

        --name=NAME              The name for the real-time logging
                                 configuration
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging newrelic list --version=VERSION [<flags>]
    List all of the New Relic Logs logging objects for a particular service and
    version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging newrelic update --name=NAME --version=VERSION [<flags>]
    Update a New Relic Logs logging object for a particular service and version

        --name=NAME              The name for the real-time logging
                                 configuration to update
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --format=FORMAT          A Fastly log format string. Must produce valid
                                 JSON that New Relic Logs can ingest
        --format-version=FORMAT-VERSION
                                 The version of the custom logging format used
                                 for the configured endpoint
        --key=KEY                The Insert API key from the Account page of
                                 your New Relic account
        --new-name=NEW-NAME      The name for the real-time logging
                                 configuration
        --placement=PLACEMENT    Where in the generated VCL the logging call
                                 should be placed
        --region=REGION          The region to which to stream logs
        --response-condition=RESPONSE-CONDITION
                                 The name of an existing condition in the
                                 configured endpoint
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging openstack create --name=NAME --version=VERSION --bucket=BUCKET --access-key=ACCESS-KEY --user=USER --url=URL [<flags>]
    Create an OpenStack logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the OpenStack logging object. Used
                                 as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --bucket=BUCKET          The name of your OpenStack container
        --access-key=ACCESS-KEY  Your OpenStack account access key
        --user=USER              The username for your OpenStack account
        --url=URL                Your OpenStack auth url
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging openstack delete --version=VERSION --name=NAME [<flags>]
    Delete an OpenStack logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the OpenStack logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging openstack describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an OpenStack logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the OpenStack logging object

  logging openstack list --version=VERSION [<flags>]
    List OpenStack logging endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging openstack update --version=VERSION --name=NAME [<flags>]
    Update an OpenStack logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the OpenStack logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging papertrail create --name=NAME --version=VERSION --address=ADDRESS [<flags>]
    Create a Papertrail logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Papertrail logging object. Used
                                 as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --address=ADDRESS        A hostname or IPv4 address
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Papertrail logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging papertrail describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Papertrail logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Papertrail logging object

  logging papertrail list --version=VERSION [<flags>]
    List Papertrail endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging papertrail update --version=VERSION --name=NAME [<flags>]
    Update a Papertrail logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Papertrail logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging s3 create --name=NAME --version=VERSION --bucket=BUCKET [<flags>]
    Create an Amazon S3 logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the S3 logging object. Used as a
                                 primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --bucket=BUCKET          Your S3 bucket name
        --access-key=ACCESS-KEY  Your S3 account access key
        --secret-key=SECRET-KEY  Your S3 account secret key
        --iam-role=IAM-ROLE      The IAM role ARN for logging
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging s3 delete --version=VERSION --name=NAME [<flags>]
    Delete a S3 logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the S3 logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging s3 describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a S3 logging endpoint on a Fastly service
    version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the S3 logging object

  logging s3 list --version=VERSION [<flags>]
    List S3 endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging s3 update --version=VERSION --name=NAME [<flags>]
    Update a S3 logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the S3 logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --new-name=NEW-NAME      New name of the S3 logging object
        --bucket=BUCKET          Your S3 bucket name
        --access-key=ACCESS-KEY  Your S3 account access key
        --secret-key=SECRET-KEY  Your S3 account secret key
        --iam-role=IAM-ROLE      The IAM role ARN for logging
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging scalyr create --name=NAME --version=VERSION --auth-token=AUTH-TOKEN [<flags>]
    Create a Scalyr logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Scalyr logging object. Used as
                                 a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --auth-token=AUTH-TOKEN  The token to use for authentication
                                 (https://www.scalyr.com/keys)
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging scalyr delete --version=VERSION --name=NAME [<flags>]
    Delete a Scalyr logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Scalyr logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging scalyr describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Scalyr logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Scalyr logging object

  logging scalyr list --version=VERSION [<flags>]
    List Scalyr endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging scalyr update --version=VERSION --name=NAME [<flags>]
    Update a Scalyr logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Scalyr logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging sftp create --name=NAME --version=VERSION --address=ADDRESS --user=USER --ssh-known-hosts=SSH-KNOWN-HOSTS [<flags>]
    Create an SFTP logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the SFTP logging object. Used as a
                                 primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --address=ADDRESS        The hostname or IPv4 addres
        --user=USER              The username for the server
        --ssh-known-hosts=SSH-KNOWN-HOSTS
                                 A list of host keys for all hosts we can
                                 connect to over SFTP
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging sftp delete --version=VERSION --name=NAME [<flags>]
    Delete an SFTP logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the SFTP logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging sftp describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about an SFTP logging endpoint on a Fastly service
    version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the SFTP logging object

  logging sftp list --version=VERSION [<flags>]
    List SFTP endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging sftp update --version=VERSION --name=NAME [<flags>]
    Update an SFTP logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the SFTP logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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
        --compression-codec=COMPRESSION-CODEC
                                 The codec used for compression of your logs.
                                 Valid values are zstd, snappy, and gzip. If the
                                 specified codec is "gzip", gzip_level will
                                 default to 3. To specify a different level,
                                 leave compression_codec blank and explicitly
                                 set the level using gzip_level. Specifying both
                                 compression_codec and gzip_level in the same
                                 API request will result in an error.

  logging splunk create --name=NAME --version=VERSION --url=URL [<flags>]
    Create a Splunk logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Splunk logging object. Used
                                   as a primary key for API access
        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
        --url=URL                  The URL to POST to
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
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

  logging splunk delete --version=VERSION --name=NAME [<flags>]
    Delete a Splunk logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Splunk logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging splunk describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Splunk logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Splunk logging object

  logging splunk list --version=VERSION [<flags>]
    List Splunk endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging splunk update --version=VERSION --name=NAME [<flags>]
    Update a Splunk logging endpoint on a Fastly service version

        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
    -n, --name=NAME                The name of the Splunk logging object
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
        --new-name=NEW-NAME        New name of the Splunk logging object
        --url=URL                  The URL to POST to.
        --tls-ca-cert=TLS-CA-CERT  A secure certificate to authenticate the
                                   server with. Must be in PEM format
        --tls-hostname=TLS-HOSTNAME
                                   The hostname used to verify the server's
                                   certificate. It can either be the Common Name
                                   or a Subject Alternative Name (SAN)
        --tls-client-cert=TLS-CLIENT-CERT
                                   The client certificate used to make
                                   authenticated requests. Must be in PEM format
        --tls-client-key=TLS-CLIENT-KEY
                                   The client private key used to make
                                   authenticated requests. Must be in PEM format
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

  logging sumologic create --name=NAME --version=VERSION --url=URL [<flags>]
    Create a Sumologic logging endpoint on a Fastly service version

    -n, --name=NAME              The name of the Sumologic logging object. Used
                                 as a primary key for API access
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --url=URL                The URL to POST to
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Sumologic logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging sumologic describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Sumologic logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Sumologic logging object

  logging sumologic list --version=VERSION [<flags>]
    List Sumologic endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging sumologic update --version=VERSION --name=NAME [<flags>]
    Update a Sumologic logging endpoint on a Fastly service version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Sumologic logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
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

  logging syslog create --name=NAME --version=VERSION --address=ADDRESS [<flags>]
    Create a Syslog logging endpoint on a Fastly service version

    -n, --name=NAME                The name of the Syslog logging object. Used
                                   as a primary key for API access
        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
        --address=ADDRESS          A hostname or IPv4 address
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -n, --name=NAME              The name of the Syslog logging object
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  logging syslog describe --version=VERSION --name=NAME [<flags>]
    Show detailed information about a Syslog logging endpoint on a Fastly
    service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -n, --name=NAME              The name of the Syslog logging object

  logging syslog list --version=VERSION [<flags>]
    List Syslog endpoints on a Fastly service version

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  logging syslog update --version=VERSION --name=NAME [<flags>]
    Update a Syslog logging endpoint on a Fastly service version

        --version=VERSION          'latest', 'active', or the number of a
                                   specific version
        --autoclone                If the selected service version is not
                                   editable, clone it and use the clone.
    -n, --name=NAME                The name of the Syslog logging object
    -s, --service-id=SERVICE-ID    Service ID (falls back to FASTLY_SERVICE_ID,
                                   then fastly.toml)
        --service-name=SERVICE-NAME
                                   The name of the service
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

  pops
    List Fastly datacenters


  profile create [<profile>]
    Create user profile


  profile delete <profile>
    Delete user profile


  profile list
    List user profiles


  profile switch <profile>
    Switch user profile


  profile token [<flags>]
    Print user token

    -u, --user=USER  Profile user to print token for

  profile update [<profile>]
    Update user profile


  purge [<flags>]
    Invalidate objects in the Fastly cache

        --all                    Purge everything from a service
        --file=FILE              Purge a service of a newline delimited list of
                                 Surrogate Keys
        --key=KEY                Purge a service of objects tagged with a
                                 Surrogate Key
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --soft                   A 'soft' purge marks affected objects as stale
                                 rather than making them inaccessible
        --url=URL                Purge an individual URL

  service create --name=NAME [<flags>]
    Create a Fastly service

    -n, --name=NAME        Service name
        --type=vcl         Service type. Can be one of "wasm" or "vcl", defaults
                           to "vcl".
        --comment=COMMENT  Human-readable comment

  service delete [<flags>]
    Delete a Fastly service

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
    -f, --force                  Force deletion of an active service

  service describe [<flags>]
    Show detailed information about a Fastly service

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  service list [<flags>]
    List Fastly services

        --direction=ascend   Direction in which to sort results
    -j, --json               Render output as JSON
        --page=PAGE          Page number of data set to fetch
        --per-page=PER-PAGE  Number of records per page
        --sort="created"     Field on which to sort

  service search --name=NAME
    Search for a Fastly service by name

    -n, --name=NAME  Service name

  service update [<flags>]
    Update a Fastly service

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
    -n, --name=NAME              Service name
        --comment=COMMENT        Human-readable comment

  service-version activate --version=VERSION [<flags>]
    Activate a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.

  service-version clone --version=VERSION [<flags>]
    Clone a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  service-version deactivate --version=VERSION [<flags>]
    Deactivate a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  service-version list [<flags>]
    List Fastly service versions

    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  service-version lock --version=VERSION [<flags>]
    Lock a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version

  service-version update --version=VERSION [<flags>]
    Update a Fastly service version

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --comment=COMMENT        Human-readable comment

  stats historical [<flags>]
    View historical stats for a Fastly service

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --from=FROM              From time, accepted formats at
                                 https://fastly.dev/reference/api/metrics-stats/historical-stats
        --to=TO                  To time
        --by=BY                  Aggregation period (minute/hour/day)
        --region=REGION          Filter by region ('stats regions' to list)
        --format=FORMAT          Output format (json)

  stats realtime [<flags>]
    View realtime stats for a Fastly service

    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --format=FORMAT          Output format (json)

  stats regions
    List stats regions


  update
    Update the CLI to the latest version


  user create --login=LOGIN --name=NAME [<flags>]
    Create a user of the Fastly API and web interface

    --login=LOGIN  The login associated with the user (typically, an email
                   address)
    --name=NAME    The real life name of the user
    --role=ROLE    The permissions role assigned to the user. Can be user,
                   billing, engineer, or superuser

  user delete --id=ID
    Delete a user of the Fastly API and web interface

    --id=ID  Alphanumeric string identifying the user

  user describe [<flags>]
    Get a specific user of the Fastly API and web interface

        --current  Get the logged in user
        --id=ID    Alphanumeric string identifying the user
    -j, --json     Render output as JSON

  user list [<flags>]
    List all users from a specified customer id

        --customer-id=CUSTOMER-ID  Alphanumeric string identifying the customer
                                   (falls back to FASTLY_CUSTOMER_ID)
    -j, --json                     Render output as JSON

  user update [<flags>]
    Update a user of the Fastly API and web interface

    --id=ID           Alphanumeric string identifying the user
    --login=LOGIN     The login associated with the user (typically, an email
                      address)
    --name=NAME       The real life name of the user
    --password-reset  Requests a password reset for the specified user
    --role=ROLE       The permissions role assigned to the user. Can be user,
                      billing, engineer, or superuser

  vcl custom create --content=CONTENT --name=NAME --version=VERSION [<flags>]
    Upload a VCL for a particular service and version

        --content=CONTENT        VCL passed as file path or content, e.g. $(<
                                 main.vcl)
        --name=NAME              The name of the VCL
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --main                   Whether the VCL is the 'main' entrypoint
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  vcl custom delete --name=NAME --version=VERSION [<flags>]
    Delete the uploaded VCL for a particular service and version

        --name=NAME              The name of the VCL to delete
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  vcl custom describe --name=NAME --version=VERSION [<flags>]
    Get the uploaded VCL for a particular service and version

        --name=NAME              The name of the VCL
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  vcl custom list --version=VERSION [<flags>]
    List the uploaded VCLs for a particular service and version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  vcl custom update --name=NAME --version=VERSION [<flags>]
    Update the uploaded VCL for a particular service and version

        --name=NAME              The name of the VCL to update
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --new-name=NEW-NAME      New name for the VCL
        --content=CONTENT        VCL passed as file path or content, e.g. $(<
                                 main.vcl)
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  vcl snippet create --content=CONTENT --name=NAME --version=VERSION --type=TYPE [<flags>]
    Create a snippet for a particular service and version

        --content=CONTENT        VCL snippet passed as file path or content,
                                 e.g. $(< snippet.vcl)
        --name=NAME              The name of the VCL snippet
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --type=TYPE              The location in generated VCL where the snippet
                                 should be placed
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --dynamic                Whether the VCL snippet is dynamic or versioned
    -p, --priority=PRIORITY      Priority determines execution order. Lower
                                 numbers execute first
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  vcl snippet delete --name=NAME --version=VERSION [<flags>]
    Delete a specific snippet for a particular service and version

        --name=NAME              The name of the VCL snippet to delete
        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  vcl snippet describe --version=VERSION [<flags>]
    Get the uploaded VCL snippet for a particular service and version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --dynamic                Whether the VCL snippet is dynamic or versioned
    -j, --json                   Render output as JSON
        --name=NAME              The name of the VCL snippet
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --snippet-id=SNIPPET-ID  Alphanumeric string identifying a VCL Snippet

  vcl snippet list --version=VERSION [<flags>]
    List the uploaded VCL snippets for a particular service and version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
    -j, --json                   Render output as JSON
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service

  vcl snippet update --version=VERSION [<flags>]
    Update a VCL snippet for a particular service and version

        --version=VERSION        'latest', 'active', or the number of a specific
                                 version
        --autoclone              If the selected service version is not
                                 editable, clone it and use the clone.
        --content=CONTENT        VCL snippet passed as file path or content,
                                 e.g. $(< snippet.vcl)
        --dynamic                Whether the VCL snippet is dynamic or versioned
        --name=NAME              The name of the VCL snippet to update
        --new-name=NEW-NAME      New name for the VCL snippet
    -p, --priority=PRIORITY      Priority determines execution order. Lower
                                 numbers execute first
    -s, --service-id=SERVICE-ID  Service ID (falls back to FASTLY_SERVICE_ID,
                                 then fastly.toml)
        --service-name=SERVICE-NAME
                                 The name of the service
        --snippet-id=SNIPPET-ID  Alphanumeric string identifying a VCL Snippet
        --type=TYPE              The location in generated VCL where the snippet
                                 should be placed

  version
    Display version information for the Fastly CLI


  whoami
    Get information about the currently authenticated account


SEE ALSO
  https://developer.fastly.com/reference/cli/

For help on a specific command, try e.g.

	fastly help profile
	fastly profile --help
`) + "\n\n"
