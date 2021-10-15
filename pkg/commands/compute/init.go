package compute

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/file"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/text"
)

var (
	gitRepositoryRegEx        = regexp.MustCompile(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`)
	fastlyOrgRegEx            = regexp.MustCompile(`^https:\/\/github\.com\/fastly`)
	fastlyFileIgnoreListRegEx = regexp.MustCompile(`\.github|LICENSE|SECURITY\.md|CHANGELOG\.md|screenshot\.png`)
)

// InitCommand initializes a Compute@Edge project package on the local machine.
type InitCommand struct {
	cmd.Base
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
func NewInitCommand(parent cmd.Registerer, client api.HTTPClient, globals *config.Data, data manifest.Data) *InitCommand {
	var c InitCommand
	c.Globals = globals
	c.client = client
	c.manifest = data
	c.CmdClause = parent.Command("init", "Initialize a new Compute@Edge package locally")
	c.CmdClause.Flag("name", "Name of package, defaulting to directory name of the --path destination").Short('n').StringVar(&c.manifest.File.Name)
	c.CmdClause.Flag("description", "Description of the package").Short('d').StringVar(&c.manifest.File.Description)
	c.CmdClause.Flag("author", "Author(s) of the package").Short('a').StringsVar(&c.manifest.File.Authors)
	c.CmdClause.Flag("language", "Language of the package").Short('l').StringVar(&c.language)
	c.CmdClause.Flag("from", "Git repository URL, or URL referencing a .zip/.tar.gz file, containing a package template").Short('f').StringVar(&c.from)
	c.CmdClause.Flag("branch", "Git branch name to clone from package template repository").Hidden().StringVar(&c.branch)
	c.CmdClause.Flag("tag", "Git tag name to clone from package template repository").Hidden().StringVar(&c.tag)
	c.CmdClause.Flag("path", "Destination to write the new package, defaulting to the current directory").Short('p').StringVar(&c.path)
	c.CmdClause.Flag("force", "Skip non-empty directory verification step and force new project creation").BoolVar(&c.forceNonEmpty)

	return &c
}

// Exec implements the command interface.
func (c *InitCommand) Exec(in io.Reader, out io.Writer) (err error) {
	var introContext string
	if c.from != "" {
		introContext = " (using --from to locate package template)"
	}

	text.Break(out)
	text.Output(out, "Creating a new Compute@Edge project%s.", introContext)
	text.Break(out)
	text.Output(out, "Press ^C at any time to quit.")
	text.Break(out)

	if !c.forceNonEmpty {
		cont, err := verifyDirectory(out, in)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			return err
		}
		if !cont {
			return errors.RemediationError{
				Inner:       fmt.Errorf("project directory not empty"),
				Remediation: errors.ExistingDirRemediation,
			}
		}
	}

	var progress text.Progress
	if c.Globals.Verbose() {
		progress = text.NewVerboseProgress(out)
	} else {
		// Use a null progress writer whilst gathering input.
		progress = text.NewNullProgress()
	}
	defer func(errLog errors.LogInterface) {
		if err != nil {
			errLog.Add(err)
			progress.Fail() // progress.Done is handled inline
		}
	}(c.Globals.ErrLog)

	wd, err := os.Getwd()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error determining current directory: %w", err)
	}

	mf := c.manifest.File
	if c.path == "" && !mf.Exists() {
		fmt.Fprintf(progress, "--path not specified, using current directory\n")
		c.path = wd
	}

	dst, err := verifyDestination(c.path, progress)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Path": c.path,
		})
		return err
	}
	c.path = dst

	name, desc, authors, err := promptOrReturn(c.manifest, c.path, c.Globals.File.User.Email, in, out)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Authors":     authors,
			"Description": desc,
			"Email":       c.Globals.File.User.Email,
			"Name":        name,
			"Path":        c.path,
		})
		return err
	}

	languages := NewLanguages(c.Globals.File.StarterKits, c.client, c.Globals)
	language, err := selectLanguage(c.language, languages, in, out)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Language": c.language,
		})
		return err
	}

	var from, branch, tag string

	if noProjectFiles(c.from, language.Name, mf) {
		from, branch, tag, err = promptForStarterKit(language.StarterKits, in, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"From":           c.from,
				"Branch":         c.branch,
				"Tag":            c.tag,
				"Manifest Exist": false,
			})
			return err
		}
	}

	text.Break(out)
	if !c.Globals.Verbose() {
		progress = text.NewProgress(out, false)
	}

	if from != "" && !mf.Exists() {
		err = pkgFetch(from, branch, tag, c.path, file.Archives, progress, c.client, out, c.Globals.ErrLog)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
				"From":   from,
				"Branch": branch,
				"Tag":    tag,
				"Path":   c.path,
			})
			return err
		}
	}

	mf, err = updateManifest(mf, progress, c.path, name, desc, authors, language)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Path":        c.path,
			"Name":        name,
			"Description": desc,
			"Authors":     authors,
			"Language":    language,
		})
		return err
	}

	language, err = initializeLanguage(progress, language, languages, mf.Language, wd, c.path)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error initializing package: %w", err)
	}

	progress.Done()
	displayOutput(mf.Name, dst, language.Name, out)
	return nil
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

		label := fmt.Sprintf("The current directory isn't empty. Are you sure you want to initialize a Compute@Edge project in %s? [y/N] ", dir)
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

// verifyDestination checks the provided path exists and is a directory.
//
// NOTE: For validating user permissions it will create a temporary file within
// the directory and then remove it before returning the absolute path to the
// directory itself.
func verifyDestination(path string, verbose io.Writer) (dst string, err error) {
	dst, err = filepath.Abs(path)
	if err != nil {
		return dst, err
	}

	fi, err := os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return dst, fmt.Errorf("couldn't verify package directory: %w", err) // generic error
	}
	if err == nil && !fi.IsDir() {
		return dst, fmt.Errorf("package destination is not a directory") // specific problem
	}
	if err != nil && os.IsNotExist(err) { // normal-ish case
		fmt.Fprintf(verbose, "Creating %s...\n", dst)
		if err := os.MkdirAll(dst, 0700); err != nil {
			return dst, fmt.Errorf("error creating package destination: %w", err)
		}
	}

	tmpname := make([]byte, 16)
	n, err := rand.Read(tmpname)
	if err != nil {
		return dst, fmt.Errorf("error generating random filename: %w", err)
	}
	if n != 16 {
		return dst, fmt.Errorf("failed to generate enough entropy (%d/%d)", n, 16)
	}

	f, err := os.Create(filepath.Join(dst, fmt.Sprintf("tmp_%x", tmpname)))
	if err != nil {
		return dst, fmt.Errorf("error creating file in package destination: %w", err)
	}

	if err := f.Close(); err != nil {
		return dst, fmt.Errorf("error closing file in package destination: %w", err)
	}

	if err := os.Remove(f.Name()); err != nil {
		return dst, fmt.Errorf("error removing file in package destination: %w", err)
	}

	return dst, nil
}

// promptOrReturn will prompt the user for information missing from the
// fastly.toml manifest file, otherwise if it already exists then the value is
// returned as is.
func promptOrReturn(m manifest.Data, path, email string, in io.Reader, out io.Writer) (name, description string, authors []string, err error) {
	name, _ = m.Name()
	name, err = pkgName(name, path, in, out)
	if err != nil {
		return name, description, authors, err
	}

	description, _ = m.Description()
	description, err = pkgDesc(description, in, out)
	if err != nil {
		return name, description, authors, err
	}

	authors, _ = m.Authors()
	authors, err = pkgAuthors(authors, email, in, out)
	if err != nil {
		return name, description, authors, err
	}

	return name, description, authors, nil
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

// selectLanguage decides whether to prompt the user for a language if none
// defined or try and match the --language flag against available languages.
func selectLanguage(langFlag string, ls []*Language, in io.Reader, out io.Writer) (*Language, error) {
	if langFlag == "" {
		return promptForLanguage(ls, in, out)
	}

	for _, language := range ls {
		if strings.EqualFold(langFlag, language.Name) {
			return language, nil
		}
	}

	return nil, fmt.Errorf("error looking up specified language: '%s' not supported", langFlag)
}

// promptForLanguage prompts the user for a package language unless already
// defined either via the corresponding CLI flag or the manifest file.
func promptForLanguage(languages []*Language, in io.Reader, out io.Writer) (*Language, error) {
	var language *Language

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

	return language, nil
}

// validateLanguageOption ensures the user selects an appropriate value from
// the prompt options displayed.
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

// noProjectFiles indicates if the user needs to be prompted to select a
// Starter Kit for their chosen language.
func noProjectFiles(from, language string, mf manifest.File) bool {
	return from == "" && language != "other" && !mf.Exists()
}

// promptForStarterKit prompts the user for a package starter kit.
//
// It returns the path to the starter kit, and the corresponding branch/tag,
func promptForStarterKit(kits []config.StarterKit, in io.Reader, out io.Writer) (from string, branch string, tag string, err error) {
	text.Output(out, "%s", text.Bold("Starter kit:"))
	for i, kit := range kits {
		fmt.Fprintf(out, "[%d] %s\n", i+1, text.Bold(kit.Name))
		text.Indent(out, 4, "%s\n%s", kit.Description, kit.Path)
	}
	option, err := text.Input(out, "Choose option or paste git URL: [1] ", in, validateTemplateOptionOrURL(kits))
	if err != nil {
		return "", "", "", fmt.Errorf("error reading input: %w", err)
	}
	if option == "" {
		option = "1"
	}

	var i int
	if i, err = strconv.Atoi(option); err != nil {
		return "", "", "", fmt.Errorf("error parsing input: %w", err)
	}

	template := kits[i-1]

	return template.Path, template.Branch, template.Tag, nil
}

func validateTemplateOptionOrURL(templates []config.StarterKit) func(string) error {
	return func(input string) error {
		msg := "must be a valid option or git URL"
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

// pkgFetch will determine if the package code should be fetched from GitHub
// using the git binary to clone the source or a HTTP request that uses
// content-negotiation to determine the type of archive format used.
func pkgFetch(
	from, branch, tag, dst string,
	archives []file.Archive,
	progress text.Progress,
	client api.HTTPClient,
	out io.Writer,
	errLog errors.LogInterface) error {

	progress.Step("Fetching package template...")

	_, err := url.Parse(from)
	if err != nil {
		errLog.Add(err)
		return fmt.Errorf("error parsing --from as URL: %w", err)
	}

	req, err := http.NewRequest("GET", from, nil)
	if err != nil {
		errLog.Add(err)
		return fmt.Errorf("failed to construct package request URL: %w", err)
	}

	for _, archive := range archives {
		for _, mime := range archive.MimeTypes() {
			req.Header.Add("Accept", mime)
		}
	}

	res, err := client.Do(req)
	if err != nil {
		errLog.Add(err)
		return fmt.Errorf("failed to get package: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errLog.Add(err)
		return fmt.Errorf("failed to get package: %s", res.Status)
	}

	filename := filepath.Base(from)
	f, err := os.Create(filename)
	if err != nil {
		errLog.Add(err)
		return fmt.Errorf("failed to create local %s archive: %w", filename, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			errLog.Add(err)
		}
	}()
	defer func(base string) {
		err := os.Remove(filename)
		if err != nil {
			errLog.Add(err)
			text.Break(out)
			text.Info(out, "We were unable to clean-up the local %s file (it can be safely removed)", filename)
		}
	}(filename)

	_, err = io.Copy(f, res.Body)
	if err != nil {
		errLog.Add(err)
		return fmt.Errorf("failed to write %s archive to disk: %w", filename, err)
	}

	var archive file.Archive

mimes:
	for _, mimetype := range res.Header.Values("Content-Type") {
		for _, a := range archives {
			for _, mime := range a.MimeTypes() {
				if mimetype == mime {
					archive = a
					break mimes
				}
			}
		}
	}

	if archive == nil {
		ext := filepath.Ext(filename)

		for _, a := range archives {
			if ext == a.Extension() {
				archive = a
				break
			}
		}
	}

	if archive != nil {
		archive.SetDestination(dst)
		archive.SetFile(f)
		archive.SetFilename(filename)

		err = archive.Extract()
		if err != nil {
			errLog.Add(err)
			return fmt.Errorf("failed to extract %s archive content: %w", filename, err)
		}

		return nil
	}

	return pkgClone(from, branch, tag, dst)
}

// pkgClone clones the given repo (from) into a temp directory, then copies
// specific files to the destination directory (path).
func pkgClone(from string, branch string, tag string, dst string) error {
	_, err := exec.LookPath("git")
	if err != nil {
		return errors.RemediationError{
			Inner:       fmt.Errorf("`git` not found in $PATH"),
			Remediation: fmt.Sprintf("The Fastly CLI requires a local installation of git.  For installation instructions for your operating system see:\n\n\t$ %s", text.Bold("https://git-scm.com/book/en/v2/Getting-Started-Installing-Git")),
		}
	}

	tempdir, err := tempDir("package-init")
	if err != nil {
		return fmt.Errorf("error creating temporary path for package template: %w", err)
	}
	defer os.RemoveAll(tempdir)

	if branch != "" && tag != "" {
		return fmt.Errorf("cannot use both git branch and tag name")
	}

	args := []string{
		"clone",
		"--depth",
		"1",
	}
	var ref string
	if branch != "" {
		ref = branch
	}
	if tag != "" {
		ref = tag
	}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, from, tempdir)

	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as there should be no vulnerability to cloning a remote repo.
	/* #nosec */
	cmd := exec.Command("git", args...)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error fetching package template: %w\n\n%s", err, stdoutStderr)
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

		dst := filepath.Join(dst, rel)
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

// updateManifest updates the manifest with data acquired from various sources.
// e.g. prompting the user, existing manifest file.
//
// NOTE: The lang argument might be nil (if the user passes --from flag).
func updateManifest(m manifest.File, progress text.Progress, path string, name string, desc string, authors []string, lang *Language) (manifest.File, error) {
	progress.Step("Updating package manifest...")

	mp := filepath.Join(path, manifest.Filename)

	if err := m.Read(mp); err != nil {
		if lang != nil {
			if lang.Name == "other" {
				// We create a fastly.toml manifest on behalf of the user if they're
				// bringing their own pre-compiled Wasm binary to be packaged.
				m.ManifestVersion = manifest.ManifestLatestVersion
				m.Name = name
				m.Description = desc
				m.Authors = authors
				m.Language = lang.Name
				if err := m.Write(mp); err != nil {
					return m, fmt.Errorf("error saving package manifest: %w", err)
				}
				return m, nil
			}
		}
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

	if lang != nil {
		fmt.Fprintf(progress, "Setting language in manifest to %s...\n", lang.Name)
		m.Language = lang.Name
	}

	if err := m.Write(mp); err != nil {
		return m, fmt.Errorf("error saving package manifest: %w", err)
	}

	return m, nil
}

// initializeLanguage for newly cloned package.
func initializeLanguage(progress text.Progress, language *Language, languages []*Language, name, wd, path string) (*Language, error) {
	progress.Step("Initializing package...")

	if wd != path {
		err := os.Chdir(path)
		if err != nil {
			return nil, fmt.Errorf("error changing to your project directory: %w", err)
		}
	}

	// Language will not be set if user provides the --from flag. So we'll check
	// the manifest content and ensure what's set there is the language instance
	// used for the sake of `compute build` operations.
	if language == nil {
		var match bool
		for _, l := range languages {
			if strings.EqualFold(name, l.Name) {
				language = l
				match = true
				break
			}
		}
		if !match {
			return nil, fmt.Errorf("unrecognised package language")
		}
	}

	if language.Name != "other" {
		if err := language.Initialize(progress); err != nil {
			return nil, err
		}
	}

	return language, nil
}

// displayOutput of package information and useful links.
func displayOutput(name, dst, language string, out io.Writer) {
	text.Break(out)
	text.Description(out, fmt.Sprintf("Initialized package %s to", text.Bold(name)), dst)

	if language == "other" {
		text.Description(out, "To package a pre-compiled Wasm binary for deployment, run", "fastly compute pack")
		text.Description(out, "To deploy the package, run", "fastly compute deploy")
	} else {
		text.Description(out, "To publish the package (build and deploy), run", "fastly compute publish")
	}

	text.Description(out, "To learn about deploying Compute@Edge projects using third-party orchestration tools, visit", "https://developer.fastly.com/learning/integrations/orchestration/")
	text.Success(out, "Initialized package %s", text.Bold(name))
}
