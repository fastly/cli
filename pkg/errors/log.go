package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fastly/go-fastly/v5/fastly"
	"github.com/getsentry/sentry-go"
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
	AddWithContext(err error, ctx map[string]interface{})
	Persist(logPath string, args []string) error
}

// MockLog is a no-op Log type.
type MockLog struct{}

func (ml MockLog) Add(err error)                                        {}
func (ml MockLog) AddWithContext(err error, ctx map[string]interface{}) {}
func (ml MockLog) Persist(logPath string, args []string) error          { return nil }

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
func (l *LogEntries) AddWithContext(err error, ctx map[string]interface{}) {
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
	instrument(l)

	errMsg := "error accessing audit log file: %w"

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}

	if fi, err := f.Stat(); err == nil {
		if fi.Size() >= FileRotationSize {
			f.Close()

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

	cmd := "\nCOMMAND:\nfastly " + strings.Join(args, " ") + "\n\n"
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

	f.Write([]byte("------------------------------\n\n"))

	return nil
}

// instrument reports errors to our error analysis platform.
func instrument(l LogEntries) {
	for _, entry := range l {
		var (
			file string
			line int
		)
		if v, ok := entry.Caller["FILE"]; ok {
			file, _ = v.(string)
		}
		if v, ok := entry.Caller["LINE"]; ok {
			line, _ = v.(int)
		}
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Message:   fmt.Sprintf("%s (file: %s, line: %d)", entry.Err, file, line),
			Timestamp: entry.Time,
			Data:      entry.Context,
		})
	}
	sentry.CaptureException(l[len(l)-1].Err)
}

// createLogEntry generates the boilerplate of a LogEntry.
func createLogEntry(err error) LogEntry {
	le := LogEntry{
		Time: Now(),
		Err:  err,
	}

	_, file, line, ok := runtime.Caller(2)
	if ok {
		le.Caller = map[string]interface{}{
			"FILE": file[strings.Index(file, "/pkg/"):],
			"LINE": line,
		}
	}

	return le
}

// LogEntry represents a single error log entry.
type LogEntry struct {
	Time    time.Time
	Err     error
	Caller  map[string]interface{}
	Context map[string]interface{}
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
// codebase when tracking errors related to `cmd.ServiceDetails`.
func ServiceVersion(v *fastly.Version) (sv int) {
	if v != nil {
		sv = v.Number
	}
	return
}
