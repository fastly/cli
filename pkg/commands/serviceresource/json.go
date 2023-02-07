package serviceresource

import (
	"encoding/json"
	"io"
	"time"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/go-fastly/v7/fastly"
)

// jsonOutput is a helper for adding a `--json` flag and encoding
// values to JSON. It can be embedded into command structs.
type jsonOutput struct {
	enabled bool // Set via flag.
}

// jsonFlag creates a flag for enabling JSON output.
func (j *jsonOutput) jsonFlag() cmd.BoolFlagOpts {
	return cmd.BoolFlagOpts{
		Name:        cmd.FlagJSONName,
		Description: cmd.FlagJSONDesc,
		Dst:         &j.enabled,
		Short:       'j',
	}
}

// WriteJSON checks whether the enabled flag is set or not. If set,
// then the given value is written as JSON to out. Otherwise, false is returned.
func (j *jsonOutput) WriteJSON(out io.Writer, value any) (bool, error) {
	if !j.enabled {
		return false, nil
	}

	// If value is a fastly.Resource (or variations of), then convert into
	// a jsonResource for improved JSON encoding.
	switch v := value.(type) {
	case *fastly.Resource:
		value = jsonResource(*v)
	case fastly.Resource:
		value = jsonResource(v)
	case []*fastly.Resource:
		cp := make([]jsonResource, len(v))
		for i := range v {
			cp[i] = jsonResource(*v[i])
		}
		value = cp
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return true, enc.Encode(value)
}

// jsonResource is a fastly.Resource with `json` field tags defined.
type jsonResource struct {
	CreatedAt      *time.Time `json:"created_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
	HREF           string     `json:"-"` // Omit this field from JSON output.
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	ResourceID     string     `json:"resource_id"`
	ResourceType   string     `json:"resource_type"`
	ServiceID      string     `json:"service_id"`
	ServiceVersion string     `json:"service_version"`
	UpdatedAt      *time.Time `json:"updated_at"`
}
