package grafanacloudlogs

import (
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

// UpdateCommand calls the Fastly API to update a Grafana Cloud Logs logging endpoint.
type UpdateCommand struct {
	argparser.Base
	Manifest manifest.Data

	// Required.
	EndpointName   string // Can't shadow argparser.Base method Name().
	ServiceName    argparser.OptionalServiceNameID
	ServiceVersion argparser.OptionalServiceVersion
	User           argparser.OptionalString
	URL            argparser.OptionalString
	Index          argparser.OptionalString
	Token          argparser.OptionalString

	// Optional.
	AutoClone         argparser.OptionalAutoClone
	Format            argparser.OptionalString
	FormatVersion     argparser.OptionalInt
	MessageType       argparser.OptionalString
	NewName           argparser.OptionalString
	ResponseCondition argparser.OptionalString
}

// NewUpdateCommand returns a usable command registered under the parent.
func NewUpdateCommand(parent argparser.Registerer, g *global.Data) *UpdateCommand {
	c := UpdateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("update", "Update a Grafana Cloud Logs logging endpoint on a Fastly service version")

	// Required.
	c.CmdClause.Flag("name", "The name of the Grafana Cloud Logs logging object").Short('n').Required().StringVar(&c.EndpointName)
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
	common.Format(c.CmdClause, &c.Format)
	common.FormatVersion(c.CmdClause, &c.FormatVersion)
	common.MessageType(c.CmdClause, &c.MessageType)
	c.CmdClause.Flag("new-name", "New name of the Grafana Cloud Logs logging object").Action(c.NewName.Set).StringVar(&c.NewName.Value)
	common.ResponseCondition(c.CmdClause, &c.ResponseCondition)
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
	c.CmdClause.Flag("user", "Your Grafana Cloud Logs User ID.").Action(c.User.Set).StringVar(&c.User.Value)
	c.CmdClause.Flag("auth-token", "Your Granana Access Policy Token").Action(c.Token.Set).StringVar(&c.Token.Value)
	c.CmdClause.Flag("url", "URL of your Grafana Instance").Action(c.URL.Set).StringVar(&c.URL.Value)
	c.CmdClause.Flag("index", "Stream identifier").Action(c.Index.Set).StringVar(&c.Index.Value)
	return &c
}

// ConstructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *UpdateCommand) ConstructInput(serviceID string, serviceVersion int) (*fastly.UpdateGrafanaCloudLogsInput, error) {
	input := fastly.UpdateGrafanaCloudLogsInput{
		ServiceID:      serviceID,
		ServiceVersion: serviceVersion,
		Name:           c.EndpointName,
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
	if c.NewName.WasSet {
		input.NewName = &c.NewName.Value
	}
	if c.ResponseCondition.WasSet {
		input.ResponseCondition = &c.ResponseCondition.Value
	}
	if c.User.WasSet {
		input.User = &c.User.Value
	}
	if c.URL.WasSet {
		input.URL = &c.URL.Value
	}
	if c.Token.WasSet {
		input.Token = &c.Token.Value
	}
	if c.Index.WasSet {
		input.Index = &c.Index.Value
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

	grafanacloudlogs, err := c.Globals.APIClient.UpdateGrafanaCloudLogs(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Success(out,
		"Updated Grafana Cloud Logs logging endpoint %s (service %s version %d)",
		fastly.ToValue(grafanacloudlogs.Name),
		fastly.ToValue(grafanacloudlogs.ServiceID),
		fastly.ToValue(grafanacloudlogs.ServiceVersion),
	)
	return nil
}
