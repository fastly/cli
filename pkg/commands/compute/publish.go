package compute

import (
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// PublishCommand produces and deploys an artifact from files on the local disk.
type PublishCommand struct {
	cmd.Base
	manifest manifest.Data
	build    *BuildCommand
	deploy   *DeployCommand

	// Build fields
	acceptCustomBuild cmd.OptionalBool
	includeSrc        cmd.OptionalBool
	lang              cmd.OptionalString
	name              cmd.OptionalString
	skipVerification  cmd.OptionalBool
	timeout           cmd.OptionalInt

	// Deploy fields
	acceptDefaults cmd.OptionalBool
	comment        cmd.OptionalString
	domain         cmd.OptionalString
	path           cmd.OptionalString
	serviceName    cmd.OptionalServiceNameID
	serviceVersion cmd.OptionalServiceVersion
}

// NewPublishCommand returns a usable command registered under the parent.
func NewPublishCommand(parent cmd.Registerer, globals *config.Data, build *BuildCommand, deploy *DeployCommand, data manifest.Data) *PublishCommand {
	var c PublishCommand
	c.Globals = globals
	c.manifest = data
	c.build = build
	c.deploy = deploy
	c.CmdClause = parent.Command("publish", "Build and deploy a Compute@Edge package to a Fastly service")

	c.CmdClause.Flag("accept-custom-build", "Do not prompt when project manifest defines [scripts.build]").Action(c.acceptCustomBuild.Set).BoolVar(&c.acceptCustomBuild.Value)
	c.CmdClause.Flag("accept-defaults", "Accept default values for all prompts and perform deploy non-interactively").Action(c.acceptDefaults.Set).BoolVar(&c.acceptDefaults.Value)
	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").Action(c.domain.Set).StringVar(&c.domain.Value)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("name", "Package name").Action(c.name.Set).StringVar(&c.name.Value)
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').Action(c.path.Set).StringVar(&c.path.Value)
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagServiceIDName,
		Description: cmd.FlagServiceIDDesc,
		Dst:         &c.manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        cmd.FlagServiceName,
		Description: cmd.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.RegisterFlag(cmd.StringFlagOpts{
		Name:        cmd.FlagVersionName,
		Description: cmd.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Action:      c.serviceVersion.Set,
	})
	c.CmdClause.Flag("skip-verification", "Skip verification steps and force build").Action(c.skipVerification.Set).BoolVar(&c.skipVerification.Value)
	c.CmdClause.Flag("timeout", "Timeout, in seconds, for the build compilation step").Action(c.timeout.Set).IntVar(&c.timeout.Value)

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
	if c.acceptCustomBuild.WasSet {
		c.build.Flags.AcceptCustomBuild = c.acceptCustomBuild.Value
	}
	if c.includeSrc.WasSet {
		c.build.Flags.IncludeSrc = c.includeSrc.Value
	}
	if c.lang.WasSet {
		c.build.Flags.Lang = c.lang.Value
	}
	if c.name.WasSet {
		c.build.Flags.PackageName = c.name.Value
	}
	if c.skipVerification.WasSet {
		c.build.Flags.SkipVerification = c.skipVerification.Value
	}
	if c.timeout.WasSet {
		c.build.Flags.Timeout = c.timeout.Value
	}
	c.build.Manifest = c.manifest

	err = c.build.Exec(in, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Break(out)

	// Reset the fields on the DeployCommand based on PublishCommand values.
	if c.name.WasSet {
		c.manifest.Flag.Name = c.name.Value
	}
	if c.acceptDefaults.WasSet {
		c.deploy.AcceptDefaults = c.acceptDefaults.Value
	}
	if c.path.WasSet {
		c.deploy.Path = c.path.Value
	}
	if c.serviceName.WasSet {
		c.deploy.ServiceName = c.serviceName // deploy's field is a cmd.OptionalServiceNameID
	}
	if c.serviceVersion.WasSet {
		c.deploy.ServiceVersion = c.serviceVersion // deploy's field is a cmd.OptionalServiceVersion
	}
	if c.domain.WasSet {
		c.deploy.Domain = c.domain.Value
	}
	if c.comment.WasSet {
		c.deploy.Comment = c.comment
	}
	c.deploy.Manifest = c.manifest

	err = c.deploy.Exec(in, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	return nil
}
