package syslog

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"

	"github.com/fastly/cli/v10/pkg/argparser"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// DescribeCommand calls the Fastly API to describe a Syslog logging endpoint.
type DescribeCommand struct {
	argparser.Base
	argparser.JSONOutput

	Input          fastly.GetSyslogInput
	serviceName    argparser.OptionalServiceNameID
	serviceVersion argparser.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent argparser.Registerer, g *global.Data) *DescribeCommand {
	c := DescribeCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("describe", "Show detailed information about a Syslog logging endpoint on a Fastly service version").Alias("get")

	// Required.
	c.CmdClause.Flag("name", "The name of the Syslog logging object").Short('n').Required().StringVar(&c.Input.Name)
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
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
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

	o, err := c.Globals.APIClient.GetSyslog(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	lines := text.Lines{
		"Address":                fastly.ToValue(o.Address),
		"Format version":         fastly.ToValue(o.FormatVersion),
		"Format":                 fastly.ToValue(o.Format),
		"Hostname":               fastly.ToValue(o.Hostname),
		"IPV4":                   fastly.ToValue(o.IPV4),
		"Message type":           fastly.ToValue(o.MessageType),
		"Name":                   fastly.ToValue(o.Name),
		"Placement":              fastly.ToValue(o.Placement),
		"Port":                   fastly.ToValue(o.Port),
		"Response condition":     fastly.ToValue(o.ResponseCondition),
		"TLS CA certificate":     fastly.ToValue(o.TLSCACert),
		"TLS client certificate": fastly.ToValue(o.TLSClientCert),
		"TLS client key":         fastly.ToValue(o.TLSClientKey),
		"TLS hostname":           fastly.ToValue(o.TLSHostname),
		"Token":                  fastly.ToValue(o.Token),
		"Use TLS":                fastly.ToValue(o.UseTLS),
		"Version":                fastly.ToValue(o.ServiceVersion),
	}
	if !c.Globals.Verbose() {
		lines["Service ID"] = fastly.ToValue(o.ServiceID)
	}
	text.PrintLines(out, lines)

	return nil
}
