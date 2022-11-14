package kafka

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
)

// CreateCommand calls the Fastly API to create a Kafka logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	Topic          string
	Brokers        string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// optional
	AutoClone         cmd.OptionalAutoClone
	UseTLS            cmd.OptionalBool
	CompressionCodec  cmd.OptionalString
	RequiredACKs      cmd.OptionalString
	TLSCACert         cmd.OptionalString
	TLSClientCert     cmd.OptionalString
	TLSClientKey      cmd.OptionalString
	TLSHostname       cmd.OptionalString
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	Placement         cmd.OptionalString
	ResponseCondition cmd.OptionalString
	ParseLogKeyvals   cmd.OptionalBool
	RequestMaxBytes   cmd.OptionalInt
	UseSASL           cmd.OptionalBool
	AuthMethod        cmd.OptionalString
	User              cmd.OptionalString
	Password          cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("create", "Create a Kafka logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the Kafka logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("topic", "The Kafka topic to send logs to").Required().StringVar(&c.Topic)
	c.CmdClause.Flag("brokers", "A comma-separated list of IP addresses or hostnames of Kafka brokers").Required().StringVar(&c.Brokers)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.ServiceName.Value,
	})
	c.CmdClause.Flag("compression-codec", "The codec used for compression of your logs. One of: gzip, snappy, lz4").Action(c.CompressionCodec.Set).StringVar(fastly.String(c.CompressionCodec.Value))
	c.CmdClause.Flag("required-acks", "The Number of acknowledgements a leader must receive before a write is considered successful. One of: 1 (default) One server needs to respond. 0	No servers need to respond. -1	Wait for all in-sync replicas to respond").Action(c.RequiredACKs.Set).StringVar(&c.RequiredACKs.Value)
	c.CmdClause.Flag("use-tls", "Whether to use TLS for secure logging. Can be either true or false").Action(c.UseTLS.Set).BoolVar(&c.UseTLS.Value)
	common.TLSCACert(c.CmdClause, &c.TLSCACert)
	common.TLSClientCert(c.CmdClause, &c.TLSClientCert)
	common.TLSClientKey(c.CmdClause, &c.TLSClientKey)
	common.TLSHostname(c.CmdClause, &c.TLSHostname)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.Placement(c.CmdClause, &c.Placement)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.CmdClause.Flag("parse-log-keyvals", "Parse key-value pairs within the log format").Action(c.ParseLogKeyvals.Set).BoolVar(&c.ParseLogKeyvals.Value)
	c.CmdClause.Flag("max-batch-size", "The maximum size of the log batch in bytes").Action(c.RequestMaxBytes.Set).IntVar(&c.RequestMaxBytes.Value)
	c.CmdClause.Flag("use-sasl", "Enable SASL authentication. Requires --auth-method, --username, and --password to be specified").Action(c.UseSASL.Set).BoolVar(&c.UseSASL.Value)
	c.CmdClause.Flag("auth-method", "SASL authentication method. Valid values are: plain, scram-sha-256, scram-sha-512").Action(c.AuthMethod.Set).HintOptions("plain", "scram-sha-256", "scram-sha-512").EnumVar(&c.AuthMethod.Value, "plain", "scram-sha-256", "scram-sha-512")
	c.CmdClause.Flag("username", "SASL authentication username. Required if --auth-method is specified").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("password", "SASL authentication password. Required if --auth-method is specified").Action(c.Password.Set).StringVar(&c.Password.Value)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateKafkaInput, error) {
	var input fastly.CreateKafkaInput

	if c.UseSASL.WasSet && c.UseSASL.Value && (c.AuthMethod.Value == "" || c.User.Value == "" || c.Password.Value == "") {
		return nil, fmt.Errorf("the --auth-method, --username, and --password flags must be present when using the --use-sasl flag")
	}

	if !c.UseSASL.Value && (c.AuthMethod.Value != "" || c.User.Value != "" || c.Password.Value != "") {
		return nil, fmt.Errorf("the --auth-method, --username, and --password options are only valid when the --use-sasl flag is specified")
	}

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Name = fastly.String(c.EndpointName)
	input.Topic = fastly.String(c.Topic)
	input.Brokers = fastly.String(c.Brokers)

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = fastly.String(c.CompressionCodec.Value)
	}

	if c.RequiredACKs.WasSet {
		input.RequiredACKs = fastly.String(c.RequiredACKs.Value)
	}

	if c.UseTLS.WasSet {
		input.UseTLS = fastly.CBool(c.UseTLS.Value)
	}

	if c.TLSCACert.WasSet {
		input.TLSCACert = fastly.String(c.TLSCACert.Value)
	}

	if c.TLSClientCert.WasSet {
		input.TLSClientCert = fastly.String(c.TLSClientCert.Value)
	}

	if c.TLSClientKey.WasSet {
		input.TLSClientKey = fastly.String(c.TLSClientKey.Value)
	}

	if c.TLSHostname.WasSet {
		input.TLSHostname = fastly.String(c.TLSHostname.Value)
	}

	if c.Format.WasSet {
		input.Format = fastly.String(c.Format.Value)
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = fastly.Int(c.FormatVersion.Value)
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = fastly.String(c.ResponseCondition.Value)
	}

	if c.Placement.WasSet {
		input.Placement = fastly.String(c.Placement.Value)
	}

	if c.ParseLogKeyvals.WasSet {
		input.ParseLogKeyvals = fastly.CBool(c.ParseLogKeyvals.Value)
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = fastly.Int(c.RequestMaxBytes.Value)
	}

	if c.AuthMethod.WasSet {
		input.AuthMethod = fastly.String(c.AuthMethod.Value)
	}

	if c.User.WasSet {
		input.User = fastly.String(c.User.Value)
	}

	if c.Password.WasSet {
		input.Password = fastly.String(c.Password.Value)
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, serviceVersion.Number)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	d, err := c.Globals.APIClient.CreateKafka(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created Kafka logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
