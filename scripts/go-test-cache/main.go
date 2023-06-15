// This script is based on https://github.com/airplanedev/blog-examples/blob/main/go-test-caching/update_file_timestamps.py?ref=airplane.ghost.io
// and was generated automatically using AI.
package main

import (
	"crypto/sha1"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	BUF_SIZE    = 65536
	BASE_DATE   = 1684178360
	TIME_FORMAT = "2006-01-02 15:04:05"
)

func main() {
	repoRoot := "."
	allDirs := make([]string, 0)

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dirPath := filepath.Join(repoRoot, path)
			relPath, _ := filepath.Rel(repoRoot, dirPath)

			if strings.HasPrefix(relPath, ".") {
				return nil
			}

			allDirs = append(allDirs, dirPath)
		} else {
			filePath := filepath.Join(repoRoot, path)
			relPath, _ := filepath.Rel(repoRoot, filePath)

			if strings.HasPrefix(relPath, ".") {
				return nil
			}

			sha1Hash, err := getFileSHA1(filePath)
			if err != nil {
				return err
			}

			modTime := getModifiedTime(sha1Hash)

			log.Printf("Setting modified time of file %s to %s\n", relPath, modTime.Format(TIME_FORMAT))
			err = os.Chtimes(filePath, modTime, modTime)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal("Error:", err)
	}

	sort.Slice(allDirs, func(i, j int) bool {
		return len(allDirs[i]) > len(allDirs[j]) || (len(allDirs[i]) == len(allDirs[j]) && allDirs[i] < allDirs[j])
	})

	for _, dirPath := range allDirs {
		relPath, _ := filepath.Rel(repoRoot, dirPath)

		log.Printf("Setting modified time of directory %s to %s\n", relPath, time.Unix(BASE_DATE, 0).Format(TIME_FORMAT))
		err := os.Chtimes(dirPath, time.Unix(BASE_DATE, 0), time.Unix(BASE_DATE, 0))
		if err != nil {
			log.Fatal("Error:", err)
		}
	}

	log.Println("Done")
}

func getFileSHA1(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.CopyBuffer(hash, file, make([]byte, BUF_SIZE)); err != nil {
		return "", err
	}

	return string(hash.Sum(nil)), nil
}

func getModifiedTime(sha1Hash string) time.Time {
	hashBytes := []byte(sha1Hash)
	lastFiveBytes := hashBytes[:5]
	lastFiveValue := int64(0)

	for _, b := range lastFiveBytes {
		lastFiveValue = (lastFiveValue << 8) + int64(b)
	}

	modTime := BASE_DATE - (lastFiveValue % 10000)
	return time.Unix(modTime, 0)
}
