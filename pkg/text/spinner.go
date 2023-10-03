package text

import (
	"fmt"
	"io"
	"time"

	"github.com/theckman/yacspin"
)

// SpinnerErrWrapper is a generic error message the caller can interpolate their
// own error into.
const SpinnerErrWrapper = "failed to stop spinner (error: %w) when handling the error: %w"

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
	WrapErr(err error) error
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
	err error
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
		spinErr := sp.StopFail()
		if spinErr != nil {
			return fmt.Errorf("failed to stop spinner (error: %w) when handling the error: %w", spinErr, err)
		}
		return err
	}

	sp.StopMessage(msg)
	return sp.Stop()
}

// Start is a proxy to the underlying spinner implementation. It sets the
// internal err to the error that is returned from a call to the underlying
// Start() so it can be used by a call subsequent to WrapErr().
func (sp *SpinnerWrapper) Start() error {
	err := sp.Spinner.Start()
	if err != nil {
		sp.err = err
	}
	return err
}

// StopFail is a proxy to the underlying spinner implementation. It sets the
// internal err to the error that is returned from a call to the underlying
// StopFail() so it can be used by a call subsequent to WrapErr().
func (sp *SpinnerWrapper) StopFail() error {
	err := sp.Spinner.StopFail()
	if err != nil {
		sp.err = err
	}
	return err
}

// Stop is a proxy to the underlying spinner implementation. It sets the
// internal err to the error that is returned from a call to the underlying
// Stop() so it can be used by a call subsequent to WrapErr().
func (sp *SpinnerWrapper) Stop() error {
	err := sp.Spinner.Stop()
	if err != nil {
		sp.err = err
	}
	return err
}

// WrapErr returns a spinner error that wraps err.
func (sp *SpinnerWrapper) WrapErr(err error) error {
	return fmt.Errorf(SpinnerErrWrapper, sp.err, err)
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

	return &SpinnerWrapper{
		Spinner: spinner,
		err:     nil,
	}, nil
}
