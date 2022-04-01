package profile

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
)

// DoesNotExist describes an output error/warning message.
const DoesNotExist = "the specified profile does not exist"

// Exist reports whether the given profile exists.
func Exist(name string, p config.Profiles) bool {
	for k := range p {
		if k == name {
			return true
		}
	}
	return false
}

// Default returns the default profile.
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

// Set configures the named profile to be the default.
//
// NOTE: The type assigned to the config.Profiles map key value is a struct.
// Structs are passed by value and so we must return the mutated type.
func Set(name string, p config.Profiles) (config.Profiles, bool) {
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
// NOTE: The type assigned to the config.Profiles map key value is a struct.
// Structs are passed by value and so we must return the mutated type.
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
func Init(token string, data *manifest.Data, globals *config.Data, command cmd.Command, in io.Reader, out io.Writer) (string, error) {
	if data.File.Profile != "" {
		if name, p := Get(data.File.Profile, globals.File.Profiles); name != "" {
			token = p.Token
		}
	}

	if globals.Flag.Profile != "" && command.Name() != "configure" {
		if exist := Exist(globals.Flag.Profile, globals.File.Profiles); exist {
			// Persist the permanent switch of profiles.
			var ok bool
			if globals.File.Profiles, ok = Set(globals.Flag.Profile, globals.File.Profiles); ok {
				if err := globals.File.Write(globals.Path); err != nil {
					return token, err
				}
			}
		} else {
			msg := DoesNotExist

			name, p := Default(globals.File.Profiles)
			if name == "" {
				msg = fmt.Sprintf("%s (no account profiles configured)", msg)
				return token, fsterr.RemediationError{
					Inner:       fmt.Errorf(msg),
					Remediation: fsterr.ProfileRemediation,
				}
			}

			msg = fmt.Sprintf("%s The default profile '%s' (%s) will be used.", msg, name, p.Email)
			text.Warning(out, msg)

			label := "\nWould you like to continue? [y/N] "
			cont, err := text.AskYesNo(out, label, in)
			if err != nil {
				return token, err
			}
			if !cont {
				return token, nil
			}
			token = p.Token
		}
	}

	return token, nil
}
