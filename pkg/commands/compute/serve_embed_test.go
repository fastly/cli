//go:build viceroy_embed

package compute_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/argparser"
	"github.com/fastly/cli/pkg/commands/compute"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/embedded/viceroy"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/testutil"
	"github.com/fastly/cli/pkg/text"
)

func TestGetViceroyUsesEmbedded(t *testing.T) {
	if !viceroy.Supported() {
		t.Skipf("no embedded Viceroy for %s/%s; skipping", runtime.GOOS, runtime.GOARCH)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	rootdir := testutil.NewEnv(testutil.EnvOpts{
		T:    t,
		Dirs: []string{"install"},
		Write: []testutil.FileIO{
			{Src: "", Dst: config.FileName},
		},
	})
	installDir := filepath.Join(rootdir, "install")
	configPath := filepath.Join(rootdir, config.FileName)
	defer os.RemoveAll(rootdir)

	if err := os.Chdir(rootdir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	github.InstallDir = installDir

	var out bytes.Buffer

	// DownloadOK=false makes any accidental fallback to the download path
	// surface as a test failure.
	av := mock.AssetVersioner{
		AssetVersion:   viceroy.Version(),
		BinaryFilename: "viceroy",
		DownloadOK:     false,
	}

	var file config.File
	if err := file.Read("example", strings.NewReader("yes"), &out, fsterr.MockLog{}, false); err != nil {
		t.Fatal(err)
	}

	spinner, err := text.NewSpinner(&out)
	if err != nil {
		t.Fatal(err)
	}

	c := &compute.ServeCommand{
		Base: argparser.Base{
			Globals: &global.Data{
				Config:     file,
				ConfigPath: configPath,
				ErrLog:     fsterr.MockLog{},
			},
		},
		ViceroyVersioner: av,
	}

	bin, err := c.GetViceroy(spinner, &out, "fastly.toml")
	if err != nil {
		t.Fatalf("GetViceroy() error = %v", err)
	}

	wantDir := filepath.Join(installDir, "viceroy-embedded", "v"+viceroy.Version())
	if filepath.Dir(bin) != wantDir {
		t.Errorf("extracted binary in unexpected dir: got %q, want under %q", bin, wantDir)
	}

	if info, err := os.Stat(bin); err != nil {
		t.Fatalf("extracted binary missing: %v", err)
	} else if info.Size() == 0 {
		t.Error("extracted binary is empty")
	}

	if strings.Contains(out.String(), "Fetching Viceroy release") {
		t.Errorf("embedded path unexpectedly triggered a download: %s", out.String())
	}
}
