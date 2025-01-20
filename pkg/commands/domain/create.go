package domain

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"4d63.com/optional"
	"github.com/fastly/go-fastly/v9/fastly"
	v1 "github.com/fastly/go-fastly/v9/fastly/domains/v1"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// CreateCommand calls the Fastly API to create domains.
type CreateCommand struct {
	argparser.Base

	// Required.
	serviceVersion argparser.OptionalServiceVersion

	// Optional.
	apiVersion  argparser.OptionalString
	autoClone   argparser.OptionalAutoClone
	comment     argparser.OptionalString
	fqdn        argparser.OptionalString
	name        argparser.OptionalString
	serviceName argparser.OptionalServiceNameID
}

// NewCreateCommand returns a usable command registered under the parent.
func NewCreateCommand(parent argparser.Registerer, g *global.Data) *CreateCommand {
	c := CreateCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("create", "Create a domain").Alias("add")

	// Optional.
	c.CmdClause.Flag("api-version", fmt.Sprintf("The Fastly API version (%s)", strings.Join(APIVersions, ","))).HintOptions(APIVersions...).Action(c.apiVersion.Set).EnumVar(&c.apiVersion.Value, APIVersions...)
	c.RegisterAutoCloneFlag(argparser.AutoCloneFlagOpts{
		Action: c.autoClone.Set,
		Dst:    &c.autoClone.Value,
	})
	c.CmdClause.Flag("comment", "A descriptive note").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("fqdn", "The fully qualified domain name (version support: v1)").Action(c.fqdn.Set).StringVar(&c.fqdn.Value)
	c.CmdClause.Flag("name", "Domain name").Short('n').Action(c.name.Set).StringVar(&c.name.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *CreateCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.apiVersion.Value == "v1" {
		return c.v1(out)
	}
	return c.v0(out)
}

func (c *CreateCommand) v1(out io.Writer) error {
	if !c.fqdn.WasSet {
		return errors.New("--fqdn required when using --api-version")
	}
	input := &v1.CreateInput{
		FQDN: &c.fqdn.Value,
	}

	serviceID, _, _, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err == nil {
		input.ServiceID = &serviceID
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	d, err := v1.Create(fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"FQDN":       c.fqdn.Value,
			"Service ID": serviceID,
		})
		return err
	}

	serviceOutput := ""
	if d.ServiceID != nil {
		serviceOutput = fmt.Sprintf(" (service-id: %s)", *d.ServiceID)
	}

	text.Success(out, "Created domain '%s' (domain-id: %s)%s", d.FQDN, d.DomainID, serviceOutput)
	return nil
}

func (c *CreateCommand) v0(out io.Writer) error {
	serviceID, serviceVersion, err := argparser.ServiceDetails(argparser.ServiceDetailsOpts{
		Active:             optional.Of(false),
		Locked:             optional.Of(false),
		AutoCloneFlag:      c.autoClone,
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
	input := fastly.CreateDomainInput{
		ServiceID:      serviceID,
		ServiceVersion: fastly.ToValue(serviceVersion.Number),
	}
	if c.name.WasSet {
		input.Name = &c.name.Value
	}
	if c.comment.WasSet {
		input.Comment = &c.comment.Value
	}
	d, err := c.Globals.APIClient.CreateDomain(&input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Service ID":      serviceID,
			"Service Version": fastly.ToValue(serviceVersion.Number),
		})
		return err
	}

	text.Success(out, "Created domain %s (service %s version %d)", fastly.ToValue(d.Name), fastly.ToValue(d.ServiceID), fastly.ToValue(d.ServiceVersion))
	return nil
}
