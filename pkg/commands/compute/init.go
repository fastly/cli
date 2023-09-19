package compute

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	cp "github.com/otiai10/copy"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/file"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/profile"
	"github.com/fastly/cli/pkg/text"
)

var (
	gitRepositoryRegEx        = regexp.MustCompile(`((git|ssh|http(s)?)|(git@[\w\.]+))(:(//)?)([\w\.@\:/\-~]+)(\.git)?(/)?`)
	fastlyOrgRegEx            = regexp.MustCompile(`^https:\/\/github\.com\/fastly`)
	fastlyFileIgnoreListRegEx = regexp.MustCompile(`\.github|LICENSE|SECURITY\.md|CHANGELOG\.md|screenshot\.png`)
)

// InitCommand initializes a Compute project package on the local machine.
type InitCommand struct {
	cmd.Base

	branch             string
	dir                string
	cloneFrom          string
	language           string
	manifest           manifest.Data
	tag                string
	wasmtoolsVersioner github.AssetVersioner
}

// Languages is a list of supported language options.
var Languages = []string{"rust", "javascript", "go", "other"}

// NewInitCommand returns a usable command registered under the parent.
func NewInitCommand(parent cmd.Registerer, g *global.Data, m manifest.Data, wasmtoolsVersioner github.AssetVersioner) *InitCommand {
	var c InitCommand
	c.Globals = g
	c.manifest = m
	c.wasmtoolsVersioner = wasmtoolsVersioner

	c.CmdClause = parent.Command("init", "Initialize a new Compute package locally")
	c.CmdClause.Flag("directory", "Destination to write the new package, defaulting to the current directory").Short('p').StringVar(&c.dir)
	c.CmdClause.Flag("author", "Author(s) of the package").Short('a').StringsVar(&c.manifest.File.Authors)
	c.CmdClause.Flag("language", "Language of the package").Short('l').HintOptions(Languages...).EnumVar(&c.language, Languages...)
	c.CmdClause.Flag("from", "Local project directory, or Git repository URL, or URL referencing a .zip/.tar.gz file, containing a package template").Short('f').StringVar(&c.cloneFrom)
	c.CmdClause.Flag("branch", "Git branch name to clone from package template repository").Hidden().StringVar(&c.branch)
	c.CmdClause.Flag("tag", "Git tag name to clone from package template repository").Hidden().StringVar(&c.tag)

	return &c
}

// Exec implements the command interface.
func (c *InitCommand) Exec(in io.Reader, out io.Writer) (err error) {
	var introContext string
	if c.cloneFrom != "" {
		introContext = " (using --from to locate package template)"
	}

	text.Output(out, "Creating a new Compute project%s.\n\n", introContext)
	text.Output(out, "Press ^C at any time to quit.")

	if c.cloneFrom != "" && c.language == "" {
		text.Warning(out, "\nWhen using the --from flag, the project language cannot be inferred. Please either use the --language flag to explicitly set the language or ensure the project's fastly.toml sets a valid language.")
	}

	text.Break(out)
	cont, notEmpty, err := verifyDirectory(c.Globals.Flags, c.dir, out, in)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}
	if !cont {
		return fsterr.RemediationError{
			Inner:       fmt.Errorf("project directory not empty"),
			Remediation: fsterr.ExistingDirRemediation,
		}
	}

	defer func(errLog fsterr.LogInterface) {
		if err != nil {
			errLog.Add(err)
		}
	}(c.Globals.ErrLog)

	wd, err := os.Getwd()
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error determining current directory: %w", err)
	}

	mf := c.manifest.File
	if c.Globals.Flags.Quiet {
		mf.SetQuiet(true)
	}
	if c.dir == "" && !mf.Exists() && c.Globals.Verbose() {
		text.Info(out, "--directory not specified, using current directory\n\n")
		c.dir = wd
	}

	spinner, err := text.NewSpinner(out)
	if err != nil {
		return err
	}

	dst, err := verifyDestination(c.dir, spinner)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Directory": c.dir,
		})
		return err
	}
	c.dir = dst

	if notEmpty {
		text.Break(out)
	}
	err = spinner.Process("Validating directory permissions", validateDirectoryPermissions(dst))
	if err != nil {
		return err
	}

	// Assign the default profile email if available.
	email := ""
	profileName, p := profile.Default(c.Globals.Config.Profiles)
	if profileName != "" {
		email = p.Email
	}

	name, desc, authors, err := promptOrReturn(c.Globals.Flags, c.manifest, c.dir, email, in, out)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Description": desc,
			"Directory":   c.dir,
		})
		return err
	}

	languages := NewLanguages(c.Globals.Config.StarterKits)

	var language *Language

	if c.language == "" && c.cloneFrom == "" && c.Globals.Manifest.File.Language == "" {
		language, err = promptForLanguage(c.Globals.Flags, languages, in, out)
		if err != nil {
			return err
		}
	}

	// NOTE: The --language flag is an EnumVar, meaning it's already validated.
	if c.language != "" || mf.Language != "" {
		l := c.language
		if c.language == "" {
			l = mf.Language
		}
		for _, recognisedLanguage := range languages {
			if strings.EqualFold(l, recognisedLanguage.Name) {
				language = recognisedLanguage
			}
		}
	}

	var from, branch, tag string

	// If the user doesn't tell us where to clone from, or there is already a
	// fastly.toml manifest, or the language they selected was "other" (meaning
	// they're bringing their own project code), then we'll prompt the user to
	// select a starter kit project.
	triggerStarterKitPrompt := c.cloneFrom == "" && !mf.Exists() && language.Name != "other"
	if triggerStarterKitPrompt {
		from, branch, tag, err = promptForStarterKit(c.Globals.Flags, language.StarterKits, in, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"From":           c.cloneFrom,
				"Branch":         c.branch,
				"Tag":            c.tag,
				"Manifest Exist": false,
			})
			return err
		}
		c.cloneFrom = from
	}

	// We only want to fetch a remote package if c.cloneFrom has been set.
	// This can happen in two ways:
	//
	// 1. --from flag is set
	// 2. user selects starter kit when prompted
	//
	// We don't fetch if the user has indicated their language of choice is
	// "other" because this means they intend on handling the compilation of code
	// that isn't natively supported by the platform.
	if c.cloneFrom != "" {
		err = fetchPackageTemplate(c, branch, tag, file.Archives, spinner, out)
		if err != nil {
			c.Globals.ErrLog.AddWithContext(err, map[string]any{
				"From":      from,
				"Branch":    branch,
				"Tag":       tag,
				"Directory": c.dir,
			})
			return err
		}
	}

	// If the user was prompted to fill the name/desc/authors/lang, then we insert
	// a line break so the following spinner instances have spacing. But only if
	// the starter kit wasn't prompted for as that already handles spacing.
	if (mf.Name == "" || mf.Description == "" || mf.Language == "" || len(mf.Authors) == 0) && !triggerStarterKitPrompt {
		text.Break(out)
	}

	mf, err = updateManifest(mf, spinner, c.dir, name, desc, authors, language)
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]any{
			"Directory":   c.dir,
			"Description": desc,
			"Language":    language,
		})
		return err
	}

	language, err = initializeLanguage(spinner, language, languages, mf.Language, wd, c.dir)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return fmt.Errorf("error initializing package: %w", err)
	}

	var md manifest.Data
	err = md.File.Read(manifest.Filename)
	if err != nil {
		return fmt.Errorf("failed to read manifest after initialisation: %w", err)
	}

	postInit := md.File.Scripts.PostInit
	if postInit != "" {
		if !c.Globals.Flags.AutoYes && !c.Globals.Flags.NonInteractive {
			msg := fmt.Sprintf(CustomPostScriptMessage, "init", manifest.Filename)
			err := promptForPostInitContinue(msg, postInit, out, in)
			if err != nil {
				if errors.Is(err, fsterr.ErrPostInitStopped) {
					displayInitOutput(mf.Name, dst, language.Name, out)
					return nil
				}
				return err
			}
		}

		if c.Globals.Flags.Verbose && len(md.File.Scripts.EnvVars) > 0 {
			text.Description(out, "Environment variables set", strings.Join(md.File.Scripts.EnvVars, " "))
		}

		// If we're in verbose mode, the command output is shown.
		// So in that case we don't want to have a spinner as it'll interweave output.
		// In non-verbose mode we have a spinner running while the command execution is happening.
		msg := "Running [scripts.post_init]..."
		if !c.Globals.Flags.Verbose {
			err = spinner.Start()
			if err != nil {
				return err
			}
			spinner.Message(msg)
		}

		s := Shell{}
		command, args := s.Build(postInit)
		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
		// Disabling as we require the user to provide this command.
		// #nosec
		// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command
		err := fstexec.Command(fstexec.CommandOpts{
			Args:           args,
			Command:        command,
			Env:            md.File.Scripts.EnvVars,
			ErrLog:         c.Globals.ErrLog,
			Output:         out,
			Spinner:        spinner,
			SpinnerMessage: msg,
			Timeout:        0, // zero indicates no timeout
			Verbose:        c.Globals.Flags.Verbose,
		})
		if err != nil {
			// In verbose mode we'll have the failure status AFTER the error output.
			// But we can't just call StopFailMessage() without first starting the spinner.
			if c.Globals.Flags.Verbose {
				text.Break(out)
				spinErr := spinner.Start()
				if spinErr != nil {
					return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
				spinner.Message(msg + "...")
				spinner.StopFailMessage(msg)
				spinErr = spinner.StopFail()
				if spinErr != nil {
					return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
			}
			return err
		}

		// In verbose mode we'll have the failure status AFTER the error output.
		// But we can't just call StopMessage() without first starting the spinner.
		if c.Globals.Flags.Verbose {
			err = spinner.Start()
			if err != nil {
				return err
			}
			spinner.Message(msg + "...")
			text.Break(out)
		}

		spinner.StopMessage(msg)
		err = spinner.Stop()
		if err != nil {
			return err
		}
	}

	displayInitOutput(mf.Name, dst, language.Name, out)
	return nil
}

// verifyDirectory indicates if the user wants to continue with the execution
// flow when presented with a prompt that suggests the current directory isn't
// empty.
func verifyDirectory(flags global.Flags, dir string, out io.Writer, in io.Reader) (cont, notEmpty bool, err error) {
	if dir == "" {
		dir = "."
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return false, false, err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return false, false, err
	}

	if strings.Contains(dir, " ") && !flags.Quiet {
		text.Warning(out, "Your project path contains spaces. In some cases this can result in issues with your installed language toolchain, e.g. `npm`. Consider removing any spaces.\n\n")
	}

	if len(files) > 0 && !flags.AutoYes && !flags.NonInteractive {
		label := fmt.Sprintf("The current directory isn't empty. Are you sure you want to initialize a Compute project in %s? [y/N] ", dir)
		result, err := text.AskYesNo(out, label, in)
		if err != nil {
			return false, true, err
		}
		return result, true, nil
	}

	return true, false, nil
}

// verifyDestination checks the provided path exists and is a directory.
//
// NOTE: For validating user permissions it will create a temporary file within
// the directory and then remove it before returning the absolute path to the
// directory itself.
func verifyDestination(path string, spinner text.Spinner) (dst string, err error) {
	dst, err = filepath.Abs(path)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(dst)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return dst, fmt.Errorf("couldn't verify package directory: %w", err) // generic error
	}
	if err == nil && !fi.IsDir() {
		return dst, fmt.Errorf("package destination is not a directory") // specific problem
	}
	if err != nil && errors.Is(err, fs.ErrNotExist) { // normal-ish case
		err := spinner.Process(fmt.Sprintf("Creating %s", dst), func(_ *text.SpinnerWrapper) error {
			if err := os.MkdirAll(dst, 0o700); err != nil {
				return fmt.Errorf("error creating package destination: %w", err)
			}
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	return dst, nil
}

func validateDirectoryPermissions(dst string) text.SpinnerProcess {
	return func(_ *text.SpinnerWrapper) error {
		tmpname := make([]byte, 16)
		n, err := rand.Read(tmpname)
		if err != nil {
			return fmt.Errorf("error generating random filename: %w", err)
		}
		if n != 16 {
			return fmt.Errorf("failed to generate enough entropy (%d/%d)", n, 16)
		}

		// gosec flagged this:
		// G304 (CWE-22): Potential file inclusion via variable
		//
		// Disabling as the input is determined by our own package.
		// #nosec
		f, err := os.Create(filepath.Join(dst, fmt.Sprintf("tmp_%x", tmpname)))
		if err != nil {
			return fmt.Errorf("error creating file in package destination: %w", err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("error closing file in package destination: %w", err)
		}

		if err := os.Remove(f.Name()); err != nil {
			return fmt.Errorf("error removing file in package destination: %w", err)
		}
		return nil
	}
}

// promptOrReturn will prompt the user for information missing from the
// fastly.toml manifest file, otherwise if it already exists then the value is
// returned as is.
func promptOrReturn(
	flags global.Flags,
	m manifest.Data,
	path, email string,
	in io.Reader,
	out io.Writer,
) (name, description string, authors []string, err error) {
	name, _ = m.Name()
	description, _ = m.Description()
	authors, _ = m.Authors()

	if name == "" && !flags.AcceptDefaults && !flags.NonInteractive {
		text.Break(out)
	}
	name, err = promptPackageName(flags, name, path, in, out)
	if err != nil {
		return "", description, authors, err
	}

	if description == "" && !flags.AcceptDefaults && !flags.NonInteractive {
		text.Break(out)
	}
	description, err = promptPackageDescription(flags, description, in, out)
	if err != nil {
		return name, "", authors, err
	}

	if len(authors) == 0 && !flags.AcceptDefaults && !flags.NonInteractive {
		text.Break(out)
	}
	authors, err = promptPackageAuthors(flags, authors, email, in, out)
	if err != nil {
		return name, description, []string{}, err
	}

	return name, description, authors, nil
}

// promptPackageName prompts the user for a package name unless already defined either
// via the corresponding CLI flag or the manifest file.
//
// It will use a default of the current directory path if no value provided by
// the user via the prompt.
func promptPackageName(flags global.Flags, name string, dirPath string, in io.Reader, out io.Writer) (string, error) {
	defaultName := filepath.Base(dirPath)

	if name == "" && (flags.AcceptDefaults || flags.NonInteractive) {
		return defaultName, nil
	}

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

// promptPackageDescription prompts the user for a package description unless already
// defined either via the corresponding CLI flag or the manifest file.
func promptPackageDescription(flags global.Flags, desc string, in io.Reader, out io.Writer) (string, error) {
	if desc == "" && (flags.AcceptDefaults || flags.NonInteractive) {
		return desc, nil
	}

	if desc == "" {
		var err error

		desc, err = text.Input(out, "Description: ", in)
		if err != nil {
			return "", fmt.Errorf("error reading input: %w", err)
		}
	}

	return desc, nil
}

// promptPackageAuthors prompts the user for a package name unless already defined
// either via the corresponding CLI flag or the manifest file.
//
// It will use a default of the user's email found within the manifest, if set
// there, otherwise the value will be an empty slice.
//
// FIXME: Handle prompting for multiple authors.
func promptPackageAuthors(flags global.Flags, authors []string, manifestEmail string, in io.Reader, out io.Writer) ([]string, error) {
	defaultValue := []string{manifestEmail}
	if len(authors) == 0 && (flags.AcceptDefaults || flags.NonInteractive) {
		return defaultValue, nil
	}
	if len(authors) == 0 {
		label := "Author (email): "

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
			authors = defaultValue
		}
	}

	return authors, nil
}

// promptForLanguage prompts the user for a package language unless already
// defined either via the corresponding CLI flag or the manifest file.
func promptForLanguage(flags global.Flags, languages []*Language, in io.Reader, out io.Writer) (*Language, error) {
	var (
		language *Language
		option   string
		err      error
	)

	if !flags.AcceptDefaults && !flags.NonInteractive {
		text.Output(out, "\n%s", text.Bold("Language:"))
		text.Output(out, "(Find out more about language support at https://developer.fastly.com/learning/compute)")
		for i, lang := range languages {
			text.Output(out, "[%d] %s", i+1, lang.DisplayName)
		}

		text.Break(out)
		option, err = text.Input(out, "Choose option: [1] ", in, validateLanguageOption(languages))
		if err != nil {
			return nil, fmt.Errorf("reading input %w", err)
		}
	}

	if option == "" {
		option = "1"
	}

	i, err := strconv.Atoi(option)
	if err != nil {
		return nil, fmt.Errorf("failed to identify chosen language")
	}
	language = languages[i-1]

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

// promptForStarterKit prompts the user for a package starter kit.
//
// It returns the path to the starter kit, and the corresponding branch/tag,
func promptForStarterKit(flags global.Flags, kits []config.StarterKit, in io.Reader, out io.Writer) (from string, branch string, tag string, err error) {
	var option string

	if !flags.AcceptDefaults && !flags.NonInteractive {
		text.Output(out, "\n%s", text.Bold("Starter kit:"))
		for i, kit := range kits {
			fmt.Fprintf(out, "[%d] %s\n", i+1, text.Bold(kit.Name))
			text.Indent(out, 4, "%s\n%s", kit.Description, kit.Path)
		}
		text.Info(out, "\nFor a complete list of Starter Kits:")
		text.Indent(out, 4, "https://developer.fastly.com/solutions/starters/")
		text.Break(out)

		option, err = text.Input(out, "Choose option or paste git URL: [1] ", in, validateTemplateOptionOrURL(kits))
		if err != nil {
			return "", "", "", fmt.Errorf("error reading input: %w", err)
		}
		text.Break(out)
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
	c *InitCommand,
	branch, tag string,
	archives []file.Archive,
	spinner text.Spinner,
	out io.Writer,
) error {
	err := spinner.Start()
	if err != nil {
		return err
	}
	msg := "Fetching package template"
	spinner.Message(msg + "...")

	// If the user has provided a local file path, we'll recursively copy the
	// directory to c.dir.
	if fi, err := os.Stat(c.cloneFrom); err == nil && fi.IsDir() {
		if err := cp.Copy(c.cloneFrom, c.dir); err != nil {
			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
			}
			return err
		}
		spinner.StopMessage(msg)
		return spinner.Stop()
	}
	c.Globals.ErrLog.Add(err)

	req, err := http.NewRequest("GET", c.cloneFrom, nil)
	if err != nil {
		err = fmt.Errorf("failed to construct package request URL: %w", err)
		c.Globals.ErrLog.Add(err)

		if gitRepositoryRegEx.MatchString(c.cloneFrom) {
			if err := clonePackageFromEndpoint(c.cloneFrom, branch, tag, c.dir); err != nil {
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
				return err
			}
			spinner.StopMessage(msg)
			return spinner.Stop()
		}

		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
		}
		return err
	}

	for _, archive := range archives {
		for _, mime := range archive.MimeTypes() {
			req.Header.Add("Accept", mime)
		}
	}

	res, err := c.Globals.HTTPClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get package: %w", err)
		c.Globals.ErrLog.Add(err)
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
		}
		return err
	}
	defer res.Body.Close() // #nosec G307

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("failed to get package: %s", res.Status)
		c.Globals.ErrLog.Add(err)
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
		}
		return err
	}

	filename := filepath.Base(c.cloneFrom)
	ext := filepath.Ext(filename)

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as we require a user to configure their own environment.
	/* #nosec */
	f, err := os.Create(filename)
	if err != nil {
		err = fmt.Errorf("failed to create local %s archive: %w", filename, err)
		c.Globals.ErrLog.Add(err)
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
		}
		return err
	}
	defer func() {
		// NOTE: Later on we rename the file to include an extension and the
		// following call to os.Remove works still because the `filename` variable
		// that is still in scope is also updated to include the extension.
		err := os.Remove(filename)
		if err != nil {
			c.Globals.ErrLog.Add(err)
			text.Info(out, "We were unable to clean-up the local %s file (it can be safely removed)", filename)
		}
	}()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		err = fmt.Errorf("failed to write %s archive to disk: %w", filename, err)
		c.Globals.ErrLog.Add(err)
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
		}
		return err
	}

	// NOTE: We used to `defer` the closing of the file after its creation but
	// realised that this caused issues on Windows as it was unable to rename the
	// file as we still have the descriptor `f` open.
	if err := f.Close(); err != nil {
		c.Globals.ErrLog.Add(err)
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
				c.Globals.ErrLog.Add(err)
				spinner.StopFailMessage(msg)
				spinErr := spinner.StopFail()
				if spinErr != nil {
					return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
				}
				return err
			}
			filename = filenameWithExt
		}

		archive.SetDestination(c.dir)
		archive.SetFilename(filename)

		err = archive.Extract()
		if err != nil {
			err = fmt.Errorf("failed to extract %s archive content: %w", filename, err)
			c.Globals.ErrLog.Add(err)
			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
			}
			return err
		}

		spinner.StopMessage(msg)
		return spinner.Stop()
	}

	if err := clonePackageFromEndpoint(c.cloneFrom, branch, tag, c.dir); err != nil {
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
		}
		return err
	}

	spinner.StopMessage(msg)
	return spinner.Stop()
}

// clonePackageFromEndpoint clones the given repo (from) into a temp directory,
// then copies specific files to the destination directory (path).
func clonePackageFromEndpoint(
	from string,
	branch string,
	tag string,
	dst string,
) error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fsterr.RemediationError{
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
	c := exec.Command("git", args...)

	// nosemgrep (invalid-usage-of-modified-variable)
	stdoutStderr, err := c.CombinedOutput()
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

		return filesystem.CopyFile(path, dst)
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
	spinner text.Spinner,
	path, name, desc string,
	authors []string,
	language *Language,
) (manifest.File, error) {
	var returnEarly bool
	mp := filepath.Join(path, manifest.Filename)

	err := spinner.Process("Reading fastly.toml", func(_ *text.SpinnerWrapper) error {
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
						return fmt.Errorf("error saving fastly.toml: %w", err)
					}
					returnEarly = true
					return nil // EXIT updateManifest
				}
			}
			return fmt.Errorf("error reading fastly.toml: %w", err)
		}
		return nil
	})
	if err != nil {
		return m, err
	}
	if returnEarly {
		return m, nil
	}

	err = spinner.Process(fmt.Sprintf("Setting package name in manifest to %q", name), func(_ *text.SpinnerWrapper) error {
		m.Name = name
		return nil
	})
	if err != nil {
		return m, err
	}

	var descMsg string
	if desc != "" {
		descMsg = " to '" + desc + "'"
	}

	err = spinner.Process(fmt.Sprintf("Setting description in manifest%s", descMsg), func(_ *text.SpinnerWrapper) error {
		// NOTE: We allow an empty description to be set.
		m.Description = desc
		return nil
	})
	if err != nil {
		return m, err
	}

	if len(authors) > 0 {
		err = spinner.Process(fmt.Sprintf("Setting authors in manifest to '%s'", strings.Join(authors, ", ")), func(_ *text.SpinnerWrapper) error {
			m.Authors = authors
			return nil
		})
		if err != nil {
			return m, err
		}
	}

	if language != nil {
		err = spinner.Process(fmt.Sprintf("Setting language in manifest to '%s'", language.Name), func(_ *text.SpinnerWrapper) error {
			m.Language = language.Name
			return nil
		})
		if err != nil {
			return m, err
		}
	}

	err = spinner.Process("Saving manifest changes", func(_ *text.SpinnerWrapper) error {
		if err := m.Write(mp); err != nil {
			return fmt.Errorf("error saving fastly.toml: %w", err)
		}
		return nil
	})
	if err != nil {
		return m, err
	}
	return m, nil
}

// initializeLanguage for newly cloned package.
func initializeLanguage(spinner text.Spinner, language *Language, languages []*Language, name, wd, path string) (*Language, error) {
	err := spinner.Process("Initializing package", func(_ *text.SpinnerWrapper) error {
		if wd != path {
			err := os.Chdir(path)
			if err != nil {
				return fmt.Errorf("error changing to your project directory: %w", err)
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
				return fmt.Errorf("unrecognised package language")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return language, nil
}

// promptForPostInitContinue ensures the user is happy to continue with running
// the define post_init script in the fastly.toml manifest file.
func promptForPostInitContinue(msg, script string, out io.Writer, in io.Reader) error {
	text.Info(out, "\n%s:\n", msg)
	text.Indent(out, 4, "%s", script)

	label := "\nDo you want to run this now? [y/N] "
	answer, err := text.AskYesNo(out, label, in)
	if err != nil {
		return err
	}
	if !answer {
		return fsterr.ErrPostInitStopped
	}
	text.Break(out)
	return nil
}

// displayInitOutput of package information and useful links.
func displayInitOutput(name, dst, language string, out io.Writer) {
	text.Break(out)
	text.Description(out, fmt.Sprintf("Initialized package %s to", text.Bold(name)), dst)

	if language == "other" {
		text.Description(out, "To package a pre-compiled Wasm binary for deployment, run", "fastly compute pack")
		text.Description(out, "To deploy the package, run", "fastly compute deploy")
	} else {
		text.Description(out, "To publish the package (build and deploy), run", "fastly compute publish")
	}

	text.Description(out, "To learn about deploying Compute projects using third-party orchestration tools, visit", "https://developer.fastly.com/learning/integrations/orchestration/")
	text.Success(out, "Initialized package %s", text.Bold(name))
}
