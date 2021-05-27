package common

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/go-fastly/v3/fastly"
	"github.com/fastly/kingpin"
)

// Command is an interface that abstracts over all of the concrete command
// structs. The Name method lets us select which command should be run, and the
// Exec method invokes whatever business logic the command should do.
type Command interface {
	Name() string
	Exec(in io.Reader, out io.Writer) error
}

// SelectCommand chooses the command matching name, if it exists.
func SelectCommand(name string, commands []Command) (Command, bool) {
	for _, command := range commands {
		if command.Name() == name {
			return command, true
		}
	}
	return nil, false
}

// Registerer abstracts over a kingpin.App and kingpin.CmdClause. We pass it to
// each concrete command struct's constructor as the "parent" into which the
// command should install itself.
type Registerer interface {
	Command(name, help string) *kingpin.CmdClause
}

// Globals are flags and other stuff that's useful to every command. Globals are
// passed to each concrete command's constructor as a pointer, and are populated
// after a call to Parse. A concrete command's Exec method can use any of the
// information in the globals.
type Globals struct {
	Token   string
	Verbose bool
	Client  api.Interface
}

// Base is stuff that should be included in every concrete command.
type Base struct {
	CmdClause *kingpin.CmdClause
	Globals   *config.Data
}

// Name implements the Command interface, and returns the FullCommand from the
// kingpin.Command that's used to select which command to actually run.
func (b Base) Name() string {
	return b.CmdClause.FullCommand()
}

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
	clause := b.CmdClause.Flag("version", "Number of service version, 'latest' or 'active'")
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

// Optional models an optional type that consumers can use to assert whether the
// inner value has been set and is therefore valid for use.
type Optional struct {
	WasSet bool
}

// Set implements kingpin.Action and is used as callback to set that the optional
// inner value is valid.
func (o *Optional) Set(e *kingpin.ParseElement, c *kingpin.ParseContext) error {
	o.WasSet = true
	return nil
}

// OptionalString models an optional string flag value.
type OptionalString struct {
	Optional
	Value string
}

// OptionalStringSlice models an optional string slice flag value.
type OptionalStringSlice struct {
	Optional
	Value []string
}

// OptionalBool models an optional boolean flag value.
type OptionalBool struct {
	Optional
	Value bool
}

// OptionalUint models an optional uint flag value.
type OptionalUint struct {
	Optional
	Value uint
}

// OptionalUint8 models an optional unit8 flag value.
type OptionalUint8 struct {
	Optional
	Value uint8
}

// OptionalInt models an optional int flag value.
type OptionalInt struct {
	Optional
	Value int
}

// OptionalServiceVersion represents a Fastly service version.
type OptionalServiceVersion struct {
	OptionalString
}

// Parse returns a service version based on the given user input.
func (v *OptionalServiceVersion) Parse(sid string, c api.Interface) (*fastly.Version, error) {
	vs, err := c.ListVersions(&fastly.ListVersionsInput{
		ServiceID: sid,
	})
	if err != nil || len(vs) == 0 {
		return nil, fmt.Errorf("error listing service versions: %w", err)
	}

	// NOTE: The decision to sort the versions by 'most recently updated' was
	// originally discussed/agreed in github.com/fastly/cli/pull/50
	sort.Slice(vs, func(i, j int) bool {
		return vs[i].UpdatedAt.Before(*vs[j].UpdatedAt)
	})

	var version *fastly.Version

	switch strings.ToLower(v.Value) {
	case "active":
		version, err = getLatestActiveVersion(vs)
	case "latest":
		version, err = getLatestVersion(vs)
	case "":
		version, err = getLatestActiveVersion(vs)
		if err != nil {
			version, err = getLatestVersion(vs)
		}
	default:
		version, err = getSpecifiedVersion(sid, v.Value, c)
	}
	if err != nil {
		return nil, err
	}

	return version, nil
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
	if v.Active || v.Locked {
		version, err := c.CloneVersion(&fastly.CloneVersionInput{
			ServiceID:      sid,
			ServiceVersion: v.Number,
		})
		if err != nil {
			return nil, fmt.Errorf("error cloning latest service version: %w", err)
		}
		return version, nil
	}

	// Treat the function as a no-op if the version is editable.
	return v, nil
}

// getLatestVersion returns the latest 'locked' version (that isn't 'active'),
// otherwise it will return the latest 'draft' version.
//
// The reason we don't consider an 'active' version as part of this algorithm
// is because the --version flag accepts 'active' as a distinct requirement for
// which we have an explicit function to handle that behaviour.
//
// NOTE: We iterate over the slice in reverse as we would expect the latest
// locked version to be nearer the end of the slice (i.e. nearer to a more
// recently updated version than at the start of the slice) and so with this
// implementation we can short-circuit the loop rather than iterating over the
// entire collection, which worst case would be O(n).
func getLatestVersion(vs []*fastly.Version) (*fastly.Version, error) {
	var locked, latest *fastly.Version

	for i := len(vs) - 1; i >= 0; i-- {
		v := vs[i]

		// NOTE: The decision to prefer 'locked' over a 'draft' (i.e. an editable
		// service version) was discussed/agreed in github.com/fastly/cli/pull/50
		if v.Locked && !v.Active {
			locked = v
			break
		}
		if !v.Active && latest == nil {
			latest = v
		}
	}

	v := latest
	if locked != nil {
		v = locked
	}

	if v == nil {
		return nil, fmt.Errorf("error finding latest service version")
	}
	return v, nil
}

// getLatestActiveVersion returns the latest active service version.
//
// NOTE: We iterate over the slice in reverse as we would expect the latest
// active version to be nearer the end of the slice (i.e. nearer to a more
// recently updated version than at the start of the slice).
func getLatestActiveVersion(vs []*fastly.Version) (*fastly.Version, error) {
	for i := len(vs) - 1; i >= 0; i-- {
		if vs[i].Active {
			return vs[i], nil
		}
	}
	return nil, fmt.Errorf("error locating latest active service version")
}

// getSpecifiedVersion returns the specified service version.
func getSpecifiedVersion(sid string, version string, c api.Interface) (*fastly.Version, error) {
	i, err := strconv.Atoi(version)
	if err != nil {
		return nil, err
	}

	vs, err := c.GetVersion(&fastly.GetVersionInput{
		ServiceID:      sid,
		ServiceVersion: i,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting specified service version: %w", err)
	}
	return vs, nil
}
