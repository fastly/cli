//go:build !windows

package auth_test

import (
	"errors"
	"testing"

	"github.com/creack/pty"

	fsterr "github.com/fastly/cli/pkg/errors"
)

func TestToken_TTY_Refused(t *testing.T) {
	// Create a PTY pair so we have a writable *os.File that
	// term.IsTerminal recognises as a terminal. This runs reliably
	// on Unix CI (no /dev/tty required) and, unlike os.Stdout, never
	// risks leaking a token to the developer's real terminal.
	ptm, pts, err := pty.Open()
	if err != nil {
		t.Fatalf("failed to open pty: %v", err)
	}
	defer ptm.Close()
	defer pts.Close()

	cmd := newTokenCommand(globalDataWithToken("secret-token"))
	err = cmd.Exec(nil, pts)
	if err == nil {
		t.Fatal("expected error when stdout is a terminal")
	}
	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("expected RemediationError, got %T: %v", err, err)
	}
	if re.Inner == nil || re.Inner.Error() != "refusing to print token to a terminal" {
		t.Errorf("unexpected inner error: %v", re.Inner)
	}
}
