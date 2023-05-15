//go:build windows

package kvstoreentry

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// isHiddenFile checks if a file or directory is a hidden dotfile
func isHiddenFile(filename string) (bool, error) {
	attrs, err := getFileAttributes(filename)
	if err != nil {
		return false, err
	}
	isHidden := attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0 || matchHiddenDotfilePattern(filename)
	return isHidden, nil
}

// matchHiddenDotfilePattern checks if a file name matches the pattern of hidden files on Windows
func matchHiddenDotfilePattern(filename string) bool {
	match, _ := filepath.Match(".*", filename)
	return match
}

// getFileAttributes returns the file attributes for the specified file on Windows
func getFileAttributes(filename string) (uint32, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}

	sys, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return 0, fmt.Errorf("unsupported file attribute data type")
	}

	return sys.FileAttributes, nil
}
