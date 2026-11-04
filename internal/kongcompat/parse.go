package kongcompat

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
)

// Parse parses args the way kingpin's Application.Parse did and returns the
// full name of the selected command, such as "db migrate".
//
// It also handles what kingpin applications got for free.
// --help and the help command print usage with uc, or with kingpin's default
// template when uc is nil.
// --completion-bash, --completion-script-bash and --completion-script-zsh
// answer the shell.
// Errors use kingpin's wording.
// Usage goes to the parser's Stdout, and the parser's Exit is called
// wherever kingpin would have exited.
func Parse(k *kong.Kong, args []string, uc *UsageContext) (string, error) {
	if version == 2 {
		expanded, err := ExpandArgs(args)
		if err != nil {
			return "", err
		}
		args = expanded
	}
	// After "--" everything is an argument, completion flags included.
	flags := args
	for i, a := range args {
		if a == "--" {
			flags = args[:i]
			break
		}
	}
	for _, a := range flags {
		switch a {
		case "--help-long", "--help-man":
			if version != 2 {
				continue
			}
			tmpl := pick(v2LegacyLongHelpTemplate, v2LongHelpTemplate, v2LongHelpTemplate)
			if a == "--help-man" {
				tmpl = pick(v2LegacyManPageTemplate, v2ManPageTemplate, v2ManPageTemplate)
			}
			// A syntax error still comes first, values are not checked.
			s, err := scan(k, without(args, a), true)
			if err != nil {
				return "", err
			}
			if err := usageAt(alwaysStdout(k), k, s.path, &UsageContext{Template: tmpl}); err != nil {
				return "", err
			}
			k.Exit(0)
			return "", nil
		case "--completion-script-bash", "--completion-script-zsh":
			shell := strings.TrimPrefix(a, "--completion-script-")
			if err := CompletionScript(alwaysStdout(k), k.Model.Name, shell); err != nil {
				return "", err
			}
			k.Exit(0)
			return "", nil
		}
	}
	for i, a := range flags {
		if a == "--completion-bash" {
			rest := append(append([]string{}, args[:i]...), args[i+1:]...)
			fmt.Fprint(alwaysStdout(k), strings.Join(Complete(k, rest), "\n"))
			k.Exit(0)
			return "", nil
		}
	}

	// kingpin printed the version before checking any value or handling
	// --help, as long as the command line up to the flag made sense.
	// kong prints it from a hook that runs before any check, so the flag is
	// parsed alone.
	if i := versionFlagIndex(k, flags); i >= 0 {
		if _, err := scan(k, args[:i+1], true); err == nil {
			_, err := k.Parse(args[i : i+1])
			return "", err
		}
	}

	// Values must be decoded only once, by kong's Parse, since a value's Set
	// may do more than store it.
	// Until then, just follow the command line the way kingpin read it.
	scanned, err := scan(k, args, true)
	if err != nil {
		// kingpin checked the defaults even before the syntax.
		if _, ferr := scan(k, args, false); ferr != nil {
			return "", ferr
		}
		return "", err
	}
	path := scanned.path
	if scanned.sawHelp {
		if err := usageAt(k.Stdout, k, path, uc); err != nil {
			return "", err
		}
		// kingpin v3 exited with status 1 after --help, v2 with 0.
		k.Exit(pickStatus(0, 1))
	}
	if len(path) > 1 && isHelpCommand(path[1]) {
		sel := path[1]
		// kingpin still decoded the values before running help, so a bad
		// one is an error.
		s := newChecker(args, false)
		s.noRequired = true
		s.enter(k.Model.Node)
		if err := s.run(); err != nil {
			return "", err
		}
		// help [<command>...] shows the usage of the named command.
		target := scanned.helpArgs
		// kingpin's help command ignored default commands.
		// With no command named, it showed the application's usage, and it
		// failed on an unknown command.
		// When the program turned exiting off, as the Fastly CLI does,
		// kingpin went on to show the usage of what it recognized.
		var targetPath []*kong.Node
		if len(target) > 0 {
			s := newChecker(target, true)
			s.sawHelp = true
			s.enter(k.Model.Node)
			if err := s.run(); err != nil {
				fmt.Fprintf(k.Stderr, "%s: error: %s\n", k.Model.Name, err)
				k.Exit(1)
			}
			targetPath = s.path
		}
		if err := usageAt(k.Stdout, k, targetPath, uc); err != nil {
			return "", err
		}
		k.Exit(0)
		return sel.Name, nil
	}
	if err := Check(k, args); err != nil {
		if errors.Is(err, ErrCommandNotSpecified) {
			// kingpin printed only the usage when no command was given.
			if uerr := usageAt(k.Stdout, k, path, uc); uerr != nil {
				return "", uerr
			}
			k.Exit(pickStatus(0, 1))
		}
		return "", err
	}
	// keepUnselected goes first and clears the variables of unselected
	// values, so readLikeKingpin never looks them up.
	keep := keepUnselected(k, path)
	restore := readLikeKingpin(k, scanned.given)
	kctx, err := k.Parse(dashAsEmpty(args))
	restore()
	keep()
	if err != nil {
		return "", valueError(k, err)
	}
	return FullCommand(kctx), nil
}

// keepUnselected stops kong from changing the values of commands that args
// do not select.
// kong resets every value in the grammar from its default or environment
// variable, while kingpin left those values alone.
// Until restore is called, they have no default and no variable, and
// restore puts back what they held.
// A value that shares its Go variable with a selected command is left to
// kong.
func keepUnselected(k *kong.Kong, path []*kong.Node) (restore func()) {
	selected := map[*kong.Node]bool{}
	for _, n := range path {
		selected[n] = true
	}
	parsed := map[uintptr]bool{}
	var others []*kong.Value
	nodes(k.Model.Node, func(n *kong.Node) {
		for _, v := range ownValues(n) {
			switch {
			case !v.Target.CanAddr():
			case selected[n]:
				parsed[v.Target.Addr().Pointer()] = true
			default:
				others = append(others, v)
			}
		}
	})
	type saved struct {
		v          *kong.Value
		target     reflect.Value
		hasDefault bool
		def        string
		envs       []string
	}
	var keep []saved
	for _, v := range others {
		if parsed[v.Target.Addr().Pointer()] {
			continue
		}
		// kong works on a spare value instead, so the real one stays as it
		// is all along, hooks included.
		keep = append(keep, saved{v, v.Target, v.HasDefault, v.Default, v.Tag.Envs})
		v.Target = reflect.New(v.Target.Type()).Elem()
		v.HasDefault, v.Default, v.Tag.Envs = false, "", nil
	}
	return func() {
		for _, s := range keep {
			s.v.Target = s.target
			s.v.HasDefault, s.v.Default, s.v.Tag.Envs = s.hasDefault, s.def, s.envs
		}
	}
}

// dashAsEmpty replaces each lone "-" before "--" with an empty value, as
// kingpin read it.
func dashAsEmpty(args []string) []string {
	out := args
	for i, a := range args {
		if a == "--" {
			break
		}
		if a == "-" {
			if &out[0] == &args[0] {
				out = append([]string(nil), args...)
			}
			out[i] = ""
		}
	}
	return out
}

// Given reports whether the flag named name was given on the command line,
// rather than taking its default or an environment variable.
func Given(ctx *kong.Context, name string) bool {
	for _, p := range ctx.Path {
		if p.Flag != nil && p.Flag.Name == name && !p.Resolved {
			return true
		}
	}
	return false
}

// fillMaps gives each map[string]string flag left nil an empty map, as
// kingpin's StringMap allocated one even when the flag was absent.
// Custom map values stay nil, as kingpin left them.
// Values runs it after kong resets the values and before the actions.
func fillMaps(k *kong.Kong) {
	values(k.Model.Node, func(v *kong.Value) {
		if t := v.Target; t.IsValid() && t.Type() == reflect.TypeOf(map[string]string{}) && t.IsNil() && t.CanSet() {
			t.Set(reflect.MakeMap(t.Type()))
		}
	})
}

// readLikeKingpin makes kong read values the way kingpin did, until restore
// is called.
// kingpin ignored empty environment variables, and those of values given on
// the command line, so a bad variable could not break a command line that
// overrides it.
// It also ignored the defaults of given values, and only split environment
// variables and default lists at newlines, never a given value.
// A list's environment variable was split into lines, without the line
// endings at the end.
// Only the parser's view changes, the environment is left as is.
func readLikeKingpin(k *kong.Kong, given map[*kong.Value]bool) (restore func()) {
	type saved struct {
		envs        []string
		hasDefault  bool
		def         string
		sep, mapSep rune
	}
	changed := map[*kong.Value]saved{}
	values(k.Model.Node, func(v *kong.Value) {
		before := saved{v.Tag.Envs, v.HasDefault, v.Default, v.Tag.Sep, v.Tag.MapSep}
		var keep []string
		for _, e := range v.Tag.Envs {
			if val, ok := os.LookupEnv(e); !ok || val != "" && !given[v] {
				keep = append(keep, e)
			}
		}
		if len(keep) != len(v.Tag.Envs) {
			v.Tag.Envs = keep
			changed[v] = before
		}
		// kingpin split a list's environment variable into lines, ignoring
		// the line endings at the end, so kong gets those lines as a
		// default.
		if text, ok := envValue(v); ok && !given[v] && (v.Tag.Sep == '\n' || v.Tag.MapSep == '\n') {
			if lines := strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n"); lines != text {
				v.Tag.Envs, v.HasDefault, v.Default = nil, true, lines
				changed[v] = before
			}
		}
		if given[v] {
			v.HasDefault = false
			if v.Tag.Sep == '\n' {
				v.Tag.Sep = -1
			}
			if v.Tag.MapSep == '\n' {
				v.Tag.MapSep = -1
			}
			changed[v] = before
		}
	})
	return func() {
		for v, s := range changed {
			v.Tag.Envs, v.HasDefault, v.Default, v.Tag.Sep, v.Tag.MapSep = s.envs, s.hasDefault, s.def, s.sep, s.mapSep
		}
	}
}

// MustParse is Parse that stops on errors, like kingpin.MustParse.
// It prints "<program>: error: <message>, try --help" and exits with
// status 1.
//
// kingpin reported this through kingpin.CommandLine, not the application.
// So, as with Fatalf, the name comes from the executable, the message goes
// to the stderr the program started with, and k's Exit is not used.
func MustParse(k *kong.Kong, args []string, uc *UsageContext) string {
	cmd, err := Parse(k, args, uc)
	if err != nil {
		Fatalf("%s, try --help", err)
	}
	return cmd
}

// NoHelpShortFlag removes the -h short form kong gives its help flag, for
// programs whose kingpin help flag had none.
func NoHelpShortFlag() kong.Option {
	return HelpShortFlag(0)
}

// HelpShortFlag gives the help flag the short form kingpin's had, such as
// -H, instead of kong's -h.
func HelpShortFlag(short rune) kong.Option {
	return kong.PostBuild(func(k *kong.Kong) error {
		for _, f := range k.Model.Flags {
			if f.Name == "help" {
				f.Short = short
			}
		}
		return nil
	})
}

// PrintUsage prints the usage of the command args select, as kingpin's
// Application.Usage did.
// It stops the program if kingpin could not have read the command line.
func PrintUsage(k *kong.Kong, args []string, uc *UsageContext) {
	if version == 2 {
		// kingpin v2 read @file arguments here too.
		expanded, err := ExpandArgs(args)
		k.FatalIfErrorf(err)
		args = expanded
	}
	s, err := scan(k, args, true)
	k.FatalIfErrorf(err)
	k.FatalIfErrorf(usageAt(k.Stdout, k, s.path, uc))
}

// HideHelp hides the help flag from usage, as the program did with kingpin's.
func HideHelp() kong.Option {
	return kong.PostBuild(func(k *kong.Kong) error {
		for _, f := range k.Model.Flags {
			if f.Name == "help" {
				f.Hidden = true
			}
		}
		return nil
	})
}

// HelpEnvar shows an environment variable next to the help flag in usage.
// kingpin's DefaultEnvars gave one to every flag, --help included, but never
// read it, so the parser does not either.
func HelpEnvar(name string) kong.Option {
	return kong.PostBuild(func(k *kong.Kong) error {
		helpEnvars.Store(k.Model, name)
		return nil
	})
}

// helpEnvars maps each application to its help flag's environment variable.
// It cannot go in the flag's own list: kong rewrites that list after the
// options run, and would read the variable.
var helpEnvars sync.Map

// HelpPrinter returns a kong help printer rendering kingpin's usage, for
// parsers whose --help kong handles itself.
func HelpPrinter(uc *UsageContext) kong.HelpPrinter {
	return func(_ kong.HelpOptions, ctx *kong.Context) error {
		return Usage(ctx.Stdout, ctx.Kong, ctx, uc)
	}
}

// versionFlagIndex returns the position in args of the application's
// kong.VersionFlag, or -1.
func versionFlagIndex(k *kong.Kong, args []string) int {
	for _, f := range k.Model.Flags {
		if f.Target.Type() != reflect.TypeOf(kong.VersionFlag(false)) {
			continue
		}
		for i, a := range args {
			if a == "--"+f.Name || f.Short != 0 && a == "-"+string(f.Short) {
				return i
			}
		}
	}
	return -1
}

func isHelpCommand(n *kong.Node) bool {
	if !n.Target.IsValid() {
		return false
	}
	_, ok := n.Target.Addr().Interface().(*HelpCommand)
	return ok
}

func without(args []string, drop string) []string {
	var out []string
	for _, a := range args {
		if a != drop {
			out = append(out, a)
		}
	}
	return out
}

// alwaysStdout is where completion output and, in kingpin v2, long help and
// man pages go.
// kingpin wrote them to the process's standard output, whatever writers the
// application had.
func alwaysStdout(*kong.Kong) io.Writer {
	return os.Stdout
}

// ExpandArgs replaces each @file argument before "--" with the lines of that
// file, as kingpin v2 did.
// Lines starting with # are skipped, and so are blank lines since v2.3.0.
// The lines read can themselves be @file arguments.
func ExpandArgs(args []string) ([]string, error) {
	var out []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return append(out, args[i:]...), nil
		}
		if !strings.HasPrefix(a, "@") {
			out = append(out, a)
			continue
		}
		lines, err := argsFromFile(a[1:])
		if err != nil {
			return nil, err
		}
		rest, err := ExpandArgs(append(lines, args[i+1:]...))
		if err != nil {
			return nil, err
		}
		return append(out, rest...), nil
	}
	return out, nil
}

func argsFromFile(name string) ([]string, error) {
	if name == "" {
		return nil, errors.New("expected @ file to expand arguments from")
	}
	f, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open arguments file %q: %s", name, err)
	}
	defer f.Close()
	var out []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || !legacy && strings.TrimSpace(line) == "" {
			continue
		}
		out = append(out, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read arguments from %q: %s", name, err)
	}
	return out, nil
}

// UsageToStderr sends the parser's standard output to standard error, where
// kingpin v2 printed help and the version.
// Errors already go there.
func UsageToStderr() kong.Option {
	return kong.Writers(os.Stderr, os.Stderr)
}

func pickStatus(v2, v3 int) int {
	if version == 2 {
		return v2
	}
	return v3
}

// isPath reports whether v holds an existing file or directory.
func isPath(v *kong.Value) bool {
	return v.Tag.Type == "existingfile" || v.Tag.Type == "existingdir"
}

// checkDefault checks the environment variable or, failing that, the
// default of a value missing from the command line.
// An empty default counts.
func checkDefault(v *kong.Value) error {
	text, ok := envValue(v)
	if !ok {
		text, ok = v.Default, v.HasDefault
	}
	if !ok {
		return nil
	}
	for _, t := range valueParts(v, text) {
		if err := checkValue(v, t); err != nil {
			return err
		}
	}
	return nil
}
