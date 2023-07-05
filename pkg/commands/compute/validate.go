package compute

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"

	"github.com/fastly/cli/pkg/cmd"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// NewValidateCommand returns a usable command registered under the parent.
func NewValidateCommand(parent cmd.Registerer, g *global.Data, m manifest.Data) *ValidateCommand {
	var c ValidateCommand
	c.Globals = g
	c.manifest = m
	c.CmdClause = parent.Command("validate", "Validate a Compute@Edge package")
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.path)
	return &c
}

// Exec implements the command interface.
func (c *ValidateCommand) Exec(_ io.Reader, out io.Writer) error {
	packagePath := c.path
	if packagePath == "" {
		projectName, source := c.manifest.Name()
		if source == manifest.SourceUndefined {
			return fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to read project name: %w", fsterr.ErrReadingManifest),
				Remediation: "Run `fastly compute build` to produce a Compute@Edge package, alternatively use the --package flag to reference a package outside of the current project.",
			}
		}
		packagePath = filepath.Join("pkg", fmt.Sprintf("%s.tar.gz", sanitize.BaseName(projectName)))
	}

	p, err := filepath.Abs(packagePath)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path": c.path,
		})
		return fmt.Errorf("error reading file path: %w", err)
	}

	if err := validatePackageContent(p); err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Path": c.path,
		})
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("failed to validate package: %w", err),
			Remediation: "Run `fastly compute build` to produce a Compute@Edge package, alternatively use the --package flag to reference a package outside of the current project.",
		}
	}

	text.Success(out, "Validated package %s", p)
	return nil
}

// ValidateCommand validates a package archive.
type ValidateCommand struct {
	cmd.Base
	manifest manifest.Data
	path     string
}

// validatePackageContent is a utility function to determine whether a package
// is valid. It walks through the package files checking the filename against a
// list of required files. If one of the files doesn't exist it returns an error.
//
// NOTE: This function is also called by the `deploy` command.
func validatePackageContent(pkgPath string) error {
	files := map[string]bool{
		"fastly.toml": false,
		"main.wasm":   false,
	}

	if err := packageFiles(pkgPath, func(f archiver.File) error {
		for k := range files {
			if k == f.Name() {
				files[k] = true
			}
		}
		return nil
	}); err != nil {
		return err
	}

	for k, found := range files {
		if !found {
			return fmt.Errorf("error validating package: package must contain a %s file", k)
		}
	}

	return nil
}

// packageFiles is a utility function to iterate over the package content.
// It attempts to unarchive and read a tar.gz file from a specific path,
// calling fn on each file in the archive.
func packageFiles(path string, fn func(archiver.File) error) error {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("error reading package: %w", err)
	}
	defer file.Close() // #nosec G307

	tr := archiver.NewTarGz()
	err = tr.Open(file, 0)
	if err != nil {
		return fmt.Errorf("error unarchiving package: %w", err)
	}
	defer tr.Close()

	// Track overall package size
	var pkgSize int64

	for {
		f, err := tr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading package: %w", err)
		}

		// Avoids G110: Potential DoS vulnerability via decompression bomb (gosec).
		pkgSize += f.Size()
		if pkgSize > MaxPackageSize {
			f.Close()
			return fmt.Errorf("package size exceeded 100MB limit")
		}

		header, ok := f.Header.(*tar.Header)
		if !ok || header.Typeflag != tar.TypeReg {
			f.Close()
			continue
		}

		if err = fn(f); err != nil {
			f.Close()
			return err
		}

		err = f.Close()
		if err != nil {
			return fmt.Errorf("error closing file: %w", err)
		}
	}

	return nil
}
