package viceroy

import (
	"strings"
	"testing"
)

func TestVersionNonEmpty(t *testing.T) {
	v := Version()
	if v == "" {
		t.Fatal("Version() returned empty string")
	}
	if strings.ContainsAny(v, " \t\n\r") {
		t.Errorf("Version() = %q contains whitespace", v)
	}
}
