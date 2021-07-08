package errors

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
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
	Persist(logPath string, args []string) error
}

// Log is ...
var Log = new(LogEntries)

// LogEntries represents a list of recorded log entries.
type LogEntries []LogEntry

// Add adds a new log entry.
func (l *LogEntries) Add(err error) {
	logMutex.Lock()
	*l = append(*l, LogEntry{
		Time: Now(),
		Err:  err,
	})
	logMutex.Unlock()
}

// Persist persists recorded log entries to disk.
func (l LogEntries) Persist(logPath string, args []string) error {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}

	// G307 (CWE-): Deferring unsafe method "*os.File" on type "Close".
	// gosec flagged this:
	// Disabling because this file isn't critical to the functioning of the CLI
	// and we only attempt to close it at the end of the user's execution flow.
	/* #nosec */
	defer f.Close()

	cmd := "\nCOMMAND: " + strings.Join(args, " ") + "\n"
	if _, err := f.Write([]byte(cmd)); err != nil {
		return err
	}

	record := `
{{.Time}}
{{.Err}}
--------------------
`
	t := template.Must(template.New("record").Parse(record))
	for _, entry := range l {
		err := t.Execute(f, entry)
		if err != nil {
			return err
		}
	}

	return nil
}

// LogEntry represents a single error log entry.
type LogEntry struct {
	Time time.Time
	Err  error
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
