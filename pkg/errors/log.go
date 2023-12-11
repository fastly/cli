package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fastly/go-fastly/v8/fastly"
)

// LogPath is the location of the fastly CLI error log.
var LogPath = func() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "fastly", "errors.log")
	}
	if dir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(dir, ".fastly", "errors.log")
	}
	panic("unable to deduce user config dir or user home dir")
}()

// LogInterface represents the LogEntries behaviours.
type LogInterface interface {
	Add(err error)
	AddWithContext(err error, ctx map[string]any)
	Persist(logPath string, args []string) error
}

// MockLog is a no-op Log type.
type MockLog struct{}

// Add adds an error to the mock log.
func (ml MockLog) Add(_ error) {}

// AddWithContext adds an error and context to the mock log.
func (ml MockLog) AddWithContext(_ error, _ map[string]any) {}

// Persist writes the error data to logPath.
func (ml MockLog) Persist(_ string, _ []string) error {
	return nil
}

// Log is the primary interface for consumers.
var Log = new(LogEntries)

// LogEntries represents a list of recorded log entries.
type LogEntries []LogEntry

// Add adds a new log entry.
func (l *LogEntries) Add(err error) {
	logMutex.Lock()
	*l = append(*l, createLogEntry(err))
	logMutex.Unlock()
}

// AddWithContext adds a new log entry with extra contextual data.
func (l *LogEntries) AddWithContext(err error, ctx map[string]any) {
	le := createLogEntry(err)
	le.Context = ctx

	logMutex.Lock()
	*l = append(*l, le)
	logMutex.Unlock()
}

// Persist persists recorded log entries to disk.
func (l LogEntries) Persist(logPath string, args []string) error {
	if len(l) == 0 {
		return nil
	}
	cmd := "fastly " + strings.Join(args, " ")
	errMsg := "error accessing audit log file: %w"

	// gosec flagged this:
	// G304 (CWE-22): Potential file inclusion via variable
	//
	// Disabling as the input is determined from our own package.
	/* #nosec */
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if fi, err := f.Stat(); err == nil {
		if fi.Size() >= FileRotationSize {
			err = f.Close()
			if err != nil {
				return err
			}

			// gosec flagged this:
			// G304 (CWE-22): Potential file inclusion via variable
			//
			// Disabling as the input is determined from our own package.
			/* #nosec */
			f, err = os.Create(logPath)
			if err != nil {
				return fmt.Errorf(errMsg, err)
			}
		}
	}

	// G307 (CWE-): Deferring unsafe method "*os.File" on type "Close".
	// gosec flagged this:
	// Disabling because this file isn't critical to the functioning of the CLI
	// and we only attempt to close it at the end of the user's execution flow.
	/* #nosec */
	defer f.Close()

	cmd = "\nCOMMAND:\n" + cmd + "\n\n"
	if _, err := f.Write([]byte(cmd)); err != nil {
		return err
	}

	record := `TIMESTAMP:
{{.Time}}

ERROR:
{{.Err}}
{{ range $key, $value := .Caller }}
{{ $key }}:
{{ $value }}
{{ end }}
{{ range $key, $value := .Context }}
  {{ $key }}: {{ $value }}
{{ end }}
`
	t := template.Must(template.New("record").Parse(record))
	for _, entry := range l {
		err := t.Execute(f, entry)
		if err != nil {
			return err
		}
	}

	if _, err := f.Write([]byte("------------------------------\n\n")); err != nil {
		return err
	}

	return nil
}

var (
	// TokenRegEx matches a Token as part of the error output (https://regex101.com/r/ulIw1m/1)
	TokenRegEx = regexp.MustCompile(`Token ([\w-]+)`)
	// TokenFlagRegEx matches the token flag (https://regex101.com/r/YNr78Q/1)
	TokenFlagRegEx = regexp.MustCompile(`(-t|--token)(\s*=?\s*['"]?)([\w-]+)(['"]?)`)
)

// FilterToken replaces any matched patterns with "REDACTED".
//
// EXAMPLE: https://go.dev/play/p/cT4BwIh9Asa
func FilterToken(input string) (inputFiltered string) {
	inputFiltered = TokenRegEx.ReplaceAllString(input, "Token REDACTED")
	inputFiltered = TokenFlagRegEx.ReplaceAllString(inputFiltered, "${1}${2}REDACTED${4}")
	return inputFiltered
}

// createLogEntry generates the boilerplate of a LogEntry.
func createLogEntry(err error) LogEntry {
	le := LogEntry{
		Time: Now(),
		Err:  err,
	}

	_, file, line, ok := runtime.Caller(2)
	if ok {
		idx := strings.Index(file, "/pkg/")
		if idx == -1 {
			idx = 0
		}
		le.Caller = map[string]any{
			"FILE": file[idx:],
			"LINE": line,
		}
	}

	return le
}

// LogEntry represents a single error log entry.
type LogEntry struct {
	Time    time.Time
	Err     error
	Caller  map[string]any
	Context map[string]any
}

// Caller represents where an error occurred.
type Caller struct {
	File string
	Line int
}

// Appending to a slice isn't threadsafe, and although we currently don't
// expect this to be a problem we can't predict future logic requirements that
// might result in more asynchronous operations, so we play it safe and utilise
// a lock before updating the LogEntries.
var logMutex sync.Mutex

// Now is exposed so that we may mock it from our test file.
//
// NOTE: The ideal way to deal with time is to inject it as a dependency and
// then the caller can provide a stubbed value, but in this case we don't want
// to have the CLI's business logic littered with lots of calls to time.Now()
// when that call can be handled internally by the .Add() method.
var Now = time.Now

// FileRotationSize represents the size the log file needs to be before we
// truncate it.
//
// NOTE: To enable easier testing of the log rotation logic, we don't define
// this as a constant but as a variable so the test file can mutate the value
// to something much smaller, meaning we can commit a small test file as part
// of the testing logic that will trigger a 'over the threshold' scenario.
var FileRotationSize int64 = 5242880 // 5mb

// ServiceVersion returns an integer regardless of whether the given argument
// is a nil pointer or not. It helps to reduce the boilerplate found across the
// codebase when tracking errors related to `argparser.ServiceDetails`.
func ServiceVersion(v *fastly.Version) int {
	var sv int
	if v != nil {
		sv = fastly.ToValue(v.Number)
	}
	return sv
}
