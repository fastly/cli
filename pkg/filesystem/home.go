package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	// UnixHome is the home directory for a unix system.
	UnixHome = "$HOME"
	// UnixHomeShort is the 'short' home directory for a unix system.
	UnixHomeShort = "~"
	// WindowsHome is the home directory for a Windows system.
	WindowsHome = "%USERPROFILE%"
)

// ResolveAbs returns an absolute path with the user home directory resolved.
//
// EXAMPLE (unix):
// $HOME/.gitignore -> /Users/<USER>/.gitignore
// ~/.gitignore     -> /Users/<USER>/.gitignore
func ResolveAbs(path string) string {
	var uhd string
	if strings.HasPrefix(path, UnixHome) {
		uhd = UnixHome
	}
	if strings.HasPrefix(path, UnixHomeShort) {
		uhd = UnixHomeShort
	}
	if strings.HasPrefix(path, WindowsHome) {
		uhd = WindowsHome
	}

	if uhd != "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = strings.Replace(path, uhd, "", 1)
		return filepath.Join(home, path)
	}

	s, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return s
}
