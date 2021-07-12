package compute

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
)

// PublishCommand produces and deploys an artifact from files on the local disk.
type PublishCommand struct {
	cmd.Base
	manifest manifest.Data
	build    *BuildCommand
	deploy   *DeployCommand

	// Deploy fields
	path           cmd.OptionalString
	domain         cmd.OptionalString
	backend        cmd.OptionalString
	backendPort    cmd.OptionalUint
	serviceVersion cmd.OptionalServiceVersion

	// Build fields
	name       cmd.OptionalString
	lang       cmd.OptionalString
	includeSrc cmd.OptionalBool
	force      cmd.OptionalBool
	timeout    cmd.OptionalInt
}

// NewPublishCommand returns a usable command registered under the parent.
func NewPublishCommand(parent cmd.Registerer, globals *config.Data, build *BuildCommand, deploy *DeployCommand) *PublishCommand {
	var c PublishCommand
	c.Globals = globals
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.build = build
	c.deploy = deploy
	c.CmdClause = parent.Command("publish", "Build and deploy a Compute@Edge package to a Fastly service")

	// Build flags
	c.CmdClause.Flag("name", "Package name").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("force", "Skip verification steps and force build").Action(c.force.Set).BoolVar(&c.force.Value)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").Action(c.timeout.Set).IntVar(&c.timeout.Value)

	// Deploy flags
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Action:   c.serviceVersion.Set,
		Dst:      &c.serviceVersion.Value,
		Optional: true,
	})
	c.CmdClause.Flag("path", "Path to package").Short('p').Action(c.path.Set).StringVar(&c.path.Value)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").Action(c.domain.Set).StringVar(&c.domain.Value)
	c.CmdClause.Flag("backend", "A hostname, IPv4, or IPv6 address for the package backend").Action(c.backend.Set).StringVar(&c.backend.Value)
	c.CmdClause.Flag("backend-port", "A port number for the package backend").Action(c.backendPort.Set).UintVar(&c.backendPort.Value)

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
	if c.name.WasSet {
		c.build.PackageName = c.name.Value
	}
	if c.lang.WasSet {
		c.build.Lang = c.lang.Value
	}
	if c.includeSrc.WasSet {
		c.build.IncludeSrc = c.includeSrc.Value
	}
	if c.force.WasSet {
		c.build.Force = c.force.Value
	}
	if c.timeout.WasSet {
		c.build.Timeout = c.timeout.Value
	}

	err = c.build.Exec(in, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Break(out)

	// Reset the fields on the DeployCommand based on PublishCommand values.
	if c.path.WasSet {
		c.deploy.Path = c.path.Value
	}
	if c.serviceVersion.WasSet {
		c.deploy.ServiceVersion = c.serviceVersion // deploy's field is a cmd.OptionalServiceVersion
	}
	if c.domain.WasSet {
		c.deploy.Domain = c.domain.Value
	}
	if c.backend.WasSet {
		c.deploy.Backend = c.backend.Value
	}
	if c.backendPort.WasSet {
		c.deploy.BackendPort = c.backendPort.Value
	}
	c.deploy.Manifest = c.manifest

	err = c.deploy.Exec(in, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	return nil
}
