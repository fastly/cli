package kinesis

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v3/fastly"
)

// ListCommand calls the Fastly API to list Amazon Kinesis logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest manifest.Data
	Input    fastly.ListKinesisInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Kinesis endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').Envar(env.ServiceID).StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.ServiceVersion)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.ServiceID = serviceID

	kineses, err := c.Globals.Client.ListKinesis(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, kinesis := range kineses {
			tw.AddLine(kinesis.ServiceID, kinesis.ServiceVersion, kinesis.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.ServiceID)
	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, kinesis := range kineses {
		fmt.Fprintf(out, "\tKinesis %d/%d\n", i+1, len(kineses))
		fmt.Fprintf(out, "\t\tService ID: %s\n", kinesis.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", kinesis.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", kinesis.Name)
		fmt.Fprintf(out, "\t\tStream name: %s\n", kinesis.StreamName)
		fmt.Fprintf(out, "\t\tRegion: %s\n", kinesis.Region)
		if kinesis.AccessKey != "" || kinesis.SecretKey != "" {
			fmt.Fprintf(out, "\t\tAccess key: %s\n", kinesis.AccessKey)
			fmt.Fprintf(out, "\t\tSecret key: %s\n", kinesis.SecretKey)
		}
		if kinesis.IAMRole != "" {
			fmt.Fprintf(out, "\t\tIAM role: %s\n", kinesis.IAMRole)
		}
		fmt.Fprintf(out, "\t\tFormat: %s\n", kinesis.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", kinesis.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", kinesis.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", kinesis.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
