//go:build windows

package kvstoreentry

import (
	"path/filepath"
	"syscall"
)

const dotCharacter = 46

func isHiddenFile(filename string) (bool, error) {
	// pointer, err := syscall.UTF16PtrFromString(filename)
	// if err != nil {
	// 	return false, err
	// }
	// attributes, err := syscall.GetFileAttributes(pointer)
	// if err != nil {
	// 	return false, err
	// }
	// return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil

	// dotfiles also count as hidden (if you want)
	if filename[0] == dotCharacter {
		return true, nil
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return false, err
	}

	// Appending `\\?\` to the absolute path helps with
	// preventing 'Path Not Specified Error' when accessing
	// long paths and filenames
	// https://docs.microsoft.com/en-us/windows/win32/fileio/maximum-file-path-limitation?tabs=cmd
	pointer, err := syscall.UTF16PtrFromString(`\\?\` + absPath)
	if err != nil {
		return false, err
	}

	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
