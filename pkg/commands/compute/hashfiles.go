package compute

import (
	"archive/tar"
	"bytes"
	"crypto/sha512"
	"fmt"
	"io"
	"path/filepath"
	"sort"

	"github.com/kennygrant/sanitize"
	"github.com/mholt/archiver/v3"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// MaxPackageSize represents the max package size that can be uploaded to the
// Fastly Package API endpoint.
//
// NOTE: This is variable not a constant for the sake of test manipulations.
// https://developer.fastly.com/learning/compute/#limitations-and-constraints
var MaxPackageSize int64 = 100000000 // 100MB in bytes

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

	hash, err := getFilesHash(pkg)
	if err != nil {
		return err
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

// getFilesHash returns a hash of all the files in the package in sorted filename order.
func getFilesHash(pkgPath string) (string, error) {
	contents := make(map[string]*bytes.Buffer)
	if err := validate(pkgPath, func(f archiver.File) error {
		// This is safe to do - we already verified it in validate().
		filename := f.Header.(*tar.Header).Name
		contents[filename] = &bytes.Buffer{}
		if _, err := io.Copy(contents[filename], f); err != nil {
			return fmt.Errorf("error reading %s: %w", filename, err)
		}
		return nil
	}); err != nil {
		return "", err
	}

	keys := make([]string, 0, len(contents))
	for k := range contents {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha512.New()
	for _, fname := range keys {
		if _, err := io.Copy(h, contents[fname]); err != nil {
			return "", fmt.Errorf("failed to generate hash from package files: %w", err)
		}
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
