package azureblob

import (
	"io"

	"github.com/fastly/go-fastly/v8/fastly"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// DescribeCommand calls the Fastly API to describe an Azure Blob Storage logging endpoint.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	Input          fastly.GetBlobStorageInput
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: cmd.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about an Azure Blob Storage logging endpoint on a Fastly service version").Alias("get")

	// Required.
	c.CmdClause.Flag("name", "The name of the Azure Blob Storage logging object").Short('n').Required().StringVar(&c.Input.Name)
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
		Dst:         &g.Manifest.Flag.ServiceID,
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
	c.Input.ServiceVersion = serviceVersion.Number

	o, err := c.Globals.APIClient.GetBlobStorage(&c.Input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": serviceVersion.Number,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	lines := text.Lines{
		"Account name":       o.AccountName,
		"Compression codec":  o.CompressionCodec,
		"Container":          o.Container,
		"File max bytes":     o.FileMaxBytes,
		"Format version":     o.FormatVersion,
		"Format":             o.Format,
		"GZip level":         o.GzipLevel,
		"Message type":       o.MessageType,
		"Name":               o.Name,
		"Path":               o.Path,
		"Period":             o.Period,
		"Placement":          o.Placement,
		"Public key":         o.PublicKey,
		"Response condition": o.ResponseCondition,
		"SAS token":          o.SASToken,
		"Timestamp format":   o.TimestampFormat,
		"Version":            o.ServiceVersion,
	}
	if !c.Globals.Verbose() {
		lines["Service ID"] = o.ServiceID
	}
	text.PrintLines(out, lines)

	return nil
}
