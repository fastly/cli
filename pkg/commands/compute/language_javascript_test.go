package compute

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	fsterr "github.com/fastly/cli/pkg/errors"
)

// createFakeRuntime creates a fake executable that outputs the given string.
func createFakeRuntime(t *testing.T, dir, name, output string) {
	t.Helper()
	var script string
	if runtime.GOOS == "windows" {
		script = "@echo off\r\necho " + output
		name += ".bat"
	} else {
		script = "#!/bin/sh\necho '" + output + "'"
	}
	path := filepath.Join(dir, name)
	// G306 (CWE-276): Expect WriteFile permissions to be 0600 or less
	// Disabling as executables must be executable.
	// #nosec G306
	err := os.WriteFile(path, []byte(script), 0o755)
	if err != nil {
		t.Fatal(err)
	}
}

func TestJavaScript_detectRuntime_NoRuntime(t *testing.T) {
	// Create a temp directory with no executables
	tmpDir := t.TempDir()
	t.Setenv("PATH", tmpDir)

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	_, err := j.detectRuntime()
	if err == nil {
		t.Fatal("expected error when no runtime is found")
	}

	// Check it's a RemediationError with helpful message
	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("expected RemediationError, got %T", err)
	}

	if re.Remediation == "" {
		t.Error("expected remediation message")
	}
}

func TestJavaScript_detectRuntime_NodeFound(t *testing.T) {
	tmpDir := t.TempDir()
	createFakeRuntime(t, tmpDir, "node", "v24.13.0")
	createFakeRuntime(t, tmpDir, "npm", "11.7.0")
	t.Setenv("PATH", tmpDir)

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	rt, err := j.detectRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rt.Name != "node" {
		t.Errorf("expected runtime name 'node', got %q", rt.Name)
	}
	if rt.PkgMgr != "npm" {
		t.Errorf("expected package manager 'npm', got %q", rt.PkgMgr)
	}
}

func TestJavaScript_detectRuntime_BunFound(t *testing.T) {
	tmpDir := t.TempDir()
	createFakeRuntime(t, tmpDir, "bun", "1.3.7")
	t.Setenv("PATH", tmpDir)

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	rt, err := j.detectRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rt.Name != "bun" {
		t.Errorf("expected runtime name 'bun', got %q", rt.Name)
	}
	if rt.PkgMgr != "bun" {
		t.Errorf("expected package manager 'bun', got %q", rt.PkgMgr)
	}
}

func TestJavaScript_detectRuntime_NodePreferredByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	createFakeRuntime(t, tmpDir, "bun", "1.3.7")
	createFakeRuntime(t, tmpDir, "node", "v24.13.0")
	createFakeRuntime(t, tmpDir, "npm", "11.7.0")
	t.Setenv("PATH", tmpDir)

	// Create project dir without bun.lockb (npm project)
	projectDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	rt, err := j.detectRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Node should be preferred by default (no bun.lockb)
	if rt.Name != "node" {
		t.Errorf("expected runtime name 'node' (default), got %q", rt.Name)
	}
}

func TestJavaScript_detectRuntime_BunPreferredWithLockfile(t *testing.T) {
	tmpDir := t.TempDir()
	createFakeRuntime(t, tmpDir, "bun", "1.3.7")
	createFakeRuntime(t, tmpDir, "node", "v24.13.0")
	createFakeRuntime(t, tmpDir, "npm", "11.7.0")
	t.Setenv("PATH", tmpDir)

	// Create project dir with package.json and bun.lockb (bun project)
	projectDir := t.TempDir()
	// #nosec G306
	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// #nosec G306
	if err := os.WriteFile(filepath.Join(projectDir, "bun.lockb"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	rt, err := j.detectRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bun should be used when bun.lockb exists alongside package.json
	if rt.Name != "bun" {
		t.Errorf("expected runtime name 'bun' (bun.lockb detected), got %q", rt.Name)
	}
}

func TestJavaScript_detectRuntime_BunLockfileInParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	createFakeRuntime(t, tmpDir, "bun", "1.3.7")
	createFakeRuntime(t, tmpDir, "node", "v24.13.0")
	createFakeRuntime(t, tmpDir, "npm", "11.7.0")
	t.Setenv("PATH", tmpDir)

	// Create project structure: projectDir/subdir with package.json and bun.lockb in projectDir
	projectDir := t.TempDir()
	subDir := filepath.Join(projectDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// #nosec G306
	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// #nosec G306
	if err := os.WriteFile(filepath.Join(projectDir, "bun.lockb"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	// Run from subdir - should detect bun.lockb alongside package.json in parent
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	rt, err := j.detectRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bun should be detected from project root (where package.json is)
	if rt.Name != "bun" {
		t.Errorf("expected runtime name 'bun' (bun.lockb with package.json), got %q", rt.Name)
	}
}

func TestJavaScript_detectRuntime_BunWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	createFakeRuntime(t, tmpDir, "bun", "1.3.7")
	createFakeRuntime(t, tmpDir, "node", "v24.13.0")
	createFakeRuntime(t, tmpDir, "npm", "11.7.0")
	t.Setenv("PATH", tmpDir)

	// Create Bun workspace structure:
	// workspace/package.json (workspace root)
	// workspace/bun.lockb
	// workspace/packages/myapp/package.json (subpackage - we run from here)
	workspaceDir := t.TempDir()
	subpkgDir := filepath.Join(workspaceDir, "packages", "myapp")
	if err := os.MkdirAll(subpkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Workspace root package.json
	// #nosec G306
	if err := os.WriteFile(filepath.Join(workspaceDir, "package.json"), []byte(`{"workspaces":["packages/*"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// #nosec G306
	if err := os.WriteFile(filepath.Join(workspaceDir, "bun.lockb"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	// Subpackage package.json
	// #nosec G306
	if err := os.WriteFile(filepath.Join(subpkgDir, "package.json"), []byte(`{"name":"myapp"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Run from subpackage
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(subpkgDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	rt, err := j.detectRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Bun should be detected from workspace root (bun.lockb + package.json)
	if rt.Name != "bun" {
		t.Errorf("expected runtime name 'bun' (workspace detected), got %q", rt.Name)
	}
}

func TestJavaScript_detectRuntime_IgnoresUnrelatedBunLockfile(t *testing.T) {
	tmpDir := t.TempDir()
	createFakeRuntime(t, tmpDir, "bun", "1.3.7")
	createFakeRuntime(t, tmpDir, "node", "v24.13.0")
	createFakeRuntime(t, tmpDir, "npm", "11.7.0")
	t.Setenv("PATH", tmpDir)

	// Create structure: parentDir/bun.lockb (unrelated) and parentDir/project/package.json (npm project)
	parentDir := t.TempDir()
	projectDir := filepath.Join(parentDir, "project")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Unrelated bun.lockb in parent (not alongside package.json)
	// #nosec G306
	if err := os.WriteFile(filepath.Join(parentDir, "bun.lockb"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}
	// Project's package.json (no bun.lockb here)
	// #nosec G306
	if err := os.WriteFile(filepath.Join(projectDir, "package.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	rt, err := j.detectRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use Node because project root has no bun.lockb (parent's is unrelated)
	if rt.Name != "node" {
		t.Errorf("expected runtime name 'node' (unrelated bun.lockb ignored), got %q", rt.Name)
	}
}

func TestJavaScript_detectRuntime_NodeMissingNpm(t *testing.T) {
	tmpDir := t.TempDir()
	createFakeRuntime(t, tmpDir, "node", "v24.13.0")
	// npm is NOT created
	t.Setenv("PATH", tmpDir)

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
	}

	_, err := j.detectRuntime()
	if err == nil {
		t.Fatal("expected error when npm is missing")
	}

	// Check for specific error message
	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("expected RemediationError, got %T", err)
	}

	if !errors.Is(re.Inner, ErrNpmMissing) {
		t.Errorf("expected ErrNpmMissing, got %v", re.Inner)
	}
}

func TestJavaScript_findNodeModules(t *testing.T) {
	// Create directory structure: project/subdir with node_modules in project
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	subDir := filepath.Join(projectDir, "subdir")
	nodeModulesDir := filepath.Join(projectDir, "node_modules")

	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(nodeModulesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{}

	// Should find node_modules in parent directory
	found, path := j.findNodeModules(subDir, tmpDir)
	if !found {
		t.Error("expected to find node_modules")
	}
	if path != nodeModulesDir {
		t.Errorf("expected path %q, got %q", nodeModulesDir, path)
	}

	// Should find node_modules in current directory
	found, path = j.findNodeModules(projectDir, tmpDir)
	if !found {
		t.Error("expected to find node_modules")
	}
	if path != nodeModulesDir {
		t.Errorf("expected path %q, got %q", nodeModulesDir, path)
	}

	// Should not find node_modules above home
	found, _ = j.findNodeModules(tmpDir, tmpDir)
	if found {
		t.Error("expected not to find node_modules above home")
	}
}

func TestJavaScript_verifyDependencies_NoPackageJson(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := t.TempDir()
	createFakeRuntime(t, binDir, "node", "v24.13.0")
	createFakeRuntime(t, binDir, "npm", "11.7.0")
	t.Setenv("PATH", binDir)

	// Change to temp dir with no package.json
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
		runtime: &JSRuntime{Name: "node", PkgMgr: "npm"},
	}

	err := j.verifyDependencies()
	if err == nil {
		t.Fatal("expected error when package.json not found")
	}

	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("expected RemediationError, got %T", err)
	}
}

func TestJavaScript_verifyDependencies_NoNodeModules(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := t.TempDir()
	createFakeRuntime(t, binDir, "node", "v24.13.0")
	createFakeRuntime(t, binDir, "npm", "11.7.0")
	t.Setenv("PATH", binDir)

	// Create package.json but no node_modules
	// #nosec G306
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
		runtime: &JSRuntime{Name: "node", PkgMgr: "npm"},
	}

	err := j.verifyDependencies()
	if err == nil {
		t.Fatal("expected error when node_modules not found")
	}

	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("expected RemediationError, got %T", err)
	}
}

func TestJavaScript_verifyJsComputeRuntime_NotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	if err := os.MkdirAll(nodeModulesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:         &bytes.Buffer{},
		verbose:        false,
		nodeModulesDir: nodeModulesDir,
		runtime:        &JSRuntime{Name: "node", PkgMgr: "npm"},
	}

	err := j.verifyJsComputeRuntime()
	if err == nil {
		t.Fatal("expected error when @fastly/js-compute not found")
	}

	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("expected RemediationError, got %T", err)
	}
}

func TestJavaScript_verifyJsComputeRuntime_Installed(t *testing.T) {
	tmpDir := t.TempDir()
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	runtimeDir := filepath.Join(nodeModulesDir, "@fastly", "js-compute")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:         &bytes.Buffer{},
		verbose:        false,
		nodeModulesDir: nodeModulesDir,
		runtime:        &JSRuntime{Name: "node", PkgMgr: "npm"},
	}

	err := j.verifyJsComputeRuntime()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJavaScript_getDefaultBuildCommand_NodeWithWebpack(t *testing.T) {
	tmpDir := t.TempDir()
	// #nosec G306
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"devDependencies":{"webpack":"5.0.0"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
		runtime: &JSRuntime{Name: "node", PkgMgr: "npm"},
	}

	cmd, err := j.getDefaultBuildCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmd != JsDefaultBuildCommandForWebpack {
		t.Errorf("expected webpack command, got %q", cmd)
	}
}

func TestJavaScript_getDefaultBuildCommand_NodeNoWebpack(t *testing.T) {
	tmpDir := t.TempDir()
	// #nosec G306
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
		runtime: &JSRuntime{Name: "node", PkgMgr: "npm"},
	}

	cmd, err := j.getDefaultBuildCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmd != JsDefaultBuildCommand {
		t.Errorf("expected default command, got %q", cmd)
	}
}

func TestJavaScript_getDefaultBuildCommand_Bun(t *testing.T) {
	tmpDir := t.TempDir()
	// #nosec G306
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	j := &JavaScript{
		output:  &bytes.Buffer{},
		verbose: false,
		runtime: &JSRuntime{Name: "bun", PkgMgr: "bun"},
	}

	cmd, err := j.getDefaultBuildCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should use bunx instead of npm exec
	if cmd == JsDefaultBuildCommand {
		t.Errorf("expected bun command, got npm command %q", cmd)
	}
	if !bytes.Contains([]byte(cmd), []byte("bunx")) {
		t.Errorf("expected command to contain 'bunx', got %q", cmd)
	}
}
