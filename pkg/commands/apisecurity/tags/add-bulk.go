package tags

import (
	"context"
	"errors"
	"io"

	"github.com/fastly/go-fastly/v13/fastly"
	"github.com/fastly/go-fastly/v13/fastly/apisecurity/operations"

	"github.com/fastly/cli/pkg/argparser"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// AddBulkCommand calls the Fastly API to add tags to multiple operations.
type AddBulkCommand struct {
	argparser.Base
	argparser.JSONOutput

	// Required.
	operationIDs []string
	tagIDs       []string

	// Optional.
	serviceName argparser.OptionalServiceNameID
}

// NewAddBulkCommand returns a usable command registered under the parent.
func NewAddBulkCommand(parent argparser.Registerer, g *global.Data) *AddBulkCommand {
	c := AddBulkCommand{
		Base: argparser.Base{
			Globals: g,
		},
	}

	c.CmdClause = parent.Command("add-bulk", "Add tags to multiple operations")

	// Required.
	c.CmdClause.Flag("operation-id", "Operation ID. Set flag multiple times to include multiple operations").Required().StringsVar(&c.operationIDs)
	c.CmdClause.Flag("tag-id", "Tag ID to add. Set flag multiple times to include multiple tags").Required().StringsVar(&c.tagIDs)

	// Optional.
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &g.Manifest.Flag.ServiceID,
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceNameDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlagBool(c.JSONFlag())

	return &c
}

// Exec invokes the application logic for the command.
func (c *AddBulkCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Globals.Verbose() && c.JSONOutput.Enabled {
		return fsterr.ErrInvalidVerboseJSONCombo
	}

	serviceID, source, flag, err := argparser.ServiceID(c.serviceName, *c.Globals.Manifest, c.Globals.APIClient, c.Globals.ErrLog)
	if err != nil {
		return err
	}
	if c.Globals.Verbose() {
		argparser.DisplayServiceID(serviceID, flag, source, out)
	}

	if serviceID == "" {
		return errors.New("service-id is required")
	}

	if len(c.operationIDs) == 0 {
		return errors.New("at least one operation-id must be provided")
	}
	if len(c.tagIDs) == 0 {
		return errors.New("at least one tag-id must be provided")
	}

	fc, ok := c.Globals.APIClient.(*fastly.Client)
	if !ok {
		return errors.New("failed to convert interface to a fastly client")
	}

	result, err := operations.BulkAddTags(context.TODO(), fc, &operations.BulkAddTagsInput{
		ServiceID:    &serviceID,
		OperationIDs: c.operationIDs,
		TagIDs:       c.tagIDs,
	})
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if ok, err := c.WriteJSON(out, result); ok {
		return err
	}

	text.Success(out, "Bulk add tags completed. Processed %d operations", len(result.Data))
	return nil
}
