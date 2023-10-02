package compute

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archiver/v3"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// PackCommand takes a .wasm and builds the required tar/gzip package ready to be uploaded.
type PackCommand struct {
	cmd.Base
	manifest   manifest.Data
	wasmBinary string
}

// NewPackCommand returns a usable command registered under the parent.
func NewPackCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *PackCommand {
	var c PackCommand
	c.Globals = g
	c.manifest = m

	c.CmdClause = parent.Command("pack", "Package a pre-compiled Wasm binary for a Fastly Compute@Edge service")
	c.CmdClause.Flag("wasm-binary", "Path to a pre-compiled Wasm binary").Short('w').Required().StringVar(&c.wasmBinary)

	return &c
}

// Exec implements the command interface.
//
// NOTE: The bin/manifest is placed in a 'package' folder within the tar.gz.
func (c *PackCommand) Exec(_ io.Reader, out io.Writer) (err error) {
	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	defer func(errLog fsterr.LogInterface) {
		os.RemoveAll("pkg/package")
		if err != nil {
			errLog.Add(err)
		}
	}(c.Globals.ErrLog)

	if err = c.manifest.File.ReadError(); err != nil {
		return err
	}

	bin := "pkg/package/bin/main.wasm"
	bindir := filepath.Dir(bin)

	err = filesystem.MakeDirectoryIfNotExists(bindir)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Wasm directory (relative)": bindir,
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

	dst, err := filepath.Abs(bin)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Wasm destination (relative)": bin,
		})
		return err
	}

	err = spinner.Process("Copying wasm binary", func(_ *text.SpinnerWrapper) error {
		if err := filesystem.CopyFile(src, dst); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Path (absolute)":             src,
				"Wasm destination (absolute)": dst,
			})
			return fmt.Errorf("error copying wasm binary to '%s': %w", dst, err)
		}

		if !filesystem.FileExists(bin) {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("no wasm binary found"),
				Remediation: "Run `fastly compute pack --path </path/to/wasm/binary>` to copy your wasm binary to the required location",
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = spinner.Process("Copying manifest", func(_ *text.SpinnerWrapper) error {
		src = manifest.Filename
		dst = fmt.Sprintf("pkg/package/%s", manifest.Filename)
		if err := filesystem.CopyFile(src, dst); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"Manifest (destination)": dst,
				"Manifest (source)":      src,
			})
			return fmt.Errorf("error copying manifest to '%s': %w", dst, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return spinner.Process("Creating package.tar.gz file", func(_ *text.SpinnerWrapper) error {
		tar := archiver.NewTarGz()
		tar.OverwriteExisting = true
		{
			dir := "pkg/package"
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
		return nil
	})
}
