package file

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fastly/cli/pkg/errors"
	"github.com/mholt/archiver"
)

// Archives is a collection of supported archive formats.
var Archives = []Archive{TarGz, Zip}

// Archive represents the associated behaviour for a collection of files
// contained inside an archive format.
type Archive interface {
	Extensions() []string
	Extract() error
	Filename() string
	MimeTypes() []string
	SetDestination(d string)
	SetFilename(n string)
}

// TarGz represents an instance of a tar.gz archive file.
var TarGz = &ArchiveGzip{
	ArchiveBase{
		Exts:  []string{".tgz", ".gz"},
		Mimes: []string{"application/gzip", "application/x-gzip", "application/x-tar"},
	},
}

// Zip represents an instance of a zip archive file.
var Zip = &ArchiveZip{
	ArchiveBase{
		Exts:  []string{".zip"},
		Mimes: []string{"application/zip", "application/x-zip"},
	},
}

// ArchiveGzip represents a container for the .tar.gz file format.
type ArchiveGzip struct {
	ArchiveBase
}

// ArchiveZip represents a container for the .zip file format.
type ArchiveZip struct {
	ArchiveBase
}

// ArchiveBase represents a container for a collection of files.
type ArchiveBase struct {
	Dst   string
	Exts  []string
	File  io.ReadSeeker
	Mimes []string
	Name  string
}

// Extensions returns the accepted file extensions.
func (a ArchiveBase) Extensions() []string {
	return a.Exts
}

// MimeTypes returns all valid  mime types for the format.
func (a ArchiveBase) MimeTypes() []string {
	return a.Mimes
}

// Filename returns the file name.
func (a ArchiveBase) Filename() string {
	return a.Name
}

// SetDestination sets the destination for where files should be extracted.
func (a *ArchiveBase) SetDestination(d string) {
	a.Dst = d
}

// SetFilename sets the name of the local archive file.
//
// NOTE: This archive file is the 'container' of the archived files that will
// be extracted separately.
func (a *ArchiveBase) SetFilename(n string) {
	a.Name = n
}

// Extract all files and folders from the collection.
func (a ArchiveBase) Extract() error {
	if err := archiver.Unarchive(a.Filename(), a.Dst); err != nil {
		return fmt.Errorf("error extracting contents from archive: %w", err)
	}

	if _, err := os.Stat("fastly.toml"); err == nil {
		return nil
	}

	// Looks like the package files are contained within a top-level directory
	// that now need to be extracted.
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error determining current directory: %w", err)
	}

	var dirContentToMove string

	err = filepath.WalkDir(wd, func(path string, entry fs.DirEntry, err error) error {
		// WalkDir() triggered an error
		if err != nil {
			return err
		}
		// We already check if the current directory had a manifest so skip it
		if entry.IsDir() && path == wd {
			return nil
		}
		// We expect there to be a directory that contains the manifest
		if entry.IsDir() {
			if _, err := os.Stat(filepath.Join(path, "fastly.toml")); err == nil {
				dirContentToMove = path
				return errors.ErrStopWalk
			}
		}
		return nil
	})

	if err != nil && err != errors.ErrStopWalk {
		return err
	}
	if dirContentToMove == "" {
		return errors.ErrInvalidArchive
	}

	files, err := filepath.Glob(filepath.Join(dirContentToMove, "*"))
	if err != nil {
		return err
	}

	// Move files from within package directory into its parent directory
	for _, path := range files {
		dir, file := filepath.Split(path)
		if strings.HasSuffix(dir, string(os.PathSeparator)) {
			dir = dir[:len(dir)-1]
		}
		parent := filepath.Dir(dir)
		err := os.Rename(path, filepath.Join(parent, file))
		if err != nil {
			return err
		}
	}

	os.RemoveAll(dirContentToMove)

	return nil
}
