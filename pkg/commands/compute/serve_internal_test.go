package compute

import (
	"testing"

	"github.com/fsnotify/fsnotify"
)

// TestIsContentChange checks that content/existence events trigger a rebuild and
// attribute-only (Chmod) events do not.
func TestIsContentChange(t *testing.T) {
	scenarios := []struct {
		name string
		op   fsnotify.Op
		want bool
	}{
		{name: "create", op: fsnotify.Create, want: true},
		{name: "write", op: fsnotify.Write, want: true},
		{name: "remove", op: fsnotify.Remove, want: true},
		{name: "rename", op: fsnotify.Rename, want: true},
		{name: "chmod only", op: fsnotify.Chmod, want: false},
		{name: "write combined with chmod", op: fsnotify.Write | fsnotify.Chmod, want: true},
		{name: "no op", op: 0, want: false},
	}
	for _, testcase := range scenarios {
		t.Run(testcase.name, func(t *testing.T) {
			if got := isContentChange(testcase.op); got != testcase.want {
				t.Errorf("isContentChange(%s) = %v, want %v", testcase.op, got, testcase.want)
			}
		})
	}
}
