package profile

import (
	"github.com/fastly/cli/pkg/config"
)

// DefaultName is the default profile name.
const DefaultName = "user"

// DoesNotExist describes an output error/warning message.
const DoesNotExist = "the profile '%s' does not exist"

// NoDefaults describes an output warning message.
const NoDefaults = "At least one account profile should be set as the 'default'. Run `fastly profile update <NAME>` and ensure the profile is set to be the default."

// TokenExpired is a token expiration error message.
const TokenExpired = "the token in profile '%s' expired at '%s'"

// TokenWillExpire is a token expiration error message.
const TokenWillExpire = "the token in profile '%s' will expire at '%s'"

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
	return "", nil
}

// Get returns the specified profile.
func Get(name string, p config.Profiles) *config.Profile {
	for k, v := range p {
		if k == name {
			return v
		}
	}
	return nil
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

// SetADefault sets one of the profiles to be the default.
//
// NOTE: This is used by the `sso` command.
// The reason it exists is because there could be profiles that for some reason
// the user has set them all to not be a default. So to avoid errors in the CLI
// we require at least one profile to be a default and this function makes it
// easy to just pick the first profile and generically set it as the default.
func SetADefault(p config.Profiles) (string, config.Profiles) {
	var profileName string
	for k, v := range p {
		profileName = k
		v.Default = true
		break
	}
	return profileName, p
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
