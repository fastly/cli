package usagemetrics

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v19/fastly"
	"github.com/fastly/go-fastly/v19/fastly/airuntimecontrol/v1/usagemetrics"

	"github.com/fastly/cli/pkg/argparser"
)

// ExportCommand calls the Fastly API to export usage metrics as CSV.
type ExportCommand struct {
	argparser.Base

	// Optional.
	Filters filterFlags `embed:""`
}

// Exec invokes the application logic for the command.
func (c *ExportCommand) Exec(_ io.Reader, out io.Writer) error {
	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	data, err := usagemetrics.Export(context.TODO(), fc, c.Filters.exportInput())
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	_, err = out.Write(data)
	return err
}
