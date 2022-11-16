package s3

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

// CreateCommand calls the Fastly API to create an Amazon S3 logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	EndpointName   string // Can't shadow cmd.Base method Name().
	BucketName     string
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// mutual exclusions
	// AccessKey + SecretKey or IAMRole must be provided
	AccessKey cmd.OptionalString
	SecretKey cmd.OptionalString
	IAMRole   cmd.OptionalString

	// optional
	AutoClone                    cmd.OptionalAutoClone
	Domain                       cmd.OptionalString
	Path                         cmd.OptionalString
	Period                       cmd.OptionalInt
	GzipLevel                    cmd.OptionalInt
	Format                       cmd.OptionalString
	FormatVersion                cmd.OptionalInt
	MessageType                  cmd.OptionalString
	ResponseCondition            cmd.OptionalString
	TimestampFormat              cmd.OptionalString
	Placement                    cmd.OptionalString
	Redundancy                   cmd.OptionalString
	PublicKey                    cmd.OptionalString
	ServerSideEncryption         cmd.OptionalString
	ServerSideEncryptionKMSKeyID cmd.OptionalString
	CompressionCodec             cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("create", "Create an Amazon S3 logging endpoint on a Fastly service version").Alias("add")
	c.CmdClause.Flag("name", "The name of the S3 logging object. Used as a primary key for API access").Short('n').Required().StringVar(&c.EndpointName)
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
	c.CmdClause.Flag("bucket", "Your S3 bucket name").Required().StringVar(&c.BucketName)
	c.CmdClause.Flag("access-key", "Your S3 account access key").Action(c.AccessKey.Set).StringVar(&c.AccessKey.Value)
	c.CmdClause.Flag("secret-key", "Your S3 account secret key").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("iam-role", "The IAM role ARN for logging").Action(c.IAMRole.Set).StringVar(&c.IAMRole.Value)
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
	c.CmdClause.Flag("domain", "The domain of the S3 endpoint").Action(c.Domain.Set).StringVar(&c.Domain.Value)
	common.Path(c.CmdClause, &c.Path)
	common.Period(c.CmdClause, &c.Period)
	common.GzipLevel(c.CmdClause, &c.GzipLevel)
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.MessageType(c.CmdClause, &c.MessageType)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	common.TimestampFormat(c.CmdClause, &c.TimestampFormat)
	c.CmdClause.Flag("redundancy", "The S3 storage class. One of: standard, intelligent_tiering, standard_ia, onezone_ia, glacier, glacier_ir, deep_archive, or reduced_redundancy").Action(c.Redundancy.Set).EnumVar(&c.Redundancy.Value, string(fastly.S3RedundancyStandard), string(fastly.S3RedundancyIntelligentTiering), string(fastly.S3RedundancyStandardIA), string(fastly.S3RedundancyOneZoneIA), string(fastly.S3RedundancyGlacierFlexibleRetrieval), string(fastly.S3RedundancyGlacierInstantRetrieval), string(fastly.S3RedundancyGlacierDeepArchive), string(fastly.S3RedundancyReduced))
	common.Placement(c.CmdClause, &c.Placement)
	common.PublicKey(c.CmdClause, &c.PublicKey)
	c.CmdClause.Flag("server-side-encryption", "Set to enable S3 Server Side Encryption. Can be either AES256 or aws:kms").Action(c.ServerSideEncryption.Set).EnumVar(&c.ServerSideEncryption.Value, string(fastly.S3ServerSideEncryptionAES), string(fastly.S3ServerSideEncryptionKMS))
	c.CmdClause.Flag("server-side-encryption-kms-key-id", "Server-side KMS Key ID. Must be set if server-side-encryption is set to aws:kms").Action(c.ServerSideEncryptionKMSKeyID.Set).StringVar(&c.ServerSideEncryptionKMSKeyID.Value)
	common.CompressionCodec(c.CmdClause, &c.CompressionCodec)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateS3Input, error) {
	var input fastly.CreateS3Input

	input.ServiceID = serviceID
	input.ServiceVersion = serviceVersion
	input.Name = &c.EndpointName
	input.BucketName = &c.BucketName

	// The following block checks for invalid permutations of the ways in
	// which the AccessKey + SecretKey and IAMRole flags can be
	// provided. This is necessary because either the AccessKey and
	// SecretKey or the IAMRole is required, but they are mutually
	// exclusive. The kingpin library lacks a way to express this constraint
	// via the flag specification API so we enforce it manually here.
	if !c.AccessKey.WasSet && !c.SecretKey.WasSet && !c.IAMRole.WasSet {
		return nil, fmt.Errorf("error parsing arguments: the --access-key and --secret-key flags or the --iam-role flag must be provided")
	} else if (c.AccessKey.WasSet || c.SecretKey.WasSet) && c.IAMRole.WasSet {
		// Enforce mutual exclusion
		return nil, fmt.Errorf("error parsing arguments: the --access-key and --secret-key flags are mutually exclusive with the --iam-role flag")
	} else if c.AccessKey.WasSet && !c.SecretKey.WasSet {
		return nil, fmt.Errorf("error parsing arguments: required flag --secret-key not provided")
	} else if !c.AccessKey.WasSet && c.SecretKey.WasSet {
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
	return
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

	d, err := c.Globals.APIClient.CreateS3(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created S3 logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
