package cmd

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/global"
	"github.com/fastly/cli/pkg/manifest"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v7/fastly"
	"github.com/fastly/kingpin"
)

// Command is an interface that abstracts over all of the concrete command
// structs. The Name method lets us select which command should be run, and the
// Exec method invokes whatever business logic the command should do.
type Command interface {
	Name() string
	Exec(in io.Reader, out io.Writer) error
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
	Globals   *global.Data
}

// Name implements the Command interface, and returns the FullCommand from the
// kingpin.Command that's used to select which command to actually run.
func (b Base) Name() string {
	return b.CmdClause.FullCommand()
}

// Optional models an optional type that consumers can use to assert whether the
// inner value has been set and is therefore valid for use.
type Optional struct {
	WasSet bool
}

// Set implements kingpin.Action and is used as callback to set that the optional
// inner value is valid.
func (o *Optional) Set(_ *kingpin.ParseElement, _ *kingpin.ParseContext) error {
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
	APIClient          api.Interface
	Manifest           manifest.Data
	Out                io.Writer
	ServiceNameFlag    OptionalServiceNameID
	ServiceVersionFlag OptionalServiceVersion
	VerboseMode        bool
	ErrLog             fsterr.LogInterface
}

// ServiceDetails returns the Service ID and Service Version.
func ServiceDetails(opts ServiceDetailsOpts) (serviceID string, serviceVersion *fastly.Version, err error) {
	serviceID, source, flag, err := ServiceID(opts.ServiceNameFlag, opts.Manifest, opts.APIClient, opts.ErrLog)
	if err != nil {
		return serviceID, serviceVersion, err
	}
	if opts.VerboseMode {
		DisplayServiceID(serviceID, flag, source, opts.Out)
	}

	v, err := opts.ServiceVersionFlag.Parse(serviceID, opts.APIClient)
	if err != nil {
		return serviceID, serviceVersion, err
	}

	if opts.AutoCloneFlag.WasSet {
		currentVersion := v
		v, err = opts.AutoCloneFlag.Parse(currentVersion, serviceID, opts.VerboseMode, opts.Out, opts.APIClient)
		if err != nil {
			return serviceID, currentVersion, err
		}
	} else if !opts.AllowActiveLocked && (v.Active || v.Locked) {
		err = fsterr.RemediationError{
			Inner:       fmt.Errorf("service version %d is not editable", v.Number),
			Remediation: fsterr.AutoCloneRemediation,
		}
		return serviceID, v, err
	}

	return serviceID, v, nil
}

// ServiceID returns the Service ID and the source of that information.
//
// NOTE: If Service ID not provided then check if Service Name provided and use
// that information to acquire the Service ID.
func ServiceID(serviceName OptionalServiceNameID, data manifest.Data, client api.Interface, li fsterr.LogInterface) (serviceID string, source manifest.Source, flag string, err error) {
	flag = "--service-id"
	serviceID, source = data.ServiceID()

	if source == manifest.SourceUndefined {
		if !serviceName.WasSet {
			err = fsterr.ErrNoServiceID
			if li != nil {
				li.Add(err)
			}
			return serviceID, source, flag, err
		}

		serviceID, err = serviceName.Parse(client)
		if err != nil && li != nil {
			li.Add(err)
		}
		flag = "--service-name"
		source = manifest.SourceFlag
	}

	return serviceID, source, flag, err
}

// DisplayServiceID acquires the Service ID (if provided) and displays both it
// and its source location.
func DisplayServiceID(sid, flag string, s manifest.Source, out io.Writer) {
	var via string
	switch s {
	case manifest.SourceFlag:
		via = fmt.Sprintf(" (via %s)", flag)
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
// `help --format=json` or `help --format json`.
func ArgsIsHelpJSON(args []string) bool {
	switch len(args) {
	case 2:
		if args[0] == "help" && args[1] == "--format=json" {
			return true
		}
	case 3:
		if args[0] == "help" && args[1] == "--format" && args[2] == "json" {
			return true
		}
	}
	return false
}

// IsHelpOnly indicates if the user called `fastly help [...]`.
func IsHelpOnly(args []string) bool {
	return len(args) > 0 && args[0] == "help"
}

// IsHelpFlagOnly indicates if the user called `fastly --help [...]`.
func IsHelpFlagOnly(args []string) bool {
	return len(args) > 0 && args[0] == "--help"
}

// IsVerboseAndQuiet indicates if the user called `fastly --verbose --quiet`.
// These flags are mutually exclusive.
func IsVerboseAndQuiet(args []string) bool {
	matches := map[string]bool{}
	for _, a := range args {
		if a == "--verbose" || a == "-v" {
			matches["--verbose"] = true
		}
		if a == "--quiet" || a == "-q" {
			matches["--quiet"] = true
		}
	}
	return len(matches) > 1
}

// IsGlobalFlagsOnly indicates if the user called the binary with any
// permutation order of the globally defined flags.
//
// NOTE: Some global flags accept a value while others do not. The following
// algorithm takes this into account by mapping the flag to an expected value.
// For example, --verbose doesn't accept a value so is set to zero.
//
// EXAMPLES:
//
// The following would return false as a command was specified:
//
// args: [--verbose -v --endpoint ... --token ... -t ... --endpoint ...  version] 11
// total: 10
//
// The following would return true as only global flags were specified:
//
// args: [--verbose -v --endpoint ... --token ... -t ... --endpoint ...] 10
// total: 10
func IsGlobalFlagsOnly(args []string) bool {
	// Global flags are defined in ../app/run.go
	globals := map[string]int{
		"--accept-defaults": 0,
		"-d":                0,
		"--auto-yes":        0,
		"-y":                0,
		"--endpoint":        1,
		"--help":            0,
		"--non-interactive": 0,
		"-i":                0,
		"--profile":         1,
		"-o":                1,
		"--quiet":           0,
		"-q":                0,
		"--token":           1,
		"-t":                1,
		"--verbose":         0,
		"-v":                0,
	}
	var total int
	for _, a := range args {
		for k := range globals {
			if a == k {
				total++
				total += globals[k]
			}
		}
	}
	return len(args) == total
}
