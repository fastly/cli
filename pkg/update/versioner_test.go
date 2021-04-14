package update

import (
	"context"
	"testing"
)

// TestName validates that the Name method returns the expected binary name.
func TestName(t *testing.T) {
	want := "binary"
	gh := NewGitHub(context.Background(), "org", "repo", want)

	if have := gh.Name(); have != want {
		t.Fatalf("want: %s, have: %s", want, have)
	}
}

// TestRename validates that the Name method returns the expected binary name
// when the instance has been configured to rename the binary.
func TestRename(t *testing.T) {
	want := "foobar"

	gh := NewGitHub(context.Background(), "org", "repo", "binary")
	gh.RenameLocalBinary(want)

	if have := gh.Name(); have != want {
		t.Fatalf("want: %s, have: %s", want, have)
	}
}
