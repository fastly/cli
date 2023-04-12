package ratelimit

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/lookup"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/go-fastly/v8/fastly"
)

// NewDescribeCommand returns a usable command registered under the parent.
func NewDescribeCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *DescribeCommand {
	var c DescribeCommand
	c.CmdClause = parent.Command("describe", "Get a rate limiter by its ID").Alias("get")
	c.Globals = g
	c.manifest = m

	// Required.
	c.CmdClause.Flag("id", "Alphanumeric string identifying the rate limiter").Required().StringVar(&c.id)

	// Optional.
	c.RegisterFlagBool(c.JSONFlag()) // --json

	return &c
}

// DescribeCommand calls the Fastly API to describe an appropriate resource.
type DescribeCommand struct {
	cmd.Base
	cmd.JSONOutput

	id       string
	manifest manifest.Data
}

// Exec invokes the application logic for the command.
func (c *DescribeCommand) Exec(_ io.Reader, out io.Writer) error {
	_, s := c.Globals.Token()
	if s == lookup.SourceUndefined {
		return fsterr.ErrNoToken
	}

	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input, err := c.constructInput()
	if err != nil {
		return err
	}

	o, err := c.Globals.APIClient.GetERL(input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, o); ok {
		return err
	}

	c.print(out, o)
	return nil
}

// constructInput transforms values parsed from CLI flags into an object to be used by the API client library.
func (c *DescribeCommand) constructInput() (*fastly.GetERLInput, error) {
	var input fastly.GetERLInput

	input.ERLID = c.id

	return &input, nil
}

// print displays the information returned from the API.
func (c *DescribeCommand) print(out io.Writer, o *fastly.ERL) {
	fmt.Fprintf(out, "\nAction: %+v\n", o.Action)
	fmt.Fprintf(out, "Client Key: %+v\n", o.ClientKey)
	fmt.Fprintf(out, "Feature Revision: %+v\n", o.FeatureRevision)
	fmt.Fprintf(out, "HTTP Methods: %+v\n", o.HTTPMethods)
	fmt.Fprintf(out, "ID: %+v\n", o.ID)
	fmt.Fprintf(out, "Logger Type: %+v\n", o.LoggerType)
	fmt.Fprintf(out, "Name: %+v\n", o.Name)
	fmt.Fprintf(out, "Penalty Box Duration: %+v\n", o.PenaltyBoxDuration)
	fmt.Fprintf(out, "Response: %+v\n", o.Response)
	fmt.Fprintf(out, "Response Object Name: %+v\n", o.ResponseObjectName)
	fmt.Fprintf(out, "RPS Limit: %+v\n", o.RpsLimit)
	fmt.Fprintf(out, "Service ID: %+v\n", o.ServiceID)
	fmt.Fprintf(out, "URI Dictionary Name: %+v\n", o.URIDictionaryName)
	fmt.Fprintf(out, "Version: %+v\n", o.Version)
	fmt.Fprintf(out, "WindowSize: %+v\n", o.WindowSize)

	if o.CreatedAt != nil {
		fmt.Fprintf(out, "Created at: %s\n", o.CreatedAt)
	}
	if o.UpdatedAt != nil {
		fmt.Fprintf(out, "Updated at: %s\n", o.UpdatedAt)
	}
	if o.DeletedAt != nil {
		fmt.Fprintf(out, "Deleted at: %s\n", o.DeletedAt)
	}
}
