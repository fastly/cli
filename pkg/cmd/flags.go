package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/fastly/cli/pkg/api"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/fastly/kingpin"
)

// ServiceVersionFlagOpts enables easy configuration of the --version flag
// defined via the NewServiceVersionFlag constructor.
//
// NOTE: The reason we define an 'optional' field rather than a 'required'
// field is because 99% of the use cases where --version is defined the flag
// will be required, and so we cater for the common case. Meaning only those
// subcommands that have --version as optional will need to set that field.
type ServiceVersionFlagOpts struct {
	Dst      *string
	Optional bool
	Action   kingpin.Action
}

// NewServiceVersionFlag defines a --version flag that accepts multiple values
// such as 'latest', 'active' and numerical values which are then converted
// into the appropriate service version.
func (b Base) NewServiceVersionFlag(opts ServiceVersionFlagOpts, args ...string) {
	clause := b.CmdClause.Flag("version", "Number of service version, 'latest', 'active' or 'editable'")
	if !opts.Optional {
		clause = clause.Required()
	} else {
		clause = clause.Action(opts.Action)
	}
	clause.StringVar(opts.Dst)
}

// NewAutoCloneFlag defines a --autoclone flag that will cause a clone of the
// identified service version if its found to be active or locked.
func (b Base) NewAutoCloneFlag(action kingpin.Action, dst *bool) {
	b.CmdClause.Flag("autoclone", "Automatically clone the identified service version if it's 'locked' or 'active'").Action(action).BoolVar(dst)
}

// OptionalServiceVersion represents a Fastly service version.
type OptionalServiceVersion struct {
	OptionalString
}

// Parse returns a service version based on the given user input.
func (sv *OptionalServiceVersion) Parse(sid string, c api.Interface) (*fastly.Version, error) {
	vs, err := c.ListVersions(&fastly.ListVersionsInput{
		ServiceID: sid,
	})
	if err != nil || len(vs) == 0 {
		return nil, fmt.Errorf("error listing service versions: %w", err)
	}

	// Sort versions into descending order.
	sort.Slice(vs, func(i, j int) bool {
		return vs[i].Number > vs[j].Number
	})

	var v *fastly.Version

	switch strings.ToLower(sv.Value) {
	case "latest":
		return vs[0], nil
	case "active":
		v, err = getLatestActiveVersion(vs)
	case "editable":
		v, err = getLatestEditableVersion(vs)
	case "":
		v, err = getLatestEditableVersion(vs)
	default:
		v, err = getSpecifiedVersion(vs, sv.Value)
	}
	if err != nil {
		return nil, err
	}

	return v, nil
}

// OptionalAutoClone defines a method set for abstracting the logic required to
// identify if a given service version needs to be cloned.
type OptionalAutoClone struct {
	OptionalBool
}

// Parse returns a service version.
//
// The returned version is either the same as the input argument `v` or it's a
// cloned version if the input argument was either active or locked.
func (ac *OptionalAutoClone) Parse(v *fastly.Version, sid string, c api.Interface) (*fastly.Version, error) {
	// if user didn't provide --autoclone flag
	if !ac.Value && (v.Active || v.Locked) {
		return nil, fmt.Errorf("service version %d is not editable", v.Number)
	}
	if ac.Value && (v.Active || v.Locked) {
		version, err := c.CloneVersion(&fastly.CloneVersionInput{
			ServiceID:      sid,
			ServiceVersion: v.Number,
		})
		if err != nil {
			return nil, fmt.Errorf("error cloning service version: %w", err)
		}
		return version, nil
	}

	// Treat the function as a no-op if the version is editable.
	return v, nil
}

// getLatestActiveVersion returns the latest active service version.
func getLatestActiveVersion(vs []*fastly.Version) (*fastly.Version, error) {
	for _, v := range vs {
		if v.Active {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no active service version found")
}

// getLatestEditableVersion returns the latest editable service version.
//
// The returned version will be ahead of the latest active version. If there is
// no editable version above the latest active, an error will be returned.
//
// When no --version flag is provided, this algorithm helps handle cases where
// a command (such as `backend create`) is executable multiple times, as the
// latest editable version will be reused and subsequently will contain each
// new backend created from the prior executions.
//
// There could be a scenario where a customer has a single service version
// which is activated and so there would be no match for an editable version
// without us first cloning it. The act of cloning should be a decision made by
// the user (i.e. --autoclone) and so this function will return an error if no
// editable version available.
func getLatestEditableVersion(vs []*fastly.Version) (*fastly.Version, error) {
	for _, v := range vs {
		if v.Active {
			break
		}
		if !v.Locked {
			return v, nil
		}
	}
	return nil, fsterr.RemediationError{
		Inner:       errors.New("error retrieving an editable service version: no editable version found"),
		Remediation: "Please retry with --version=[latest|active] and --autoclone",
	}
}

// getSpecifiedVersion returns the specified service version.
func getSpecifiedVersion(vs []*fastly.Version, version string) (*fastly.Version, error) {
	i, err := strconv.Atoi(version)
	if err != nil {
		return nil, err
	}

	for _, v := range vs {
		if v.Number == i {
			return v, nil
		}
	}

	return nil, fmt.Errorf("specified service version not found: %s", version)
}
