package compute

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
)

// PublishCommand produces and deploys an artifact from files on the local disk.
type PublishCommand struct {
	common.Base
	build  *BuildCommand
	deploy *DeployCommand
}

// NewPublishCommand returns a usable command registered under the parent.
func NewPublishCommand(parent common.Registerer, globals *config.Data, build *BuildCommand, deploy *DeployCommand) *PublishCommand {
	var c PublishCommand
	c.Globals = globals
	c.build = build
	c.deploy = deploy
	c.CmdClause = parent.Command("publish", "Runs \"build\" then \"deploy\" using fastly.toml manifest configuration")
	return &c
}

// Exec implements the command interface.
//
// NOTE: unlike other non-aggregate commands that initialize a new
// text.Progress type for displaying progress information to the user, we don't
// use that in this command because the nested commands overlap the output in
// non-deterministic ways. It's best to leave those nested commands to handle
// the progress indicator.
func (c *PublishCommand) Exec(in io.Reader, out io.Writer) (err error) {
	err = c.build.Exec(in, out)
	if err != nil {
		return fmt.Errorf("error building package: %w", err)
	}

	text.Break(out)

	err = c.deploy.Exec(in, out)
	if err != nil {
		return fmt.Errorf("error deploying package: %w", err)
	}

	return nil
}
