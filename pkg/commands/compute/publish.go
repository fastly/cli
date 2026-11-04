package compute

import (
	"fmt"
	"io"
	"os"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// PublishCommand produces and deploys an artifact from files on the local disk.
type PublishCommand struct {
	argparser.Base
	build  *BuildCommand
	deploy *DeployCommand

	// Build fields
	Dir                   argparser.OptionalString `short:"C" help:"Project directory to build (default: current directory)"`
	IncludeSrc            argparser.OptionalBool   `name:"include-source" help:"Include source code in built package"`
	Lang                  argparser.OptionalString `name:"language" help:"Language type"`
	MetadataDisable       argparser.OptionalBool   `help:"Disable Wasm binary metadata annotations"`
	MetadataFilterEnvVars argparser.OptionalString `name:"metadata-filter-envvars" help:"Redact specified environment variables from [scripts.env_vars] using comma-separated list"`
	MetadataShow          argparser.OptionalBool   `help:"Inspect the Wasm binary metadata"`
	PackageName           argparser.OptionalString `help:"Package name"`
	Timeout               argparser.OptionalInt    `help:"Timeout, in seconds, for the build compilation step"`

	// Deploy fields
	Comment            argparser.OptionalString         `help:"Human-readable comment"`
	Domain             argparser.OptionalString         `help:"The name of the domain associated to the package"`
	Env                argparser.OptionalString         `help:"The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')"`
	NoDefaultDomain    argparser.OptionalBool           `help:"Skip default domain creation"`
	Pkg                argparser.OptionalString         `name:"package" short:"p" help:"Path to a package tar.gz"`
	ServiceName        argparser.OptionalServiceNameID  `help:"The name of the service"`
	ServiceVersion     argparser.OptionalServiceVersion `name:"version" help:"'latest', 'active', 'staged', or the number of a specific Fastly service version (default: the active version, or the latest version if none is active)"`
	StatusCheckCode    int                              `help:"Set the expected status response for the service availability check to the root path"`
	StatusCheckOff     bool                             `help:"Disable the service availability check"`
	StatusCheckPath    string                           `default:"/" help:"Specify the URL path for the service availability check"`
	StatusCheckTimeout int                              `default:"120" help:"Set a timeout (in seconds) for the service availability check"`

	// Publish private fields
	projectDir string

	ServiceID *string `short:"s" help:"Service ID (falls back to FASTLY_SERVICE_ID, then fastly.toml)"`
}

// NewPublishCommand returns a PublishCommand set up as it was before parsing.
// Its flags now come from the struct's tags.
func NewPublishCommand(g *global.Data, build *BuildCommand, deploy *DeployCommand) *PublishCommand {
	var c PublishCommand
	c.Globals = g
	c.build = build
	c.deploy = deploy

	return &c
}

// BeforeApply sets the command up with the values bound when the parser
// was built, before kong applies the command line.
func (c *PublishCommand) BeforeApply(g *global.Data) error {
	c.build = &BuildCommand{}
	c.build.Globals = g
	c.deploy = &DeployCommand{}
	c.deploy.Globals = g
	return nil
}

// AfterApply hands the parsed flags over to the fields the command reads them
// from.
func (c *PublishCommand) AfterApply() error {
	if c.ServiceID != nil {
		c.Globals.Manifest.Flag.ServiceID = *c.ServiceID
	}
	return nil
}

// Exec implements the command interface.
//
// NOTE: unlike other non-aggregate commands that initialize a new
// text.Progress type for displaying progress information to the user, we don't
// use that in this command because the nested commands overlap the output in
// non-deterministic ways. It's best to leave those nested commands to handle
// the progress indicator.
func (c *PublishCommand) Exec(in io.Reader, out io.Writer) (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()

	c.projectDir, err = ChangeProjectDirectory(c.Dir.Value)
	if err != nil {
		return err
	}
	if c.projectDir != "" {
		if c.Globals.Verbose() {
			text.Info(out, ProjectDirMsg, c.projectDir)
		}
	}

	err = c.Build(in, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Break(out)

	err = c.Deploy(in, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	return nil
}

// Build constructs and executes the build logic.
func (c *PublishCommand) Build(in io.Reader, out io.Writer) error {
	// Reset the fields on the BuildCommand based on PublishCommand values.
	if c.Dir.WasSet {
		c.build.Flags.Dir = c.Dir.Value
	}
	if c.Env.WasSet {
		c.build.Flags.Env = c.Env.Value
	}
	if c.IncludeSrc.WasSet {
		c.build.Flags.IncludeSrc = c.IncludeSrc.Value
	}
	if c.Lang.WasSet {
		c.build.Flags.Lang = c.Lang.Value
	}
	if c.PackageName.WasSet {
		c.build.Flags.PackageName = c.PackageName.Value
	}
	if c.Timeout.WasSet {
		c.build.Flags.Timeout = c.Timeout.Value
	}
	if c.MetadataDisable.WasSet {
		c.build.MetadataDisable = c.MetadataDisable.Value
	}
	if c.MetadataFilterEnvVars.WasSet {
		c.build.MetadataFilterEnvVars = c.MetadataFilterEnvVars.Value
	}
	if c.MetadataShow.WasSet {
		c.build.MetadataShow = c.MetadataShow.Value
	}
	if c.projectDir != "" {
		c.build.SkipChangeDir = true // we've already changed directory
	}
	return c.build.Exec(in, out)
}

// Deploy constructs and executes the deploy logic.
func (c *PublishCommand) Deploy(in io.Reader, out io.Writer) error {
	// Reset the fields on the DeployCommand based on PublishCommand values.
	if c.Dir.WasSet {
		c.deploy.Dir = c.Dir.Value
	}
	if c.Pkg.WasSet {
		c.deploy.PackagePath = c.Pkg.Value
	}
	if c.ServiceName.WasSet {
		c.deploy.ServiceName = c.ServiceName // deploy's field is an argparser.OptionalServiceNameID
	}
	if c.ServiceVersion.WasSet {
		c.deploy.ServiceVersion = c.ServiceVersion // deploy's field is an argparser.OptionalServiceVersion
	}
	if c.Domain.WasSet {
		c.deploy.Domain = c.Domain.Value
	}
	if c.Env.WasSet {
		c.deploy.Env = c.Env.Value
	}
	if c.NoDefaultDomain.WasSet {
		c.deploy.NoDefaultDomain = c.NoDefaultDomain
	}
	if c.Comment.WasSet {
		c.deploy.Comment = c.Comment
	}
	if c.StatusCheckCode > 0 {
		c.deploy.StatusCheckCode = c.StatusCheckCode
	}
	if c.StatusCheckOff {
		c.deploy.StatusCheckOff = c.StatusCheckOff
	}
	if c.StatusCheckTimeout > 0 {
		c.deploy.StatusCheckTimeout = c.StatusCheckTimeout
	}
	c.deploy.StatusCheckPath = c.StatusCheckPath
	if c.projectDir != "" {
		c.build.SkipChangeDir = true // we've already changed directory
	}
	return c.deploy.Exec(in, out)
}

// FlagOrder lists the flags in the order kingpin showed them in help, which
// the fields do not follow.
func (c *PublishCommand) FlagOrder() []string {
	return []string{
		"comment", "dir", "domain", "env", "include-source",
		"no-default-domain", "language", "metadata-disable",
		"metadata-filter-envvars", "metadata-show", "package", "package-name",
		"service-id", "service-name", "status-check-code", "status-check-off",
		"status-check-path", "status-check-timeout", "version", "timeout",
	}
}
