package revision

import "testing"

func TestSemVer(t *testing.T) {
	got := SemVer("v1.0.0-xyz")
	want := "1.0.0"

	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}
