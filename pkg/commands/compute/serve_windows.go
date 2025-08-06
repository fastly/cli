//go:build windows

package compute

import (
	"os/exec"
	"strconv"
)

func killProcess(pid int) error {
	// This is safe as the pid is obtained internally
	// nolint:gosec
	return exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid)).Run()
}
