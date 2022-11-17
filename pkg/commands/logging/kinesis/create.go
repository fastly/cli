package kinesis

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

// CreateCommand calls the Fastly API to create an Amazon Kinesis logging endpoint.
type CreateCommand struct {
	cmd.Base
	Manifest manifest.Data

	// required
	ServiceName    cmd.OptionalServiceNameID
	ServiceVersion cmd.OptionalServiceVersion

	// mutual exclusions
	// AccessKey + SecretKey or IAMRole must be provided
	AccessKey cmd.OptionalString
	SecretKey cmd.OptionalString
	IAMRole   cmd.OptionalString

	// optional
	AutoClone         cmd.OptionalAutoClone
	EndpointName      cmd.OptionalString // Can't shadow cmd.Base method Name().
	Format            cmd.OptionalString
	FormatVersion     cmd.OptionalInt
	Placement         cmd.OptionalString
	Region            cmd.OptionalString
	ResponseCondition cmd.OptionalString
	StreamName        cmd.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *CreateCommand {
	var c CreateCommand
	c.Globals = globals
	c.Manifest = data
	c.CmdClause = parent.Command("create", "Create an Amazon Kinesis logging endpoint on a Fastly service version").Alias("add")

	// required
	c.CmdClause.Flag("name", "The name of the Kinesis logging object. Used as a primary key for API access").Short('n').Action(c.EndpointName.Set).StringVar(&c.EndpointName.Value)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})
	c.CmdClause.Flag("stream-name", "The Amazon Kinesis stream to send logs to").Action(c.StreamName.Set).StringVar(&c.StreamName.Value)
	c.CmdClause.Flag("region", "The AWS region where the Kinesis stream exists").Action(c.Region.Set).StringVar(&c.Region.Value)

	// required, but mutually exclusive
	c.CmdClause.Flag("access-key", "The access key associated with the target Amazon Kinesis stream").Action(c.AccessKey.Set).StringVar(&c.AccessKey.Value)
	c.CmdClause.Flag("secret-key", "The secret key associated with the target Amazon Kinesis stream").Action(c.SecretKey.Set).StringVar(&c.SecretKey.Value)
	c.CmdClause.Flag("iam-role", "The IAM role ARN for logging").Action(c.IAMRole.Set).StringVar(&c.IAMRole.Value)

	// optional
	c.RegisterAutoCloneFlag(cmd.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
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
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
	common.Placement(c.CmdClause, &c.Placement)

	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateKinesisInput, error) {
	var input fastly.CreateKinesisInput

	input.ServiceID = serviceID
	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
	}
	if c.StreamName.WasSet {
		input.StreamName = &c.StreamName.Value
	}
	if c.Region.WasSet {
		input.Region = &c.Region.Value
	}
	input.ServiceVersion = serviceVersion

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

	if c.AccessKey.WasSet {
		input.AccessKey = &c.AccessKey.Value
	}

	if c.SecretKey.WasSet {
		input.SecretKey = &c.SecretKey.Value
	}

	if c.IAMRole.WasSet {
		input.IAMRole = &c.IAMRole.Value
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

	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
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

	d, err := c.Globals.APIClient.CreateKinesis(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out, "Created Kinesis logging endpoint %s (service %s version %d)", d.Name, d.ServiceID, d.ServiceVersion)
	return nil
}
