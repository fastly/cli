package testutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/errors"
	"github.com/google/go-cmp/cmp"
)

// AssertEqual fatals a test if the parameters aren't equal.
func AssertEqual(t *testing.T, want, have interface{}) {
	t.Helper()
	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatal(diff)
	}
}

// AssertBool fatals a test if the parameters aren't equal.
func AssertBool(t *testing.T, want, have bool) {
	t.Helper()
	if want != have {
		t.Fatalf("want %v, have %v", want, have)
	}
}

// AssertString fatals a test if the parameters aren't equal.
func AssertString(t *testing.T, want, have string) {
	t.Helper()
	if want != have {
		t.Fatal(cmp.Diff(want, have))
	}
}

// AssertStringContains fatals a test if the string doesn't contain a substring.
func AssertStringContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("%q doesn't contain %q", s, substr)
	}
}

// AssertStringDoesntContain fatals a test if the string does contain a substring.
func AssertStringDoesntContain(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("%q contains %q", s, substr)
	}
}

// AssertNoError fatals a test if the error is not nil.
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertErrorContains fatals a test if the error's Error string doesn't contain
// target. As a special case, if target is the empty string, we assume the error
// should be nil.
func AssertErrorContains(t *testing.T, err error, target string) {
	t.Helper()
	switch {
	case err == nil && target == "":
		return // great
	case err == nil && target != "":
		t.Fatalf("want %q, have no error", target)
	case err != nil && target == "":
		t.Fatalf("want no error, have %v", err)
	case err != nil && target != "":
		if want, have := target, err.Error(); !strings.Contains(have, want) {
			t.Fatalf("want %q, have %q", want, have)
		}
	}
}

// AssertRemediationErrorContains fatals a test if the error's RemediationError
// remediation string doesn't contain target. As a special case, if target is
// the empty string, we assume the error should be nil.
func AssertRemediationErrorContains(t *testing.T, err error, target string) {
	t.Helper()

	re, ok := err.(errors.RemediationError)

	switch {
	case err == nil && target == "":
		return // great
	case err == nil && target != "":
		t.Fatalf("want %q, have no error", target)
	case err != nil && target != "" && !ok:
		t.Fatal("have no RemediationError")
	case err != nil && target != "":
		if want, have := target, re.Remediation; !strings.Contains(have, want) {
			t.Fatalf("want %q, have %q", want, have)
		}
	}
}

// AssertPathContentFlag errors a test scenario if the given flag value hasn't
// been parsed as expected.
//
// Example: Some flags will internally be passed to `cmd.Content` to acquire
// the value. If passed a file path, then we expect the testdata/<fixture> to
// have been read, otherwise we expect the given flag value to have been used.
func AssertPathContentFlag(flag string, wantError string, args []string, fixture string, content string, t *testing.T) {
	if wantError == "" {
		for i, a := range args {
			if a == fmt.Sprintf("--%s", flag) {
				want := args[i+1]
				if want == fmt.Sprintf("./testdata/%s", fixture) {
					want = cmd.Content(want)
				}
				if content != want {
					t.Errorf("wanted %s, have %s", want, content)
				}
				break
			}
		}
	}
}
