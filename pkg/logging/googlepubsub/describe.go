package googlepubsub

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
)

// DescribeCommand calls the Fastly API to describe a Google Cloud Pub/Sub logging endpoint.
type DescribeCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.GetPubsubInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, globals *config.Data) *DescribeCommand {
	var c DescribeCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("describe", "Show detailed information about a Google Cloud Pub/Sub logging endpoint on a Fastly service version").Alias("get")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	c.CmdClause.Flag("name", "The name of the Google Cloud Pub/Sub logging object").Short('n').Required().StringVar(&c.Input.Name)
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

	googlepubsub, err := c.Globals.Client.GetPubsub(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	fmt.Fprintf(out, "Service ID: %s\n", googlepubsub.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", googlepubsub.ServiceVersion)
	fmt.Fprintf(out, "Name: %s\n", googlepubsub.Name)
	fmt.Fprintf(out, "User: %s\n", googlepubsub.User)
	fmt.Fprintf(out, "Secret key: %s\n", googlepubsub.SecretKey)
	fmt.Fprintf(out, "Project ID: %s\n", googlepubsub.ProjectID)
	fmt.Fprintf(out, "Topic: %s\n", googlepubsub.Topic)
	fmt.Fprintf(out, "Format: %s\n", googlepubsub.Format)
	fmt.Fprintf(out, "Format version: %d\n", googlepubsub.FormatVersion)
	fmt.Fprintf(out, "Response condition: %s\n", googlepubsub.ResponseCondition)
	fmt.Fprintf(out, "Placement: %s\n", googlepubsub.Placement)

	return nil
}
