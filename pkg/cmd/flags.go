package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
	"github.com/fastly/kingpin"
)

var (
	completionRegExp       = regexp.MustCompile("completion-bash$")
	completionScriptRegExp = regexp.MustCompile("completion-script-(?:bash|zsh)$")
)

// StringFlagOpts enables easy configuration of a flag.
type StringFlagOpts struct {
	Action      kingpin.Action
	Description string
	Dst         *string
	Name        string
	Required    bool
	Short       rune
}

// RegisterFlag defines a flag.
func (b Base) RegisterFlag(opts StringFlagOpts) {
	clause := b.CmdClause.Flag(opts.Name, opts.Description)
	if opts.Short > 0 {
		clause = clause.Short(opts.Short)
	}
	if opts.Required {
		clause = clause.Required()
	}
	if opts.Action != nil {
		clause = clause.Action(opts.Action)
	}
	clause.StringVar(opts.Dst)
}

// BoolFlagOpts enables easy configuration of a flag.
type BoolFlagOpts struct {
	Action      kingpin.Action
	Description string
	Dst         *bool
	Name        string
	Required    bool
	Short       rune
}

// RegisterFlagBool defines a boolean flag.
//
// TODO: Use generics support in go 1.18 to remove the need for two functions.
func (b Base) RegisterFlagBool(opts BoolFlagOpts) {
	clause := b.CmdClause.Flag(opts.Name, opts.Description)
	if opts.Short > 0 {
		clause = clause.Short(opts.Short)
	}
	if opts.Required {
		clause = clause.Required()
	}
	if opts.Action != nil {
		clause = clause.Action(opts.Action)
	}
	clause.BoolVar(opts.Dst)
}

// OptionalServiceVersion represents a Fastly service version.
type OptionalServiceVersion struct {
	OptionalString
}

// Parse returns a service version based on the given user input.
func (sv *OptionalServiceVersion) Parse(sid string, client api.Interface) (*fastly.Version, error) {
	vs, err := client.ListVersions(&fastly.ListVersionsInput{
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
		v, err = GetActiveVersion(vs)
	case "":
		return vs[0], nil
	default:
		v, err = GetSpecifiedVersion(vs, sv.Value)
	}
	if err != nil {
		return nil, err
	}

	return v, nil
}

// OptionalServiceNameID represents a mapping between a Fastly service name and
// its ID.
type OptionalServiceNameID struct {
	OptionalString
}

// Parse returns a service ID based off the given service name.
func (sv *OptionalServiceNameID) Parse(client api.Interface) (serviceID string, err error) {
	services, err := client.ListServices(&fastly.ListServicesInput{})
	if err != nil {
		return serviceID, fmt.Errorf("error listing services: %w", err)
	}
	for _, s := range services {
		if s.Name == sv.Value {
			return s.ID, nil
		}
	}
	return serviceID, errors.New("error matching service name with available services")
}

// OptionalCustomerID represents a Fastly customer ID.
type OptionalCustomerID struct {
	OptionalString
}

// Parse returns a customer ID either from a flag or from a user defined
// environment variable (see pkg/env/env.go).
//
// NOTE: Will fallback to FASTLY_CUSTOMER_ID environment variable if no flag value set.
func (sv *OptionalCustomerID) Parse() error {
	if sv.Value == "" {
		if e := os.Getenv(env.CustomerID); e != "" {
			sv.Value = e
			return nil
		}
		return fsterr.ErrNoCustomerID
	}
	return nil
}

// AutoCloneFlagOpts enables easy configuration of the --autoclone flag defined
// via the RegisterAutoCloneFlag constructor.
type AutoCloneFlagOpts struct {
	Action kingpin.Action
	Dst    *bool
}

// RegisterAutoCloneFlag defines a --autoclone flag that will cause a clone of the
// identified service version if it's found to be active or locked.
func (b Base) RegisterAutoCloneFlag(opts AutoCloneFlagOpts) {
	b.CmdClause.Flag("autoclone", "If the selected service version is not editable, clone it and use the clone.").Action(opts.Action).BoolVar(opts.Dst)
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
func (ac *OptionalAutoClone) Parse(v *fastly.Version, sid string, verbose bool, out io.Writer, client api.Interface) (*fastly.Version, error) {
	// if user didn't provide --autoclone flag
	if !ac.Value && (v.Active || v.Locked) {
		return nil, fsterr.RemediationError{
			Inner:       fmt.Errorf("service version %d is not editable", v.Number),
			Remediation: fsterr.AutoCloneRemediation,
		}
	}
	if ac.Value && (v.Active || v.Locked) {
		version, err := client.CloneVersion(&fastly.CloneVersionInput{
			ServiceID:      sid,
			ServiceVersion: v.Number,
		})
		if err != nil {
			return nil, fmt.Errorf("error cloning service version: %w", err)
		}
		if verbose {
			msg := fmt.Sprintf("Service version %d is not editable, so it was automatically cloned because --autoclone is enabled. Now operating on version %d.", v.Number, version.Number)
			text.Output(out, msg)
		}
		return version, nil
	}

	// Treat the function as a no-op if the version is editable.
	return v, nil
}

// GetActiveVersion returns the active service version.
func GetActiveVersion(vs []*fastly.Version) (*fastly.Version, error) {
	for _, v := range vs {
		if v.Active {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no active service version found")
}

// GetSpecifiedVersion returns the specified service version.
func GetSpecifiedVersion(vs []*fastly.Version, version string) (*fastly.Version, error) {
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

// Content determines if the given flag value is a file path, and if so read
// the contents from disk, otherwise presume the given value is the content.
func Content(flagval string) string {
	content := flagval
	if path, err := filepath.Abs(flagval); err == nil {
		if _, err := os.Stat(path); err == nil {
			if data, err := os.ReadFile(path); err == nil {
				content = string(data)
			}
		}
	}
	return content
}

// IntToBool converts a binary 0|1 to a boolean.
func IntToBool(i int) bool {
	return i > 0
}

// ContextHasHelpFlag asserts whether a given kingpin.ParseContext contains a
// `help` flag.
func ContextHasHelpFlag(ctx *kingpin.ParseContext) bool {
	_, ok := ctx.Elements.FlagMap()["help"]
	return ok
}

// IsCompletionScript determines whether the supplied command arguments are for
// shell completion output that is then eval()'ed by the user's shell.
func IsCompletionScript(args []string) bool {
	var found bool
	for _, arg := range args {
		if completionScriptRegExp.MatchString(arg) {
			found = true
		}
	}
	return found
}

// IsCompletion determines whether the supplied command arguments are for
// shell completion (i.e. --completion-bash) that should produce output that
// the user's shell can utilise for handling autocomplete behaviour.
func IsCompletion(args []string) bool {
	var found bool
	for _, arg := range args {
		if completionRegExp.MatchString(arg) {
			found = true
		}
	}
	return found
}
