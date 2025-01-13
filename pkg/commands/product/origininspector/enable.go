package origininspector

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly/products/origininspector"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// EnableCommand calls the Fastly API to disable the product.
type EnableCommand struct {
	productcore.Enable[*origininspector.EnableOutput]
}

// NewEnableCommand returns a usable command registered under the parent.
func NewEnableCommand(parent argparser.Registerer, g *global.Data) *EnableCommand {
	c := EnableCommand{}
	c.Init(parent, g, origininspector.ProductID, origininspector.ProductName, &EnablementHooks)
	return &c
}

// Exec invokes the application logic for the command.
func (cmd *EnableCommand) Exec(_ io.Reader, out io.Writer) error {
	return cmd.Enable.Exec(out)
}
