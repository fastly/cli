package s3

import (
	"io"

	"github.com/fastly/go-fastly/v10/fastly"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe an Amazon S3 logging endpoint.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.GetS3Input
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a S3 logging endpoint on a Fastly service version").Alias("get")

	// Required.
	c.CmdClause.Flag("name", "The name of the S3 logging object").Short('n').Required().StringVar(&c.Input.Name)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		APIClient:          c.Globals.APIClient,
		Manifest:           *c.Globals.Manifest,
		Out:                out,
		ServiceNameFlag:    c.serviceName,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flags.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fsterr.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = fastly.ToValue(serviceVersion.Number)

	o, err := c.Globals.APIClient.GetS3(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	lines := text.Lines{
		"Bucket":                            fastly.ToValue(o.BucketName),
		"Compression codec":                 fastly.ToValue(o.CompressionCodec),
		"File max bytes":                    fastly.ToValue(o.FileMaxBytes),
		"Format version":                    fastly.ToValue(o.FormatVersion),
		"Format":                            fastly.ToValue(o.Format),
		"GZip level":                        fastly.ToValue(o.GzipLevel),
		"Message type":                      fastly.ToValue(o.MessageType),
		"Name":                              fastly.ToValue(o.Name),
		"Path":                              fastly.ToValue(o.Path),
		"Period":                            fastly.ToValue(o.Period),
		"Placement":                         fastly.ToValue(o.Placement),
		"Processing region":                 fastly.ToValue(o.ProcessingRegion),
		"Public key":                        fastly.ToValue(o.PublicKey),
		"Redundancy":                        fastly.ToValue(o.Redundancy),
		"Response condition":                fastly.ToValue(o.ResponseCondition),
		"Server-side encryption KMS key ID": fastly.ToValue(o.ServerSideEncryption),
		"Server-side encryption":            fastly.ToValue(o.ServerSideEncryption),
		"Timestamp format":                  fastly.ToValue(o.TimestampFormat),
		"Version":                           fastly.ToValue(o.ServiceVersion),
	}
	if o.AccessKey != nil || o.SecretKey != nil {
		lines["Access key"] = fastly.ToValue(o.AccessKey)
		lines["Secret key"] = fastly.ToValue(o.SecretKey)
	}
	if o.IAMRole != nil {
		lines["IAM role"] = fastly.ToValue(o.IAMRole)
	}
	if !c.Globals.Verbose() {
		lines["Service ID"] = fastly.ToValue(o.ServiceID)
	}
	text.PrintLines(out, lines)

	return nil
}
