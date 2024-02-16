package s3

import (
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// ListCommand calls the Fastly API to list Amazon S3 logging endpoints.
type ListCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.ListS3sInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent argparser.Registerer, g *global.Data) *ListCommand {
	c := ListCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("list", "List S3 endpoints on a Fastly service version")

	// Required.
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
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		AllowActiveLocked:  true,
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

	o, err := c.Globals.APIClient.ListS3s(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, s3 := range o {
			tw.AddLine(
				fastly.ToValue(s3.ServiceID),
				fastly.ToValue(s3.ServiceVersion),
				fastly.ToValue(s3.Name),
			)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, s3 := range o {
		fmt.Fprintf(out, "\tS3 %d/%d\n", i+1, len(o))
		fmt.Fprintf(out, "\t\tService ID: %s\n", fastly.ToValue(s3.ServiceID))
		fmt.Fprintf(out, "\t\tVersion: %d\n", fastly.ToValue(s3.ServiceVersion))
		fmt.Fprintf(out, "\t\tName: %s\n", fastly.ToValue(s3.Name))
		fmt.Fprintf(out, "\t\tBucket: %s\n", fastly.ToValue(s3.BucketName))
		if s3.AccessKey != nil || s3.SecretKey != nil {
			fmt.Fprintf(out, "\t\tAccess key: %s\n", fastly.ToValue(s3.AccessKey))
			fmt.Fprintf(out, "\t\tSecret key: %s\n", fastly.ToValue(s3.SecretKey))
		}
		if s3.IAMRole != nil {
			fmt.Fprintf(out, "\t\tIAM role: %s\n", fastly.ToValue(s3.IAMRole))
		}
		fmt.Fprintf(out, "\t\tPath: %s\n", fastly.ToValue(s3.Path))
		fmt.Fprintf(out, "\t\tPeriod: %d\n", fastly.ToValue(s3.Period))
		fmt.Fprintf(out, "\t\tGZip level: %d\n", fastly.ToValue(s3.GzipLevel))
		fmt.Fprintf(out, "\t\tFormat: %s\n", fastly.ToValue(s3.Format))
		fmt.Fprintf(out, "\t\tFormat version: %d\n", fastly.ToValue(s3.FormatVersion))
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", fastly.ToValue(s3.ResponseCondition))
		fmt.Fprintf(out, "\t\tMessage type: %s\n", fastly.ToValue(s3.MessageType))
		fmt.Fprintf(out, "\t\tTimestamp format: %s\n", fastly.ToValue(s3.TimestampFormat))
		fmt.Fprintf(out, "\t\tPlacement: %s\n", fastly.ToValue(s3.Placement))
		fmt.Fprintf(out, "\t\tPublic key: %s\n", fastly.ToValue(s3.PublicKey))
		fmt.Fprintf(out, "\t\tRedundancy: %s\n", fastly.ToValue(s3.Redundancy))
		fmt.Fprintf(out, "\t\tServer-side encryption: %s\n", fastly.ToValue(s3.ServerSideEncryption))
		fmt.Fprintf(out, "\t\tServer-side encryption KMS key ID: %s\n", fastly.ToValue(s3.ServerSideEncryption))
		fmt.Fprintf(out, "\t\tFile max bytes: %d\n", fastly.ToValue(s3.FileMaxBytes))
		fmt.Fprintf(out, "\t\tCompression codec: %s\n", fastly.ToValue(s3.CompressionCodec))
	}
	fmt.Fprintln(out)

	return nil
}
