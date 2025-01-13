package argparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/fastly/kingpin"

	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
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
// TODO: Use generics support in go 1.18 to remove the need for multiple functions.
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

// IntFlagOpts enables easy configuration of a flag.
type IntFlagOpts struct {
	Action      kingpin.Action
	Default     int
	Description string
	Dst         *int
	Name        string
	Required    bool
	Short       rune
}

// RegisterFlagInt defines an integer flag.
func (b Base) RegisterFlagInt(opts IntFlagOpts) {
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
	if opts.Default != 0 {
		clause = clause.Default(strconv.Itoa(opts.Default))
	}
	clause.IntVar(opts.Dst)
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
	if err != nil {
		return nil, fmt.Errorf("error listing service versions: %w", err)
	}
	if len(vs) == 0 {
		return nil, errors.New("error listing service versions: no versions available")
	}

	// Sort versions into descending order.
	sort.Slice(vs, func(i, j int) bool {
		return fastly.ToValue(vs[i].Number) > fastly.ToValue(vs[j].Number)
	})

	var v *fastly.Version

	switch strings.ToLower(sv.Value) {
	case "latest":
		return vs[0], nil
	case "active":
		v, err = GetActiveVersion(vs)
	case "": // no --version flag provided
		v, err = GetActiveVersion(vs)
		if err != nil {
			return vs[0], nil //lint:ignore nilerr if no active version, return latest version
		}
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
	paginator := client.GetServices(&fastly.GetServicesInput{})
	var services []*fastly.Service
	for paginator.HasNext() {
		data, err := paginator.GetNext()
		if err != nil {
			return serviceID, fmt.Errorf("error listing services: %w", err)
		}
		services = append(services, data...)
	}
	for _, s := range services {
		if fastly.ToValue(s.Name) == sv.Value {
			return fastly.ToValue(s.ServiceID), nil
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
	if !ac.Value && (fastly.ToValue(v.Active) || fastly.ToValue(v.Locked)) {
		return nil, fsterr.RemediationError{
			Inner:       fmt.Errorf("service version %d is not editable", fastly.ToValue(v.Number)),
			Remediation: fsterr.AutoCloneRemediation,
		}
	}
	if ac.Value && (v.Active != nil && *v.Active || v.Locked != nil && *v.Locked) {
		version, err := client.CloneVersion(&fastly.CloneVersionInput{
			ServiceID:      sid,
			ServiceVersion: fastly.ToValue(v.Number),
		})
		if err != nil {
			return nil, fmt.Errorf("error cloning service version: %w", err)
		}
		if verbose {
			msg := "Service version %d is not editable, so it was automatically cloned because --autoclone is enabled. Now operating on version %d.\n\n"
			format := fmt.Sprintf(msg, fastly.ToValue(v.Number), fastly.ToValue(version.Number))
			text.Info(out, format)
		}
		return version, nil
	}

	// Treat the function as a no-op if the version is editable.
	return v, nil
}

// GetActiveVersion returns the active service version.
func GetActiveVersion(vs []*fastly.Version) (*fastly.Version, error) {
	for _, v := range vs {
		if fastly.ToValue(v.Active) {
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
		if fastly.ToValue(v.Number) == i {
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
			if data, err := os.ReadFile(path); err == nil /* #nosec */ {
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

// JSONOutput is a helper for adding a `--json` flag and encoding
// values to JSON. It can be embedded into command structs.
type JSONOutput struct {
	Enabled bool // Set via flag.
}

// JSONFlag creates a flag for enabling JSON output.
func (j *JSONOutput) JSONFlag() BoolFlagOpts {
	return BoolFlagOpts{
		Name:        FlagJSONName,
		Description: FlagJSONDesc,
		Dst:         &j.Enabled,
		Short:       'j',
	}
}

// WriteJSON checks whether the enabled flag is set or not. If set,
// then the given value is written as JSON to out. Otherwise, false is returned.
func (j *JSONOutput) WriteJSON(out io.Writer, value any) (bool, error) {
	if !j.Enabled {
		return false, nil
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return true, enc.Encode(value)
}
