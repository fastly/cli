package compute

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// MetadataCommand controls what metadata is collected for a Wasm binary.
type MetadataCommand struct {
	argparser.Base

	Disable        bool `help:"Disable all metadata"`
	DisableBuild   bool `help:"Disable metadata for information regarding the time taken for builds and compilation processes"`
	DisableMachine bool `help:"Disable metadata for general, non-identifying system specifications (CPU, RAM, operating system)"`
	DisablePackage bool `help:"Disable metadata for packages and libraries utilized in your source code"`
	DisableScript  bool `help:"Disable metadata for script info from the fastly.toml manifest (i.e. [scripts] section)."`
	Enable         bool `help:"Enable all metadata"`
	EnableBuild    bool `help:"Enable metadata for information regarding the time taken for builds and compilation processes"`
	EnableMachine  bool `help:"Enable metadata for general, non-identifying system specifications (CPU, RAM, operating system)"`
	EnablePackage  bool `help:"Enable metadata for packages and libraries utilized in your source code"`
	EnableScript   bool `help:"Enable metadata for script info from the fastly.toml manifest (i.e. [scripts] section)."`
}

// Exec implements the command interface.
func (c *MetadataCommand) Exec(_ io.Reader, out io.Writer) error {
	if c.Disable && c.Enable {
		return fsterr.ErrInvalidEnableDisableFlagCombo
	}

	var modified bool

	// Global enable/disable
	if c.Enable {
		c.Globals.Config.WasmMetadata = toggleAll("enable")
		modified = true
	}
	if c.Disable {
		c.Globals.Config.WasmMetadata = toggleAll("disable")
		modified = true
	}

	// Specific enablement
	if c.EnableBuild {
		c.Globals.Config.WasmMetadata.BuildInfo = "enable"
		modified = true
	}
	if c.EnableMachine {
		c.Globals.Config.WasmMetadata.MachineInfo = "enable"
		modified = true
	}
	if c.EnablePackage {
		c.Globals.Config.WasmMetadata.PackageInfo = "enable"
		modified = true
	}
	if c.EnableScript {
		c.Globals.Config.WasmMetadata.ScriptInfo = "enable"
		modified = true
	}

	// Specific disablement
	if c.DisableBuild {
		c.Globals.Config.WasmMetadata.BuildInfo = "disable"
		modified = true
	}
	if c.DisableMachine {
		c.Globals.Config.WasmMetadata.MachineInfo = "disable"
		modified = true
	}
	if c.DisablePackage {
		c.Globals.Config.WasmMetadata.PackageInfo = "disable"
		modified = true
	}
	if c.DisableScript {
		c.Globals.Config.WasmMetadata.ScriptInfo = "disable"
		modified = true
	}

	if modified {
		if c.Disable && (c.EnableBuild || c.EnableMachine || c.EnablePackage || c.EnableScript) {
			text.Info(out, "We will disable all metadata except for the specified `--enable-*` flags")
			text.Break(out)
		}
		if c.Enable && (c.DisableBuild || c.DisableMachine || c.DisablePackage || c.DisableScript) {
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
