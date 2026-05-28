package viceroy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"

	fstruntime "github.com/fastly/cli/pkg/runtime"
)

const extractSubdir = "viceroy-embedded"

func binaryName() string {
	if fstruntime.Windows {
		return "viceroy.exe"
	}
	return "viceroy"
}

// Extract decompresses the embedded Viceroy binary into installDir and
// returns the absolute path to the executable. If a non-empty file
// already exists at the expected versioned location, Extract is a
// no-op and returns the existing path. The function is safe to call
// concurrently from multiple processes thanks to a temp-file + atomic
// rename.
//
// Extract returns ErrUnsupported when Supported reports false. Old
// versioned directories from a previous CLI install are intentionally
// left in place: a separate cleanup path (rather than this hot one)
// should handle them, to avoid removing a binary that another fastly
// process may still be exec'ing.
func Extract(installDir string) (string, error) {
	if !Supported() {
		return "", ErrUnsupported
	}

	versionDir := filepath.Join(installDir, extractSubdir, "v"+Version())
	binPath := filepath.Join(versionDir, binaryName())

	if info, err := os.Stat(binPath); err == nil && !info.IsDir() && info.Size() > 0 {
		return binPath, nil
	}

	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return "", fmt.Errorf("viceroy: create extract dir: %w", err)
	}

	tmpPath, err := writeTempBinary(versionDir)
	if err != nil {
		return "", err
	}

	if err := os.Rename(tmpPath, binPath); err != nil {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("viceroy: install extracted binary: %w", err)
	}
	return binPath, nil
}

func writeTempBinary(dir string) (string, error) {
	f, err := os.CreateTemp(dir, "."+binaryName()+".tmp.*")
	if err != nil {
		return "", fmt.Errorf("viceroy: create temp file: %w", err)
	}
	tmpPath := f.Name()

	dec, err := zstd.NewReader(nil)
	if err != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("viceroy: init zstd: %w", err)
	}
	defer dec.Close()

	src, err := dec.DecodeAll(binaryZstd, nil)
	if err != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("viceroy: decompress: %w", err)
	}

	if _, err := f.Write(src); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("viceroy: write temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("viceroy: close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("viceroy: chmod temp file: %w", err)
	}

	return tmpPath, nil
}
