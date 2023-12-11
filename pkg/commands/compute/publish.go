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
	dir                   argparser.OptionalString
	includeSrc            argparser.OptionalBool
	lang                  argparser.OptionalString
	metadataDisable       argparser.OptionalBool
	metadataFilterEnvVars argparser.OptionalString
	metadataShow          argparser.OptionalBool
	packageName           argparser.OptionalString
	timeout               argparser.OptionalInt

	// Deploy fields
	comment            argparser.OptionalString
	domain             argparser.OptionalString
	env                argparser.OptionalString
	pkg                argparser.OptionalString
	serviceName        argparser.OptionalServiceNameID
	serviceVersion     argparser.OptionalServiceVersion
	statusCheckCode    int
	statusCheckOff     bool
	statusCheckPath    string
	statusCheckTimeout int
}

// NewPublishCommand returns a usable command registered under the parent.
func NewPublishCommand(parent argparser.Registerer, g *global.Data, build *BuildCommand, deploy *DeployCommand) *PublishCommand {
	var c PublishCommand
	c.Globals = g
	c.build = build
	c.deploy = deploy
	c.CmdClause = parent.Command("publish", "Build and deploy a Compute package to a Fastly service")

	c.CmdClause.Flag("comment", "Human-readable comment").Action(c.comment.Set).StringVar(&c.comment.Value)
	c.CmdClause.Flag("dir", "Project directory to build (default: current directory)").Short('C').Action(c.dir.Set).StringVar(&c.dir.Value)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").Action(c.domain.Set).StringVar(&c.domain.Value)
	c.CmdClause.Flag("env", "The manifest environment config to use (e.g. 'stage' will attempt to read 'fastly.stage.toml')").Action(c.env.Set).StringVar(&c.env.Value)
	c.CmdClause.Flag("include-source", "Include source code in built package").Action(c.includeSrc.Set).BoolVar(&c.includeSrc.Value)
	c.CmdClause.Flag("language", "Language type").Action(c.lang.Set).StringVar(&c.lang.Value)
	c.CmdClause.Flag("metadata-disable", "Disable Wasm binary metadata annotations").Action(c.metadataDisable.Set).BoolVar(&c.metadataDisable.Value)
	c.CmdClause.Flag("metadata-filter-envvars", "Redact specified environment variables from [scripts.env_vars] using comma-separated list").Action(c.metadataFilterEnvVars.Set).StringVar(&c.metadataFilterEnvVars.Value)
	c.CmdClause.Flag("metadata-show", "Inspect the Wasm binary metadata").Action(c.metadataShow.Set).BoolVar(&c.metadataShow.Value)
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').Action(c.pkg.Set).StringVar(&c.pkg.Value)
	c.CmdClause.Flag("package-name", "Package name").Action(c.packageName.Set).StringVar(&c.packageName.Value)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagServiceIDName,
		Description: argparser.FlagServiceIDDesc,
		Dst:         &c.Globals.Manifest.Flag.ServiceID,
		Short:       's',
	})
	c.RegisterFlag(argparser.StringFlagOpts{
		Action:      c.serviceName.Set,
		Name:        argparser.FlagServiceName,
		Description: argparser.FlagServiceDesc,
		Dst:         &c.serviceName.Value,
	})
	c.CmdClause.Flag("status-check-code", "Set the expected status response for the service availability check to the root path").IntVar(&c.statusCheckCode)
	c.CmdClause.Flag("status-check-off", "Disable the service availability check").BoolVar(&c.statusCheckOff)
	c.CmdClause.Flag("status-check-path", "Specify the URL path for the service availability check").Default("/").StringVar(&c.statusCheckPath)
	c.CmdClause.Flag("status-check-timeout", "Set a timeout (in seconds) for the service availability check").Default("120").IntVar(&c.statusCheckTimeout)
	c.RegisterFlag(argparser.StringFlagOpts{
		Name:        argparser.FlagVersionName,
		Description: argparser.FlagVersionDesc,
		Dst:         &c.serviceVersion.Value,
		Action:      c.serviceVersion.Set,
	})
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
	if c.dir.WasSet {
		c.build.Flags.Dir = c.dir.Value
	}
	if c.env.WasSet {
		c.build.Flags.Env = c.env.Value
	}
	if c.includeSrc.WasSet {
		c.build.Flags.IncludeSrc = c.includeSrc.Value
	}
	if c.lang.WasSet {
		c.build.Flags.Lang = c.lang.Value
	}
	if c.packageName.WasSet {
		c.build.Flags.PackageName = c.packageName.Value
	}
	if c.timeout.WasSet {
		c.build.Flags.Timeout = c.timeout.Value
	}
	if c.metadataDisable.WasSet {
		c.build.MetadataDisable = c.metadataDisable.Value
	}
	if c.metadataFilterEnvVars.WasSet {
		c.build.MetadataFilterEnvVars = c.metadataFilterEnvVars.Value
	}
	if c.metadataShow.WasSet {
		c.build.MetadataShow = c.metadataShow.Value
	}

	err = c.build.Exec(in, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	text.Break(out)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	defer func() {
		_ = os.Chdir(wd)
	}()

	projectDir, err := ChangeProjectDirectory(c.dir.Value)
	if err != nil {
		return err
	}
	if projectDir != "" {
		if c.Globals.Verbose() {
			text.Info(out, ProjectDirMsg, projectDir)
		}
	}

	// Reset the fields on the DeployCommand based on PublishCommand values.
	if c.dir.WasSet {
		c.deploy.Dir = c.dir.Value
	}
	if c.pkg.WasSet {
		c.deploy.PackagePath = c.pkg.Value
	}
	if c.serviceName.WasSet {
		c.deploy.ServiceName = c.serviceName // deploy's field is a argparser.OptionalServiceNameID
	}
	if c.serviceVersion.WasSet {
		c.deploy.ServiceVersion = c.serviceVersion // deploy's field is a argparser.OptionalServiceVersion
	}
	if c.domain.WasSet {
		c.deploy.Domain = c.domain.Value
	}
	if c.env.WasSet {
		c.deploy.Env = c.env.Value
	}
	if c.comment.WasSet {
		c.deploy.Comment = c.comment
	}
	if c.statusCheckCode > 0 {
		c.deploy.StatusCheckCode = c.statusCheckCode
	}
	if c.statusCheckOff {
		c.deploy.StatusCheckOff = c.statusCheckOff
	}
	if c.statusCheckTimeout > 0 {
		c.deploy.StatusCheckTimeout = c.statusCheckTimeout
	}
	c.deploy.StatusCheckPath = c.statusCheckPath

	err = c.deploy.Exec(in, out)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	return nil
}
