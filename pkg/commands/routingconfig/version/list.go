package version

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/domainmanagement/v1/routingconfigs/versions"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// ListCommand calls the Fastly API to list a routing config's inactive
// versions.
type ListCommand struct {
	argparser.Base

	// Required.
	RoutingConfigID string `required:"" help:"The Routing Config Identifier"`

	argparser.JSONOutput

	// Optional.
	Sort argparser.OptionalString `help:"The order in which to list the results"`
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	input := &versions.ListInput{
		RoutingConfigID: &c.RoutingConfigID,
	}
	if c.Sort.WasSet {
		input.Sort = &c.Sort.Value
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	cl, err := versions.List(context.TODO(), fc, input)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Routing Config ID": c.RoutingConfigID,
		})
		return err
	}

	if ok, err := c.WriteJSON(out, cl); ok {
		return err
	}

	t := text.NewTable(out)
	t.AddHeader("VERSION ID", "COMMENT", "CREATED AT", "ACTIVATED AT")
	for _, v := range cl {
		var activatedAt string
		if v.ActivatedAt != nil {
			activatedAt = v.ActivatedAt.String()
		}
		t.AddLine(v.VersionID, v.Comment, v.CreatedAt, activatedAt)
	}
	t.Print()
	return nil
}
