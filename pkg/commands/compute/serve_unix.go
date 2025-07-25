//go:build !windows

package compute

import (
	"syscall"
)

func killProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}
