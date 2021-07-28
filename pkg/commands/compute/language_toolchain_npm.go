package compute

import (
	"os/exec"
	"strings"
)

func getNpmBinPath() (string, error) {
	path, err := exec.Command("npm", "bin").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(path)), nil
}

func checkPackageDependencyExists(name string) bool {
	// gosec flagged this:
	// G204 (CWE-78): Subprocess launched with variable
	// Disabling as the variables come from trusted sources.
	/* #nosec */
	err := exec.Command("npm", "list", "--json", "--depth", "0", name).Run()
	return err == nil
}
