package kvstoreentry

func isHiddenFile(filename string) bool {
	return filename[0] == '.'
}
