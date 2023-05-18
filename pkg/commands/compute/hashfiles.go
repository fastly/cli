package compute

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/kennygrant/sanitize"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

const (
	maxPackageSize = 100000000 // 100MB in bytes
)

// HashFilesCommand produces a deployable artifact from files on the local disk.
type HashFilesCommand struct {
	cmd.Base

	buildCmd  *BuildCommand
	Manifest  manifest.Data
	Package   string
	SkipBuild bool
}

// NewHashFilesCommand returns a usable command registered under the parent.
func NewHashFilesCommand(parent cmd.Registerer, g *global.Data, build *BuildCommand, m manifest.Data) *HashFilesCommand {
	var c HashFilesCommand
	c.buildCmd = build
	c.Globals = g
	c.Manifest = m
	c.CmdClause = parent.Command("hash-files", "Generate a SHA512 digest from the contents of the Compute@Edge package")
	c.CmdClause.Flag("package", "Path to a package tar.gz").Short('p').StringVar(&c.Package)
	c.CmdClause.Flag("skip-build", "Skip the build step").BoolVar(&c.SkipBuild)
	return &c
}

// Exec implements the command interface.
func (c *HashFilesCommand) Exec(in io.Reader, out io.Writer) (err error) {
	if !c.SkipBuild {
		err = c.Build(in, out)
		if err != nil {
			return err
		}
	}

	pkgName := fmt.Sprintf("%s.tar.gz", sanitize.BaseName(c.Globals.Manifest.File.Name))
	pkg := filepath.Join("pkg", pkgName)

	if c.Package != "" {
		pkg, err = filepath.Abs(c.Package)
		if err != nil {
			return fmt.Errorf("failed to locate package path '%s': %w", c.Package, err)
		}
	}

	var r io.Reader
	r, err = os.Open(pkg)
	if err != nil {
		return fmt.Errorf("failed to open package '%s': %w", pkg, err)
	}

	zr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to create a gzip reader: %w", err)
	}

	files, err := c.ReadFilesFromPackage(tar.NewReader(zr))
	if err != nil {
		return fmt.Errorf("failed to read files within the package: %w", err)
	}

	hash, err := c.GetFilesHash(files)
	if err != nil {
		return fmt.Errorf("failed to generate hash from package files: %w", err)
	}

	text.Output(out, hash)
	return nil
}

// Build constructs and executes the build logic.
func (c *HashFilesCommand) Build(in io.Reader, out io.Writer) error {
	output := out
	if !c.Globals.Verbose() {
		output = io.Discard
	}
	return c.buildCmd.Exec(in, output)
}

func (c *HashFilesCommand) ReadFilesFromPackage(tr *tar.Reader) (map[string]*bytes.Buffer, error) {
	// Store the content of every file within the package.
	contents := make(map[string]*bytes.Buffer)

	// Track overall package size.
	var pkgSize int64

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Avoids G110: Potential DoS vulnerability via decompression bomb (gosec).
		pkgSize += hdr.Size
		if pkgSize > maxPackageSize {
			return nil, errors.New("package size exceeded 100MB limit")
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		contents[hdr.Name] = &bytes.Buffer{}

		_, err = io.CopyN(contents[hdr.Name], tr, hdr.Size)
		if err != nil {
			return nil, err
		}
	}

	return contents, nil
}

func (c *HashFilesCommand) GetFilesHash(contents map[string]*bytes.Buffer) (string, error) {
	keys := make([]string, 0, len(contents))
	for k := range contents {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha512.New()
	for _, fname := range keys {
		if _, err := io.Copy(h, contents[fname]); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
