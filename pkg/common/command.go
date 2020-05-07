package common

import (
	"io"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/config"
	"gopkg.in/alecthomas/kingpin.v3-unstable"
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

// Optional models an optional type that consumers can use to assert whether the
// inner value has been set and is therefore valid for use.
type Optional struct {
	Valid bool
}

// Set implements kingpin.Action and is used as callback to set that the optional
// inner value is valid.
func (o *Optional) Set(e *kingpin.ParseElement, c *kingpin.ParseContext) error {
	o.Valid = true
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

// OptionalInt models an optional int flag value.
type OptionalInt struct {
	Optional
	Value int
}
