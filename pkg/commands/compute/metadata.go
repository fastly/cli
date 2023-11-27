package compute

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

// MetadataCommand controls what metadata is collected for a Wasm binary.
type MetadataCommand struct {
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

// NewMetadataCommand returns a new command registered in the parent.
func NewMetadataCommand(parent cmd.Registerer, g *global.Data) *MetadataCommand {
	var c MetadataCommand
	c.Globals = g
	c.CmdClause = parent.Command("metadata", "Control what metadata is collected")
	c.CmdClause.Flag("disable", "Disable all metadata").BoolVar(&c.disable)
	c.CmdClause.Flag("disable-build", "Disable metadata for information regarding the time taken for builds and compilation processes").BoolVar(&c.disableBuild)
	c.CmdClause.Flag("disable-machine", "Disable metadata for general, non-identifying system specifications (CPU, RAM, operating system)").BoolVar(&c.disableMachine)
	c.CmdClause.Flag("disable-package", "Disable metadata for packages and libraries utilized in your source code").BoolVar(&c.disablePackage)
	c.CmdClause.Flag("enable", "Enable all metadata").BoolVar(&c.enable)
	c.CmdClause.Flag("enable-build", "Enable metadata for information regarding the time taken for builds and compilation processes").BoolVar(&c.enableBuild)
	c.CmdClause.Flag("enable-machine", "Enable metadata for general, non-identifying system specifications (CPU, RAM, operating system)").BoolVar(&c.enableMachine)
	c.CmdClause.Flag("enable-package", "Enable metadata for packages and libraries utilized in your source code").BoolVar(&c.enablePackage)
	return &c
}

// Exec implements the command interface.
func (c *MetadataCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.disable && c.enable {
		return fsterr.ErrInvalidEnableDisableFlagCombo
	}
	if c.disable {
		c.Globals.Config.WasmMetadata = toggleAll("disable")
	}
	if c.enable {
		c.Globals.Config.WasmMetadata = toggleAll("enable")
	}
	if c.disable && (c.enableBuild || c.enableMachine || c.enablePackage) {
		text.Info(out, "We will disable all metadata except for the specified `--enable-*` flags")
		text.Break(out)
	}
	if c.enable && (c.disableBuild || c.disableMachine || c.disablePackage) {
		text.Info(out, "We will enable all metadata except for the specified `--disable-*` flags")
		text.Break(out)
	}
	if c.enableBuild {
		c.Globals.Config.WasmMetadata.BuildInfo = "enable"
	}
	if c.enableMachine {
		c.Globals.Config.WasmMetadata.MachineInfo = "enable"
	}
	if c.enablePackage {
		c.Globals.Config.WasmMetadata.PackageInfo = "enable"
	}
	if c.disableBuild {
		c.Globals.Config.WasmMetadata.BuildInfo = "disable"
	}
	if c.disableMachine {
		c.Globals.Config.WasmMetadata.MachineInfo = "disable"
	}
	if c.disablePackage {
		c.Globals.Config.WasmMetadata.PackageInfo = "disable"
	}
	err := c.Globals.Config.Write(c.Globals.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to persist metadata choices to disk: %w", err)
	}
	text.Success(out, "configuration updated (see: `fastly config`)")
	text.Break(out)
	text.Output(out, "Build Information: %s", c.Globals.Config.WasmMetadata.BuildInfo)
	text.Output(out, "Machine Information: %s", c.Globals.Config.WasmMetadata.MachineInfo)
	text.Output(out, "Package Information: %s", c.Globals.Config.WasmMetadata.PackageInfo)
	return nil
}

func toggleAll(state string) config.WasmMetadata {
	var t config.WasmMetadata
	t.BuildInfo = state
	t.MachineInfo = state
	t.PackageInfo = state
	return t
}
