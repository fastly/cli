package resourcelink

import (
	"encoding/json"
	"io"

	"github.com/fastly/cli/pkg/cmd"
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

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return true, enc.Encode(value)
}
