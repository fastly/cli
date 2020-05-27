package googlepubsub

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
)

// ListCommand calls the Fastly API to list Google Cloud Pub/Sub logging endpoints.
type ListCommand struct {
	common.Base
	manifest manifest.Data
	Input    fastly.ListPubsubsInput
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent common.Registerer, globals *config.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("list", "List Google Cloud Pub/Sub endpoints on a Fastly service version")
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of service version").Required().IntVar(&c.Input.Version)
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, source := c.manifest.ServiceID()
	if source == manifest.SourceUndefined {
		return errors.ErrNoServiceID
	}
	c.Input.Service = serviceID

	googlepubsubs, err := c.Globals.Client.ListPubsubs(&c.Input)
	if err != nil {
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, googlepubsub := range googlepubsubs {
			tw.AddLine(googlepubsub.ServiceID, googlepubsub.Version, googlepubsub.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Service ID: %s\n", c.Input.Service)
	fmt.Fprintf(out, "Version: %d\n", c.Input.Version)
	for i, googlepubsub := range googlepubsubs {
		fmt.Fprintf(out, "\tGoogle Cloud Pub/Sub %d/%d\n", i+1, len(googlepubsubs))
		fmt.Fprintf(out, "\t\tService ID: %s\n", googlepubsub.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", googlepubsub.Version)
		fmt.Fprintf(out, "\t\tName: %s\n", googlepubsub.Name)
		fmt.Fprintf(out, "\t\tUser: %s\n", googlepubsub.User)
		fmt.Fprintf(out, "\t\tSecret key: %s\n", googlepubsub.SecretKey)
		fmt.Fprintf(out, "\t\tProject ID: %s\n", googlepubsub.ProjectID)
		fmt.Fprintf(out, "\t\tTopic: %s\n", googlepubsub.Topic)
		fmt.Fprintf(out, "\t\tFormat: %s\n", googlepubsub.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", googlepubsub.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", googlepubsub.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", googlepubsub.Placement)

	}
	fmt.Fprintln(out)

	return nil
}
