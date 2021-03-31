package compute

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
)

// PublishCommand produces and deploys an artifact from files on the local disk.
type PublishCommand struct {
	common.Base
	manifest manifest.Data
	build    *BuildCommand
	deploy   *DeployCommand

	// Deploy fields
	path    string
	version common.OptionalInt

	// Build fields
	name       string
	lang       string
	includeSrc bool
	force      bool
}

// NewPublishCommand returns a usable command registered under the parent.
func NewPublishCommand(parent common.Registerer, globals *config.Data, build *BuildCommand, deploy *DeployCommand) *PublishCommand {
	var c PublishCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.build = build
	c.deploy = deploy
	c.CmdClause = parent.Command("publish", "Composite of \"build\", then \"deploy\"")

	// Deploy flags
	c.CmdClause.Flag("service-id", "Service ID").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("version", "Number of version to activate").Action(c.version.Set).IntVar(&c.version.Value)
	c.CmdClause.Flag("path", "Path to package").Short('p').StringVar(&c.path)

	// Build flags
	c.CmdClause.Flag("name", "Package name").StringVar(&c.name)
	c.CmdClause.Flag("language", "Language type").StringVar(&c.lang)
	c.CmdClause.Flag("include-source", "Include source code in built package").BoolVar(&c.includeSrc)
	c.CmdClause.Flag("force", "Skip verification steps and force build").BoolVar(&c.force)

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
	// Reset the fields on the BuildCommand based on PublishCommand values.
	c.build.PackageName = c.name
	c.build.Lang = c.lang
	c.build.IncludeSrc = c.includeSrc
	c.build.Force = c.force

	err = c.build.Exec(in, out)
	if err != nil {
		return fmt.Errorf("error building package: %w", err)
	}

	text.Break(out)

	// Reset the fields on the DeployCommand based on PublishCommand values.
	c.deploy.Path = c.path
	c.deploy.Version = c.version

	err = c.deploy.Exec(in, out)
	if err != nil {
		return fmt.Errorf("error deploying package: %w", err)
	}

	return nil
}
