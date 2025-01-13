package ddosprotection

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly/products/ddosprotection"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// DisableCommand calls the Fastly API to disable the product.
type DisableCommand struct {
	productcore.Disable[*ddosprotection.EnableOutput]
}

// NewDisableCommand returns a usable command registered under the parent.
func NewDisableCommand(parent argparser.Registerer, g *global.Data) *DisableCommand {
	c := DisableCommand{}
	c.Init(parent, g, ddosprotection.ProductID, ddosprotection.ProductName, &EnablementHooks)
	return &c
}

// Exec invokes the application logic for the command.
func (cmd *DisableCommand) Exec(_ io.Reader, out io.Writer) error {
	return cmd.Disable.Exec(out)
}
