package update

import (
	"testing"
)

// TestBinary validates that the Binary method returns the expected binary name.
func TestName(t *testing.T) {
	want := "binary"
	gh := NewGitHub(GitHubOpts{"org", "repo", want})

	if have := gh.Binary(); have != want {
		t.Fatalf("want: %s, have: %s", want, have)
	}
}
