//go:build !windows

package kvstoreentry

// isHiddenFile checks if a file or directory is a hidden dotfile
func isHiddenFile(filename string) (bool, error) {
	return filename[0] == '.', nil
}
