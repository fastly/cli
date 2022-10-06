package common

import (
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/kingpin"
)

// Format defines the format flag
func Format(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("format", "Apache style log formatting. Your log must produce valid JSON").Action(c.Set).StringVar(&c.Value)
}

// GzipLevel defines the gzip flag
func GzipLevel(cmd *kingpin.CmdClause, c *cmd.OptionalUint8) {
	cmd.Flag("gzip-level", "What level of GZIP encoding to have when dumping logs (default 0, no compression)").Action(c.Set).Uint8Var(&c.Value)
}

// Path defines the path flag
func Path(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("path", "The path to upload logs to").Action(c.Set).StringVar(&c.Value)
}

// MessageType defines the path flag
func MessageType(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("message-type", "How the message should be formatted. One of: classic (default), loggly, logplex or blank").Action(c.Set).StringVar(&c.Value)
}

// Period defines the period flag
func Period(cmd *kingpin.CmdClause, c *cmd.OptionalUint) {
	cmd.Flag("period", "How frequently log files are finalized so they can be available for reading (in seconds, default 3600)").Action(c.Set).UintVar(&c.Value)
}

// FormatVersion defines the format-version flag
func FormatVersion(cmd *kingpin.CmdClause, c *cmd.OptionalUint) {
	cmd.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (the default, version 2 log format) or 1 (the version 1 log format). The logging call gets placed by default in vcl_log if format_version is set to 2 and in vcl_deliver if format_version is set to 1").Action(c.Set).UintVar(&c.Value)
}

// CompressionCodec defines the compression-codec flag
func CompressionCodec(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("compression-codec", `The codec used for compression of your logs. Valid values are zstd, snappy, and gzip. If the specified codec is "gzip", gzip_level will default to 3. To specify a different level, leave compression_codec blank and explicitly set the level using gzip_level. Specifying both compression_codec and gzip_level in the same API request will result in an error.`).Action(c.Set).StringVar(&c.Value)
}

// Placement defines the placement flag
func Placement(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug. This field is not required and has no default value").Action(c.Set).StringVar(&c.Value)
}

// ResponseCondition defines the response-condition flag
func ResponseCondition(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.Set).StringVar(&c.Value)
}

// TimestampFormat defines the timestamp-format flag
func TimestampFormat(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("timestamp-format", `strftime specified timestamp formatting (default "%Y-%m-%dT%H:%M:%S.000")`).Action(c.Set).StringVar(&c.Value)
}

// PublicKey defines the public-key flag
func PublicKey(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("public-key", "A PGP public key that Fastly will use to encrypt your log files before writing them to disk").Action(c.Set).StringVar(&c.Value)
}

// TLSCACert defines the tls-ca-cert flag
func TLSCACert(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").Action(c.Set).StringVar(&c.Value)
}

// TLSHostname defines the tls-hostname flag
func TLSHostname(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("tls-hostname", "Used during the TLS handshake to validate the certificate").Action(c.Set).StringVar(&c.Value)
}

// TLSClientCert defines the tls-client-cert flag
func TLSClientCert(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("tls-client-cert", "The client certificate used to make authenticated requests. Must be in PEM format").Action(c.Set).StringVar(&c.Value)
}

// TLSClientKey defines the tls-client-key flag
func TLSClientKey(cmd *kingpin.CmdClause, c *cmd.OptionalString) {
	cmd.Flag("tls-client-key", "The client private key used to make authenticated requests. Must be in PEM format").Action(c.Set).StringVar(&c.Value)
}
