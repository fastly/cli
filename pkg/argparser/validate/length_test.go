package validate_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/argparser/validate"
	fsterr "github.com/fastly/cli/pkg/errors"
)

func TestLengthValidators(t *testing.T) {
	scenarios := []struct {
		name            string
		action          kingpin.Action
		value           string
		wantError       string
		wantRemediation string
	}{
		{
			name:   "LengthBetween accepts a value at the minimum",
			action: validate.LengthBetween(2, 4),
			value:  "ab",
		},
		{
			name:   "LengthBetween accepts a value at the maximum",
			action: validate.LengthBetween(2, 4),
			value:  "abcd",
		},
		{
			name:            "LengthBetween rejects a value below the minimum",
			action:          validate.LengthBetween(2, 4),
			value:           "a",
			wantError:       "--flag must be between 2 and 4 characters, got 1",
			wantRemediation: "Provide a value for --flag that is between 2 and 4 characters long.",
		},
		{
			name:            "LengthBetween rejects a value above the maximum",
			action:          validate.LengthBetween(2, 4),
			value:           "abcde",
			wantError:       "--flag must be between 2 and 4 characters, got 5",
			wantRemediation: "Provide a value for --flag that is between 2 and 4 characters long.",
		},
		{
			name:   "MaxLength accepts an empty value",
			action: validate.MaxLength(4),
			value:  "",
		},
		{
			name:            "MaxLength rejects a value above the maximum",
			action:          validate.MaxLength(4),
			value:           "abcde",
			wantError:       "--flag must be between 0 and 4 characters, got 5",
			wantRemediation: "Provide a value for --flag that is between 0 and 4 characters long.",
		},
		{
			name:   "MinLength has no upper bound",
			action: validate.MinLength(2),
			value:  strings.Repeat("a", 10000),
		},
		{
			name:            "MinLength rejects a value below the minimum",
			action:          validate.MinLength(2),
			value:           "a",
			wantError:       "--flag must be at least 2 characters, got 1",
			wantRemediation: "Provide a value for --flag that is at least 2 characters long.",
		},
		{
			name:            "MinLength(1) reports an empty value",
			action:          validate.MinLength(1),
			value:           "",
			wantError:       "--flag must not be empty",
			wantRemediation: "Provide a value for --flag that is not empty.",
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			app := kingpin.New("test", "")
			var dst string
			app.Flag("flag", "").Action(s.action).StringVar(&dst)

			_, err := app.Parse([]string{"--flag=" + s.value})
			assertValidationError(t, err, s.wantError, s.wantRemediation)
		})
	}
}

func TestLengthValidatorsNamePositionalArgs(t *testing.T) {
	app := kingpin.New("test", "")
	var dst string
	app.Arg("thing", "").Action(validate.MinLength(1)).StringVar(&dst)

	_, err := app.Parse([]string{""})
	assertValidationError(t, err, "thing must not be empty", "Provide a value for thing that is not empty.")
}

func assertValidationError(t *testing.T, err error, wantError, wantRemediation string) {
	t.Helper()
	if wantError == "" {
		if err != nil {
			t.Fatalf("want no error, have %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("want error %q, have nil", wantError)
	}
	if have := err.Error(); have != wantError {
		t.Fatalf("want error %q, have %q", wantError, have)
	}
	var re fsterr.RemediationError
	if !errors.As(err, &re) {
		t.Fatalf("want a RemediationError, have %T", err)
	}
	if re.Remediation != wantRemediation {
		t.Fatalf("want remediation %q, have %q", wantRemediation, re.Remediation)
	}
}
