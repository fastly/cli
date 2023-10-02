package text

import (
	"fmt"
	"io"
	"time"

	"github.com/theckman/yacspin"
)

// Spinner represents a terminal prompt status indicator.
type Spinner interface {
	Status() yacspin.SpinnerStatus
	Start() error
	Message(message string)
	StopFailMessage(message string)
	StopFail() error
	StopMessage(message string)
	Stop() error
	Process(msg string, fn SpinnerProcess) error
}

// SpinnerProcess is the logic to execute in between the spinner start/stop.
//
// NOTE: The `sp` SpinnerWrapper is passed in to handle more complex scenarios.
// For example, the logic inside the SpinnerProcess might want to control the
// Start/Stop mechanisms outside of the basic flow provided by `Process()`.
type SpinnerProcess func(sp *SpinnerWrapper) error

// SpinnerWrapper implements the Spinner interface.
type SpinnerWrapper struct {
	*yacspin.Spinner
}

// Process starts/stops the spinner with `msg` and executes `fn` in between.
func (sp *SpinnerWrapper) Process(msg string, fn SpinnerProcess) error {
	err := sp.Start()
	if err != nil {
		return err
	}
	sp.Message(msg + "...")

	err = fn(sp)
	if err != nil {
		sp.StopFailMessage(msg)
		spinStopErr := sp.StopFail()
		if spinStopErr != nil {
			return fmt.Errorf("failed to stop spinner (error: %w) when handling the error: %w", spinStopErr, err)
		}
		return err
	}

	sp.StopMessage(msg)
	return sp.Stop()
}

// NewSpinner returns a new instance of a terminal prompt spinner.
func NewSpinner(out io.Writer) (Spinner, error) {
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

	return &SpinnerWrapper{spinner}, nil
}
