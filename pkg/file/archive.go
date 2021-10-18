package file

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TarGz represents an instance of a tar.gz archive file.
var TarGz = &ArchiveGzip{
	ArchiveBase{
		Ext:   ".gz",
		Mimes: []string{"application/gzip", "application/x-gzip"},
	},
}

// Zip represents an instance of a zip archive file.
var Zip = &ArchiveZip{
	ArchiveBase{
		Ext:   ".zip",
		Mimes: []string{"application/zip", "application/x-zip"},
	},
}

// Archives is a collection of supported archive formats.
var Archives = []Archive{TarGz, Zip}

// Archive represents the associated behaviour for a collection of files
// contained inside an archive format.
type Archive interface {
	Extension() string
	Extract() error
	Filename() string
	MimeTypes() []string
	SetDestination(d string)
	SetFile(r io.ReadSeeker)
	SetFilename(n string)
}

// ArchiveBase represents a container for a collection of files.
type ArchiveBase struct {
	Dst   string
	Ext   string
	File  io.ReadSeeker
	Mimes []string
	Name  string
}

// Extension returns the file extension.
func (a ArchiveBase) Extension() string {
	return a.Ext
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

// SetFile sets the local file descriptor where the archive should be written.
//
// NOTE: This archive file is the 'container' of the archived files that will
// be extracted separately.
func (a *ArchiveBase) SetFile(r io.ReadSeeker) {
	a.File = r
}

// SetFilename sets the name of the local archive file.
//
// NOTE: This archive file is the 'container' of the archived files that will
// be extracted separately.
func (a *ArchiveBase) SetFilename(n string) {
	a.Name = n
}

// ArchiveGzip represents a container for the .tar.gz file format.
type ArchiveGzip struct {
	ArchiveBase
}

// Extract all files and folders from the collection.
func (a ArchiveGzip) Extract() error {
	// NOTE: After the os.File has content written to it via io.Copy() inside
	// pkgFetch(), we find the cursor index position is updated. This causes an
	// EOF error when passing the file into gzip.NewReader() and so we need to
	// first reset the cursor index.
	_, err := a.File.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to the start of archive: %w", err)
	}

	gr, err := gzip.NewReader(a.File)
	if err != nil {
		return fmt.Errorf("error creating gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		header, err := tr.Next()

		switch {

		// If no more files are found return
		case err == io.EOF:
			return nil

		// Return any other error
		case err != nil:
			return err

		// If the header is nil, skip it
		case header == nil:
			continue

		// Skip the any files duplicated as hidden files
		case strings.HasPrefix(header.Name, "._"):
			continue
		}

		// The target location where the dir/file should be created
		segs := splitArchivePaths(header.Name)
		segs = segs[1:]
		target := filepath.Join(a.Dst, filepath.Join(segs...))

		fi := header.FileInfo()

		if fi.IsDir() {
			os.MkdirAll(target, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}

		fd, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fi.Mode())
		if err != nil {
			return err
		}

		// NOTE: We use looped CopyN() not Copy() to avoid gosec G110 (CWE-409):
		// Potential DoS vulnerability via decompression bomb.
		for {
			_, err := io.CopyN(fd, tr, 1024)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}

		fd.Close()
	}
}

// ArchiveZip represents a container for the .zip file format.
type ArchiveZip struct {
	ArchiveBase
}

// Extract all files and folders from the collection.
func (a ArchiveZip) Extract() error {
	r, err := zip.OpenReader(a.Name)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// The zip contains a folder, and inside that folder are the files we're
		// interested in. So while looping over the files (whose .Name field is the
		// full path including the containing folder) we strip out the first path
		// segment to ensure the files we need are extracted to the current directory.
		segs := splitArchivePaths(f.Name)
		segs = segs[1:]
		target := filepath.Join(a.Dst, filepath.Join(segs...))

		if f.FileInfo().IsDir() {
			err := os.MkdirAll(target, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(target), os.ModePerm); err != nil {
			return err
		}

		fd, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		// NOTE: We use looped CopyN() not Copy() to avoid gosec G110 (CWE-409):
		// Potential DoS vulnerability via decompression bomb.
		for {
			_, err := io.CopyN(fd, rc, 1024)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
		}

		fd.Close()
		rc.Close()
	}

	return nil
}

// splitArchivePaths splits a path into segments.
//
// The algorithm takes into account archives containing files created by either
// Windows or unix based system such as macOS or Linux. Specifically the
// filepath.Separator isn't reliable as the binary could be running on one OS
// while trying to use an archive created via a different OS.
//
// NOTE: We expect the archive to contain a single directory that contains the
// package files/directories. This means when splitting the file path into
// segments the length should be at least two (e.g. 'compute-package/...').
func splitArchivePaths(filepath string) (segments []string) {
	unix := `/`
	win := `\`

	segments = strings.Split(filepath, unix)

	if len(segments) < 2 {
		segments = strings.Split(filepath, win)
	}

	return segments
}
