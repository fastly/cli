package azureblob

import (
	"context"
	"fmt"
	"io"

	"github.com/fastly/go-fastly/v11/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/logflags"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create an Azure Blob Storage logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion

	// Optional.
	EndpointName      argparser.OptionalString
	Container         argparser.OptionalString
	AccountName       argparser.OptionalString
	SASToken          argparser.OptionalString
	AutoClone         argparser.OptionalAutoClone
	Path              argparser.OptionalString
	Period            argparser.OptionalInt
	GzipLevel         argparser.OptionalInt
	MessageType       argparser.OptionalString
	Format            argparser.OptionalString
	FormatVersion     argparser.OptionalInt
	ResponseCondition argparser.OptionalString
	TimestampFormat   argparser.OptionalString
	Placement         argparser.OptionalString
	ProcessingRegion  argparser.OptionalString
	PublicKey         argparser.OptionalString
	FileMaxBytes      argparser.OptionalInt
	CompressionCodec  argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create an Azure Blob Storage logging endpoint on a Fastly service version").Alias("add")

	// Required.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.CmdClause.Flag("account-name", "The unique Azure Blob Storage namespace in which your data objects are stored").Action(c.AccountName.Set).StringVar(&c.AccountName.Value)
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	logflags.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	c.CmdClause.Flag("container", "The name of the Azure Blob Storage container in which to store logs").Action(c.Container.Set).StringVar(&c.Container.Value)
	c.CmdClause.Flag("file-max-bytes", "The maximum size of a log file in bytes").Action(c.FileMaxBytes.Set).IntVar(&c.FileMaxBytes.Value)
	logflags.Format(c.CmdClause, &c.Format)
	logflags.FormatVersion(c.CmdClause, &c.FormatVersion)
	logflags.GzipLevel(c.CmdClause, &c.GzipLevel)
	logflags.MessageType(c.CmdClause, &c.MessageType)
	c.CmdClause.Flag("name", "The name of the Azure Blob Storage logging object. Used as a primary key for API access").Short('n').Action(c.EndpointName.Set).StringVar(&c.EndpointName.Value)
	logflags.Path(c.CmdClause, &c.Path)
	logflags.Period(c.CmdClause, &c.Period)
	logflags.Placement(c.CmdClause, &c.Placement)
	logflags.ProcessingRegion(c.CmdClause, &c.ProcessingRegion, "Azure Blob Storage")
	logflags.PublicKey(c.CmdClause, &c.PublicKey)
	logflags.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.CmdClause.Flag("sas-token", "The Azure shared access signature providing write access to the blob service objects. Be sure to update your token before it expires or the logging functionality will not work").Action(c.SASToken.Set).StringVar(&c.SASToken.Value)
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
	logflags.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateBlobStorageInput, error) {
	var input fastly.CreateBlobStorageInput

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion

	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.Container.WasSet {
		input.Container = &c.Container.Value
	}
	if c.AccountName.WasSet {
		input.AccountName = &c.AccountName.Value
	}
	if c.SASToken.WasSet {
		input.SASToken = &c.SASToken.Value
	}

	// The following blocks enforces the mutual exclusivity of the
	// CompressionCodec and GzipLevel flags.
	if c.CompressionCodec.WasSet && c.GzipLevel.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag")
	}

	if c.Path.WasSet {
		input.Path = &c.Path.Value
	}
	if c.Period.WasSet {
		input.Period = &c.Period.Value
	}
	if c.GzipLevel.WasSet {
		input.GzipLevel = &c.GzipLevel.Value
	}
	if c.Format.WasSet {
		input.Format = fastly.ToPointer(argparser.Content(c.Format.Value))
	}
	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}
	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}
	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
	}
	if c.TimestampFormat.WasSet {
		input.TimestampFormat = &c.TimestampFormat.Value
	}
	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}
	if c.ProcessingRegion.WasSet {
		input.ProcessingRegion = &c.ProcessingRegion.Value
	}
	if c.PublicKey.WasSet {
		input.PublicKey = &c.PublicKey.Value
	}
	if c.FileMaxBytes.WasSet {
		input.FileMaxBytes = &c.FileMaxBytes.Value
	}
	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
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
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	d, err := c.Globals.APIClient.CreateBlobStorage(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out,
		"Created Azure Blob Storage logging endpoint %s (service %s version %d)",
		fastly.ToValue(d.Name),
		fastly.ToValue(d.ServiceID),
		fastly.ToValue(d.ServiceVersion),
	)
	return nil
}
