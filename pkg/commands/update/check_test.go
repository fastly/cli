package update_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/commands/update"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/mock"
	"github.com/google/go-cmp/cmp"
)

func TestCheck(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		current     string
		latest      github.Versioner
		wantCurrent semver.Version
		wantLatest  semver.Version
		wantUpdate  bool
	}{
		{
			name:    "empty current version",
			current: "",
			latest:  mock.Versioner{},
		},
		{
			name:    "invalid current version",
			current: "unknown",
			latest:  mock.Versioner{},
		},
		{
			name:        "same version",
			current:     "v1.2.3",
			latest:      mock.Versioner{AssetVersion: "1.2.3"},
			wantCurrent: semver.MustParse("1.2.3"),
			wantLatest:  semver.MustParse("1.2.3"),
			wantUpdate:  false,
		},
		{
			name:        "new version",
			current:     "v1.2.3",
			latest:      mock.Versioner{AssetVersion: "1.2.4"},
			wantCurrent: semver.MustParse("1.2.3"),
			wantLatest:  semver.MustParse("1.2.4"),
			wantUpdate:  true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			current, latest, shouldUpdate := update.Check(testcase.current, testcase.latest)
			if want, have := testcase.wantCurrent, current; !want.Equals(have) {
				t.Fatalf("current version: want %s, have %s", want, have)
			}
			if want, have := testcase.wantLatest, latest; !want.Equals(have) {
				t.Fatalf("latest version: want %s, have %s", want, have)
			}
			if want, have := testcase.wantUpdate, shouldUpdate; want != have {
				t.Fatalf("should update: want %v, have %v", want, have)
			}
		})
	}
}

func TestCheckAsync(t *testing.T) {
	for _, testcase := range []struct {
		name           string
		file           config.File
		currentVersion string
		cliVersioner   github.Versioner
		wantOutput     string
	}{
		{
			name:           "no last_check same version",
			currentVersion: "0.0.1",
			cliVersioner:   mock.Versioner{AssetVersion: "0.0.1"},
		},
		{
			name:           "no last_check new version",
			currentVersion: "0.0.1",
			cliVersioner:   mock.Versioner{AssetVersion: "0.0.2"},
			wantOutput:     "\nA new version of the Fastly CLI is available.\nCurrent version: 0.0.1\nLatest version: 0.0.2\nRun `fastly update` to get the latest version.\n\n",
		},
		{
			name:           "recent last_check new version",
			currentVersion: "0.0.1",
			cliVersioner:   mock.Versioner{AssetVersion: "0.0.2"},
			wantOutput:     "\nA new version of the Fastly CLI is available.\nCurrent version: 0.0.1\nLatest version: 0.0.2\nRun `fastly update` to get the latest version.\n\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			configFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("fastly_TestCheckAsync_%d", time.Now().UnixNano()))
			defer os.RemoveAll(configFilePath)

			var buf bytes.Buffer
			f := update.CheckAsync(
				testcase.currentVersion,
				testcase.cliVersioner,
				false,
			)
			f(&buf)

			if want, have := testcase.wantOutput, buf.String(); want != have {
				t.Error(cmp.Diff(want, have))
			}
		})
	}
}
