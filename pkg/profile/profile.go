package profile

import (
	"github.com/fastly/cli/pkg/config"
)

// DoesNotExist describes an output warning message.
const DoesNotExist = "The specified profile does not exist."

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
		if v.Default == true {
			return k, v
		}
	}
	return "", new(config.Profile)
}

// Set configures the named profile to be the default.
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
