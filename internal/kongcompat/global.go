package kongcompat

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
)

// ErrorReporter prints errors and exits the way kingpin's Application did.
//
// kong has methods with the same names, but they behave a bit differently.
// They indent multi-line messages and return the parser, and FatalIfErrorf
// takes the exit status from the error and may print the usage.
type ErrorReporter struct {
	k *kong.Kong
}

// Errors returns the error reporter for k.
func Errors(k *kong.Kong) ErrorReporter {
	return ErrorReporter{k}
}

// Name and writer used when there is no kong.Kong.
// Like kingpin's default application, they are set once at startup.
var (
	defaultName             = filepath.Base(os.Args[0])
	defaultStderr io.Writer = os.Stderr
)

// Errorf prints "<name>: error: <message>" on the error writer.
func (r ErrorReporter) Errorf(format string, args ...any) {
	name, w := defaultName, defaultStderr
	if r.k != nil {
		name, w = r.k.Model.Name, r.k.Stderr
	}
	fmt.Fprintf(w, name+": error: "+format+"\n", args...)
}

// Fatalf prints an error as Errorf does and exits with status 1.
func (r ErrorReporter) Fatalf(format string, args ...any) {
	r.Errorf(format, args...)
	r.exit(1)
}

// FatalIfError does nothing if err is nil.
// Otherwise it prints err, prefixed with the formatted message when format
// is not empty, and exits with status 1.
func (r ErrorReporter) FatalIfError(err error, format string, args ...any) {
	if err == nil {
		return
	}
	prefix := ""
	if format != "" {
		prefix = fmt.Sprintf(format, args...) + ": "
	}
	r.Errorf(prefix+"%s", err)
	r.exit(1)
}

func (r ErrorReporter) exit(code int) {
	if r.k != nil {
		r.k.Exit(code)
		return
	}
	os.Exit(code)
}

// Errorf replaces kingpin.Errorf.
// Like kingpin's default application, it prints under the program's name.
func Errorf(format string, args ...any) {
	ErrorReporter{}.Errorf(format, args...)
}

// Fatalf replaces kingpin.Fatalf.
func Fatalf(format string, args ...any) {
	ErrorReporter{}.Fatalf(format, args...)
}

// FatalIfError replaces kingpin.FatalIfError.
func FatalIfError(err error, format string, args ...any) {
	ErrorReporter{}.FatalIfError(err, format, args...)
}

// Must is kingpin's package-level MustParse, for the result of
// Application.Parse: on an error it prints it, as kingpin did, and exits
// with status 1.
func Must(command string, err error) string {
	if err != nil {
		Fatalf("%s, try --help", err)
	}
	return command
}
