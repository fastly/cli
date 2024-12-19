package bot_management

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/go-fastly/v9/fastly/products/bot_management"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// DisableFn is a dependency-injection point for unit tests to provide
// a mock implementation of the API operation.
var DisableFn = func(client api.Interface, serviceID string) error {
	return bot_management.Disable(client.(*fastly.Client), serviceID)
}

// DisableCommand calls the Fastly API to disable the product.
type DisableCommand struct {
	productcore.Disable
}

// NewDisableCommand returns a usable command registered under the parent.
func NewDisableCommand(parent argparser.Registerer, g *global.Data) *DisableCommand {
	c := DisableCommand{}
	c.Init(parent, g, bot_management.ProductName)
	return &c
}

// Exec invokes the application logic for the command.
func (cmd *DisableCommand) Exec(_ io.Reader, out io.Writer) error {
	return cmd.Disable.Exec(out, DisableFn)
}
