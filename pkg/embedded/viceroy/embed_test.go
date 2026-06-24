//go:build viceroy_embed

package viceroy

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func onlyOnSupportedPlatform(t *testing.T) {
	t.Helper()
	if !Supported() {
		t.Skipf("no embedded Viceroy for %s/%s; skipping", runtime.GOOS, runtime.GOARCH)
	}
}

func TestExtractCreatesExecutableFile(t *testing.T) {
	onlyOnSupportedPlatform(t)

	dir := t.TempDir()
	path, err := Extract(dir)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat extracted file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("extracted file is empty")
	}

	if runtime.GOOS != "windows" {
		if info.Mode().Perm()&0o111 == 0 {
			t.Errorf("extracted file is not executable: mode=%v", info.Mode())
		}
	}

	expectedSuffix := binaryName()
	if filepath.Base(path) != expectedSuffix {
		t.Errorf("extracted basename = %q, want %q", filepath.Base(path), expectedSuffix)
	}

	wantDir := filepath.Join(dir, extractSubdir, "v"+Version())
	if filepath.Dir(path) != wantDir {
		t.Errorf("extracted dir = %q, want %q", filepath.Dir(path), wantDir)
	}
}

func TestExtractIdempotent(t *testing.T) {
	onlyOnSupportedPlatform(t)

	dir := t.TempDir()
	first, err := Extract(dir)
	if err != nil {
		t.Fatalf("first Extract() error = %v", err)
	}
	infoBefore, err := os.Stat(first)
	if err != nil {
		t.Fatal(err)
	}

	second, err := Extract(dir)
	if err != nil {
		t.Fatalf("second Extract() error = %v", err)
	}
	if first != second {
		t.Errorf("second Extract() returned different path: %q vs %q", first, second)
	}

	infoAfter, err := os.Stat(second)
	if err != nil {
		t.Fatal(err)
	}
	if !infoBefore.ModTime().Equal(infoAfter.ModTime()) {
		t.Errorf("Extract() rewrote the file on second call: before=%v after=%v",
			infoBefore.ModTime(), infoAfter.ModTime())
	}
}

func TestExtractConcurrent(t *testing.T) {
	onlyOnSupportedPlatform(t)

	dir := t.TempDir()
	const N = 8

	var wg sync.WaitGroup
	results := make([]string, N)
	errs := make([]error, N)
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = Extract(dir)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: Extract() error = %v", i, err)
		}
	}
	for i := 1; i < N; i++ {
		if results[i] != results[0] {
			t.Errorf("Extract() paths diverged: %q vs %q", results[0], results[i])
		}
	}
}

func TestExtractLeavesStaleSiblingVersionsAlone(t *testing.T) {
	onlyOnSupportedPlatform(t)

	dir := t.TempDir()
	stale := filepath.Join(dir, extractSubdir, "v0.0.0-stale")
	if err := os.MkdirAll(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	staleMarker := filepath.Join(stale, "marker")
	if err := os.WriteFile(staleMarker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Extract(dir); err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if _, err := os.Stat(stale); err != nil {
		t.Errorf("Extract() removed an unrelated sibling directory: %v", err)
	}
}

func TestExtractRejectsEmptyExistingFile(t *testing.T) {
	onlyOnSupportedPlatform(t)

	dir := t.TempDir()
	versionDir := filepath.Join(dir, extractSubdir, "v"+Version())
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(versionDir, binaryName())
	if err := os.WriteFile(binPath, nil, 0o755); err != nil {
		t.Fatal(err)
	}

	path, err := Extract(dir)
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("Extract() returned a zero-byte file instead of overwriting it")
	}
}
