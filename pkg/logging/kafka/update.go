package kafka

import (
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// UpdateCommand calls the Fastly API to update Kafka logging endpoints.
type UpdateCommand struct {
	common.Base
	manifest manifest.Data

	// required
	EndpointName string // Can't shaddow common.Base method Name().
	Version      int

	// optional
	NewName           common.OptionalString
	Index             common.OptionalString
	Topic             common.OptionalString
	Brokers           common.OptionalString
	UseTLS            common.OptionalBool
	CompressionCodec  common.OptionalString
	RequiredACKs      common.OptionalString
	TLSCACert         common.OptionalString
	TLSClientCert     common.OptionalString
	TLSClientKey      common.OptionalString
	TLSHostname       common.OptionalString
	Format            common.OptionalString
	FormatVersion     common.OptionalUint
	Placement         common.OptionalString
	ResponseCondition common.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent common.Registerer, globals *config.Data) *UpdateCommand {
	var c UpdateCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("update", "Update a Kafka logging endpoint on a Fastly service version")

	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Version)
	c.CmdClause.Flag("name", "The name of the Kafka logging object").Short('n').Required().StringVar(&c.EndpointName)

	c.CmdClause.Flag("new-name", "New name of the Kafka logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	c.CmdClause.Flag("topic", "The Kafka topic to send logs to").Action(c.Topic.Set).StringVar(&c.Topic.Value)
	c.CmdClause.Flag("brokers", "A comma-separated list of IP addresses or hostnames of Kafka brokers").Action(c.Brokers.Set).StringVar(&c.Brokers.Value)
	c.CmdClause.Flag("compression-codec", "The codec used for compression of your logs. One of: gzip, snappy, lz4").Action(c.CompressionCodec.Set).StringVar(&c.CompressionCodec.Value)
	c.CmdClause.Flag("required-acks", "The Number of acknowledgements a leader must receive before a write is considered successful. One of: 1 (default) One server needs to respond. 0	No servers need to respond. -1	Wait for all in-sync replicas to respond").Action(c.RequiredACKs.Set).StringVar(&c.RequiredACKs.Value)
	c.CmdClause.Flag("use-tls", "Whether to use TLS for secure logging. Can be either true or false").Action(c.UseTLS.Set).BoolVar(&c.UseTLS.Value)
	c.CmdClause.Flag("tls-ca-cert", "A secure certificate to authenticate the server with. Must be in PEM format").Action(c.TLSCACert.Set).StringVar(&c.TLSCACert.Value)
	c.CmdClause.Flag("tls-client-cert", "The client certificate used to make authenticated requests. Must be in PEM format").Action(c.TLSClientCert.Set).StringVar(&c.TLSClientCert.Value)
	c.CmdClause.Flag("tls-client-key", "The client private key used to make authenticated requests. Must be in PEM format").Action(c.TLSClientKey.Set).StringVar(&c.TLSClientKey.Value)
	c.CmdClause.Flag("tls-hostname", "The hostname used to verify the server's certificate. It can either be the Common Name or a Subject Alternative Name (SAN)").Action(c.TLSHostname.Set).StringVar(&c.TLSHostname.Value)
	c.CmdClause.Flag("format", "Apache style log formatting. Your log must produce valid JSON that Kafka can ingest").Action(c.Format.Set).StringVar(&c.Format.Value)
	c.CmdClause.Flag("format-version", "The version of the custom logging format used for the configured endpoint. Can be either 2 (default) or 1").Action(c.FormatVersion.Set).UintVar(&c.FormatVersion.Value)
	c.CmdClause.Flag("placement", "Where in the generated VCL the logging call should be placed, overriding any format_version default. Can be none or waf_debug").Action(c.Placement.Set).StringVar(&c.Placement.Value)
	c.CmdClause.Flag("response-condition", "The name of an existing condition in the configured endpoint, or leave blank to always execute").Action(c.ResponseCondition.Set).StringVar(&c.ResponseCondition.Value)

	return &c
}

// createInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) createInput() (*fastly.UpdateKafkaInput, error) {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return nil, errors.ErrNoServiceID
	}

	kafka, err := c.Globals.Client.GetKafka(&fastly.GetKafkaInput{
		Service: serviceID,
		Name:    c.EndpointName,
		Version: c.Version,
	})
	if err != nil {
		return nil, err
	}

	input := fastly.UpdateKafkaInput{
		Service:           kafka.ServiceID,
		Version:           kafka.Version,
		Name:              kafka.Name,
		NewName:           fastly.String(kafka.Name),
		Brokers:           fastly.String(kafka.Brokers),
		Topic:             fastly.String(kafka.Topic),
		RequiredACKs:      fastly.String(kafka.RequiredACKs),
		UseTLS:            fastly.CBool(kafka.UseTLS),
		CompressionCodec:  fastly.String(kafka.CompressionCodec),
		Format:            fastly.String(kafka.Format),
		FormatVersion:     fastly.Uint(kafka.FormatVersion),
		ResponseCondition: fastly.String(kafka.ResponseCondition),
		Placement:         fastly.String(kafka.Placement),
		TLSCACert:         fastly.String(kafka.TLSCACert),
		TLSHostname:       fastly.String(kafka.TLSHostname),
		TLSClientCert:     fastly.String(kafka.TLSClientCert),
		TLSClientKey:      fastly.String(kafka.TLSClientKey),
	}

	if c.NewName.Valid {
		input.NewName = fastly.String(c.NewName.Value)
	}

	if c.Topic.Valid {
		input.Topic = fastly.String(c.Topic.Value)
	}

	if c.Brokers.Valid {
		input.Brokers = fastly.String(c.Brokers.Value)
	}

	if c.CompressionCodec.Valid {
		input.CompressionCodec = fastly.String(c.CompressionCodec.Value)
	}

	if c.RequiredACKs.Valid {
		input.RequiredACKs = fastly.String(c.RequiredACKs.Value)
	}

	if c.UseTLS.Valid {
		input.UseTLS = fastly.CBool(c.UseTLS.Value)
	}

	if c.TLSCACert.Valid {
		input.TLSCACert = fastly.String(c.TLSCACert.Value)
	}

	if c.TLSClientCert.Valid {
		input.TLSClientCert = fastly.String(c.TLSClientCert.Value)
	}

	if c.TLSClientKey.Valid {
		input.TLSClientKey = fastly.String(c.TLSClientKey.Value)
	}

	if c.TLSHostname.Valid {
		input.TLSHostname = fastly.String(c.TLSHostname.Value)
	}

	if c.Format.Valid {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.Valid {
		input.FormatVersion = fastly.Uint(c.FormatVersion.Value)
	}

	if c.ResponseCondition.Valid {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.Placement.Valid {
		input.Placement = fastly.String(c.Placement.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(in io.Reader, out io.Writer) error {
	input, err := c.createInput()
	if err != nil {
		return err
	}

	kafka, err := c.Globals.Client.UpdateKafka(input)
	if err != nil {
		return err
	}

	text.Success(out, "Updated Kafka logging endpoint %s (service %s version %d)", kafka.Name, kafka.ServiceID, kafka.Version)
	return nil
}
