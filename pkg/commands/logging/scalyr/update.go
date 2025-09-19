package scalyr

import (
	"context"
	"io"

	"github.com/fastly/go-fastly/v12/fastly"

	"4d63.com/optional"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/logflags"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// UpdateCommand calls the Fastly API to update Scalyr logging endpoints.
type UpdateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	EndpointName   string // Can't shadow argparser.Base method Name().
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion

	// Optional.
	AutoClone         argparser.OptionalAutoClone
	Format            argparser.OptionalString
	FormatVersion     argparser.OptionalInt
	NewName           argparser.OptionalString
	Placement         argparser.OptionalString
	ProcessingRegion  argparser.OptionalString
	ProjectID         argparser.OptionalString
	Region            argparser.OptionalString
	ResponseCondition argparser.OptionalString
	Token             argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Scalyr logging endpoint on a Fastly service version")

	// Required.
	c.CmdClause.Flag("name", "The name of the Scalyr logging object").Short('n').Required().StringVar(&c.EndpointName)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})

	// Optional.
	c.CmdClause.Flag("auth-token", "The token to use for authentication (https://www.scalyr.com/keys)").Action(c.Token.Set).StringVar(&c.Token.Value)
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	logflags.Format(c.CmdClause, &c.Format)
	logflags.FormatVersion(c.CmdClause, &c.FormatVersion)
	c.CmdClause.Flag("new-name", "New name of the Scalyr logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	logflags.Placement(c.CmdClause, &c.Placement)
	logflags.ProcessingRegion(c.CmdClause, &c.ProcessingRegion, "Scalyr")
	c.CmdClause.Flag("project-id", "The name of the logfile field sent to Scalyr").Action(c.ProjectID.Set).StringVar(&c.ProjectID.Value)
	c.CmdClause.Flag("region", "The region where logs are received and stored by Scalyr. Either US or EU. Defaults to US if undefined").Action(c.Region.Set).StringVar(&c.Region.Value)
	logflags.ResponseCondition(c.CmdClause, &c.ResponseCondition)
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
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateScalyrInput, error) {
	input := fastly.UpdateScalyrInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
	}
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}
	if c.Format.WasSet {
		input.Format = fastly.ToPointer(argparser.Content(c.Format.Value))
	}
	if c.FormatVersion.WasSet {
		input.FormatVersion = &c.FormatVersion.Value
	}
	if c.Token.WasSet {
		input.Token = &c.Token.Value
	}
	if c.Region.WasSet {
		input.Region = &c.Region.Value
	}
	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}
	if c.Placement.WasSet {
		input.Placement = &c.Placement.Value
	}
	if c.ProcessingRegion.WasSet {
		input.ProcessingRegion = &c.ProcessingRegion.Value
	}
	if c.ProjectID.WasSet {
		input.ProjectID = &c.ProjectID.Value
	}
	return &input, nil
}

// Exec invokes the application logic for the command.
func (c *UpdateCommand) Exec(_ io.Reader, out io.Writer) error {
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

	scalyr, err := c.Globals.APIClient.UpdateScalyr(context.TODO(), input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Updated Scalyr logging endpoint %s (service %s version %d)",
		fastly.ToValue(scalyr.Name),
		fastly.ToValue(scalyr.ServiceID),
		fastly.ToValue(scalyr.ServiceVersion),
	)
	return nil
}
