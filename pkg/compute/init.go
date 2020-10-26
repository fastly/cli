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
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/fastly"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

const (
	defaultTopLevelDomain = "edgecompute.app"
	manageServiceBaseURL  = "https://manage.fastly.com/configure/services/"
)

var (
	gitRepositoryRegEx        = regexp.MustCompile(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`)
	domainNameRegEx           = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)
	fastlyOrgRegEx            = regexp.MustCompile(`^https:\/\/github\.com\/fastly`)
	fastlyFileIgnoreListRegEx = regexp.MustCompile(`\.github|LICENSE|SECURITY\.md|CHANGELOG\.md|screenshot\.png`)
	starterKits               = map[string][]StarterKit{
		"assemblyscript": {
			{
				Name: "Default",
				Path: "https://github.com/fastly/compute-starter-kit-assemblyscript-default",
				Tag:  "v0.1.0",
			},
		},
		"rust": {
			{
				Name:   "Default",
				Path:   "https://github.com/fastly/fastly-template-rust-default.git",
				Branch: "0.4.0",
			},
		},
	}
)

// StarterKit models a Compute@Edge package template and its git location.
type StarterKit struct {
	Name   string
	Path   string
	Branch string
	Tag    string
}

// InitCommand initializes a Compute@Edge project package on the local machine.
type InitCommand struct {
	common.Base
	client   api.HTTPClient
	manifest manifest.Data
	language string
	from     string
	branch   string
	tag      string
	path     string
	domain   string
	backend  string
}

// NewInitCommand returns a usable command registered under the parent.
func NewInitCommand(parent common.Registerer, client api.HTTPClient, globals *config.Data) *InitCommand {
	var c InitCommand
	c.Globals = globals
	c.client = client
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("init", "Initialize a new Compute@Edge package locally")
	c.CmdClause.Flag("service-id", "Existing service ID to use. By default, this command creates a new service").Short('s').StringVar(&c.manifest.Flag.ServiceID)
	c.CmdClause.Flag("name", "Name of package, defaulting to directory name of the --path destination").Short('n').StringVar(&c.manifest.File.Name)
	c.CmdClause.Flag("description", "Description of the package").Short('d').StringVar(&c.manifest.File.Description)
	c.CmdClause.Flag("author", "Author(s) of the package").Short('a').StringsVar(&c.manifest.File.Authors)
	c.CmdClause.Flag("language", "Language of the package").Short('l').StringVar(&c.language)
	c.CmdClause.Flag("from", "Git repository containing package template").Short('f').StringVar(&c.from)
	c.CmdClause.Flag("branch", "Git branch name to clone from package template repository").Hidden().StringVar(&c.branch)
	c.CmdClause.Flag("tag", "Git tag name to clone from package template repository").Hidden().StringVar(&c.tag)
	c.CmdClause.Flag("path", "Destination to write the new package, defaulting to the current directory").Short('p').StringVar(&c.path)
	c.CmdClause.Flag("domain", "The name of the domain associated to the package").StringVar(&c.path)
	c.CmdClause.Flag("backend", "A hostname, IPv4, or IPv6 address for the package backend").StringVar(&c.path)

	return &c
}

// Exec implements the command interface.
func (c *InitCommand) Exec(in io.Reader, out io.Writer) (err error) {
	// Exit early if no token configured.
	_, s := c.Globals.Token()
	if s == config.SourceUndefined {
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
	defer func() {
		if err != nil {
			progress.Fail() // progress.Done is handled inline
		}
	}()

	undoStack := common.NewUndoStack()
	defer func() { undoStack.RunIfError(out, err) }()

	var (
		source      manifest.Source
		serviceID   string
		service     *fastly.Service
		version     int
		name        string
		description string
		authors     []string
		language    *Language
	)

	languages := []*Language{
		NewLanguage(&LanguageOptions{
			Name:        "rust",
			DisplayName: "Rust",
			StarterKits: starterKits["rust"],
			Toolchain:   NewRust(c.client),
		}),
		NewLanguage(&LanguageOptions{
			Name:        "assemblyscript",
			DisplayName: "AssemblyScript (beta)",
			StarterKits: starterKits["assemblyscript"],
			Toolchain:   NewAssemblyScript(),
		}),
	}

	name, _ = c.manifest.Name()
	description, _ = c.manifest.Description()
	authors, _ = c.manifest.Authors()

	serviceID, source = c.manifest.ServiceID()
	if source != manifest.SourceUndefined {
		service, err = c.Globals.Client.GetService(&fastly.GetServiceInput{
			ID: serviceID,
		})
		if err != nil {
			return fmt.Errorf("error fetching service details: %w", err)
		}

		name = service.Name
		description = service.Comment
	}

	if c.path == "" && !c.manifest.File.Exists() {
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

	if name == "" {
		defaultName := filepath.Base(c.path)
		name, err = text.Input(out, fmt.Sprintf("Name: [%s] ", defaultName), in)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
		if name == "" {
			name = defaultName
		}
	}

	if description == "" {
		description, err = text.Input(out, "Description: ", in)
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}
	}

	if len(authors) == 0 {
		label := "Author: "
		var defaultEmail string
		if email := c.Globals.File.Email; email != "" {
			defaultEmail = email
			label = fmt.Sprintf("%s[%s] ", label, email)
		}

		author, err := text.Input(out, label, in)
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		if author != "" {
			authors = []string{author}
		} else {
			authors = []string{defaultEmail}
		}
	}

	if c.language == "" {
		text.Output(out, "%s", text.Bold("Language:"))
		for i, lang := range languages {
			text.Output(out, "[%d] %s", i+1, lang.DisplayName)
		}
		option, err := text.Input(out, "Choose option: [1] ", in, validateLanguageOption(languages))
		if err != nil {
			return fmt.Errorf("reading input %w", err)
		}
		if option == "" {
			option = "1"
		}
		if i, err := strconv.Atoi(option); err == nil {
			language = languages[i-1]
		} else {
			return fmt.Errorf("selecting language")
		}
	} else {
		for _, l := range languages {
			if strings.EqualFold(c.language, l.Name) {
				language = l
			}
		}
	}

	if c.from == "" && !c.manifest.File.Exists() {
		text.Output(out, "%s", text.Bold("Starter kit:"))
		for i, kit := range language.StarterKits {
			text.Output(out, "[%d] %s (%s)", i+1, kit.Name, kit.Path)
		}
		option, err := text.Input(out, "Choose option or type URL: [1] ", in, validateTemplateOptionOrURL(language.StarterKits))
		if err != nil {
			return fmt.Errorf("error reading input %w", err)
		}
		if option == "" {
			option = "1"
		}

		if i, err := strconv.Atoi(option); err == nil {
			template := language.StarterKits[i-1]
			c.from = template.Path
			c.branch = template.Branch
			c.tag = template.Tag
		} else {
			c.from = option
		}
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

	// If we have an existing service, get the ideal version.
	// Otherwise create a new service and set version to 1.
	if serviceID != "" {
		versions, err := c.Globals.Client.ListVersions(&fastly.ListVersionsInput{
			Service: serviceID,
		})
		if err != nil {
			return fmt.Errorf("error listing service versions: %w", err)
		}

		v, err := getLatestIdealVersion(versions)
		if err != nil {
			return fmt.Errorf("error finding latest service version")
		}

		if v.Active || v.Locked {
			progress.Step("Cloning latest version...")
			v, err = c.Globals.Client.CloneVersion(&fastly.CloneVersionInput{
				Service: serviceID,
				Version: v.Number,
			})
			if err != nil {
				return fmt.Errorf("error cloning latest service version: %w", err)
			}
		}

		version = v.Number
	} else {
		progress.Step("Creating service...")
		service, err = c.Globals.Client.CreateService(&fastly.CreateServiceInput{
			Name:    name,
			Type:    "wasm",
			Comment: description,
		})
		if err != nil {
			return fmt.Errorf("error creating service: %w", err)
		}
		version = 1
		undoStack.Push(func() error {
			return c.Globals.Client.DeleteService(&fastly.DeleteServiceInput{
				ID: service.ID,
			})
		})
	}

	progress.Step("Creating domain...")
	_, err = c.Globals.Client.CreateDomain(&fastly.CreateDomainInput{
		Service: service.ID,
		Version: version,
		Name:    c.domain,
	})
	if err != nil {
		return fmt.Errorf("error creating domain: %w", err)
	}
	undoStack.Push(func() error {
		return c.Globals.Client.DeleteDomain(&fastly.DeleteDomainInput{
			Service: service.ID,
			Version: version,
			Name:    c.domain,
		})
	})

	progress.Step("Creating backend...")
	_, err = c.Globals.Client.CreateBackend(&fastly.CreateBackendInput{
		Service: service.ID,
		Version: version,
		Name:    c.backend,
		Address: c.backend,
	})
	if err != nil {
		return fmt.Errorf("error creating backend: %w", err)
	}
	undoStack.Push(func() error {
		return c.Globals.Client.DeleteBackend(&fastly.DeleteBackendInput{
			Service: service.ID,
			Version: version,
			Name:    c.backend,
		})
	})

	if c.from != "" && !c.manifest.File.Exists() {
		progress.Step("Fetching package template...")
		tempdir, err := tempDir("package-init")
		if err != nil {
			return fmt.Errorf("error creating temporary path for package template: %w", err)
		}
		defer os.RemoveAll(tempdir)

		if c.branch != "" && c.tag != "" {
			return fmt.Errorf("cannot use both git branch and tag name")
		}

		var ref plumbing.ReferenceName
		if c.branch != "" {
			ref = plumbing.NewBranchReferenceName(c.branch)
		}
		if c.tag != "" {
			ref = plumbing.NewTagReferenceName(c.tag)
		}

		if _, err := git.PlainClone(tempdir, false, &git.CloneOptions{
			URL:           c.from,
			ReferenceName: ref,
			Depth:         1,
			Progress:      progress,
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
	}

	progress.Step("Updating package manifest...")

	var m manifest.File
	if err := m.Read(filepath.Join(c.path, ManifestFilename)); err != nil {
		return fmt.Errorf("error reading package manifest: %w", err)
	}

	fmt.Fprintf(progress, "Setting package name in manifest to %q...\n", name)
	m.Name = name

	if description != "" {
		fmt.Fprintf(progress, "Setting description in manifest to %s...\n", description)
		m.Description = description
	}

	if len(authors) > 0 {
		fmt.Fprintf(progress, "Setting authors in manifest to %s...\n", strings.Join(authors, ", "))
		m.Authors = authors
	}

	fmt.Fprintf(progress, "Setting service ID in manifest to %q...\n", service.ID)
	m.ServiceID = service.ID

	fmt.Fprintf(progress, "Setting version in manifest to %d...\n", version)
	m.Version = version

	fmt.Fprintf(progress, "Setting language in manifest to %s...\n", language.Name)
	m.Language = language.Name

	if err := m.Write(filepath.Join(c.path, ManifestFilename)); err != nil {
		return fmt.Errorf("error saving package manifest: %w", err)
	}

	progress.Step("Initializing package...")

	if err := language.Initialize(progress); err != nil {
		fmt.Println(err)
		return fmt.Errorf("error initializing package: %w", err)
	}

	progress.Done()

	text.Break(out)

	text.Description(out, fmt.Sprintf("Initialized package %s to", text.Bold(m.Name)), abspath)
	text.Description(out, "Manage this service at", fmt.Sprintf("%s%s", manageServiceBaseURL, service.ID))
	text.Description(out, "To compile the package, run", "fastly compute build")
	text.Description(out, "To deploy the package, run", "fastly compute deploy")

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

func validateLanguageOption(languages []*Language) func(string) error {
	return func(input string) error {
		errMsg := fmt.Errorf("must be a valid option")
		if input == "" {
			return nil
		}
		if option, err := strconv.Atoi(input); err == nil {
			if option > len(languages) {
				return errMsg
			}
			return nil
		}
		return errMsg
	}
}

func validateTemplateOptionOrURL(templates []StarterKit) func(string) error {
	return func(input string) error {
		msg := "must be a valid option or Git URL"
		if input == "" {
			return nil
		}
		if option, err := strconv.Atoi(input); err == nil {
			if option > len(templates) {
				return fmt.Errorf(msg)
			}
			return nil
		}
		if !gitRepositoryRegEx.MatchString(input) {
			return fmt.Errorf(msg)
		}
		return nil
	}
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
