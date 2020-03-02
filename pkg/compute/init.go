package compute

import (
	"crypto/rand"
	"fmt"
	"io"
	mathRand "math/rand"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dustinkirkland/golang-petname"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
	"gopkg.in/src-d/go-git.v4"
)

type template struct {
	Name string
	Path string
}

const (
	defaultTemplate       = "https://github.com/fastly/fastly-template-rust-default.git"
	defaultTopLevelDomain = "edgecompute.app"
	manageServiceBaseURL  = "https://manage.fastly.com/configure/services/"
)

var (
	gitRepositoryRegEx        = regexp.MustCompile(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`)
	domainNameRegEx           = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)
	fastlyOrgRegEx            = regexp.MustCompile(`^https:\/\/github\.com\/fastly`)
	fastlyFileIgnoreListRegEx = regexp.MustCompile(`\.github|LICENSE|SECURITY\.md`)
	defaultTemplates          = map[int]template{
		1: {
			Name: "Starter kit",
			Path: defaultTemplate,
		},
	}
)

// InitCommand initializes a Compute@Edge project package on the local machine.
type InitCommand struct {
	common.Base
	name        string
	description string
	author      string
	from        string
	path        string
	domain      string
	backend     string
}

// NewInitCommand returns a usable command registered under the parent.
func NewInitCommand(parent common.Registerer, globals *config.Data) *InitCommand {
	var c InitCommand
	c.Globals = globals
	c.CmdClause = parent.Command("init", "Initialize a new Compute@Edge package locally")
	c.CmdClause.Flag("name", "Name of package, defaulting to directory name of the --path destination").Short('n').StringVar(&c.name)
	c.CmdClause.Flag("description", "Description of the package").Short('d').StringVar(&c.description)
	c.CmdClause.Flag("author", "Author of the package").Short('a').StringVar(&c.author)
	c.CmdClause.Flag("from", "Git repository containing package template").Short('f').StringVar(&c.from)
	c.CmdClause.Flag("path", "Destination to write the new package, defaulting to the current directory").Short('p').StringVar(&c.path)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").StringVar(&c.path)
	c.CmdClause.Flag("backend", "A hostname, IPv4, or IPv6 address for the package backend").StringVar(&c.path)

	return &c
}

// Exec implements the command interface.
func (c *InitCommand) Exec(in io.Reader, out io.Writer) (err error) {
	// Exit early if no token configured.
	_, source := c.Globals.Token()
	if source == config.SourceUndefined {
		return errors.ErrNoToken
	}

	text.Output(out, "This utility will walk you through creating a Compute@Edge project. It only covers the most common items, and tries to guess sensible defaults.")
	text.Break(out)
	text.Output(out, "Press ^C at any time to quit.")
	text.Break(out)

	var progress text.Progress
	if c.Globals.Verbose() {
		progress = text.NewVerboseProgress(out)
	} else {
		// Use a null progress writer whilst gathering input.
		progress = text.NewNullProgress()
	}

	undoStack := common.NewUndoStack()

	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
			// Unwind the undo stack
			for undoStack.Len() != 0 {
				if err := undoStack.Pop()(); err != nil {
					// TODO(phamann): What should we do with the error?!
					break
				}
			}
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
		c.name = filepath.Base(c.path)
		fmt.Fprintf(progress, "--name not specified, using %s\n\n", c.name)

		name, err := text.Input(out, fmt.Sprintf("Name: [%s] ", c.name), in)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
		if name != "" {
			c.name = name
		}
	}

	if c.description == "" {
		c.description, err = text.Input(out, "Description: ", in)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
	}

	if c.author == "" {
		var defaultEmail string
		if email := c.Globals.File.Email; email != "" {
			defaultEmail = fmt.Sprintf(" [%s]", email)
		}

		c.author, err = text.Input(out, fmt.Sprintf("Author:%s ", defaultEmail), in)
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		if c.author == "" {
			c.author = defaultEmail
		}
	}

	if c.from == "" {
		text.Output(out, "%s", text.Bold("Template:"))
		for i, template := range defaultTemplates {
			text.Output(out, "[%d] %s", i, template.Name)
		}
		template, err := text.Input(out, "Choose option or type URL: [1] ", in, validateTemplateOptionOrURL)
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		if template == "" {
			template = "1"
		}
		if i, err := strconv.Atoi(template); err == nil {
			template = defaultTemplates[i].Path
		}
		c.from = template
	}

	if c.domain == "" {
		mathRand.Seed(time.Now().UnixNano())
		defaultDomain := fmt.Sprintf("%s.%s", petname.Generate(3, "-"), defaultTopLevelDomain)
		c.domain, err = text.Input(out, fmt.Sprintf("Domain: [%s] ", defaultDomain), in, validateDomain)
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		if c.domain == "" {
			c.domain = defaultDomain
		}
	}

	if c.backend == "" {
		c.backend, err = text.Input(out, "Backend (originless, hostname or IP address): [originless] ", in, validateBackend)
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		if c.backend == "" || c.backend == "originless" {
			c.backend = "127.0.0.1"
		}
	}

	text.Break(out)

	if !c.Globals.Verbose() {
		progress = text.NewQuietProgress(out)
	}

	progress.Step("Creating service...")
	service, err := c.Globals.Client.CreateService(&fastly.CreateServiceInput{
		Name:    c.name,
		Type:    "wasm",
		Comment: c.description,
	})
	if err != nil {
		return fmt.Errorf("error creating service: %w", err)
	}
	undoStack.Push(func() error {
		return c.Globals.Client.DeleteService(&fastly.DeleteServiceInput{
			ID: service.ID,
		})
	})

	progress.Step("Creating domain...")
	_, err = c.Globals.Client.CreateDomain(&fastly.CreateDomainInput{
		Service: service.ID,
		Version: 1,
		Name:    c.domain,
	})
	if err != nil {
		return fmt.Errorf("error creating domain: %w", err)
	}
	undoStack.Push(func() error {
		return c.Globals.Client.DeleteDomain(&fastly.DeleteDomainInput{
			Service: service.ID,
			Version: 1,
			Name:    c.domain,
		})
	})

	progress.Step("Creating backend...")
	_, err = c.Globals.Client.CreateBackend(&fastly.CreateBackendInput{
		Service: service.ID,
		Version: 1,
		Name:    c.backend,
		Address: c.backend,
	})
	if err != nil {
		return fmt.Errorf("error creating backend: %w", err)
	}
	undoStack.Push(func() error {
		return c.Globals.Client.DeleteBackend(&fastly.DeleteBackendInput{
			Service: service.ID,
			Version: 1,
			Name:    c.backend,
		})
	})

	progress.Step("Fetching package template...")
	tempdir, err := tempDir("package-init")
	if err != nil {
		return fmt.Errorf("error creating temporary path for package template: %w", err)
	}
	defer os.RemoveAll(tempdir)

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

	fmt.Fprintf(progress, "Setting package name in manifest to %q...\n", c.name)
	m.Name = c.name

	if c.description != "" {
		fmt.Fprintf(progress, "Setting description in manifest to %s...\n", c.description)
		m.Description = c.description
	}

	if c.author != "" {
		fmt.Fprintf(progress, "Setting author in manifest to %s...\n", c.author)
		m.Authors = []string{c.author}
	}

	fmt.Fprintf(progress, "Setting service ID in manifest to %q...\n", service.ID)
	m.ServiceID = service.ID

	fmt.Fprintf(progress, "Setting version in manifest to 1...\n")
	m.Version = 1

	if err := m.Write(filepath.Join(c.path, ManifestFilename)); err != nil {
		return fmt.Errorf("error saving package manifest: %w", err)
	}

	progress.Done()

	text.Break(out)

	fmt.Fprintf(out, "Initialized package %s to:\n\t%s\n\n", text.Bold(m.Name), text.Bold(abspath))
	fmt.Fprintf(out, "Manage this service at:\n\t%s\n\n", text.Bold(fmt.Sprintf("%s%s", manageServiceBaseURL, service.ID)))
	fmt.Fprintf(out, "To compile the package, run:\n\t%s\n\n", text.Bold("fastly compute build"))
	fmt.Fprintf(out, "To deploy the package, run:\n\t%s\n\n", text.Bold("fastly compute deploy"))

	text.Success(out, "Initialized service %s", service.ID)
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

func validateTemplateOptionOrURL(input string) error {
	msg := "must be a valid option or Git URL"
	if input == "" {
		return nil
	}
	if option, err := strconv.Atoi(input); err == nil {
		if _, ok := defaultTemplates[option]; !ok {
			return fmt.Errorf(msg)
		}
		return nil
	}
	if !gitRepositoryRegEx.MatchString(input) {
		return fmt.Errorf(msg)
	}
	return nil
}

func validateBackend(input string) error {
	var isHost bool
	if _, err := net.LookupHost(input); err == nil {
		isHost = true
	}
	var isAddr bool
	if _, err := net.LookupAddr(input); err == nil {
		isHost = true
	}
	isEmpty := input == ""
	isOriginless := strings.ToLower(input) == "originless"
	if !isEmpty && !isOriginless && !isHost && !isAddr {
		return fmt.Errorf(`must be "originless" or a valid hostname, IPv4, or IPv6 address`)
	}
	return nil
}

func validateDomain(input string) error {
	if input == "" {
		return nil
	}
	if !domainNameRegEx.MatchString(input) {
		return fmt.Errorf("must be valid domain name")
	}
	return nil
}
