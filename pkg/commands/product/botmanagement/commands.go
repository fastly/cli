package botmanagement

import (
	"context"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/go-fastly/v11/fastly"
	product "github.com/fastly/go-fastly/v11/fastly/products/botmanagement"

	"github.com/fastly/cli/internal/productcore"
	"github.com/fastly/cli/pkg/api"
)

// EnablementHooks is a structure of dependency-injection points used
// by unit tests to provide mock behaviors
var EnablementHooks = productcore.EnablementHookFuncs[product.EnableOutput]{
	DisableFunc: func(client api.Interface, serviceID string) error {
		return product.Disable(context.TODO(), client.(*fastly.Client), serviceID)
	},
	EnableFunc: func(client api.Interface, serviceID string) (product.EnableOutput, error) {
		return product.Enable(context.TODO(), client.(*fastly.Client), serviceID)
	},
	GetFunc: func(client api.Interface, serviceID string) (product.EnableOutput, error) {
		return product.Get(context.TODO(), client.(*fastly.Client), serviceID)
	},
}

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	argparser.Base
	// no flags
}

// CommandName is the string to be used to invoke this command
const CommandName = "bot_management"

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent argparser.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command(CommandName, "Enable and disable the Bot Management product")
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, _ io.Writer) error {
	panic("unreachable")
}

// EnableCommand calls the Fastly API to disable the product.
type EnableCommand struct {
	productcore.Enable[product.EnableOutput, *productcore.EnablementStatus[product.EnableOutput]]
}

// NewEnableCommand returns a usable command registered under the parent.
func NewEnableCommand(parent argparser.Registerer, g *global.Data) *EnableCommand {
	c := EnableCommand{}
	c.Init(parent, g, product.ProductName, &EnablementHooks)
	return &c
}

// Exec invokes the application logic for the command.
func (cmd *EnableCommand) Exec(_ io.Reader, out io.Writer) error {
	status := &productcore.EnablementStatus[product.EnableOutput]{ProductName: product.ProductName}
	return cmd.Enable.Exec(out, status)
}

// DisableCommand calls the Fastly API to disable the product.
type DisableCommand struct {
	productcore.Disable[product.EnableOutput, *productcore.EnablementStatus[product.EnableOutput]]
}

// NewDisableCommand returns a usable command registered under the parent.
func NewDisableCommand(parent argparser.Registerer, g *global.Data) *DisableCommand {
	c := DisableCommand{}
	c.Init(parent, g, product.ProductID, product.ProductName, &EnablementHooks)
	return &c
}

// Exec invokes the application logic for the command.
func (cmd *DisableCommand) Exec(_ io.Reader, out io.Writer) error {
	status := &productcore.EnablementStatus[product.EnableOutput]{ProductName: product.ProductName}
	return cmd.Disable.Exec(out, status)
}

// StatusCommand calls the Fastly API to get the enablement status of the product.
type StatusCommand struct {
	productcore.Status[product.EnableOutput, *productcore.EnablementStatus[product.EnableOutput]]
}

// NewStatusCommand returns a usable command registered under the parent.
func NewStatusCommand(parent argparser.Registerer, g *global.Data) *StatusCommand {
	c := StatusCommand{}
	c.Init(parent, g, product.ProductName, &EnablementHooks)
	return &c
}

// Exec invokes the application logic for the command.
func (cmd *StatusCommand) Exec(_ io.Reader, out io.Writer) error {
	status := &productcore.EnablementStatus[product.EnableOutput]{ProductName: product.ProductName}
	return cmd.Status.Exec(out, status)
}
