package compute

import (
	"os/exec"
	"strings"
)

func getJsToolchainBinPath(bin string) (string, error) {
	path, err := exec.Command(bin, "bin").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(path)), nil
}

func checkJsPackageDependencyExists(bin, name string) bool {
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	err := exec.Command(bin, "list", "--json", "--depth", "0", name).Run()
	return err == nil
}
