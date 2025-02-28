package grafanacloudlogs

import (
	"io"

	"4d63.com/optional"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/logging/common"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v9/fastly"
)

// CreateCommand calls the Fastly API to create a GrafanaCloudLogs logging endpoint.
type CreateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion
	User           argparser.OptionalString
	URL            argparser.OptionalString
	Index          argparser.OptionalString
	Token          argparser.OptionalString

	// Optional.
	AutoClone         argparser.OptionalAutoClone
	EndpointName      argparser.OptionalString // Can't shadow argparser.Base method Name().
	Format            argparser.OptionalString
	FormatVersion     argparser.OptionalInt
	MessageType       argparser.OptionalString
	ResponseCondition argparser.OptionalString
	TimestampFormat   argparser.OptionalString
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a Grafana Cloud Logs logging endpoint on a Fastly service version").Alias("add")

	// Required.
	c.CmdClause.Flag("name", "The name of the Grafana Cloud Logs logging endpoint. Used as a primary key for API access").Short('n').Action(c.EndpointName.Set).StringVar(&c.EndpointName.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.ServiceVersion.Value,
		Required:    true,
	})
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

	// Optional.
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.AutoClone.Set,
		Dst:    &c.AutoClone.Value,
	})
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.MessageType(c.CmdClause, &c.MessageType)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)

	c.CmdClause.Flag("index", `The stream identifier`).Action(c.Index.Set).StringVar(&c.Index.Value)
	c.CmdClause.Flag("url", "The URL of your Grafana instance").Action(c.URL.Set).StringVar(&c.URL.Value)
	c.CmdClause.Flag("user", "Your Grafana User ID.").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("auth-token", "Your Grafana Access Policy Token").Action(c.Token.Set).StringVar(&c.Token.Value)

	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *CreateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.CreateGrafanaCloudLogsInput, error) {
	input := fastly.CreateGrafanaCloudLogsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
	}

	if c.EndpointName.WasSet {
		input.Name = &c.EndpointName.Value
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
	if c.Index.WasSet {
		input.Index = &c.Index.Value
	}
	if c.URL.WasSet {
		input.URL = &c.URL.Value
	}
	if c.User.WasSet {
		input.User = &c.User.Value
	}
	if c.Token.WasSet {
		input.Token = &c.Token.Value
	}
	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
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

	d, err := c.Globals.APIClient.CreateGrafanaCloudLogs(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Created Grafana Cloud Logs logging endpoint %s (service %s version %d)",
		fastly.ToValue(d.Name),
		fastly.ToValue(d.ServiceID),
		fastly.ToValue(d.ServiceVersion),
	)
	return nil
}
