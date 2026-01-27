package compute

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

// JsDefaultBuildCommand is a build command compiled into the CLI binary so it
// can be used as a fallback for customer's who have an existing Compute project and
// are simply upgrading their CLI version and might not be familiar with the
// changes in the 4.0.0 release with regards to how build logic has moved to the
// fastly.toml manifest.
//
// NOTE: In the 5.x CLI releases we persisted the default to the fastly.toml
// We no longer do that. In 6.x we use the default and just inform the user.
// This makes the experience less confusing as users didn't expect file changes.
var JsDefaultBuildCommand = fmt.Sprintf("npm exec js-compute-runtime ./src/index.js %s", binWasmPath)

// JsDefaultBuildCommandForWebpack is a build command compiled into the CLI
// binary so it can be used as a fallback for customer's who have an existing
// Compute project using the 'default' JS Starter Kit, and are simply upgrading
// their CLI version and might not be familiar with the changes in the 4.0.0
// release with regards to how build logic has moved to the fastly.toml manifest.
//
// NOTE: For this variation of the build script to be added to the user's
// fastly.toml will require a successful check for the webpack dependency.
var JsDefaultBuildCommandForWebpack = fmt.Sprintf("npm exec webpack && npm exec js-compute-runtime ./bin/index.js %s", binWasmPath)

// JsSourceDirectory represents the source code directory.
const JsSourceDirectory = "src"

// ErrNpmMissing is returned when Node.js is found but npm is not installed.
var ErrNpmMissing = errors.New("node found but npm missing")

// JSRuntime represents a detected JavaScript runtime.
type JSRuntime struct {
	// Name is the runtime name (node or bun).
	Name string
	// Version is the runtime version string.
	Version string
	// PkgMgr is the package manager to use (npm or bun).
	PkgMgr string
}

// NewJavaScript constructs a new JavaScript toolchain.
func NewJavaScript(
	c *BuildCommand,
	in io.Reader,
	manifestFilename string,
	out io.Writer,
	spinner text.Spinner,
) *JavaScript {
	return &JavaScript{
		Shell: Shell{},

		autoYes:               c.Globals.Flags.AutoYes,
		build:                 c.Globals.Manifest.File.Scripts.Build,
		env:                   c.Globals.Manifest.File.Scripts.EnvVars,
		errlog:                c.Globals.ErrLog,
		input:                 in,
		manifestFilename:      manifestFilename,
		metadataFilterEnvVars: c.MetadataFilterEnvVars,
		nonInteractive:        c.Globals.Flags.NonInteractive,
		output:                out,
		postBuild:             c.Globals.Manifest.File.Scripts.PostBuild,
		spinner:               spinner,
		timeout:               c.Flags.Timeout,
		verbose:               c.Globals.Verbose(),
	}
}

// JavaScript implements a Toolchain for the JavaScript language.
type JavaScript struct {
	Shell

	// autoYes is the --auto-yes flag.
	autoYes bool
	// build is a shell command defined in fastly.toml using [scripts.build].
	build string
	// defaultBuild indicates if the default build script was used.
	defaultBuild bool
	// env is environment variables to be set.
	env []string
	// errlog is an abstraction for recording errors to disk.
	errlog fsterr.LogInterface
	// input is the user's terminal stdin stream
	input io.Reader
	// manifestFilename is the name of the manifest file.
	manifestFilename string
	// metadataFilterEnvVars is a comma-separated list of user defined env vars.
	metadataFilterEnvVars string
	// nodeModulesDir is the resolved path to node_modules (may be in parent dir for monorepos).
	nodeModulesDir string
	// nonInteractive is the --non-interactive flag.
	nonInteractive bool
	// output is the users terminal stdout stream
	output io.Writer
	// pkgDir is the resolved directory containing package.json.
	pkgDir string
	// postBuild is a custom script executed after the build but before the Wasm
	// binary is added to the .tar.gz archive.
	postBuild string
	// runtime is the detected JavaScript runtime (node or bun).
	runtime *JSRuntime
	// spinner is a terminal progress status indicator.
	spinner text.Spinner
	// timeout is the build execution threshold.
	timeout int
	// verbose indicates if the user set --verbose
	verbose bool
}

// DefaultBuildScript indicates if a custom build script was used.
func (j *JavaScript) DefaultBuildScript() bool {
	return j.defaultBuild
}

// JavaScriptPackage represents a package within a JavaScript lockfile.
type JavaScriptPackage struct {
	Version string `json:"version"`
}

// JavaScriptLockFile represents a JavaScript lockfile.
type JavaScriptLockFile struct {
	Packages map[string]JavaScriptPackage `json:"packages"`
}

// Dependencies returns all dependencies used by the project.
func (j *JavaScript) Dependencies() map[string]string {
	deps := make(map[string]string)

	lockfile := "npm-shrinkwrap.json"
	_, err := os.Stat(lockfile)
	if errors.Is(err, os.ErrNotExist) {
		lockfile = "package-lock.json"
	}

	var jlf JavaScriptLockFile
	if f, err := os.Open(lockfile); err == nil {
		if err := json.NewDecoder(f).Decode(&jlf); err == nil {
			for k, v := range jlf.Packages {
				if k != "" { // avoid "root" package
					deps[k] = v.Version
				}
			}
		}
	}

	return deps
}

// Build compiles the user's source code into a Wasm binary.
func (j *JavaScript) Build() error {
	if j.build == "" {
		// Only verify toolchain when using default build (no custom [scripts.build])
		if err := j.verifyToolchain(); err != nil {
			return err
		}
		cmd, err := j.getDefaultBuildCommand()
		if err != nil {
			return err
		}
		j.build = cmd
		j.defaultBuild = true
	}

	if j.defaultBuild && j.verbose {
		text.Info(j.output, "No [scripts.build] found in %s. The following default build command for JavaScript will be used: `%s`\n\n", j.manifestFilename, j.build)
	}

	bt := BuildToolchain{
		autoYes:               j.autoYes,
		buildFn:               j.Shell.Build,
		buildScript:           j.build,
		env:                   j.env,
		errlog:                j.errlog,
		in:                    j.input,
		manifestFilename:      j.manifestFilename,
		metadataFilterEnvVars: j.metadataFilterEnvVars,
		nonInteractive:        j.nonInteractive,
		out:                   j.output,
		postBuild:             j.postBuild,
		spinner:               j.spinner,
		timeout:               j.timeout,
		verbose:               j.verbose,
	}

	return bt.Build()
}

func (j JavaScript) checkForWebpack() (bool, error) {
	wd, err := os.Getwd()
	if err != nil {
		return false, err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	found, path, err := search("package.json", wd, home)
	if err != nil {
		return false, err
	}

	if found {
		// gosec flagged this:
		// G304 (CWE-22): Potential file inclusion via variable
		//
		// Disabling as the path is determined by our own logic.
		/* #nosec */
		data, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}

		var pkg NPMPackage

		err = json.Unmarshal(data, &pkg)
		if err != nil {
			return false, err
		}

		for k := range pkg.DevDependencies {
			if k == "webpack" {
				return true, nil
			}
		}

		for k := range pkg.Dependencies {
			if k == "webpack" {
				return true, nil
			}
		}
	}

	return false, nil
}

// search recurses up the directory tree looking for the given file.
func search(filename, wd, home string) (found bool, path string, err error) {
	parent := filepath.Dir(wd)

	var noManifest bool
	path = filepath.Join(wd, filename)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		noManifest = true
	}

	// We've found the manifest.
	if !noManifest {
		return true, path, nil
	}

	// NOTE: The first condition catches if we reach the user's 'root' directory.
	if wd != parent && wd != home {
		return search(filename, parent, home)
	}

	return false, "", nil
}

// NPMPackage represents a package.json manifest and its dependencies.
type NPMPackage struct {
	DevDependencies map[string]string `json:"devDependencies"`
	Dependencies    map[string]string `json:"dependencies"`
}

// checkBun checks if Bun is installed and returns runtime info.
func (j *JavaScript) checkBun() (*JSRuntime, error) {
	if _, err := exec.LookPath("bun"); err != nil {
		return nil, err
	}
	cmd := exec.Command("bun", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return &JSRuntime{
		Name:    "bun",
		Version: strings.TrimSpace(string(output)),
		PkgMgr:  "bun",
	}, nil
}

// checkNode checks if Node.js and npm are installed and returns runtime info.
func (j *JavaScript) checkNode() (*JSRuntime, error) {
	if _, err := exec.LookPath("node"); err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("npm"); err != nil {
		return nil, ErrNpmMissing
	}
	nodeCmd := exec.Command("node", "--version")
	nodeOutput, err := nodeCmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return &JSRuntime{
		Name:    "node",
		Version: strings.TrimSpace(string(nodeOutput)),
		PkgMgr:  "npm",
	}, nil
}

// detectProjectRuntime checks lockfiles to determine which runtime the project uses.
// Searches from package.json location upward to handle workspace setups where
// bun.lockb is at the workspace root but package.json is in a subpackage.
// Only accepts bun.lockb if it's alongside a package.json (same dir) to avoid
// picking up unrelated lockfiles in parent directories.
// Returns "bun" if bun.lockb exists, "node" otherwise (default).
func (j *JavaScript) detectProjectRuntime() string {
	wd, err := os.Getwd()
	if err != nil {
		return "node"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "node"
	}

	// Find package.json first to locate the project/subpackage root
	found, pkgPath, err := search("package.json", wd, home)
	if err != nil || !found {
		return "node"
	}

	// Search upward from package.json for bun.lockb (handles workspaces)
	// Only accept bun.lockb if the same directory also has package.json
	// (ensures we're in a proper Bun project/workspace, not picking up unrelated lockfiles)
	dir := filepath.Dir(pkgPath)
	for {
		hasBunLock := false
		for _, lockfile := range []string{"bun.lockb", "bun.lock"} {
			if _, err := os.Stat(filepath.Join(dir, lockfile)); err == nil {
				hasBunLock = true
				break
			}
		}
		// Only count bun.lockb if this directory also has package.json
		if hasBunLock {
			if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
				return "bun"
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir || dir == home {
			break
		}
		dir = parent
	}

	// Default to Node.js (npm) for package-lock.json, yarn.lock, pnpm-lock.yaml, or no lockfile
	return "node"
}

// detectRuntime checks for available JavaScript runtimes.
// Respects the project's lockfile to determine preferred runtime.
func (j *JavaScript) detectRuntime() (*JSRuntime, error) {
	projectRuntime := j.detectProjectRuntime()

	// Track errors for better messaging
	var nodeErr error
	var nodeRuntime, bunRuntime *JSRuntime

	// Check both runtimes to provide accurate error messages
	bunRuntime, _ = j.checkBun()
	nodeRuntime, nodeErr = j.checkNode()

	// Use project's preferred runtime if available
	if projectRuntime == "bun" && bunRuntime != nil {
		if j.verbose {
			text.Info(j.output, "Found Bun %s (bun.lockb detected)\n", bunRuntime.Version)
		}
		return bunRuntime, nil
	}
	if projectRuntime == "node" && nodeRuntime != nil {
		if j.verbose {
			text.Info(j.output, "Found Node.js %s with npm\n", nodeRuntime.Version)
		}
		return nodeRuntime, nil
	}

	// Fall back to any available runtime
	if nodeRuntime != nil {
		if j.verbose {
			text.Info(j.output, "Found Node.js %s with npm\n", nodeRuntime.Version)
		}
		return nodeRuntime, nil
	}
	if bunRuntime != nil {
		if j.verbose {
			text.Info(j.output, "Found Bun %s\n", bunRuntime.Version)
		}
		return bunRuntime, nil
	}

	// Provide specific error if Node exists but npm is missing
	if errors.Is(nodeErr, ErrNpmMissing) {
		return nil, fsterr.RemediationError{
			Inner: nodeErr,
			Remediation: `Node.js is installed but npm is missing.

Install npm (usually bundled with Node.js):
  - Reinstall Node.js from https://nodejs.org/
  - Or install npm separately: https://docs.npmjs.com/downloading-and-installing-node-js-and-npm

Verify: npm --version

Then retry: fastly compute build`,
		}
	}

	return nil, fsterr.RemediationError{
		Inner: fmt.Errorf("no JavaScript runtime found (node or bun)"),
		Remediation: `A JavaScript runtime is required to build Compute applications.

Install one of the following:

Option 1 - Node.js:
  Install from https://nodejs.org/ (LTS version recommended)
  Or use nvm: https://github.com/nvm-sh/nvm
  Verify: node --version && npm --version

Option 2 - Bun:
  curl -fsSL https://bun.sh/install | bash
  Verify: bun --version

Then retry: fastly compute build`,
	}
}

// findNodeModules searches for node_modules starting from startDir and moving up.
// Supports monorepo/hoisted setups where node_modules is in a parent directory.
func (j *JavaScript) findNodeModules(startDir, home string) (found bool, path string) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, "node_modules")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return true, candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir || dir == home {
			return false, ""
		}
		dir = parent
	}
}

// verifyDependencies checks that package.json and node_modules exist.
func (j *JavaScript) verifyDependencies() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	found, pkgPath, err := search("package.json", wd, home)
	if err != nil {
		return err
	}
	if !found {
		initCmd := "npm init"
		installCmd := "npm install @fastly/js-compute"
		if j.runtime != nil && j.runtime.PkgMgr == "bun" {
			initCmd = "bun init"
			installCmd = "bun add @fastly/js-compute"
		}
		return fsterr.RemediationError{
			Inner: fmt.Errorf("package.json not found"),
			Remediation: fmt.Sprintf(`A package.json file is required for JavaScript Compute projects.

Ensure you're in the correct project directory, or use --dir to specify the project root.

To initialize a new project:
  %s
  %s

Then retry: fastly compute build`, initCmd, installCmd),
		}
	}

	j.pkgDir = filepath.Dir(pkgPath)
	nodeModulesFound, nodeModulesPath := j.findNodeModules(j.pkgDir, home)
	if !nodeModulesFound {
		installCmd := "npm install"
		if j.runtime != nil && j.runtime.PkgMgr == "bun" {
			installCmd = "bun install"
		}
		return fsterr.RemediationError{
			Inner: fmt.Errorf("node_modules directory not found - dependencies not installed"),
			Remediation: fmt.Sprintf(`Dependencies have not been installed.

Run: %s

This will install all dependencies from package.json.
Then retry: fastly compute build`, installCmd),
		}
	}
	j.nodeModulesDir = nodeModulesPath

	if j.verbose {
		text.Info(j.output, "Found package.json at %s\n", pkgPath)
		text.Info(j.output, "Found node_modules at %s\n", nodeModulesPath)
	}
	return nil
}

// verifyWebpackInstalled checks that webpack is installed if used.
func (j *JavaScript) verifyWebpackInstalled() error {
	hasWebpack, err := j.checkForWebpack()
	if err != nil {
		return fmt.Errorf("failed to check for webpack in package.json: %w", err)
	}
	if !hasWebpack {
		return nil
	}

	binDir := filepath.Join(j.nodeModulesDir, ".bin")
	for _, name := range []string{"webpack", "webpack.cmd"} {
		if _, err := os.Stat(filepath.Join(binDir, name)); err == nil {
			if j.verbose {
				text.Info(j.output, "Found webpack in node_modules\n")
			}
			return nil
		}
	}

	installCmd := "npm install"
	installSpecific := "npm install webpack webpack-cli --save-dev"
	verifyCmd := "npx webpack --version"
	bunTip := ""
	if j.runtime != nil && j.runtime.PkgMgr == "bun" {
		installCmd = "bun install"
		installSpecific = "bun add -d webpack webpack-cli"
		verifyCmd = "bunx webpack --version"
		bunTip = "\n\nTip: Bun has a built-in bundler. You may not need webpack at all."
	}
	return fsterr.RemediationError{
		Inner: fmt.Errorf("webpack is listed in package.json but not installed"),
		Remediation: fmt.Sprintf(`Your project uses webpack but it's not installed.

Run: %s
Or specifically: %s
Verify with: %s

Then retry: fastly compute build%s`, installCmd, installSpecific, verifyCmd, bunTip),
	}
}

// verifyJsComputeRuntime checks that @fastly/js-compute is installed.
func (j *JavaScript) verifyJsComputeRuntime() error {
	runtimePath := filepath.Join(j.nodeModulesDir, "@fastly", "js-compute")
	if _, err := os.Stat(runtimePath); os.IsNotExist(err) {
		installCmd := "npm install @fastly/js-compute"
		if j.runtime != nil && j.runtime.PkgMgr == "bun" {
			installCmd = "bun add @fastly/js-compute"
		}
		return fsterr.RemediationError{
			Inner: fmt.Errorf("@fastly/js-compute package not found"),
			Remediation: fmt.Sprintf(`The Fastly JavaScript Compute runtime is not installed.

Run: %s

This package is required to compile JavaScript for Fastly Compute.
Then retry: fastly compute build`, installCmd),
		}
	}
	if j.verbose {
		text.Info(j.output, "Found @fastly/js-compute runtime\n")
	}
	return nil
}

// verifyToolchain checks that a JavaScript runtime is installed and accessible.
// Only called when using default build script (not custom [scripts.build]).
func (j *JavaScript) verifyToolchain() error {
	runtime, err := j.detectRuntime()
	if err != nil {
		return err
	}
	j.runtime = runtime

	if err := j.verifyDependencies(); err != nil {
		return err
	}
	if err := j.verifyWebpackInstalled(); err != nil {
		return err
	}
	if err := j.verifyJsComputeRuntime(); err != nil {
		return err
	}
	return nil
}

// getDefaultBuildCommand returns the appropriate build command for the detected runtime.
func (j *JavaScript) getDefaultBuildCommand() (string, error) {
	hasWebpack, err := j.checkForWebpack()
	if err != nil {
		return "", err
	}
	if j.runtime != nil && j.runtime.PkgMgr == "bun" {
		if hasWebpack {
			return fmt.Sprintf("bunx webpack && bunx js-compute-runtime ./bin/index.js %s", binWasmPath), nil
		}
		return fmt.Sprintf("bunx js-compute-runtime ./src/index.js %s", binWasmPath), nil
	}
	if hasWebpack {
		return JsDefaultBuildCommandForWebpack, nil
	}
	return JsDefaultBuildCommand, nil
}
