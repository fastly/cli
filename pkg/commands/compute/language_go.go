package compute

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	fstexec "github.com/fastly/cli/pkg/exec"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// GoSourceDirectory represents the source code directory (i.e. root directory).
const GoSourceDirectory = "."

// GoManifestName represents the language file for configuring dependencies.
const GoManifestName = "go.mod"

// NewGo constructs a new Go toolchain.
func NewGo(pkgName string, scripts manifest.Scripts, errlog fsterr.LogInterface, timeout int, cfg config.Go) *Go {
	return &Go{
		Shell:     Shell{},
		build:     scripts.Build,
		compiler:  "tinygo",
		config:    cfg,
		errlog:    errlog,
		pkgName:   pkgName,
		postBuild: scripts.PostBuild,
		timeout:   timeout,
		toolchain: "go",
	}
}

// Go implements a Toolchain for the TinyGo language.
//
// NOTE: Two separate tools are required to support golang development.
//
// 1. Go: for defining required packages in a go.mod project module.
// 2. TinyGo: used to compile the go project.
type Go struct {
	Shell

	// build is a custom build script defined in fastly.toml using [scripts.build].
	build string
	// compiler is a WASM/WASI capable compiler (i.e. not the standard go compiler)
	compiler string
	// config is Go configuration such as toolchain constraints.
	config config.Go
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// pkgName is the name of the package (also used as the module name).
	pkgName string
	// postBuild is a custom script executed after the build but before the WASM
	// binary is added to the .tar.gz archive.
	postBuild string
	// timeout is the build execution threshold.
	timeout int
	// toolchain is the go executable.
	toolchain string
}

// Initialize implements the Toolchain interface and initializes a newly cloned
// package by installing required dependencies.
func (g Go) Initialize(out io.Writer) error {
	// Remediation used in variation sections.
	goURL := "https://go.dev/"
	remediation := fmt.Sprintf("To fix this error, install %s by visiting:\n\n\t$ %s\n\nThen execute:\n\n\t$ fastly compute init", g.toolchain, text.Bold(goURL))

	var (
		bin string
		err error
	)

	// 1. Check go command is on $PATH.
	{
		fmt.Fprintf(out, "Checking if %s is installed...\n", g.toolchain)

		bin, err = exec.LookPath(g.toolchain)
		if err != nil {
			g.errlog.Add(err)

			return fsterr.RemediationError{
				Inner:       fmt.Errorf("`%s` not found in $PATH", g.toolchain),
				Remediation: remediation,
			}
		}

		fmt.Fprintf(out, "Found %s at %s\n", g.toolchain, bin)
	}

	// 2. Check go version is correct.
	{
		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
		// Disabling as we trust the source of the variable.
		/* #nosec */
		cmd := exec.Command(bin, "version") // e.g. go version go1.18 darwin/amd64
		stdoutStderr, err := cmd.CombinedOutput()
		output := string(stdoutStderr)
		if err != nil {
			if len(stdoutStderr) > 0 {
				err = fmt.Errorf("%w: %s", err, strings.TrimSpace(output))
			}
			g.errlog.Add(err)
			return err
		}

		segs := strings.Split(output, " ")
		if len(segs) < 3 {
			return errors.New("unexpected go version output")
		}
		version := segs[2][2:]

		v, err := semver.NewVersion(version)
		if err != nil {
			return fmt.Errorf("error parsing version output %s into a semver: %w", version, err)
		}

		c, err := semver.NewConstraint(g.config.ToolchainConstraint)
		if err != nil {
			return fmt.Errorf("error parsing toolchain constraint %s into a semver: %w", g.config.ToolchainConstraint, err)
		}

		if !c.Check(v) {
			err := fsterr.RemediationError{
				Inner:       fmt.Errorf("version %s didn't meet the constraint %s", version, g.config.ToolchainConstraint),
				Remediation: remediation,
			}
			g.errlog.Add(err)
			return err
		}
	}

	// 3. Set package name.
	{
		m, err := filepath.Abs(GoManifestName)
		if err != nil {
			g.errlog.Add(err)
			return fmt.Errorf("getting %s path: %w", JSManifestName, err)
		}

		if !filesystem.FileExists(m) {
			msg := fmt.Sprintf(fsterr.FormatTemplate, text.Bold("go mod init"))
			remediation := fmt.Sprintf("%s\n\nThen execute:\n\n\t$ fastly compute init", msg)
			err := fsterr.RemediationError{
				Inner:       fmt.Errorf("%s not found", JSManifestName),
				Remediation: remediation,
			}
			g.errlog.Add(err)
			return err
		}

		if err := g.setPackageName(GoManifestName); err != nil {
			g.errlog.Add(err)
			return fmt.Errorf("error updating %s manifest: %w", GoManifestName, err)
		}

		fmt.Fprintf(out, "Found %s at %s\n", GoManifestName, m)
	}

	// 4. Download dependencies.
	{
		fmt.Fprintf(out, "Installing package dependencies...\n")
		cmd := fstexec.Streaming{
			Command: "go",
			Args:    []string{"mod", "download"},
			Env:     os.Environ(),
			Output:  out,
		}
		return cmd.Exec()
	}
}

// Verify implements the Toolchain interface and verifies whether the Go
// language toolchain is correctly configured on the host.
func (g *Go) Verify(out io.Writer) error {
	// Remediation used in variation sections.
	tinygoURL := "https://tinygo.org"
	remediation := fmt.Sprintf("To fix this error, install %s by visiting:\n\n\t$ %s", g.compiler, text.Bold(tinygoURL))

	var (
		bin string
		err error
	)

	// 1. Check tinygo command is on $PATH.
	{
		fmt.Fprintf(out, "Checking if %s is installed...\n", g.compiler)

		bin, err = exec.LookPath(g.compiler)
		if err != nil {
			g.errlog.Add(err)

			return fsterr.RemediationError{
				Inner:       fmt.Errorf("`%s` not found in $PATH", g.compiler),
				Remediation: remediation,
			}
		}

		fmt.Fprintf(out, "Found %s at %s\n", g.compiler, bin)
	}

	// 2. Check tinygo version is correct.
	{
		// gosec flagged this:
		// G204 (CWE-78): Subprocess launched with function call as argument or cmd arguments
		// Disabling as we trust the source of the variable.
		/* #nosec */
		cmd := exec.Command(bin, "version") // e.g. tinygo version 0.24.0 darwin/amd64 (using go version go1.18 and LLVM version 14.0.0)
		stdoutStderr, err := cmd.CombinedOutput()
		output := string(stdoutStderr)
		if err != nil {
			if len(stdoutStderr) > 0 {
				err = fmt.Errorf("%w: %s", err, strings.TrimSpace(output))
			}
			g.errlog.Add(err)
			return err
		}

		segs := strings.Split(output, " ")
		if len(segs) < 3 {
			return errors.New("unexpected tinygo version output")
		}
		version := segs[2]

		v, err := semver.NewVersion(version)
		if err != nil {
			return fmt.Errorf("error parsing version output %s into a semver: %w", version, err)
		}

		c, err := semver.NewConstraint(g.config.TinyGoConstraint)
		if err != nil {
			return fmt.Errorf("error parsing toolchain constraint %s into a semver: %w", g.config.TinyGoConstraint, err)
		}

		if !c.Check(v) {
			err := fsterr.RemediationError{
				Inner:       fmt.Errorf("version %s didn't meet the constraint %s", version, g.config.TinyGoConstraint),
				Remediation: remediation,
			}
			g.errlog.Add(err)
			return err
		}
	}
	return nil
}

// Build implements the Toolchain interface and attempts to compile the package
// Go source to a Wasm binary.
func (g *Go) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	cmd := g.compiler
	args := []string{
		"build",
		"-target=wasi",
		"-wasm-abi=generic",
		"-gc=conservative",
		"-o=bin/main.wasm",
		fmt.Sprintf("./%s", GoSourceDirectory),
	}

	// A bin directory is required.
	dir, err := os.Getwd()
	if err != nil {
		g.errlog.Add(err)
		return fmt.Errorf("getting current working directory: %w", err)
	}
	binDir := filepath.Join(dir, "bin")
	if err := filesystem.MakeDirectoryIfNotExists(binDir); err != nil {
		g.errlog.Add(err)
		return fmt.Errorf("creating bin directory: %w", err)
	}

	if g.build != "" {
		cmd, args = g.Shell.Build(g.build)
	}

	err = g.execCommand(cmd, args, out, progress, verbose)
	if err != nil {
		return err
	}

	// NOTE: We set the progress indicator to Done() so that any output we now
	// print via the post_build callback doesn't get hidden by the progress status.
	// The progress is 'reset' inside the main build controller `build.go`.
	progress.Done()

	if g.postBuild != "" {
		if err = callback(); err == nil {
			cmd, args := g.Shell.Build(g.postBuild)
			err := g.execCommand(cmd, args, out, progress, verbose)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (g Go) execCommand(cmd string, args []string, out, progress io.Writer, verbose bool) error {
	s := fstexec.Streaming{
		Command:  cmd,
		Args:     args,
		Env:      os.Environ(),
		Output:   out,
		Progress: progress,
		Verbose:  verbose,
	}
	if g.timeout > 0 {
		s.Timeout = time.Duration(g.timeout) * time.Second
	}
	if err := s.Exec(); err != nil {
		g.errlog.Add(err)
		return err
	}
	return nil
}

// setPackageName into go.mod manifest.
//
// NOTE: The implementation scans the go.mod line-by-line looking for the
// module directive (typically the first line, but not guaranteed) and replaces
// the module path with the user's configured package name.
func (g Go) setPackageName(path string) (err error) {
	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as we require a user to configure their own environment.
	/* #nosec */
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	var b bytes.Buffer

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "module ") {
			line = fmt.Sprintf("module %s", g.pkgName)
		}
		b.WriteString(line + "\n")
	}
	if err := s.Err(); err != nil {
		return err
	}

	err = os.WriteFile(path, b.Bytes(), fi.Mode())
	if err != nil {
		return err
	}

	return nil
}
