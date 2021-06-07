package cmd

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/text"
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

// SetServiceVersionFlag defines a --version flag that accepts multiple values
// such as 'latest', 'active' and numerical values which are then converted
// into the appropriate service version.
func (b Base) SetServiceVersionFlag(opts ServiceVersionFlagOpts, args ...string) {
	clause := b.CmdClause.Flag("version", "'latest', 'active', or the number of a specific version")
	if !opts.Optional {
		clause = clause.Required()
	} else {
		clause = clause.Action(opts.Action)
	}
	clause.StringVar(opts.Dst)
}

// SetAutoCloneFlag defines a --autoclone flag that will cause a clone of the
// identified service version if its found to be active or locked.
func (b Base) SetAutoCloneFlag(action kingpin.Action, dst *bool) {
	b.CmdClause.Flag("autoclone", "If the selected service version is not editable, clone it and use the clone.").Action(action).BoolVar(dst)
}

// OptionalServiceVersion represents a Fastly service version.
type OptionalServiceVersion struct {
	OptionalString
	Client api.Interface
}

// Parse returns a service version based on the given user input.
func (sv *OptionalServiceVersion) Parse(sid string) (*fastly.Version, error) {
	vs, err := sv.Client.ListVersions(&fastly.ListVersionsInput{
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
	case "":
		return vs[0], nil
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
	Out    io.Writer
	Client api.Interface
}

// Parse returns a service version.
//
// The returned version is either the same as the input argument `v` or it's a
// cloned version if the input argument was either active or locked.
func (ac *OptionalAutoClone) Parse(v *fastly.Version, sid string) (*fastly.Version, error) {
	// if user didn't provide --autoclone flag
	if !ac.Value && (v.Active || v.Locked) {
		return nil, fmt.Errorf("service version %d is not editable", v.Number)
	}
	if ac.Value && (v.Active || v.Locked) {
		version, err := ac.Client.CloneVersion(&fastly.CloneVersionInput{
			ServiceID:      sid,
			ServiceVersion: v.Number,
		})
		if err != nil {
			return nil, fmt.Errorf("error cloning service version: %w", err)
		}
		msg := fmt.Sprintf("Service version %d is not editable, so it was automatically cloned because --autoclone is enabled. Now operating on version %d.", v.Number, version.Number)
		text.Output(ac.Out, msg)
		text.Break(ac.Out)
		return version, nil
	}

	// Treat the function as a no-op if the version is editable.
	return v, nil
}

// getLatestActiveVersion returns the active service version.
func getLatestActiveVersion(vs []*fastly.Version) (*fastly.Version, error) {
	for _, v := range vs {
		if v.Active {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no active service version found")
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
