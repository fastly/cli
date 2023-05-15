package kvstoreentry

func isHiddenFile(filename string) (bool, error) {
	return filename[0] == '.', nil
}
