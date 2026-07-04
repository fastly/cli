package compute

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"

	"github.com/fastly/cli/pkg/check"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/filesystem"
	"github.com/fastly/cli/pkg/github"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/text"
)

var viceroyError = fsterr.RemediationError{
	Inner:       fmt.Errorf("a Viceroy version was not found"),
	Remediation: fsterr.BugRemediation,
}

// viceroyInstaller installs/updates the Viceroy binary. It is the shared
// implementation used by both `compute serve` and `compute install-tools`.
type viceroyInstaller struct {
	Globals   *global.Data
	Versioner github.AssetVersioner
	// BinPath is the user provided binary path (--viceroy-path). Empty if unset.
	BinPath string
	// ForceCheckLatest forces a check for a newer release (--viceroy-check).
	ForceCheckLatest bool
}

// get returns the path to the installed Viceroy binary.
//
// If Viceroy is installed we either update it or pin it to the version defined
// in the fastly.toml [viceroy.viceroy_version]. Otherwise, if not installed, we
// install it in the same directory as the application configuration data.
//
// In the case of a network failure we fallback to the latest installed version of the
// Viceroy binary as long as one is installed and has the correct permissions.
func (vi viceroyInstaller) get(spinner text.Spinner, out io.Writer, manifestPath string) (bin string, err error) {
	if vi.BinPath != "" {
		if vi.Globals.Verbose() {
			text.Info(out, "Using user provided install of Viceroy via --viceroy-path flag: %s\n\n", vi.BinPath)
		}
		return filepath.Abs(vi.BinPath)
	}

	// Allows a user to use a version of Viceroy that is installed in the $PATH.
	if usePath := os.Getenv("FASTLY_VICEROY_USE_PATH"); checkViceroyEnvVar(usePath) {
		path, err := exec.LookPath("viceroy")
		if err != nil {
			return "", fmt.Errorf("failed to lookup viceroy binary in user $PATH (user has set $FASTLY_VICEROY_USE_PATH): %w", err)
		}
		if vi.Globals.Verbose() {
			text.Info(out, "Using user provided install of Viceroy via $PATH lookup: %s\n\n", path)
		}
		return filepath.Abs(path)
	}

	bin = filepath.Join(github.InstallDir, vi.Versioner.BinaryName())

	// NOTE: When checking if Viceroy is installed we don't use
	// exec.LookPath("viceroy") because PATH is unreliable across OS platforms,
	// but also we actually install Viceroy in the same location as the
	// application configuration, which means it wouldn't be found looking up by
	// the PATH env var. We could pass the path for the application configuration
	// into exec.LookPath() but it's simpler to just execute the binary.
	//
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	// nosemgrep
	command := exec.Command(bin, "--version")

	var installedVersion string

	stdoutStderr, err := command.CombinedOutput()
	if err != nil {
		vi.Globals.ErrLog.Add(err)
	} else {
		// Check the version output has the expected format: `viceroy 0.1.0`
		installedVersion = strings.TrimSpace(string(stdoutStderr))
		segs := strings.Split(installedVersion, " ")
		if len(segs) < 2 {
			return bin, viceroyError
		}
		installedVersion = segs[1]
	}

	// If the user hasn't explicitly set a Viceroy version, then we'll use
	// whatever the latest version is.
	versionToInstall := "latest"
	if v := vi.Versioner.RequestedVersion(); v != "" {
		versionToInstall = v

		if _, err := semver.Parse(versionToInstall); err != nil {
			return bin, fsterr.RemediationError{
				Inner:       fmt.Errorf("failed to parse configured version as a semver: %w", err),
				Remediation: fmt.Sprintf("Ensure the %s `viceroy_version` value '%s' (under the [local_server] section) is a valid semver (https://semver.org/), e.g. `0.1.0`)", manifestPath, versionToInstall),
			}
		}
	}

	err = vi.install(installedVersion, versionToInstall, manifestPath, bin, spinner)
	if err != nil {
		vi.Globals.ErrLog.Add(err)
		return bin, err
	}

	err = github.SetBinPerms(bin)
	if err != nil {
		vi.Globals.ErrLog.Add(err)
		return bin, err
	}
	return bin, nil
}

// checkViceroyEnvVar indicates if the CLI should use a Viceroy binary exposed
// on the user's $PATH.
func checkViceroyEnvVar(value string) bool {
	switch strings.ToUpper(value) {
	case "1", "TRUE":
		return true
	}
	return false
}

// install downloads the binary from GitHub.
//
// The logic flow is as follows:
//
// 1. Check if version to install is "latest"
// 2. If so, check the latest release matches the installed version.
// 3. If not latest, check the installed version matches the expected version.
func (vi viceroyInstaller) install(
	installedVersion, versionToInstall, manifestPath, bin string,
	spinner text.Spinner,
) error {
	var (
		err         error
		msg, tmpBin string
	)

	switch {
	case installedVersion == "": // Viceroy not installed
		if vi.Globals.Verbose() {
			text.Info(vi.Globals.Output, "Viceroy is not already installed, so we will install the %s version.\n\n", versionToInstall)
		}
		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Fetching Viceroy release: %s", versionToInstall)
		spinner.Message(msg + "...")

		if versionToInstall == "latest" {
			tmpBin, err = vi.Versioner.DownloadLatest()
		} else {
			tmpBin, err = vi.Versioner.DownloadVersion(versionToInstall)
		}
	case versionToInstall != "latest":
		if installedVersion == versionToInstall {
			if vi.Globals.Verbose() {
				text.Info(vi.Globals.Output, "Viceroy is already installed, and the installed version matches the required version (%s) in the %s file.\n\n", versionToInstall, manifestPath)
			}
			return nil
		}
		if vi.Globals.Verbose() {
			text.Info(vi.Globals.Output, "Viceroy is already installed, but the installed version (%s) doesn't match the required version (%s) specified in the %s file.\n\n", installedVersion, versionToInstall, manifestPath)
		}

		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Fetching Viceroy release: %s", versionToInstall)
		spinner.Message(msg + "...")

		tmpBin, err = vi.Versioner.DownloadVersion(versionToInstall)
	case versionToInstall == "latest":
		// Viceroy is already installed, so we check if the installed version matches the latest.
		// But we'll skip that check if the TTL for the Viceroy LastChecked hasn't expired.

		stale := check.Stale(vi.Globals.Config.Viceroy.LastChecked, vi.Globals.Config.Viceroy.TTL)
		if !stale && !vi.ForceCheckLatest {
			if vi.Globals.Verbose() {
				text.Info(vi.Globals.Output, "Viceroy is installed but the CLI config (`fastly config`) shows the TTL, checking for a newer version, hasn't expired. To force a refresh, re-run the command with the `--viceroy-check` flag.\n\n")
			}
			return nil
		}

		// IMPORTANT: We declare separately so to shadow `err` from parent scope.
		var latestVersion string

		// NOTE: We won't stop the user because although we can't request the latest
		// version of the tool, the user may have a local version already installed.
		err = spinner.Process("Checking latest Viceroy release", func(_ *text.SpinnerWrapper) error {
			latestVersion, err = vi.Versioner.LatestVersion()
			if err != nil {
				return fsterr.RemediationError{
					Inner:       fmt.Errorf("error fetching latest version: %w", err),
					Remediation: fsterr.NetworkRemediation,
				}
			}
			return nil
		})
		if err != nil {
			return nil // short-circuit the rest of this function
		}

		viceroyConfig := vi.Globals.Config.Viceroy
		viceroyConfig.LatestVersion = latestVersion
		viceroyConfig.LastChecked = time.Now().Format(time.RFC3339)

		// Before attempting to write the config data back to disk we need to
		// ensure we reassign the modified struct which is a copy (not reference).
		vi.Globals.Config.Viceroy = viceroyConfig

		err = vi.Globals.Config.Write(vi.Globals.ConfigPath)
		if err != nil {
			return err
		}

		if vi.Globals.Verbose() {
			text.Info(vi.Globals.Output, "\nThe CLI config (`fastly config`) has been updated with the latest Viceroy version: %s\n\n", latestVersion)
		}

		if installedVersion != "" && installedVersion == latestVersion {
			return nil
		}

		err = spinner.Start()
		if err != nil {
			return err
		}
		msg = fmt.Sprintf("Fetching Viceroy release: %s", versionToInstall)
		spinner.Message(msg + "...")

		tmpBin, err = vi.Versioner.DownloadLatest()
	}

	// NOTE: The above `switch` needs to shadow the function-level `err` variable.
	if err != nil {
		err = fmt.Errorf("error downloading Viceroy release: %w", err)
		spinner.StopFailMessage(msg)
		spinErr := spinner.StopFail()
		if spinErr != nil {
			return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
		}
		return err
	}
	defer os.RemoveAll(tmpBin)

	if err := os.Rename(tmpBin, bin); err != nil {
		err = fmt.Errorf("failed to rename/move file: %w", err)
		if copyErr := filesystem.CopyFile(tmpBin, bin); copyErr != nil {
			err = fmt.Errorf("failed to copy file: %w (original error: %w)", copyErr, err)
			spinner.StopFailMessage(msg)
			spinErr := spinner.StopFail()
			if spinErr != nil {
				return fmt.Errorf(text.SpinnerErrWrapper, spinErr, err)
			}
			return err
		}
	}

	spinner.StopMessage(msg)
	return spinner.Stop()
}
