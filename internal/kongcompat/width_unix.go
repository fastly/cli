//go:build !appengine && (linux || freebsd || darwin || dragonfly || netbsd || openbsd)

package kongcompat

import (
	"io"
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

// guessWidth returns COLUMNS if it holds a number, else the width of the
// terminal w writes to, else 80.
func guessWidth(w io.Writer) int {
	if cols, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil {
		return cols
	}
	if t, ok := w.(*os.File); ok {
		var dimensions [4]uint16
		if _, _, err := syscall.Syscall6(
			syscall.SYS_IOCTL,
			t.Fd(),
			uintptr(syscall.TIOCGWINSZ),
			uintptr(unsafe.Pointer(&dimensions)),
			0, 0, 0,
		); err == 0 && dimensions[1] > 0 {
			return int(dimensions[1])
		}
	}
	return 80
}
