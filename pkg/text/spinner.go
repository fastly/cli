package text

import (
	"io"
	"time"

	"github.com/theckman/yacspin"
)

// NewSpinner returns a new instance of a terminal prompt spinner.
func NewSpinner(out io.Writer) (*yacspin.Spinner, error) {
	spinner, err := yacspin.New(yacspin.Config{
		CharSet:           yacspin.CharSets[9],
		Frequency:         100 * time.Millisecond,
		StopCharacter:     "✓",
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: "✗",
		StopFailColors:    []string{"fgRed"},
		Suffix:            " ",
		Writer:            out,
	})
	if err != nil {
		return nil, err
	}
	return spinner, nil
}
