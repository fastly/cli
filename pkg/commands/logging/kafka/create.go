package kafka

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"4d63.com/optional"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create a Kafka logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion

	// Optional.
	AuthMethod        argparser.OptionalString
	AutoClone         argparser.OptionalAutoClone
	Brokers           argparser.OptionalString
	CompressionCodec  argparser.OptionalString
	EndpointName      argparser.OptionalString // Can't shadow argparser.Base method Name().
	Format            argparser.OptionalString
	FormatVersion     argparser.OptionalInt
	ParseLogKeyvals   argparser.OptionalBool
	Password          argparser.OptionalString
	RequestMaxBytes   argparser.OptionalInt
	RequiredACKs      argparser.OptionalString
	ResponseCondition argparser.OptionalString
	TLSCACert         argparser.OptionalString
	TLSClientCert     argparser.OptionalString
	TLSClientKey      argparser.OptionalString
	TLSHostname       argparser.OptionalString
	Topic             argparser.OptionalString
	User              argparser.OptionalString
	UseSASL           argparser.OptionalBool
	UseTLS            argparser.OptionalBool
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a Kafka logging endpoint on a Fastly service version").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The name of the Kafka logging object. Used as a primary key for API access").Short('n').Action(c.EndpointName.Set).StringVar(&c.EndpointName.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	c.CmdClause.Flag("auth-method", "SASL authentication method. Valid values are: plain, scram-sha-256, scram-sha-512").Action(c.AuthMethod.Set).HintOptions("plain", "scram-sha-256", "scram-sha-512").EnumVar(&c.AuthMethod.Value, "plain", "scram-sha-256", "scram-sha-512")
	c.CmdClause.Flag("brokers", "A comma-separated list of IP addresses or hostnames of Kafka brokers").Action(c.Brokers.Set).StringVar(&c.Brokers.Value)
	c.CmdClause.Flag("compression-codec", "The codec used for compression of your logs. One of: gzip, snappy, lz4").Action(c.CompressionCodec.Set).StringVar(&c.CompressionCodec.Value)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	c.CmdClause.Flag("max-batch-size", "The maximum size of the log batch in bytes").Action(c.RequestMaxBytes.Set).IntVar(&c.RequestMaxBytes.Value)
	c.CmdClause.Flag("parse-log-keyvals", "Parse key-value pairs within the log format").Action(c.ParseLogKeyvals.Set).BoolVar(&c.ParseLogKeyvals.Value)
	c.CmdClause.Flag("password", "SASL authentication password. Required if --auth-method is specified").Action(c.Password.Set).StringVar(&c.Password.Value)
	c.CmdClause.Flag("required-acks", "The Number of acknowledgements a leader must receive before a write is considered successful. One of: 1 (default) One server needs to respond. 0	No servers need to respond. -1	Wait for all in-sync replicas to respond").Action(c.RequiredACKs.Set).StringVar(&c.RequiredACKs.Value)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.ServiceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.ServiceName.Value,
	})
	common.TLSCACert(c.CmdClause, &c.TLSCACert)
	common.TLSClientCert(c.CmdClause, &c.TLSClientCert)
	common.TLSClientKey(c.CmdClause, &c.TLSClientKey)
	common.TLSHostname(c.CmdClause, &c.TLSHostname)
	c.CmdClause.Flag("topic", "The Kafka topic to send logs to").Action(c.Topic.Set).StringVar(&c.Topic.Value)
	c.CmdClause.Flag("use-sasl", "Enable SASL authentication. Requires --auth-method, --username, and --password to be specified").Action(c.UseSASL.Set).BoolVar(&c.UseSASL.Value)
	c.CmdClause.Flag("use-tls", "Whether to use TLS for secure logging. Can be either true or false").Action(c.UseTLS.Set).BoolVar(&c.UseTLS.Value)
	c.CmdClause.Flag("username", "SASL authentication username. Required if --auth-method is specified").Action(c.User.Set).StringVar(&c.User.Value)
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
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.Topic.WasSet {
		input.Topic = &c.Topic.Value
	}
	if c.Brokers.WasSet {
		input.Brokers = &c.Brokers.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
	}

	if c.RequiredACKs.WasSet {
		input.RequiredACKs = &c.RequiredACKs.Value
	}

	if c.UseTLS.WasSet {
		input.UseTLS = fastly.ToPointer(fastly.Compatibool(c.UseTLS.Value))
	}

	if c.TLSCACert.WasSet {
		input.TLSCACert = &c.TLSCACert.Value
	}

	if c.TLSClientCert.WasSet {
		input.TLSClientCert = &c.TLSClientCert.Value
	}

	if c.TLSClientKey.WasSet {
		input.TLSClientKey = &c.TLSClientKey.Value
	}

	if c.TLSHostname.WasSet {
		input.TLSHostname = &c.TLSHostname.Value
	}

	if c.Format.WasSet {
		input.Format = &c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}

	if c.ParseLogKeyvals.WasSet {
		input.ParseLogKeyvals = fastly.ToPointer(fastly.Compatibool(c.ParseLogKeyvals.Value))
	}

	if c.RequestMaxBytes.WasSet {
		input.RequestMaxBytes = &c.RequestMaxBytes.Value
	}

	if c.AuthMethod.WasSet {
		input.AuthMethod = &c.AuthMethod.Value
	}

	if c.User.WasSet {
		input.User = &c.User.Value
	}

	if c.Password.WasSet {
		input.Password = &c.Password.Value
	}

	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
		AutoCloneFlag:      c.AutoClone,
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.ServiceName,
		ServiceVersionFlag: c.ServiceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	input, err := c.ConstructInput(serviceID, fastly.ToValue(serviceVersion.Number))
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	d, err := c.Globals.APIClient.CreateKafka(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Created Kafka logging endpoint %s (service %s version %d)",
		fastly.ToValue(d.Name),
		fastly.ToValue(d.ServiceID),
		fastly.ToValue(d.ServiceVersion),
	)
	return nil
}
