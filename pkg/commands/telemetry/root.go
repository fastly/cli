package telemetry

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// RootCommand is the parent command for all subcommands in this package.
// It should be installed under the primary root command.
type RootCommand struct {
	cmd.Base

	disable        bool
	disableBuild   bool
	disableMachine bool
	disablePackage bool
	enable         bool
	enableBuild    bool
	enableMachine  bool
	enablePackage  bool
}

// NewRootCommand returns a new command registered in the parent.
func NewRootCommand(parent cmd.Registerer, g *global.Data) *RootCommand {
	var c RootCommand
	c.Globals = g
	c.CmdClause = parent.Command("telemetry", "Control what telemetry data is recorded")
	c.CmdClause.Flag("disable", "Disable all telemetry").BoolVar(&c.disable)
	c.CmdClause.Flag("disable-build", "Disable telemetry for information regarding the time taken for builds and compilation processes").BoolVar(&c.disableBuild)
	c.CmdClause.Flag("disable-machine", "Disable telemetry for general, non-identifying system specifications (CPU, RAM, operating system)").BoolVar(&c.disableMachine)
	c.CmdClause.Flag("disable-package", "Disable telemetry for packages and libraries utilized in your source code").BoolVar(&c.disablePackage)
	c.CmdClause.Flag("enable", "Enable all telemetry").BoolVar(&c.enable)
	c.CmdClause.Flag("enable-build", "Enable telemetry for information regarding the time taken for builds and compilation processes").BoolVar(&c.enableBuild)
	c.CmdClause.Flag("enable-machine", "Enable telemetry for general, non-identifying system specifications (CPU, RAM, operating system)").BoolVar(&c.enableMachine)
	c.CmdClause.Flag("enable-package", "Enable telemetry for packages and libraries utilized in your source code").BoolVar(&c.enablePackage)
	return &c
}

// Exec implements the command interface.
func (c *RootCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.disable && c.enable {
		return fsterr.ErrInvalidEnableDisableFlagCombo
	}
	if c.disable {
		c.Globals.Config.Telemetry = toggleAll("disable")
	}
	if c.enable {
		c.Globals.Config.Telemetry = toggleAll("enable")
	}
	if c.disable && (c.enableBuild || c.enableMachine || c.enablePackage) {
		text.Info(out, "We will disable all telemetry except for the specified `--enable-*` flags")
		text.Break(out)
	}
	if c.enable && (c.disableBuild || c.disableMachine || c.disablePackage) {
		text.Info(out, "We will enable all telemetry except for the specified `--disable-*` flags")
		text.Break(out)
	}
	if c.enableBuild {
		c.Globals.Config.Telemetry.BuildInfo = "enable"
	}
	if c.enableMachine {
		c.Globals.Config.Telemetry.MachineInfo = "enable"
	}
	if c.enablePackage {
		c.Globals.Config.Telemetry.PackageInfo = "enable"
	}
	if c.disableBuild {
		c.Globals.Config.Telemetry.BuildInfo = "disable"
	}
	if c.disableMachine {
		c.Globals.Config.Telemetry.MachineInfo = "disable"
	}
	if c.disablePackage {
		c.Globals.Config.Telemetry.PackageInfo = "disable"
	}
	err := c.Globals.Config.Write(c.Globals.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to persist telemetry choices to disk: %w", err)
	}
	text.Success(out, "configuration updated (see: `fastly config`)")
	return nil
}

func toggleAll(state string) config.Telemetry {
	var t config.Telemetry
	t.BuildInfo = state
	t.MachineInfo = state
	t.PackageInfo = state
	return t
}
