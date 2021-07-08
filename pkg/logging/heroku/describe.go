package heroku

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Heroku logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetHerokuInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Heroku logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the Heroku logging object").Short('n').Required().StringVar(&c.Input.Name)
	return &c
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	heroku, err := c.Globals.Client.GetHeroku(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", heroku.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", heroku.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", heroku.Name)
	fmt.Fprintf(out, "URL: %s\n", heroku.URL)
	fmt.Fprintf(out, "Token: %s\n", heroku.Token)
	fmt.Fprintf(out, "Format: %s\n", heroku.Format)
	fmt.Fprintf(out, "Format version: %d\n", heroku.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", heroku.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", heroku.Placement)

	return nil
}
