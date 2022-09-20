package compute

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// GoCompilation is a language specific compilation target that converts the
// language code into a Wasm binary.
const GoCompilation = "tinygo"

// GoCompilationURL is the official TinyGo website URL.
const GoCompilationURL = "https://tinygo.org"

// GoCompilationTargetCommand is the shell command for returning the tinygo
// version.
const GoCompilationTargetCommand = "tinygo version"

// GoConstraints is the set of supported toolchain and compilation versions.
//
// NOTE: Two keys are supported: "toolchain" and "compilation", with the latter
// being optional as not all language compilation steps are separate tools from
// the toolchain itself.
var GoConstraints = make(map[string]string)

// GoInstaller is the command used to install the dependencies defined within
// the Go language manifest.
const GoInstaller = "go mod download"

// GoManifest is the manifest file for defining project configuration.
const GoManifest = "go.mod"

// GoManifestRemediation is a error remediation message for a missing manifest.
const GoManifestRemediation = "go mod init"

// GoSDK is the required Compute@Edge SDK.
// https://pkg.go.dev/github.com/fastly/compute-sdk-go
const GoSDK = "fastly/compute-sdk-go"

// GoSourceDirectory represents the source code directory.                                               │                                                           │
const GoSourceDirectory = "."

// GoToolchain is the executable responsible for managing dependencies.
const GoToolchain = "go"

// GoToolchainURL is the official Go website URL.
const GoToolchainURL = "https://go.dev/"

// GoToolchainVersionCommand is the shell command for returning the go version.
const GoToolchainVersionCommand = "go version"

// NewGo constructs a new Go toolchain.
func NewGo(
	pkgName string,
	scripts manifest.Scripts,
	errlog fsterr.LogInterface,
	timeout int,
	cfg config.Go,
	out io.Writer,
) *Go {
	GoConstraints["toolchain"] = cfg.ToolchainConstraint
	GoConstraints["compilation"] = cfg.TinyGoConstraint

	return &Go{
		Shell:     Shell{},
		build:     scripts.Build,
		errlog:    errlog,
		pkgName:   pkgName,
		postBuild: scripts.PostBuild,
		timeout:   timeout,
		validator: ToolchainValidator{
			Compilation:              GoCompilation,
			CompilationTargetCommand: GoCompilationTargetCommand,
			CompilationTargetPattern: regexp.MustCompile(`tinygo version (?P<version>\d[^\s]+)`),
			CompilationURL:           GoCompilationURL,
			Constraints:              GoConstraints,
			ErrLog:                   errlog,
			Installer:                GoInstaller,
			Manifest:                 GoManifest,
			ManifestRemediation:      GoManifestRemediation,
			Output:                   out,
			SDK:                      GoSDK,
			Toolchain:                GoToolchain,
			ToolchainVersionCommand:  GoToolchainVersionCommand,
			ToolchainVersionPattern:  regexp.MustCompile(`go version go(?P<version>\d[^\s]+)`),
			ToolchainURL:             GoToolchainURL,
		},
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

	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// pkgName is the name of the package (also used as the module name).
	pkgName string
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// timeout is the build execution threshold.
	timeout int
	// validator is an abstraction to validate required resources are installed.
	validator ToolchainValidator
}

// Initialize handles any non-build related set-up.
func (g Go) Initialize(_ io.Writer) error {
	if err := g.setPackageName(GoManifest); err != nil {
		g.errlog.Add(err)
		return fmt.Errorf("error updating %s manifest: %w", GoManifest, err)
	}
	return nil
}

// Verify ensures the user's environment has all the required resources/tools.
func (g *Go) Verify(_ io.Writer) error {
	return g.validator.Validate()
}

// Build compiles the user's source code into a Wasm binary.
func (g *Go) Build(out io.Writer, progress text.Progress, verbose bool, callback func() error) error {
	return build(language{
		buildScript: g.build,
		buildFn:     g.Shell.Build,
		errlog:      g.errlog,
		pkgName:     g.pkgName,
		postBuild:   g.postBuild,
		timeout:     g.timeout,
	}, out, progress, verbose, nil, callback)
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
