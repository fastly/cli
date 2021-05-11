package compute

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/common"
	"github.com/fastly/cli/pkg/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var (
	gitRepositoryRegEx        = regexp.MustCompile(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`)
	domainNameRegEx           = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)
	fastlyOrgRegEx            = regexp.MustCompile(`^https:\/\/github\.com\/fastly`)
	fastlyFileIgnoreListRegEx = regexp.MustCompile(`\.github|LICENSE|SECURITY\.md|CHANGELOG\.md|screenshot\.png`)
)

// InitCommand initializes a Compute@Edge project package on the local machine.
type InitCommand struct {
	common.Base
	client        api.HTTPClient
	manifest      manifest.Data
	language      string
	from          string
	branch        string
	tag           string
	path          string
	forceNonEmpty bool
}

// NewInitCommand returns a usable command registered under the parent.
func NewInitCommand(parent common.Registerer, client api.HTTPClient, globals *config.Data) *InitCommand {
	var c InitCommand
	c.Globals = globals
	c.client = client
	c.manifest.File.SetOutput(c.Globals.Output)
	c.manifest.File.Read(manifest.Filename)
	c.CmdClause = parent.Command("init", "Initialize a new Compute@Edge package locally")
	c.CmdClause.Flag("name", "Name of package, defaulting to directory name of the --path destination").Short('n').StringVar(&c.manifest.File.Name)
	c.CmdClause.Flag("description", "Description of the package").Short('d').StringVar(&c.manifest.File.Description)
	c.CmdClause.Flag("author", "Author(s) of the package").Short('a').StringsVar(&c.manifest.File.Authors)
	c.CmdClause.Flag("language", "Language of the package").Short('l').StringVar(&c.language)
	c.CmdClause.Flag("from", "Git repository containing package template").Short('f').StringVar(&c.from)
	c.CmdClause.Flag("branch", "Git branch name to clone from package template repository").Hidden().StringVar(&c.branch)
	c.CmdClause.Flag("tag", "Git tag name to clone from package template repository").Hidden().StringVar(&c.tag)
	c.CmdClause.Flag("path", "Destination to write the new package, defaulting to the current directory").Short('p').StringVar(&c.path)
	c.CmdClause.Flag("force", "Skip non-empty directory verification step and force new project creation").BoolVar(&c.forceNonEmpty)

	return &c
}

// Exec implements the command interface.
func (c *InitCommand) Exec(in io.Reader, out io.Writer) (err error) {
	text.Output(out, "Creating a new Compute@Edge project.")
	text.Break(out)
	text.Output(out, "Press ^C at any time to quit.")
	text.Break(out)

	cont, err := verifyDirectory(out, in)
	if err != nil {
		return err
	}

	if !c.forceNonEmpty && !cont {
		return errors.RemediationError{
			Inner:       fmt.Errorf("project directory not empty"),
			Remediation: errors.ExistingDirRemediation,
		}
	}

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

	var (
		name     string
		desc     string
		authors  []string
		language *Language
		from     string
	)

	languages := []*Language{
		NewLanguage(&LanguageOptions{
			Name:        "rust",
			DisplayName: "Rust",
			StarterKits: c.Globals.File.StarterKits.Rust,
			Toolchain:   NewRust(c.client, c.Globals),
		}),
		NewLanguage(&LanguageOptions{
			Name:        "assemblyscript",
			DisplayName: "AssemblyScript (beta)",
			StarterKits: c.Globals.File.StarterKits.AssemblyScript,
			Toolchain:   NewAssemblyScript(),
		}),
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

	name, _ = c.manifest.Name()
	name, err = pkgName(name, c.path, in, out)
	if err != nil {
		return err
	}

	desc, _ = c.manifest.Description()
	desc, err = pkgDesc(desc, in, out)
	if err != nil {
		return err
	}

	authors, _ = c.manifest.Authors()
	authors, err = pkgAuthors(authors, c.Globals.File.User.Email, in, out)
	if err != nil {
		return err
	}

	language, err = pkgLang(c.language, languages, in, out)
	if err != nil {
		return err
	}

	manifestExist := c.manifest.File.Exists()

	from, branch, tag, err := pkgFrom(c.from, c.branch, c.tag, manifestExist, language.StarterKits, in, out)
	if err != nil {
		return err
	}

	text.Break(out)

	if !c.Globals.Verbose() {
		progress = text.NewQuietProgress(out)
	}

	if from != "" && !manifestExist {
		err := pkgFetch(from, branch, tag, c.path, progress)
		if err != nil {
			return err
		}
	}

	m, err := updateManifest(progress, c.path, name, desc, authors, language)
	if err != nil {
		return err
	}

	progress.Step("Initializing package...")

	if err := language.Initialize(progress); err != nil {
		fmt.Println(err)
		return fmt.Errorf("error initializing package: %w", err)
	}

	progress.Done()

	text.Break(out)

	text.Description(out, fmt.Sprintf("Initialized package %s to", text.Bold(m.Name)), abspath)
	text.Description(out, "To publish the package (build and deploy), run", "fastly compute publish")
	text.Description(out, "To learn about deploying Compute@Edge projects using third-party orchestration tools, visit", "https://developer.fastly.com/learning/integrations/orchestration/")
	text.Success(out, "Initialized package %s", text.Bold(m.Name))

	return nil
}

// pkgName prompts the user for a package name unless already defined either
// via the corresponding CLI flag or the manifest file.
//
// It will use a default of the current directory path if no value provided by
// the user via the prompt.
func pkgName(name string, dirPath string, in io.Reader, out io.Writer) (string, error) {
	defaultName := filepath.Base(dirPath)

	if name == "" {
		var err error

		name, err = text.Input(out, fmt.Sprintf("Name: [%s] ", defaultName), in)
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}

		if name == "" {
			name = defaultName
		}
	}

	return name, nil
}

// pkgDesc prompts the user for a package description unless already defined
// either via the corresponding CLI flag or the manifest file.
func pkgDesc(desc string, in io.Reader, out io.Writer) (string, error) {
	if desc == "" {
		var err error

		desc, err = text.Input(out, "Description: ", in)
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}
	}

	return desc, nil
}

// pkgAuthors prompts the user for a package name unless already defined either
// via the corresponding CLI flag or the manifest file.
//
// It will use a default of the user's email found within the manifest, if set
// there, otherwise the value will be an empty slice.
func pkgAuthors(authors []string, manifestEmail string, in io.Reader, out io.Writer) ([]string, error) {
	if len(authors) == 0 {
		label := "Author: "

		if manifestEmail != "" {
			label = fmt.Sprintf("%s[%s] ", label, manifestEmail)
		}

		author, err := text.Input(out, label, in)
		if err != nil {
			return []string{}, fmt.Errorf("error reading input %w", err)
		}

		if author != "" {
			authors = []string{author}
		} else {
			authors = []string{manifestEmail}
		}
	}

	return authors, nil
}

// pkgLang prompts the user for a package language unless already defined
// either via the corresponding CLI flag or the manifest file.
func pkgLang(lang string, languages []*Language, in io.Reader, out io.Writer) (*Language, error) {
	var language *Language

	if lang == "" {
		text.Output(out, "%s", text.Bold("Language:"))
		for i, lang := range languages {
			text.Output(out, "[%d] %s", i+1, lang.DisplayName)
		}
		option, err := text.Input(out, "Choose option: [1] ", in, validateLanguageOption(languages))
		if err != nil {
			return nil, fmt.Errorf("reading input %w", err)
		}
		if option == "" {
			option = "1"
		}
		if i, err := strconv.Atoi(option); err == nil {
			language = languages[i-1]
		} else {
			return nil, fmt.Errorf("selecting language")
		}
	} else {
		for _, l := range languages {
			if strings.EqualFold(lang, l.Name) {
				language = l
			}
		}
	}

	return language, nil
}

// pkgFrom prompts the user for a package starter kit unless already defined
// either via the corresponding CLI flag or the manifest file.
//
// It returns the path to the starter kit, and the corresponding branch/tag,
// otherwise if there' is an error converting the prompt input, then the option
// number is returned along with the branch/tag that was potentially provided
// via the corresponding CLI flag or manifest content.
func pkgFrom(from string, branch string, tag string, manifestExist bool, kits []config.StarterKit, in io.Reader, out io.Writer) (string, string, string, error) {
	if from == "" && !manifestExist {
		text.Output(out, "%s", text.Bold("Starter kit:"))
		for i, kit := range kits {
			text.Output(out, "[%d] %s (%s)", i+1, kit.Name, kit.Path)
		}
		option, err := text.Input(out, "Choose option or type URL: [1] ", in, validateTemplateOptionOrURL(kits))
		if err != nil {
			return "", "", "", fmt.Errorf("error reading input %w", err)
		}
		if option == "" {
			option = "1"
		}

		if i, err := strconv.Atoi(option); err == nil {
			template := kits[i-1]
			from = template.Path
			branch = template.Branch
			tag = template.Tag
		} else {
			from = option
		}
	}

	return from, branch, tag, nil
}

// pkgFetch clones the given repo (from) into a temp directory, then copies
// specific files to the destination directory (path).
func pkgFetch(from string, branch string, tag string, fpath string, progress text.Progress) error {
	progress.Step("Fetching package template...")

	tempdir, err := tempDir("package-init")
	if err != nil {
		return fmt.Errorf("error creating temporary path for package template: %w", err)
	}
	defer os.RemoveAll(tempdir)

	if branch != "" && tag != "" {
		return fmt.Errorf("cannot use both git branch and tag name")
	}

	var ref plumbing.ReferenceName

	if branch != "" {
		ref = plumbing.NewBranchReferenceName(branch)
	}

	if tag != "" {
		ref = plumbing.NewTagReferenceName(tag)
	}

	_, err = git.PlainClone(tempdir, false, &git.CloneOptions{
		URL:           from,
		ReferenceName: ref,
		Depth:         1,
		Progress:      progress,
	})
	if err != nil {
		return fmt.Errorf("error fetching package template: %w", err)
	}

	if err := os.RemoveAll(filepath.Join(tempdir, ".git")); err != nil {
		return fmt.Errorf("error removing git metadata from package template: %w", err)
	}

	err = filepath.Walk(tempdir, func(path string, info os.FileInfo, err error) error {
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
		if fastlyOrgRegEx.MatchString(from) && fastlyFileIgnoreListRegEx.MatchString(rel) {
			return nil
		}

		dst := filepath.Join(fpath, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
			return err
		}

		if err := filesystem.CopyFile(path, dst); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error copying files from package template: %w", err)
	}

	return nil
}

// updateManifest updates the manifest with data acquired from various sources.
// e.g. prompting the user, existing manifest file.
func updateManifest(progress text.Progress, path string, name string, desc string, authors []string, lang *Language) (manifest.File, error) {
	progress.Step("Updating package manifest...")

	mp := filepath.Join(path, ManifestFilename)

	var m manifest.File
	if err := m.Read(mp); err != nil {
		return m, fmt.Errorf("error reading package manifest: %w", err)
	}

	fmt.Fprintf(progress, "Setting package name in manifest to %q...\n", name)
	m.Name = name

	if desc != "" {
		fmt.Fprintf(progress, "Setting description in manifest to %s...\n", desc)
		m.Description = desc
	}

	if len(authors) > 0 {
		fmt.Fprintf(progress, "Setting authors in manifest to %s...\n", strings.Join(authors, ", "))
		m.Authors = authors
	}

	fmt.Fprintf(progress, "Setting language in manifest to %s...\n", lang.Name)
	m.Language = lang.Name

	if err := m.Write(mp); err != nil {
		return m, fmt.Errorf("error saving package manifest: %w", err)
	}

	return m, nil
}

// verifyDirectory indicates if the user wants to continue with the execution
// flow when presented with a prompt that suggests the current directory isn't
// empty.
func verifyDirectory(out io.Writer, in io.Reader) (bool, error) {
	files, err := os.ReadDir(".")
	if err != nil {
		return false, err
	}

	if len(files) > 0 {
		dir, err := os.Getwd()
		if err != nil {
			return false, err
		}

		label := fmt.Sprintf("The current directory isn't empty. Are you sure you want to initialize a Compute@Edge project in %s? [y/n] ", dir)
		cont, err := text.Input(out, label, in)
		if err != nil {
			return false, fmt.Errorf("error reading input %w", err)
		}

		contl := strings.ToLower(cont)

		if contl == "n" || contl == "no" {
			return false, nil
		}

		if contl == "y" || contl == "yes" {
			return true, nil
		}

		// NOTE: be defensive and default to short-circuiting the execution flow if
		// the input is unrecognised.
		return false, nil
	}

	return true, nil
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

func validateTemplateOptionOrURL(templates []config.StarterKit) func(string) error {
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
