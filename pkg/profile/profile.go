package profile

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// DefaultName is the default profile name.
const DefaultName = "user"

// DoesNotExist describes an output error/warning message.
const DoesNotExist = "the profile '%s' does not exist"

// NoDefaults describes an output warning message.
const NoDefaults = "At least one account profile should be set as the 'default'. Run `fastly profile update <NAME>` and ensure the profile is set to be the default."

// Exist reports whether the given profile exists.
func Exist(name string, p config.Profiles) bool {
	for k := range p {
		if k == name {
			return true
		}
	}
	return false
}

// Default returns the default profile (which is the active profile).
func Default(p config.Profiles) (string, *config.Profile) {
	for k, v := range p {
		if v.Default {
			return k, v
		}
	}
	return "", new(config.Profile)
}

// Get returns the specified profile.
func Get(name string, p config.Profiles) (string, *config.Profile) {
	for k, v := range p {
		if k == name {
			return k, v
		}
	}
	return "", new(config.Profile)
}

// SetDefault configures the named profile to be the default.
//
// NOTE: The type assigned to the config.Profiles map key value is a struct.
// Structs are passed by value and so we must return the mutated type.
func SetDefault(name string, p config.Profiles) (config.Profiles, bool) {
	var ok bool
	for k, v := range p {
		v.Default = false
		if k == name {
			v.Default = true
			ok = true
		}
	}
	return p, ok
}

// Delete removes the named profile from the profile configuration.
func Delete(name string, p config.Profiles) bool {
	var ok bool
	for k := range p {
		if k == name {
			delete(p, k)
			ok = true
		}
	}
	return ok
}

// EditOption lets callers of Edit specify profile fields to update.
type EditOption func(*config.Profile)

// Edit modifies the named profile.
//
// IMPORTANT: We must return config.Profiles to safely update in-memory data.
// The type assigned to the config.Profiles map key value is a struct and
// structs are passed by value, so we must return the mutated type so the
// caller so they can reassign the updated struct back to the in-memory data
// and then persist that data back to disk.
func Edit(name string, p config.Profiles, opts ...EditOption) (config.Profiles, bool) {
	var ok bool
	for k, v := range p {
		if k == name {
			for _, opt := range opts {
				opt(v)
			}
			ok = true
		}
	}
	return p, ok
}

// Init checks if a profile flag is provided and potentially mutates token.
//
// NOTE: If the specified profile doesn't exist, then we'll let the user decide
// if the default profile (if available) is acceptable to use instead.
func Init(token string, m *manifest.Data, g *global.Data, in io.Reader, out io.Writer) (string, error) {
	// First check the fastly.toml manifest 'profile' field.
	profile := m.File.Profile

	// Otherwise check the --profile global flag.
	if profile == "" {
		profile = g.Flags.Profile
	}

	// If the user has specified no profile override, via flag nor manifest, then
	// we'll just return the token that has potentially been found within the
	// CLI's application configuration file.
	if profile == "" {
		return token, nil
	}

	name, p := Get(profile, g.Config.Profiles)
	if name != "" {
		return p.Token, nil
	}

	msg := fmt.Sprintf(DoesNotExist, profile)

	name, p = Default(g.Config.Profiles)
	if name == "" {
		msg = fmt.Sprintf("%s (no account profiles configured)", msg)
		return token, fsterr.RemediationError{
			Inner:       fmt.Errorf(msg),
			Remediation: fsterr.ProfileRemediation,
		}
	}

	// DoesNotExist is reused across errors and warning messages. Mostly errors
	// and so when used here for a warning message, we need to uppercase the
	// first letter so the warning reads like a proper sentence (where as golang
	// errors should always be lowercase).
	msg = fmt.Sprintf("%s%s. ", bytes.ToUpper([]byte(msg[:1])), msg[1:])

	msg = fmt.Sprintf("%sThe default profile '%s' (%s) will be used.", msg, name, p.Email)

	if !g.Flags.AutoYes {
		text.Warning(out, msg)

		label := "\nWould you like to continue? [y/N] "
		cont, err := text.AskYesNo(out, label, in)
		if err != nil {
			return token, err
		}
		if !cont {
			return token, errors.New("command execution cancelled")
		}
	}

	text.Break(out)
	return p.Token, nil
}
