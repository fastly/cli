package compute

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"
)

// PackCommand takes a .wasm and builds the required tar/gzip package ready to be uploaded.
type PackCommand struct {
	cmd.Base
	manifest   manifest.Data
	wasmBinary string
}

// NewPackCommand returns a usable command registered under the parent.
func NewPackCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *PackCommand {
	var c PackCommand
	c.Globals = globals
	c.manifest = data

	c.CmdClause = parent.Command("pack", "Package a pre-compiled Wasm binary for a Fastly Compute@Edge service")
	c.CmdClause.Flag("wasm-binary", "Path to a pre-compiled Wasm binary").Short('w').Required().StringVar(&c.wasmBinary)

	return &c
}

// Exec implements the command interface.
func (c *PackCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	progress := text.NewProgress(out, c.Globals.Verbose())

	defer func(errLog fsterr.LogInterface) {
		if err != nil {
			errLog.Add(err)
			progress.Fail()
		}
	}(c.Globals.ErrLog)

	if err = c.manifest.File.ReadError(); err != nil {
		return err
	}
	name := sanitize.BaseName(c.manifest.File.Name)
	pkg := fmt.Sprintf("pkg/%s/bin/main.wasm", name)
	dir := filepath.Dir(pkg)
	err = filesystem.MakeDirectoryIfNotExists(dir)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Wasm directory (relative)": dir,
		})
		return err
	}

	src, err := filepath.Abs(c.wasmBinary)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path (absolute)": src,
		})
		return err
	}
	dst, err := filepath.Abs(pkg)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Wasm destination (relative)": pkg,
		})
		return err
	}
	progress.Step("Copying wasm binary...")
	if err := filesystem.CopyFile(src, dst); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path (absolute)":             src,
			"Wasm destination (absolute)": dst,
		})
		return fmt.Errorf("error copying wasm binary to '%s': %w", dst, err)
	}

	if !filesystem.FileExists(pkg) {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("no wasm binary found"),
			Remediation: "Run `fastly compute pack --path </path/to/wasm/binary>` to copy your wasm binary to the required location",
		}
	}

	progress.Step("Copying manifest...")
	src = manifest.Filename
	dst = fmt.Sprintf("pkg/%s/%s", name, manifest.Filename)
	if err := filesystem.CopyFile(src, dst); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Manifest (destination)": dst,
			"Manifest (source)":      src,
		})
		return fmt.Errorf("error copying manifest to '%s': %w", dst, err)
	}

	progress.Step("Creating .tar.gz file...")
	tar := archiver.NewTarGz()
	tar.OverwriteExisting = true
	{
		dir := fmt.Sprintf("pkg/%s", name)
		src := []string{dir}
		dst := fmt.Sprintf("%s.tar.gz", dir)
		if err = tar.Archive(src, dst); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Tar source":      dir,
				"Tar destination": dst,
			})
			return err
		}
	}

	progress.Done()
	return nil
}
