package botmanagement

import (
	"io"

	"github.com/fastly/go-fastly/v9/fastly/products/botmanagement"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
)

// StatusCommand calls the Fastly API to get the enablement status of the product.
type StatusCommand struct {
	productcore.Status[*botmanagement.EnableOutput]
}

// NewStatusCommand returns a usable command registered under the parent.
func NewStatusCommand(parent argparser.Registerer, g *global.Data) *StatusCommand {
	c := StatusCommand{}
	c.Init(parent, g, botmanagement.ProductName, &EnablementHooks)
	return &c
}

// Exec invokes the application logic for the command.
func (cmd *StatusCommand) Exec(_ io.Reader, out io.Writer) error {
	return cmd.Status.Exec(out)
}
