//go:build !viceroy_embed

package viceroy

import (
	"errors"
	"testing"
)

func TestNoEmbedSupported(t *testing.T) {
	if Supported() {
		t.Fatal("Supported() = true in a build without -tags viceroy_embed")
	}
}

func TestNoEmbedExtractReturnsErrUnsupported(t *testing.T) {
	_, err := Extract(t.TempDir())
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("Extract() error = %v, want ErrUnsupported", err)
	}
}
