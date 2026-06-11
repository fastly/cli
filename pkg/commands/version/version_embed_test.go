//go:build viceroy_embed

package version_test

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/fastly/cli/pkg/app"
	"github.com/fastly/cli/pkg/commands/version"
	"github.com/fastly/cli/pkg/embedded/viceroy"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/testutil"
)

func TestVersionUsesEmbeddedViceroy(t *testing.T) {
	if !viceroy.Supported() {
		t.Skipf("no embedded Viceroy for %s/%s; skipping", runtime.GOOS, runtime.GOARCH)
	}
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("skipping non-unix variants")
	}

	rootdir := testutil.NewEnv(testutil.EnvOpts{T: t})
	orgInstallDir := github.InstallDir
	github.InstallDir = rootdir
	defer func() { github.InstallDir = orgInstallDir }()

	version.Now = func() (t time.Time) { return t }

	var stdout bytes.Buffer
	args := testutil.SplitArgs("version")
	opts := testutil.MockGlobalData(args, &stdout)
	opts.Versioners = global.Versioners{
		Viceroy: github.New(github.Opts{Org: "fastly", Repo: "viceroy", Binary: "viceroy"}),
	}
	app.Init = func(_ []string, _ io.Reader) (*global.Data, error) { return opts, nil }
	if err := app.Run(args, nil); err != nil {
		t.Fatal(err)
	}

	var mockTime time.Time
	want := strings.Join([]string{
		"Fastly CLI version v0.0.0-unknown (unknown)",
		fmt.Sprintf("Built with go version %s %s/%s (%s)", runtime.Version(), runtime.GOOS, runtime.GOARCH, mockTime.Format("2006-01-02")),
		fmt.Sprintf("Viceroy version: viceroy %s", viceroy.Version()),
		"",
	}, "\n")
	if stdout.String() != want {
		t.Errorf("unexpected output:\n got: %q\nwant: %q", stdout.String(), want)
	}
}
