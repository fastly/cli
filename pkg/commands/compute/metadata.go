package compute

import (
	"fmt"
	"io"

	"github.com/fastly/cli/v10/pkg/argparser"
	"github.com/fastly/cli/v10/pkg/config"
	fsterr "github.com/fastly/cli/v10/pkg/errors"
	"github.com/fastly/cli/v10/pkg/global"
	"github.com/fastly/cli/v10/pkg/text"
)

// MetadataCommand controls what metadata is collected for a Wasm binary.
type MetadataCommand struct {
	argparser.Base

	disable        bool
	disableBuild   bool
	disableMachine bool
	disablePackage bool
	disableScript  bool
	enable         bool
	enableBuild    bool
	enableMachine  bool
	enablePackage  bool
	enableScript   bool
}

// NewMetadataCommand returns a new command registered in the parent.
func NewMetadataCommand(parent argparser.Registerer, g *global.Data) *MetadataCommand {
	var c MetadataCommand
	c.Globals = g
	c.CmdClause = parent.Command("metadata", "Control what metadata is collected")
	c.CmdClause.Flag("disable", "Disable all metadata").BoolVar(&c.disable)
	c.CmdClause.Flag("disable-build", "Disable metadata for information regarding the time taken for builds and compilation processes").BoolVar(&c.disableBuild)
	c.CmdClause.Flag("disable-machine", "Disable metadata for general, non-identifying system specifications (CPU, RAM, operating system)").BoolVar(&c.disableMachine)
	c.CmdClause.Flag("disable-package", "Disable metadata for packages and libraries utilized in your source code").BoolVar(&c.disablePackage)
	c.CmdClause.Flag("disable-script", "Disable metadata for script info from the fastly.toml manifest (i.e. [scripts] section).").BoolVar(&c.disableScript)
	c.CmdClause.Flag("enable", "Enable all metadata").BoolVar(&c.enable)
	c.CmdClause.Flag("enable-build", "Enable metadata for information regarding the time taken for builds and compilation processes").BoolVar(&c.enableBuild)
	c.CmdClause.Flag("enable-machine", "Enable metadata for general, non-identifying system specifications (CPU, RAM, operating system)").BoolVar(&c.enableMachine)
	c.CmdClause.Flag("enable-package", "Enable metadata for packages and libraries utilized in your source code").BoolVar(&c.enablePackage)
	c.CmdClause.Flag("enable-script", "Enable metadata for script info from the fastly.toml manifest (i.e. [scripts] section).").BoolVar(&c.enableScript)
	return &c
}

// Exec implements the command interface.
func (c *MetadataCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.disable && c.enable {
		return fsterr.ErrInvalidEnableDisableFlagCombo
	}

	var modified bool

	// Global enable/disable
	if c.enable {
		c.Globals.Config.WasmMetadata = toggleAll("enable")
		modified = true
	}
	if c.disable {
		c.Globals.Config.WasmMetadata = toggleAll("disable")
		modified = true
	}

	// Specific enablement
	if c.enableBuild {
		c.Globals.Config.WasmMetadata.BuildInfo = "enable"
		modified = true
	}
	if c.enableMachine {
		c.Globals.Config.WasmMetadata.MachineInfo = "enable"
		modified = true
	}
	if c.enablePackage {
		c.Globals.Config.WasmMetadata.PackageInfo = "enable"
		modified = true
	}
	if c.enableScript {
		c.Globals.Config.WasmMetadata.ScriptInfo = "enable"
		modified = true
	}

	// Specific disablement
	if c.disableBuild {
		c.Globals.Config.WasmMetadata.BuildInfo = "disable"
		modified = true
	}
	if c.disableMachine {
		c.Globals.Config.WasmMetadata.MachineInfo = "disable"
		modified = true
	}
	if c.disablePackage {
		c.Globals.Config.WasmMetadata.PackageInfo = "disable"
		modified = true
	}
	if c.disableScript {
		c.Globals.Config.WasmMetadata.ScriptInfo = "disable"
		modified = true
	}

	if modified {
		if c.disable && (c.enableBuild || c.enableMachine || c.enablePackage || c.enableScript) {
			text.Info(out, "We will disable all metadata except for the specified `--enable-*` flags")
			text.Break(out)
		}
		if c.enable && (c.disableBuild || c.disableMachine || c.disablePackage || c.disableScript) {
			text.Info(out, "We will enable all metadata except for the specified `--disable-*` flags")
			text.Break(out)
		}
		err := c.Globals.Config.Write(c.Globals.ConfigPath)
		if err != nil {
			return fmt.Errorf("failed to persist metadata choices to disk: %w", err)
		}
		text.Success(out, "configuration updated")
		text.Break(out)
	}

	text.Output(out, "Build Information: %s", c.Globals.Config.WasmMetadata.BuildInfo)
	text.Output(out, "Machine Information: %s", c.Globals.Config.WasmMetadata.MachineInfo)
	text.Output(out, "Package Information: %s", c.Globals.Config.WasmMetadata.PackageInfo)
	text.Output(out, "Script Information: %s", c.Globals.Config.WasmMetadata.ScriptInfo)
	return nil
}

func toggleAll(state string) config.WasmMetadata {
	var t config.WasmMetadata
	t.BuildInfo = state
	t.MachineInfo = state
	t.PackageInfo = state
	t.ScriptInfo = state
	return t
}
