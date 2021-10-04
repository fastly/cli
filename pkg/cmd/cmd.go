package cmd

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/env"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
	"github.com/fastly/kingpin"
)

// Command is an interface that abstracts over all of the concrete command
// structs. The Name method lets us select which command should be run, and the
// Exec method invokes whatever business logic the command should do.
type Command interface {
	Name() string
	Exec(in io.Reader, out io.Writer) error
	Notes() string
}

// Select chooses the command matching name, if it exists.
func Select(name string, commands []Command) (Command, bool) {
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

// Notes is a no-op. It's up to each individual command to define this method.
func (b Base) Notes() string {
	return ""
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

// ServiceDetailsOpts provides data and behaviours required by the
// ServiceDetails function.
type ServiceDetailsOpts struct {
	AllowActiveLocked  bool
	AutoCloneFlag      OptionalAutoClone
	Client             api.Interface
	Manifest           manifest.Data
	Out                io.Writer
	ServiceVersionFlag OptionalServiceVersion
	VerboseMode        bool
}

// ServiceDetails returns the Service ID and Service Version.
func ServiceDetails(opts ServiceDetailsOpts) (serviceID string, serviceVersion *fastly.Version, err error) {
	serviceID, source := opts.Manifest.ServiceID()

	if opts.VerboseMode {
		DisplayServiceID(serviceID, source, opts.Out)
	}

	if source == manifest.SourceUndefined {
		return serviceID, serviceVersion, errors.ErrNoServiceID
	}

	v, err := opts.ServiceVersionFlag.Parse(serviceID, opts.Client)
	if err != nil {
		return serviceID, serviceVersion, err
	}

	if opts.AutoCloneFlag.WasSet {
		currentVersion := v
		v, err = opts.AutoCloneFlag.Parse(currentVersion, serviceID, opts.VerboseMode, opts.Out, opts.Client)
		if err != nil {
			return serviceID, currentVersion, err
		}
	} else if !opts.AllowActiveLocked && (v.Active || v.Locked) {
		err = errors.RemediationError{
			Inner:       fmt.Errorf("service version %d is not editable", v.Number),
			Remediation: errors.AutoCloneRemediation,
		}
		return serviceID, v, err
	}

	return serviceID, v, nil
}

// DisplayServiceID acquires the Service ID (if provided) and displays both it
// and its source location.
func DisplayServiceID(sid string, s manifest.Source, out io.Writer) {
	var via string
	switch s {
	case manifest.SourceFlag:
		via = " (via --service-id)"
	case manifest.SourceFile:
		via = fmt.Sprintf(" (via %s)", manifest.Filename)
	case manifest.SourceEnv:
		via = fmt.Sprintf(" (via %s)", env.ServiceID)
	case manifest.SourceUndefined:
		via = " (not provided)"
	}
	text.Output(out, "Service ID%s: %s", via, sid)
	text.Break(out)
}

// ArgsIsHelpJSON determines whether the supplied command arguments are exactly
// `help --format json`.
func ArgsIsHelpJSON(args []string) bool {
	return (len(args) == 3 &&
		args[0] == "help" &&
		args[1] == "--format" &&
		args[2] == "json")
}

// IsHelp indicates if the user called `fastly help` alone.
func IsHelp(args []string) bool {
	if args[0] == "help" && len(args) == 1 {
		return true
	}
	return false
}
