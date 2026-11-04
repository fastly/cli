package argparser

import (
	"context"
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

	"github.com/fastly/go-fastly/v19/fastly"

	"github.com/fastly/cli/internal/kongcompat"
	"github.com/fastly/cli/pkg/api"
	"github.com/fastly/cli/pkg/env"
	fsterr "github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
)

var (
	completionRegExp       = regexp.MustCompile("completion-bash$")
	completionScriptRegExp = regexp.MustCompile("completion-script-(?:bash|zsh)$")
)

// OptionalServiceVersion represents a Fastly service version.
type OptionalServiceVersion struct {
	OptionalString
}

// Parse returns a service version based on the given user input.
//
// Supported values:
//   - Numeric version (e.g., "1", "2", "42"): Returns the specified version
//   - "active": Returns the currently active version
//   - "staged": Returns the currently staged version
//   - "latest": Returns the highest version number (latest version)
//   - Omitted (no flag provided): Returns active version, falls back to latest if no active version exists
func (sv *OptionalServiceVersion) Parse(sid string, client api.Interface) (*fastly.Version, error) {
	// When no --version flag is provided (WasSet=false), default to "active" to preserve
	// the original behavior of trying active version first, with fallback to latest.
	if sv.Value == "" && !sv.WasSet {
		sv.Value = "active"
	}

	// When a specific numeric version is provided, use it directly.
	if n, err := strconv.Atoi(sv.Value); err == nil {
		return client.GetVersion(context.TODO(), &fastly.GetVersionInput{
			ServiceID:      sid,
			ServiceVersion: n,
		})
	}

	switch strings.ToLower(sv.Value) {
	case "active":
		serviceDetails, err := client.GetServiceDetails(context.TODO(), &fastly.GetServiceDetailsInput{
			ServiceID: sid,
			Filters: []fastly.ServiceDetailsFilter{
				{Key: "versions.active", Value: true},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error getting service details: %w", err)
		}
		// If active version exists, return it
		if serviceDetails.ActiveVersion != nil {
			return serviceDetails.ActiveVersion, nil
		}
		// If flag was explicitly set to "active" but no active version exists, return error
		if sv.WasSet {
			return nil, fmt.Errorf("no active service version found")
		}
		// If flag was not explicitly set and there's no active version, fall through to latest
		fallthrough
	case "latest":
		vs, err := client.ListVersions(context.TODO(), &fastly.ListVersionsInput{
			ServiceID: sid,
		})
		if err != nil {
			return nil, fmt.Errorf("error listing service versions: %w", err)
		}
		if len(vs) == 0 {
			return nil, errors.New("no service versions available")
		}
		// Sort versions into descending order to ensure we get the latest
		sort.Slice(vs, func(i, j int) bool {
			return fastly.ToValue(vs[i].Number) > fastly.ToValue(vs[j].Number)
		})
		return vs[0], nil
	case "staged":
		serviceDetails, err := client.GetServiceDetails(context.TODO(), &fastly.GetServiceDetailsInput{
			ServiceID: sid,
			Filters: []fastly.ServiceDetailsFilter{
				{Key: "versions.staged", Value: true},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error getting service details: %w", err)
		}
		if serviceDetails.Version == nil {
			return nil, fmt.Errorf("no staged service version found")
		}
		return serviceDetails.Version, nil
	default:
		return nil, fmt.Errorf("invalid version value %q: must be a version number, \"latest\", \"active\", or \"staged\"", sv.Value)
	}
}

// OptionalServiceNameID represents a mapping between a Fastly service name and
// its ID.
type OptionalServiceNameID struct {
	OptionalString
}

// Parse returns a service ID based off the given service name.
func (sv *OptionalServiceNameID) Parse(client api.Interface) (serviceID string, err error) {
	paginator := client.GetServices(context.TODO(), &fastly.GetServicesInput{})
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

// OptionalWorkspaceID represents a Fastly NGWAF Workspace ID.
type OptionalWorkspaceID struct {
	OptionalString
}

// Parse returns a workspace ID either from a flag or from a user defined
// environment variable (see pkg/env/env.go).
//
// NOTE: Will fallback to FASTLY_WORKSPACE_ID environment variable if no flag value set.
func (sv *OptionalWorkspaceID) Parse() error {
	if sv.Value == "" {
		if e := os.Getenv(env.WorkspaceID); e != "" {
			sv.Value = e
			return nil
		}
		return fsterr.ErrNoWorkspaceID
	}
	return nil
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
	stateUnknown := v.Active == nil && v.Locked == nil
	if ac.Value && (stateUnknown || v.Active != nil && *v.Active || v.Locked != nil && *v.Locked) {
		version, err := client.CloneVersion(context.TODO(), &fastly.CloneVersionInput{
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
func ContextHasHelpFlag(ctx *kongcompat.ParseContext) bool {
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
	Enabled bool `name:"json" short:"j" help:"Render output as JSON"` // Set via flag.
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

func ConvertBoolFromStringFlag(value string, argName string) (*bool, error) {
	switch value {
	case "true":
		return new(true), nil
	case "false":
		return new(false), nil
	default:
		return nil, fmt.Errorf("'%s' flag must be one of the following [true, false]", argName)
	}
}

func ConvertOrderFromStringFlag(value string, argName string) (string, error) {
	switch value {
	case "asc":
		return "", nil
	case "desc":
		return "-", nil
	default:
		return "", fmt.Errorf("'%s' flag must be one of the following [asc, desc]", argName)
	}
}
