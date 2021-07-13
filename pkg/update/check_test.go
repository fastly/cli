package update_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/mock"
	"github.com/fastly/cli/pkg/update"
	"github.com/google/go-cmp/cmp"
)

func TestCheck(t *testing.T) {
	for _, testcase := range []struct {
		name        string
		current     string
		latest      update.Versioner
		wantError   string
		wantCurrent semver.Version
		wantLatest  semver.Version
		wantUpdate  bool
	}{
		{
			name:      "empty current version",
			current:   "",
			latest:    mock.Versioner{},
			wantError: "error reading current version: Version string empty",
		},
		{
			name:      "invalid current version",
			current:   "unknown",
			latest:    mock.Versioner{},
			wantError: "error reading current version: No Major.Minor.Patch elements found",
		},
		{
			name:      "latest version check fails",
			current:   "v1.0.0",
			latest:    mock.Versioner{Error: errors.New("kaboom")},
			wantError: "error fetching latest version: kaboom",
		},
		{
			name:        "same version",
			current:     "v1.2.3",
			latest:      mock.Versioner{Version: "v1.2.3"},
			wantCurrent: semver.MustParse("1.2.3"),
			wantLatest:  semver.MustParse("1.2.3"),
			wantUpdate:  false,
		},
		{
			name:        "new version",
			current:     "v1.2.3",
			latest:      mock.Versioner{Version: "v1.2.4"},
			wantCurrent: semver.MustParse("1.2.3"),
			wantLatest:  semver.MustParse("1.2.4"),
			wantUpdate:  true,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			current, latest, shouldUpdate, err := update.Check(context.Background(), testcase.current, testcase.latest)
			if testcase.wantError != "" {
				if want, have := testcase.wantError, err; want != have.Error() {
					t.Fatalf("error: want %q, have %q", want, have.Error())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
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
		cliVersioner   update.Versioner
		wantOutput     string
	}{
		{
			name:           "no last_check same version",
			currentVersion: "0.0.1",
			cliVersioner:   mock.Versioner{Version: "0.0.1"},
		},
		{
			name: "no last_check new version",
			file: config.File{
				CLI: config.CLI{
					TTL: "24h",
				},
			},
			currentVersion: "0.0.1",
			cliVersioner:   mock.Versioner{Version: "0.0.2"},
			wantOutput:     "\nA new version of the Fastly CLI is available.\nCurrent version: 0.0.1\nLatest version: 0.0.2\nRun `fastly update` to get the latest version.\n\n",
		},
		{
			name: "recent last_check new version",
			file: config.File{
				CLI: config.CLI{
					LastChecked: time.Now().Add(-4 * time.Hour).Format(time.RFC3339),
					TTL:         "5m",
				},
			},
			currentVersion: "0.0.1",
			cliVersioner:   mock.Versioner{Version: "0.0.2"},
			wantOutput:     "\nA new version of the Fastly CLI is available.\nCurrent version: 0.0.1\nLatest version: 0.0.2\nRun `fastly update` to get the latest version.\n\n",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			configFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("fastly_TestCheckAsync_%d", time.Now().UnixNano()))
			defer os.RemoveAll(configFilePath)

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			in := strings.NewReader("user input")
			var (
				out bytes.Buffer
				buf bytes.Buffer
			)
			f := update.CheckAsync(ctx, testcase.file, configFilePath, testcase.currentVersion, testcase.cliVersioner, in, &out)
			f(&buf)

			if want, have := testcase.wantOutput, buf.String(); want != have {
				t.Error(cmp.Diff(want, have))
			}
		})
	}
}
