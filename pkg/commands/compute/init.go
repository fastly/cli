package compute

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/file"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
	cp "github.com/otiai10/copy"
)

var (
	gitRepositoryRegEx        = regexp.MustCompile(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)(/)?`)
	fastlyOrgRegEx            = regexp.MustCompile(`^https:\/\/github\.com\/fastly`)
	fastlyFileIgnoreListRegEx = regexp.MustCompile(`\.github|LICENSE|SECURITY\.md|CHANGELOG\.md|screenshot\.png`)
)

// InitCommand initializes a Compute@Edge project package on the local machine.
type InitCommand struct {
	cmd.Base

	branch           string
	dir              string
	from             string
	language         string
	manifest         manifest.Data
	skipVerification bool
	tag              string
}

// Languages is a list of supported language options.
var Languages = []string{"rust", "javascript", "go", "assemblyscript", "other"}

// NewInitCommand returns a usable command registered under the parent.
func NewInitCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *InitCommand {
	var c InitCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("init", "Initialize a new Compute@Edge package locally")
	c.CmdClause.Flag("name", "Name of package, falls back to --directory").Short('n').StringVar(&c.manifest.File.Name)
	c.CmdClause.Flag("description", "Description of the package").StringVar(&c.manifest.File.Description)
	c.CmdClause.Flag("directory", "Destination to write the new package, defaulting to the current directory").Short('p').StringVar(&c.dir)
	c.CmdClause.Flag("author", "Author(s) of the package").Short('a').StringsVar(&c.manifest.File.Authors)
	c.CmdClause.Flag("language", "Language of the package").Short('l').HintOptions(Languages...).EnumVar(&c.language, Languages...)
	c.CmdClause.Flag("from", "Local project directory, or Git repository URL, or URL referencing a .zip/.tar.gz file, containing a package template").Short('f').StringVar(&c.from)
	c.CmdClause.Flag("branch", "Git branch name to clone from package template repository").Hidden().StringVar(&c.branch)
	c.CmdClause.Flag("tag", "Git tag name to clone from package template repository").Hidden().StringVar(&c.tag)
	c.CmdClause.Flag("force", "Skip non-empty directory verification step and force new project creation").BoolVar(&c.skipVerification)

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

	if c.from != "" && c.language == "" {
		text.Warning(out, "When using the --from flag, the project language cannot be inferred. Please either use the --language flag to explicitly set the language or ensure the project's fastly.toml sets a valid language.")
	}

	text.Break(out)

	cont, err := verifyDirectory(c.dir, c.skipVerification, out, in)
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

	// NOTE: Will be a NullProgress unless --verbose is set.
	//
	// This is because we don't want any progress output until later.
	progress := instantiateProgress(c.Globals.Verbose(), out)

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
	if c.dir == "" && !mf.Exists() {
		fmt.Fprintf(progress, "--directory not specified, using current directory\n\n")
		c.dir = wd
	}

	dst, err := verifyDestination(c.dir, progress)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Directory": c.dir,
		})
		return err
	}
	c.dir = dst

	// Assign the default profile email if available.
	email := ""
	profileName, p := profile.Default(c.Globals.File.Profiles)
	if profileName != "" {
		email = p.Email
	}

	name, desc, authors, err := promptOrReturn(c.manifest, c.dir, email, in, out)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Description": desc,
			"Directory":   c.dir,
		})
		return err
	}

	languages := NewLanguages(c.Globals.File.StarterKits, c.Globals, name, mf.Scripts)
	language, err := selectLanguage(c.from, c.language, languages, mf, in, out)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Language": c.language,
		})
		return err
	}

	var from, branch, tag string

	if noProjectFiles(c.from, language, mf) {
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
		c.from = from
	}

	text.Break(out)

	// NOTE: From this point onwards we need a non-null progress regardless of
	// whether --verbose was set or not.
	progress = text.NewProgress(out, c.Globals.Verbose())

	err = fetchPackageTemplate(language, c.from, branch, tag, c.dir, mf, file.Archives, progress, c.Globals.HTTPClient, out, c.Globals.ErrLog)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"From":      from,
			"Branch":    branch,
			"Tag":       tag,
			"Directory": c.dir,
		})
		return err
	}

	mf, err = updateManifest(mf, progress, c.dir, name, desc, authors, language)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Directory":   c.dir,
			"Description": desc,
			"Language":    language,
		})
		return err
	}

	language, err = initializeLanguage(progress, language, languages, mf.Language, wd, c.dir, mf.Scripts.Build)
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
func verifyDirectory(dir string, skipVerification bool, out io.Writer, in io.Reader) (bool, error) {
	if skipVerification {
		return true, nil
	}

	if dir == "" {
		dir = "."
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}

	if len(files) > 0 {
		label := fmt.Sprintf("The current directory isn't empty. Are you sure you want to initialize a Compute@Edge project in %s? [y/N] ", dir)
		return text.AskYesNo(out, label, in)
	}

	return true, nil
}

// instantiateProgress returns an instance of a text.Progress bar.
func instantiateProgress(verbose bool, out io.Writer) text.Progress {
	if verbose {
		return text.NewVerboseProgress(out)
	}
	return text.NewNullProgress()
}

// verifyDestination checks the provided path exists and is a directory.
//
// NOTE: For validating user permissions it will create a temporary file within
// the directory and then remove it before returning the absolute path to the
// directory itself.
func verifyDestination(path string, progress text.Progress) (dst string, err error) {
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
		fmt.Fprintf(progress, "Creating %s...\n", dst)
		if err := os.MkdirAll(dst, 0o700); err != nil {
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

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as the input is determined by our own package.
	/* #nosec */
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
	name, err = packageName(name, path, in, out)
	if err != nil {
		return name, description, authors, err
	}

	description, _ = m.Description()
	description, err = packageDescription(description, in, out)
	if err != nil {
		return name, description, authors, err
	}

	authors, _ = m.Authors()
	authors, err = packageAuthors(authors, email, in, out)
	if err != nil {
		return name, description, authors, err
	}

	return name, description, authors, nil
}

// packageName prompts the user for a package name unless already defined either
// via the corresponding CLI flag or the manifest file.
//
// It will use a default of the current directory path if no value provided by
// the user via the prompt.
func packageName(name string, dirPath string, in io.Reader, out io.Writer) (string, error) {
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

// packageDescription prompts the user for a package description unless already
// defined either via the corresponding CLI flag or the manifest file.
func packageDescription(desc string, in io.Reader, out io.Writer) (string, error) {
	if desc == "" {
		var err error

		desc, err = text.Input(out, "Description: ", in)
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}
	}

	return desc, nil
}

// packageAuthors prompts the user for a package name unless already defined
// either via the corresponding CLI flag or the manifest file.
//
// It will use a default of the user's email found within the manifest, if set
// there, otherwise the value will be an empty slice.
func packageAuthors(authors []string, manifestEmail string, in io.Reader, out io.Writer) ([]string, error) {
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
func selectLanguage(from string, langFlag string, ls []*Language, mf manifest.File, in io.Reader, out io.Writer) (*Language, error) {
	if from != "" && langFlag == "" || mf.Exists() {
		return nil, nil
	}

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
func noProjectFiles(from string, language *Language, mf manifest.File) bool {
	if from != "" || language == nil || mf.Exists() {
		return false
	}
	return from == "" && language.Name != "other" && !mf.Exists()
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
	if i, err = strconv.Atoi(option); err == nil {
		template := kits[i-1]
		return template.Path, template.Branch, template.Tag, nil
	}

	return option, "", "", nil
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

// fetchPackageTemplate will determine if the package code should be fetched
// from GitHub using the git binary to clone the source or a HTTP request that
// uses content-negotiation to determine the type of archive format used.
func fetchPackageTemplate(
	language *Language,
	from, branch, tag, dst string,
	mf manifest.File,
	archives []file.Archive,
	progress text.Progress,
	client api.HTTPClient,
	out io.Writer,
	errLog errors.LogInterface,
) error {
	// We don't try to fetch a package template if the user is bringing their own
	// compiled Wasm binary (or if the directory currently already contains a
	// fastly.toml manifest file).
	if mf.Exists() || language != nil && language.Name == "other" {
		return nil
	}
	progress.Step("Fetching package template...")

	// If the user has provided a local file path, we'll recursively copy the
	// directory to dst.
	fi, err := os.Stat(from)
	if err != nil {
		errLog.Add(err)
	} else if fi.IsDir() {
		return cp.Copy(from, dst)
	}

	req, err := http.NewRequest("GET", from, nil)
	if err != nil {
		errLog.Add(err)
		if gitRepositoryRegEx.MatchString(from) {
			return clonePackageFromEndpoint(from, branch, tag, dst)
		}
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
		err := fmt.Errorf("failed to get package: %s", res.Status)
		errLog.Add(err)
		return err
	}

	filename := filepath.Base(from)
	ext := filepath.Ext(filename)

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as we require a user to configure their own environment.
	/* #nosec */
	f, err := os.Create(filename)
	if err != nil {
		errLog.Add(err)
		return fmt.Errorf("failed to create local %s archive: %w", filename, err)
	}
	defer func() {
		// NOTE: Later on we rename the file to include an extension and the
		// following call to os.Remove works still because the `filename` variable
		// that is still in scope is also updated to include the extension.
		err := os.Remove(filename)
		if err != nil {
			errLog.Add(err)
			text.Break(out)
			text.Info(out, "We were unable to clean-up the local %s file (it can be safely removed)", filename)
		}
	}()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		errLog.Add(err)
		return fmt.Errorf("failed to write %s archive to disk: %w", filename, err)
	}

	// NOTE: We used to `defer` the closing of the file after its creation but
	// realised that this caused issues on Windows as it was unable to rename the
	// file as we still have the descriptor `f` open.
	if err := f.Close(); err != nil {
		errLog.Add(err)
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
		for _, a := range archives {
			for _, e := range a.Extensions() {
				if ext == e {
					archive = a
					break
				}
			}
		}
	}

	if archive != nil {
		// Ensure there is a file extension on our filename, otherwise we won't
		// know what type of archive format we're dealing with when we come to call
		// the archive.Extract() method.
		if ext == "" {
			filenameWithExt := filename + archive.Extensions()[0]
			err := os.Rename(filename, filenameWithExt)
			if err != nil {
				errLog.Add(err)
				return err
			}
			filename = filenameWithExt
		}

		archive.SetDestination(dst)
		archive.SetFilename(filename)

		err = archive.Extract()
		if err != nil {
			errLog.Add(err)
			return fmt.Errorf("failed to extract %s archive content: %w", filename, err)
		}

		return nil
	}

	return clonePackageFromEndpoint(from, branch, tag, dst)
}

// clonePackageFromEndpoint clones the given repo (from) into a temp directory,
// then copies specific files to the destination directory (path).
func clonePackageFromEndpoint(from string, branch string, tag string, dst string) error {
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
		if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
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

	if err = os.MkdirAll(abspath, 0o750); err != nil {
		return "", err
	}

	return abspath, nil
}

// updateManifest updates the manifest with data acquired from various sources.
// e.g. prompting the user, existing manifest file.
//
// NOTE: The language argument might be nil (if the user passes --from flag).
func updateManifest(
	m manifest.File,
	progress text.Progress,
	path, name, desc string,
	authors []string,
	language *Language,
) (manifest.File, error) {
	progress.Step("Updating package manifest...")

	mp := filepath.Join(path, manifest.Filename)

	if err := m.Read(mp); err != nil {
		if language != nil {
			if language.Name == "other" {
				// We create a fastly.toml manifest on behalf of the user if they're
				// bringing their own pre-compiled Wasm binary to be packaged.
				m.ManifestVersion = manifest.ManifestLatestVersion
				m.Name = name
				m.Description = desc
				m.Authors = authors
				m.Language = language.Name
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

	if language != nil {
		fmt.Fprintf(progress, "Setting language in manifest to %s...\n", language.Name)
		m.Language = language.Name
	}

	if err := m.Write(mp); err != nil {
		return m, fmt.Errorf("error saving package manifest: %w", err)
	}

	return m, nil
}

// initializeLanguage for newly cloned package.
func initializeLanguage(progress text.Progress, language *Language, languages []*Language, name, wd, path, build string) (*Language, error) {
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

	if language.Name != "other" && build == "" {
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
