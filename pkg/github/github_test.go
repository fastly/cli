package github

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	fstruntime "github.com/fastly/cli/pkg/runtime"
)

// TestDownloadArchiveExtract validates both Windows and Unix release assets.
func TestDownloadArchiveExtract(t *testing.T) {
	scenarios := []struct {
		Platform string
		Arch     string
		Ext      string
	}{
		{
			Platform: "darwin",
			Arch:     "arm64",
			Ext:      ".tar.gz",
		},
		{
			Platform: "darwin",
			Arch:     "amd64",
			Ext:      ".tar.gz",
		},
		{
			Platform: "windows",
			Arch:     "amd64",
			Ext:      ".zip",
		},
	}

	for _, testcase := range scenarios {
		name := fmt.Sprintf("%s_%s", testcase.Platform, testcase.Arch)

		t.Run(name, func(t *testing.T) {
			// Avoid, for example, running the Windows OS scenario on non Windows OS.
			// Otherwise, the Windows OS scenario would show on Darwin an error like:
			// no asset found for your OS (darwin) and architecture (amd64)
			if runtime.GOOS != testcase.Platform || runtime.GOARCH != testcase.Arch {
				t.Skip()
			}

			binary := "fastly"
			if fstruntime.Windows {
				binary = binary + ".exe"
			}

			a := Asset{
				binary: binary,
				org:    "fastly",
				repo:   "cli",
			}

			bin, err := a.Download()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if err := os.RemoveAll(bin); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
		})
	}
}
