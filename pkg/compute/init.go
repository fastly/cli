package compute

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/text"
	"gopkg.in/src-d/go-git.v4"
)

const defaultTemplate = "https://github.com/fastly/fastly-template-rust-default"

var (
	fastlyOrgRegEx            = regexp.MustCompile(`github\.com\/fastly`)
	fastlyFileIgnoreListRegEx = regexp.MustCompile(`\.github|LICENSE|SECURITY\.md`)
)

// InitCommand initializes a Compute@Edge project package on the local machine.
type InitCommand struct {
	common.Base
	name      string
	from      string
	path      string
	serviceID string
}

// NewInitCommand returns a usable command registered under the parent.
func NewInitCommand(parent common.Registerer, globals *config.Data) *InitCommand {
	var c InitCommand
	c.Globals = globals
	c.CmdClause = parent.Command("init", "Initialize a new Compute@Edge package locally")
	c.CmdClause.Flag("name", "Name of package, defaulting to directory name of the --path destination").Short('n').StringVar(&c.name)
	c.CmdClause.Flag("from", "Git repository containing package template").Short('f').Default(defaultTemplate).StringVar(&c.from)
	c.CmdClause.Flag("path", "Destination to write the new package, defaulting to the current directory").Short('p').StringVar(&c.path)
	c.CmdClause.Flag("service-id", "Optional Fastly service ID written to the package manifest, where this package will be deployed").Short('s').StringVar(&c.serviceID)
	return &c
}

// Exec implements the command interface.
func (c *InitCommand) Exec(in io.Reader, out io.Writer) (err error) {
	var progress text.Progress
	if c.Globals.Verbose() {
		progress = text.NewVerboseProgress(out)
	} else {
		progress = text.NewQuietProgress(out)
	}

	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
		}
	}()

	if c.path == "" {
		fmt.Fprintf(progress, "--path not specified, using current directory\n")
		path, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error determining current directory: %w", err)
		}
		c.path = path
	}

	abspath, err := verifyDestination(c.path, progress)
	if err != nil {
		return err
	}
	c.path = abspath

	if c.name == "" {
		name := filepath.Base(c.path)
		fmt.Fprintf(progress, "--name not specified, using %s\n", name)
		c.name = name
	}

	tempdir, err := tempDir("package-init")
	if err != nil {
		return fmt.Errorf("error creating temporary path for package template: %w", err)
	}
	defer os.RemoveAll(tempdir)

	progress.Step("Fetching package template...")

	if _, err := git.PlainClone(tempdir, false, &git.CloneOptions{
		URL:      c.from,
		Depth:    1,
		Progress: progress,
	}); err != nil {
		return fmt.Errorf("error fetching package template: %w", err)
	}

	if err := os.RemoveAll(filepath.Join(tempdir, ".git")); err != nil {
		return fmt.Errorf("error removing git metadata from package template: %w", err)
	}

	if err := filepath.Walk(tempdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // abort
		}
		if info.IsDir() {
			return nil // descend
		}
		rel, err := filepath.Rel(tempdir, path)
		if err != nil {
			return err
		}
		// Filter any files we want to ignore in Fastly-owned templates.
		if fastlyOrgRegEx.MatchString(c.from) && fastlyFileIgnoreListRegEx.MatchString(rel) {
			return nil
		}
		dst := filepath.Join(c.path, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
			return err
		}
		if err := common.CopyFile(path, dst); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error copying files from package template: %w", err)
	}

	progress.Step("Updating package manifest...")

	var m manifest.File
	if err := m.Read(filepath.Join(c.path, ManifestFilename)); err != nil {
		return fmt.Errorf("error reading package manifest: %w", err)
	}

	if c.name != "" {
		fmt.Fprintf(progress, "Setting package name in manifest to %q...\n", c.name)
		m.Name = c.name
	}

	if c.serviceID != "" {
		fmt.Fprintf(progress, "Setting service ID in manifest to %q...\n", c.serviceID)
		m.ServiceID = c.serviceID
	}

	if err := m.Write(filepath.Join(c.path, ManifestFilename)); err != nil {
		return fmt.Errorf("error saving package manifest: %w", err)
	}

	progress.Done()
	text.Success(out, "Initialized package %s to %s", m.Name, abspath)
	return nil
}

func verifyDestination(path string, verbose io.Writer) (abspath string, err error) {
	abspath, err = filepath.Abs(path)
	if err != nil {
		return abspath, err
	}

	fi, err := os.Stat(abspath)
	if err != nil && !os.IsNotExist(err) {
		return abspath, fmt.Errorf("couldn't verify package directory: %w", err) // generic error
	}
	if err == nil && !fi.IsDir() {
		return abspath, fmt.Errorf("package destination is not a directory") // specific problem
	}
	if err != nil && os.IsNotExist(err) { // normal-ish case
		fmt.Fprintf(verbose, "Creating %s...\n", abspath)
		if err := os.MkdirAll(abspath, 0700); err != nil {
			return abspath, fmt.Errorf("error creating package destination: %w", err)
		}
	}

	if _, err := os.Stat(filepath.Join(abspath, ".git")); err == nil {
		return abspath, fmt.Errorf("package destination already contains git metadata")
	}

	if _, err := os.Stat(filepath.Join(abspath, ManifestFilename)); err == nil {
		return abspath, fmt.Errorf("package destination already contains a package manifest")
	}

	tmpname := make([]byte, 16)
	n, err := rand.Read(tmpname)
	if err != nil {
		return abspath, fmt.Errorf("error generating random filename: %w", err)
	}
	if n != 16 {
		return abspath, fmt.Errorf("failed to generate enough entropy (%d/%d)", n, 16)
	}

	f, err := os.Create(filepath.Join(abspath, fmt.Sprintf("tmp_%x", tmpname)))
	if err != nil {
		return abspath, fmt.Errorf("error creating file in package destination: %w", err)
	}

	if err := f.Close(); err != nil {
		return abspath, fmt.Errorf("error closing file in package destination: %w", err)
	}

	if err := os.Remove(f.Name()); err != nil {
		return abspath, fmt.Errorf("error removing file in package destination: %w", err)
	}

	return abspath, nil
}

func tempDir(prefix string) (abspath string, err error) {
	abspath, err = filepath.Abs(filepath.Join(
		os.TempDir(),
		fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()),
	))
	if err != nil {
		return "", err
	}

	if err = os.MkdirAll(abspath, 0750); err != nil {
		return "", err
	}

	return abspath, nil
}
