package validate

import (
	"fmt"

	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/kingpin"
)

// ValidateLength ensures the byte length of an argument.
// `minLen` and `maxLen` are inclusive bounds.
//
// We measure byte length instead of utf-8 chars / runes
// since what counts as a character varies across languages & string types,
// and the backend APIs likely measure byte length.
func validateLength(minLen int, maxLen *int) kingpin.Action {
	return func(element *kingpin.ParseElement, _ *kingpin.ParseContext) error {
		if element.Value == nil {
			return nil
		}

		name := elementName(element)
		length := len(*element.Value)
		if length < minLen || (maxLen != nil && length > *maxLen) {
			switch {
			case maxLen == nil && minLen == 1:
				return fsterr.RemediationError{
					Inner:       fmt.Errorf("%s must not be empty", name),
					Remediation: fmt.Sprintf("Provide a value for %s that is not empty.", name),
				}
			case maxLen == nil:
				return fsterr.RemediationError{
					Inner:       fmt.Errorf("%s must be at least %d bytes, got %d", name, minLen, length),
					Remediation: fmt.Sprintf("Provide a value for %s that is at least %d bytes long.", name, minLen),
				}
			default:
				return fsterr.RemediationError{
					Inner:       fmt.Errorf("%s must be between %d and %d bytes, got %d", name, minLen, *maxLen, length),
					Remediation: fmt.Sprintf("Provide a value for %s that is between %d and %d bytes long.", name, minLen, *maxLen),
				}
			}
		}
		return nil
	}
}

// MinLength validates that the arg value is at least the specified minimum byte length.
func MinLength(minLen int) kingpin.Action {
	return validateLength(minLen, nil)
}

// MaxLength validates that the arg value is at most the specified maximum byte length.
func MaxLength(maxLen int) kingpin.Action {
	return validateLength(0, &maxLen)
}

// LengthBetween validates that the arg value is between the specified minimum and maximum byte lengths.
func LengthBetween(minLen, maxLen int) kingpin.Action {
	return validateLength(minLen, &maxLen)
}

// elementName gets the model name of the provided ParseElement,
// and prefixes it with "--" if it is a flag.
func elementName(element *kingpin.ParseElement) string {
	var name string
	switch {
	case element.OneOf.Flag != nil:
		name = "--" + element.OneOf.Flag.Model().Name
	case element.OneOf.Arg != nil:
		name = element.OneOf.Arg.Model().Name
	}

	return name
}
