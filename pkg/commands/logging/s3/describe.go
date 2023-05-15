package s3

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v8/fastly"
)

// DescribeCommand calls the Fastly API to describe an Amazon S3 logging endpoint.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	manifest       manifest.Data
	Input          fastly.GetS3Input
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
		manifest: m,
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a S3 logging endpoint on a Fastly service version").Alias("get")

	// Required.
	c.CmdClause.Flag("name", "The name of the S3 logging object").Short('n').Required().StringVar(&c.Input.Name)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		APIClient:          c.Globals.APIClient,
		Manifest:           c.manifest,
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
	c.Input.ServiceVersion = serviceVersion.Number

	o, err := c.Globals.APIClient.GetS3(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	lines := text.Lines{
		"Bucket":                            o.BucketName,
		"Compression codec":                 o.CompressionCodec,
		"Format version":                    o.FormatVersion,
		"Format":                            o.Format,
		"GZip level":                        o.GzipLevel,
		"Message type":                      o.MessageType,
		"Name":                              o.Name,
		"Path":                              o.Path,
		"Period":                            o.Period,
		"Placement":                         o.Placement,
		"Public key":                        o.PublicKey,
		"Redundancy":                        o.Redundancy,
		"Response condition":                o.ResponseCondition,
		"Server-side encryption KMS key ID": o.ServerSideEncryption,
		"Server-side encryption":            o.ServerSideEncryption,
		"Timestamp format":                  o.TimestampFormat,
		"Version":                           o.ServiceVersion,
	}
	if o.AccessKey != "" || o.SecretKey != "" {
		lines["Access key"] = o.AccessKey
		lines["Secret key"] = o.SecretKey
	}
	if o.IAMRole != "" {
		lines["IAM role"] = o.IAMRole
	}
	if !c.Globals.Verbose() {
		lines["Service ID"] = o.ServiceID
	}
	text.PrintLines(out, lines)

	return nil
}
