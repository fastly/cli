package s3

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

// CreateCommand calls the Fastly API to create an Amazon S3 logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion

	// mutual exclusions
	// AccessKey + SecretKey or IAMRole must be provided
	AccessKey argparser.OptionalString
	SecretKey argparser.OptionalString
	IAMRole   argparser.OptionalString

	// Optional.
	AutoClone                    argparser.OptionalAutoClone
	BucketName                   argparser.OptionalString
	CompressionCodec             argparser.OptionalString
	Domain                       argparser.OptionalString
	EndpointName                 argparser.OptionalString // Can't shadow argparser.Base method Name().
	FileMaxBytes                 argparser.OptionalInt
	Format                       argparser.OptionalString
	FormatVersion                argparser.OptionalInt
	GzipLevel                    argparser.OptionalInt
	MessageType                  argparser.OptionalString
	Path                         argparser.OptionalString
	Period                       argparser.OptionalInt
	Placement                    argparser.OptionalString
	PublicKey                    argparser.OptionalString
	Redundancy                   argparser.OptionalString
	ResponseCondition            argparser.OptionalString
	ServerSideEncryption         argparser.OptionalString
	ServerSideEncryptionKMSKeyID argparser.OptionalString
	TimestampFormat              argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create an Amazon S3 logging endpoint on a Fastly service version").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The name of the S3 logging object. Used as a primary key for API access").Short('n').Action(c.EndpointName.Set).StringVar(&c.EndpointName.Value)
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
	c.CmdClause.Flag("access-key", "Your S3 account access key").Action(c.AccessKey.Set).StringVar(&c.AccessKey.Value)
	c.CmdClause.Flag("bucket", "Your S3 bucket name").Action(c.BucketName.Set).StringVar(&c.BucketName.Value)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	c.CmdClause.Flag("domain", "The domain of the S3 endpoint").Action(c.Domain.Set).StringVar(&c.Domain.Value)
	c.CmdClause.Flag("file-max-bytes", "The maximum size of a log file in bytes").Action(c.FileMaxBytes.Set).IntVar(&c.FileMaxBytes.Value)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	c.CmdClause.Flag("iam-role", "The IAM role ARN for logging").Action(c.IAMRole.Set).StringVar(&c.IAMRole.Value)
	common.MessageType(c.CmdClause, &c.MessageType)
	common.Path(c.CmdClause, &c.Path)
	common.Period(c.CmdClause, &c.Period)
	common.Placement(c.CmdClause, &c.Placement)
	common.PublicKey(c.CmdClause, &c.PublicKey)
	c.CmdClause.Flag("redundancy", "The S3 storage class. One of: standard, intelligent_tiering, standard_ia, onezone_ia, glacier, glacier_ir, deep_archive, or reduced_redundancy").Action(c.Redundancy.Set).EnumVar(&c.Redundancy.Value, string(fastly.S3RedundancyStandard), string(fastly.S3RedundancyIntelligentTiering), string(fastly.S3RedundancyStandardIA), string(fastly.S3RedundancyOneZoneIA), string(fastly.S3RedundancyGlacierFlexibleRetrieval), string(fastly.S3RedundancyGlacierInstantRetrieval), string(fastly.S3RedundancyGlacierDeepArchive), string(fastly.S3RedundancyReduced))
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	c.CmdClause.Flag("secret-key", "Your S3 account secret key").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("server-side-encryption", "Set to enable S3 Server Side Encryption. Can be either AES256 or aws:kms").Action(c.ServerSideEncryption.Set).EnumVar(&c.ServerSideEncryption.Value, string(fastly.S3ServerSideEncryptionAES), string(fastly.S3ServerSideEncryptionKMS))
	c.CmdClause.Flag("server-side-encryption-kms-key-id", "Server-side KMS Key ID. Must be set if server-side-encryption is set to aws:kms").Action(c.ServerSideEncryptionKMSKeyID.Set).StringVar(&c.ServerSideEncryptionKMSKeyID.Value)
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
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateS3Input, error) {
	var input fastly.CreateS3Input

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.BucketName.WasSet {
		input.BucketName = &c.BucketName.Value
	}

	// The following block checks for invalid permutations of the ways in
	// which the AccessKey + SecretKey and IAMRole flags can be
	// provided. This is necessary because either the AccessKey and
	// SecretKey or the IAMRole is required, but they are mutually
	// exclusive. The kingpin library lacks a way to express this constraint
	// via the flag specification API so we enforce it manually here.
	switch {
	case !c.AccessKey.WasSet && !c.SecretKey.WasSet && !c.IAMRole.WasSet:
		return nil, fmt.Errorf("error parsing arguments: the --access-key and --secret-key flags or the --iam-role flag must be provided")
	case (c.AccessKey.WasSet || c.SecretKey.WasSet) && c.IAMRole.WasSet:
		// Enforce mutual exclusion
		return nil, fmt.Errorf("error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag")
	case c.AccessKey.WasSet && !c.SecretKey.WasSet:
		return nil, fmt.Errorf("error parsing arguments: required flag --secret-key not provided")
	case !c.AccessKey.WasSet && c.SecretKey.WasSet:
		return nil, fmt.Errorf("error parsing arguments: required flag --access-key not provided")
	}

	// The following blocks enforces the mutual exclusivity of the
	// CompressionCodec and GzipLevel flags.
	if c.CompressionCodec.WasSet && c.GzipLevel.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --compression-codec flag is mutually exclusive with the --gzip-level flag")
	}

	if c.AccessKey.WasSet {
		input.AccessKey = &c.AccessKey.Value
	}

	if c.SecretKey.WasSet {
		input.SecretKey = &c.SecretKey.Value
	}

	if c.IAMRole.WasSet {
		input.IAMRole = &c.IAMRole.Value
	}

	if c.Domain.WasSet {
		input.Domain = &c.Domain.Value
	}

	if c.FileMaxBytes.WasSet {
		input.FileMaxBytes = &c.FileMaxBytes.Value
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
		input.Format = &c.Format.Value
	}

	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}

	if c.MessageType.WasSet {
		input.MessageType = &c.MessageType.Value
	}

	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}

	if c.TimestampFormat.WasSet {
		input.TimestampFormat = &c.TimestampFormat.Value
	}

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}

	if c.PublicKey.WasSet {
		input.PublicKey = &c.PublicKey.Value
	}

	if c.ServerSideEncryptionKMSKeyID.WasSet {
		input.ServerSideEncryptionKMSKeyID = &c.ServerSideEncryptionKMSKeyID.Value
	}

	if c.CompressionCodec.WasSet {
		input.CompressionCodec = &c.CompressionCodec.Value
	}

	if c.Redundancy.WasSet {
		redundancy, err := ValidateRedundancy(c.Redundancy.Value)
		if err == nil {
			input.Redundancy = &redundancy
		}
	}

	if c.ServerSideEncryption.WasSet {
		switch c.ServerSideEncryption.Value {
		case string(fastly.S3ServerSideEncryptionAES):
			sse := fastly.S3ServerSideEncryptionAES
			input.ServerSideEncryption = &sse
		case string(fastly.S3ServerSideEncryptionKMS):
			sse := fastly.S3ServerSideEncryptionKMS
			input.ServerSideEncryption = &sse
		}
	}

	return &input, nil
}

// ValidateRedundancy identifies the given redundancy type.
func ValidateRedundancy(val string) (redundancy fastly.S3Redundancy, err error) {
	switch val {
	case string(fastly.S3RedundancyStandard):
		redundancy = fastly.S3RedundancyStandard
	case string(fastly.S3RedundancyIntelligentTiering):
		redundancy = fastly.S3RedundancyIntelligentTiering
	case string(fastly.S3RedundancyStandardIA):
		redundancy = fastly.S3RedundancyStandardIA
	case string(fastly.S3RedundancyOneZoneIA):
		redundancy = fastly.S3RedundancyOneZoneIA
	case string(fastly.S3RedundancyGlacierInstantRetrieval):
		redundancy = fastly.S3RedundancyGlacierInstantRetrieval
	case string(fastly.S3RedundancyGlacierFlexibleRetrieval):
		redundancy = fastly.S3RedundancyGlacierFlexibleRetrieval
	case string(fastly.S3RedundancyGlacierDeepArchive):
		redundancy = fastly.S3RedundancyGlacierDeepArchive
	case string(fastly.S3RedundancyReduced):
		redundancy = fastly.S3RedundancyReduced
	default:
		err = fmt.Errorf("unknown redundancy: " + val)
	}
	return redundancy, err
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

	d, err := c.Globals.APIClient.CreateS3(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Created S3 logging endpoint %s (service %s version %d)",
		fastly.ToValue(d.Name),
		fastly.ToValue(d.ServiceID),
		fastly.ToValue(d.ServiceVersion),
	)
	return nil
}
