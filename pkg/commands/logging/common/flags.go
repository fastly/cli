package common

import (
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/kingpin"
)

// Add format flag to the CmdClause
func Format(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("format", "Apache style log formatting").Action(c.Set).StringVar(&c.Value)
}

// Add gzip-level flag to the CmdClause
func GzipLevel(cmd *kingpin.CmdClause, c *cmd.OptionalUint) {
	cmd.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.Set).UintVar(&c.Value)
}

// Add path flag to the CmdClause
func Path(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("path", "The path to upload logs to").Action(c.Set).StringVar(&c.Value)
}

// Add path flag to the CmdClause
func MessageType(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.Set).StringVar(&c.Value)
}

// Add period flag to the CmdClause
func Period(cmd *kingpin.CmdClause, c *cmd.OptionalUint) {
	cmd.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Set).UintVar(&c.Value)
}

// Add FormatVersion to the CmdClause
func FormatVersion(cmd *kingpin.CmdClause, c *cmd.OptionalUint) {
	cmd.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.Set).UintVar(&c.Value)
}

// Add CompressionCodec to the CmdClause
func CompressionCodec(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("compression-codec", `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`).Action(c.Set).StringVar(&c.Value)
}

// Add Placement to the CmdClause
func Placement(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Set).StringVar(&c.Value)
}

// Add ResponseCondition to the CmdClause
func ResponseCondition(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.Set).StringVar(&c.Value)
}

// Add TimestampFormat to the CmdClause
func TimestampFormat(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.Set).StringVar(&c.Value)
}

// Add PublicKey to the CmdClause
func PublicKey(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("public-key", "A PGP public key that Fastly will use to encrypt your log files before writing them to disk").Action(c.Set).StringVar(&c.Value)
}

// Add TLSCACert to the CmdClause
func TLSCACert(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").Action(c.Set).StringVar(&c.Value)
}

// Add TLSHostname to the CmdClause
func TLSHostname(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("tls-hostname", "Used during the TLS handshake to validate the certificate").Action(c.Set).StringVar(&c.Value)
}

// Add TLSClientCert to the CmdClause
func TLSClientCert(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("tls-client-cert", "The client certificate used to make authenticated requests. Must be in PEM format").Action(c.Set).StringVar(&c.Value)
}

// Add TOTLSClientKeyDO to the CmdClause
func TLSClientKey(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("tls-client-key", "The client private key used to make authenticated requests. Must be in PEM format").Action(c.Set).StringVar(&c.Value)
}
