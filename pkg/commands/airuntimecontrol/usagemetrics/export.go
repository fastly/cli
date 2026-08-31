package usagemetrics

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v17/fastly"
	"github.com/fastly/go-fastly/v17/fastly/airuntimecontrol/v1/usagemetrics"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// ExportCommand calls the Fastly API to export usage metrics as CSV.
type ExportCommand struct {
	argparser.Base

	filters filterFlags
}

// NewExportCommand returns a usable command registered under the parent.
func NewExportCommand(parent argparser.Registerer, g *global.Data) *ExportCommand {
	c := ExportCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}
	c.CmdClause = parent.Command("export", "Export usage metrics as CSV")

	// Optional.
	c.filters.register(c.Base)

	return &c
}

// Exec invokes the application logic for the command.
func (c *ExportCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := usagemetrics.Export(context.TODO(), fc, c.filters.input())
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	_, err = out.Write(data)
	return err
}
