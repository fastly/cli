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
		Time: time.Now(),
		Err:  err,
	})
	logMutex.Unlock()
}

// Persist persists recorded log entries to disk.
func (l LogEntries) Persist(logPath string, args []string) error {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error creating config file: %w", err)
	}
	defer f.Close()

	cmd := "COMMAND: " + strings.Join(args, " ") + "\n"
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

var logMutex sync.Mutex
