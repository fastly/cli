package compute

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/mholt/archiver/v3"
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
			fsterr.AllowInstrumentation: true,
			"Wasm directory (relative)": bindir,
		})
		return err
	}

	src, err := filepath.Abs(c.wasmBinary)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			fsterr.AllowInstrumentation: true,
			"Path (absolute)":           src,
		})
		return err
	}
	dst, err := filepath.Abs(bin)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			fsterr.AllowInstrumentation:   true,
			"Wasm destination (relative)": bin,
		})
		return err
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	msg := "Copying wasm binary"
	spinner.Message(msg + "...")

	if err := filesystem.CopyFile(src, dst); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			fsterr.AllowInstrumentation:   true,
			"Path (absolute)":             src,
			"Wasm destination (absolute)": dst,
		})

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}

		return fmt.Errorf("error copying wasm binary to '%s': %w", dst, err)
	}

	if !filesystem.FileExists(bin) {
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}

		return fsterr.RemediationError{
			Inner:       fmt.Errorf("no wasm binary found"),
			Remediation: "Run `fastly compute pack --path </path/to/wasm/binary>` to copy your wasm binary to the required location",
		}
	}

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return err
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	msg = "Copying manifest"
	spinner.Message(msg + "...")

	src = manifest.Filename
	dst = fmt.Sprintf("pkg/package/%s", manifest.Filename)
	if err := filesystem.CopyFile(src, dst); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			fsterr.AllowInstrumentation: true,
			"Manifest (destination)":    dst,
			"Manifest (source)":         src,
		})

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return spinErr
		}

		return fmt.Errorf("error copying manifest to '%s': %w", dst, err)
	}

	spinner.StopMessage(msg)
	err = spinner.Stop()
	if err != nil {
		return err
	}

	err = spinner.Start()
	if err != nil {
		return err
	}
	msg = "Creating package.tar.gz file"
	spinner.Message(msg + "...")

	tar := archiver.NewTarGz()
	tar.OverwriteExisting = true
	{
		dir := "pkg/package"
		src := []string{dir}
		dst := fmt.Sprintf("%s.tar.gz", dir)
		if err = tar.Archive(src, dst); err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				fsterr.AllowInstrumentation: true,
				"Tar source":                dir,
				"Tar destination":           dst,
			})

			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return spinErr
			}

			return err
		}
	}

	spinner.StopMessage(msg)
	return spinner.Stop()
}
