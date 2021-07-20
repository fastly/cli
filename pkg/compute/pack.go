package compute

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
	"github.com/mholt/archiver/v3"
)

// PackCommand takes a .wasm and builds the required tar/gzip package ready to be uploaded.
type PackCommand struct {
	cmd.Base
	manifest manifest.Data
	path     string
}

// NewPackCommand returns a usable command registered under the parent.
func NewPackCommand(parent cmd.Registerer, globals *config.Data) *PackCommand {
	var c PackCommand
	c.Globals = globals

	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)

	c.CmdClause = parent.Command("pack", "Package a pre-compiled Wasm binary for a Fastly Compute@Edge service")
	c.CmdClause.Flag("path", "Path to a pre-compiled Wasm binary").Short('p').Required().StringVar(&c.path)

	return &c
}

// Exec implements the command interface.
func (c *PackCommand) Exec(in io.Reader, out io.Writer) (err error) {
	var progress text.Progress
	if c.Globals.Verbose() {
		progress = text.NewVerboseProgress(out)
	} else {
		progress = text.NewQuietProgress(out)
	}

	defer func(errLog errors.LogInterface) {
		if err != nil {
			errLog.Add(err)
			progress.Fail()
		}
	}(c.Globals.ErrLog)

	pkg := fmt.Sprintf("pkg/%s/bin/main.wasm", c.manifest.File.Name)
	err = filesystem.MakeDirectoryIfNotExists(filepath.Dir(pkg))
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	src, err := filepath.Abs(c.path)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	dst, err := filepath.Abs(pkg)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	progress.Step("Copying wasm binary...")
	if err := filesystem.CopyFile(src, dst); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error copying wasm binary to '%s': %w", dst, err)
	}

	if !filesystem.FileExists(pkg) {
		return errors.RemediationError{
			Inner:       fmt.Errorf("no wasm binary found"),
			Remediation: "Run `fastly compute pack --path </path/to/wasm/binary>` to copy your wasm binary to the required location",
		}
	}

	progress.Step("Copying manifest...")
	src = manifest.Filename
	dst = fmt.Sprintf("pkg/%s/%s", c.manifest.File.Name, manifest.Filename)
	if err := filesystem.CopyFile(src, dst); err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error copying manifest to '%s': %w", dst, err)
	}

	progress.Step("Creating .tar.gz file...")
	tar := archiver.NewTarGz()
	tar.OverwriteExisting = true
	{
		dir := fmt.Sprintf("pkg/%s", c.manifest.File.Name)
		src := []string{dir}
		dst := fmt.Sprintf("%s.tar.gz", dir)
		if err = tar.Archive(src, dst); err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
	}

	progress.Done()
	return nil
}
