package kongcompat

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/kong"
)

// ErrCommandNotSpecified is kingpin's error for a command line that stops
// before reaching a command.
var ErrCommandNotSpecified = errors.New("command not specified")

// Check reads args the way kingpin did and returns the error kingpin would
// have reported, worded the same way, or nil if kingpin would have accepted
// the command line.
// It does not set any value, kong still does the actual parsing.
//
// It catches what kingpin rejected but kong lets through, such as a repeated
// single-valued flag, so a converted program accepts exactly the same
// command lines as before.
func Check(k *kong.Kong, args []string) error {
	_, err := scan(k, args, false)
	return err
}

// CheckSyntax is Check limited to what kingpin's ParseContext reported:
// unknown flags and commands, missing flag values and unexpected arguments.
// Required flags, enums and repeated flags were only checked when parsing.
func CheckSyntax(k *kong.Kong, args []string) error {
	_, err := scan(k, args, true)
	return err
}

// scan reads args the way kingpin did.
// The returned checker records what it found: the commands entered (path),
// the values given and whether --help was seen.
func scan(k *kong.Kong, args []string, syntaxOnly bool) (*checker, error) {
	s := newChecker(args, syntaxOnly)
	s.enter(k.Model.Node)
	return s, s.run()
}

func newChecker(args []string, syntaxOnly bool) *checker {
	return &checker{
		syntaxOnly: syntaxOnly,
		long:       map[string]*kong.Flag{},
		short:      map[string]*kong.Flag{},
		seen:       map[*kong.Flag]bool{},
		given:      map[*kong.Value]bool{},
		tokens:     args,
	}
}

type checker struct {
	long, short map[string]*kong.Flag
	node        *kong.Node
	path        []*kong.Node
	flags       []*kong.Flag
	args        []*kong.Positional
	nextArg     int
	seen        map[*kong.Flag]bool
	given       map[*kong.Value]bool
	tokens      []string
	pushed      []token
	argsOnly    bool
	syntaxOnly  bool
	sawHelp     bool
	valueErr    error
	// helpArgs are the arguments given to the help command.
	helpArgs []string
	// noRequired skips the checks for missing flags and arguments, as
	// kingpin did before running the help command.
	noRequired bool
	// elements are the flags, arguments and commands found, in order.
	elements ParseElements
}

type tokenKind int

const (
	tokEOL tokenKind = iota
	tokLong
	tokShort
	tokArg
)

type token struct {
	kind  tokenKind
	value string
}

func (t token) String() string {
	switch t.kind {
	case tokLong:
		return "--" + t.value
	case tokShort:
		return "-" + t.value
	case tokEOL:
		return "<EOL>"
	}
	return t.value
}

// enter brings a command's flags and arguments into scope.
func (s *checker) enter(n *kong.Node) {
	s.node = n
	s.path = append(s.path, n)
	for _, f := range commandFlags(n) {
		s.long[f.Name] = f
		if f.Short != 0 {
			s.short[string(f.Short)] = f
		}
		s.flags = append(s.flags, f)
	}
	s.args = append(s.args, n.Positional...)
}

// next returns the next token, split the way kingpin did.
// --name=value gives two tokens, and grouped short flags such as -abc are
// split as long as they are booleans.
func (s *checker) next() token {
	if len(s.pushed) > 0 {
		t := s.pushed[len(s.pushed)-1]
		s.pushed = s.pushed[:len(s.pushed)-1]
		return t
	}
	if len(s.tokens) == 0 {
		return token{kind: tokEOL}
	}
	arg := s.tokens[0]
	s.tokens = s.tokens[1:]
	if s.argsOnly {
		return token{tokArg, arg}
	}
	if arg == "--" {
		s.argsOnly = true
		return s.next()
	}
	if strings.HasPrefix(arg, "--") {
		parts := strings.SplitN(arg[2:], "=", 2)
		if len(parts) == 2 {
			s.pushed = append(s.pushed, token{tokArg, parts[1]})
		}
		return token{tokLong, parts[0]}
	}
	if strings.HasPrefix(arg, "-") {
		if len(arg) == 1 {
			return token{tokArg, ""}
		}
		r, size := utf8.DecodeRuneInString(arg[1:])
		short := string(r)
		flag, ok := s.short[short]
		if ok && !valueless(flag) {
			// A short flag with its value attached: -fVALUE.
			if len(arg) > 1+size {
				s.pushed = append(s.pushed, token{tokArg, arg[1+size:]})
			}
			return token{tokShort, short}
		}
		if len(arg) > 1+size {
			s.tokens = append([]string{"-" + arg[1+size:]}, s.tokens...)
		}
		return token{tokShort, short}
	}
	return token{tokArg, arg}
}

// run checks the command line in kingpin's order: defaults first, then
// syntax, then the values given, then what is missing.
func (s *checker) run() error {
	if err := s.scan(); err != nil {
		if derr := s.checkDefaults(); derr != nil {
			return derr
		}
		return err
	}
	return s.finish()
}

// scan reads the command line up to its end or to a syntax error.
// Bad values are kept in valueErr for later.
func (s *checker) scan() error {
	for {
		t := s.next()
		switch t.kind {
		case tokEOL:
			return nil
		case tokLong, tokShort:
			if err := s.flag(t); err != nil {
				return err
			}
		case tokArg:
			if s.nextArg < len(s.args) {
				arg := s.args[s.nextArg]
				s.given[arg] = true
				value := t.value
				if isHelpCommand(s.node) {
					s.helpArgs = append(s.helpArgs, value)
				}
				ac := newArgClause(arg)
				s.elements = append(s.elements, &ParseElement{Clause: ac, OneOf: OneOfClause{Arg: ac}, Value: &value})
				if !cumulative(arg) {
					s.nextArg++
				}
				if !s.syntaxOnly {
					s.badValue(checkValue(arg, t.value))
				}
				s.badValue(s.checkEnum(arg, t.value))
				continue
			}
			if cmd := s.command(t.value); cmd != nil {
				s.enterCommand(cmd)
				continue
			}
			// kingpin passed the argument on to the default command.
			if s.enterDefault(t) {
				continue
			}
			if hasCommands(s.node) {
				if version == 2 {
					return fmt.Errorf("expected command but got %q", t.String())
				}
				return fmt.Errorf("expected command but got %s", t)
			}
			if version == 2 {
				return fmt.Errorf("unexpected %s", t)
			}
			return fmt.Errorf("unexpected '%s'", t)
		}
	}
}

func (s *checker) flag(t token) error {
	name := t.value
	var flag *kong.Flag
	ok := false
	negated := false
	if t.kind == tokLong {
		flag, ok = s.long[name]
		if !ok && strings.HasPrefix(name, "no-") {
			flag, ok = s.long[name[3:]]
			negated = true
			if ok && (flag.Tag.Negatable == "" || !flag.IsBool()) {
				ok = false
			}
		}
		if !ok && builtinFlag(name) {
			// kingpin's own hidden flags, which kong does not declare.
			return nil
		}
		if !ok {
			if s.enterDefault(t) {
				return nil
			}
			return fmt.Errorf("unknown long flag '%s'", t)
		}
	} else {
		flag, ok = s.short[name]
		if !ok {
			if s.enterDefault(t) {
				return nil
			}
			return fmt.Errorf("unknown short flag '%s'", t)
		}
	}
	if !s.syntaxOnly && s.seen[flag] && !cumulative(flag.Value) && !isCompletionFlag(flag.Name) {
		s.badValue(fmt.Errorf("flag '%s' cannot be repeated", flag.Name))
	}
	s.seen[flag] = true
	s.given[flag.Value] = true
	if flag.Name == "help" {
		s.sawHelp = true
	}
	fc := newFlagClause(flag)
	element := &ParseElement{Clause: fc, OneOf: OneOfClause{Flag: fc}}
	s.elements = append(s.elements, element)
	if valueless(flag) {
		// kingpin recorded the boolean the flag set, false for --no-x and
		// noflag flags.
		value := strconv.FormatBool(!negated && flag.Tag.Type != "noflag")
		element.Value = &value
		return nil
	}
	v := s.next()
	if v.kind != tokArg {
		return fmt.Errorf("expected argument for flag '%s'", t)
	}
	value := v.value
	element.Value = &value
	if !s.syntaxOnly {
		s.badValue(checkValue(flag.Value, v.value))
	}
	s.badValue(s.checkEnum(flag.Value, v.value))
	return nil
}

// badValue records the first bad value, which kingpin only reported after
// reading the whole command line.
func (s *checker) badValue(err error) {
	if s.valueErr == nil {
		s.valueErr = err
	}
}

// enterDefault moves into the default command and pushes t back to try it
// there, as kingpin did with tokens the current command did not know.
func (s *checker) enterDefault(t token) bool {
	d := s.node.DefaultCmd
	if d == nil || s.sawHelp {
		return false
	}
	s.enterCommand(d)
	s.pushed = append(s.pushed, t)
	return true
}

// enterCommand enters a command and records it as an element.
func (s *checker) enterCommand(cmd *kong.Node) {
	s.enter(cmd)
	cc := newCmdClause(cmd)
	s.elements = append(s.elements, &ParseElement{Clause: cc, OneOf: OneOfClause{Cmd: cc}})
}

// builtinFlag reports whether name is one of the hidden flags kingpin gave
// every application.
func builtinFlag(name string) bool {
	for _, b := range builtinFlags() {
		if b == name {
			return true
		}
	}
	return false
}

func isCompletionFlag(name string) bool {
	return strings.HasPrefix(name, "completion-")
}

// checkEnum rejects a value outside an enum, as kingpin's EnumVar did.
func (s *checker) checkEnum(v *kong.Value, value string) error {
	if v.Enum == "" || s.syntaxOnly {
		return nil
	}
	options := v.EnumSlice()
	// The conversion adds an empty last option to an optional enum with an
	// empty default, for kong's sake.
	// kingpin rejected an empty value it did not list.
	if n := len(options); n > 0 && options[n-1] == "" && !v.Required && v.Default == "" && v.Target.Kind() != reflect.Slice {
		options = options[:n-1]
	}
	for _, o := range options {
		if o == value {
			return nil
		}
	}
	return fmt.Errorf("enum value must be one of %s, got '%s'", strings.Join(options, ","), value)
}

func (s *checker) command(name string) *kong.Node {
	for _, c := range s.node.Children {
		if c.Type != kong.CommandNode {
			continue
		}
		if c.Name == name {
			return c
		}
		for _, a := range c.Aliases {
			if a == name {
				return c
			}
		}
	}
	return nil
}

func hasCommands(n *kong.Node) bool {
	for _, c := range n.Children {
		if c.Type == kong.CommandNode {
			return true
		}
	}
	return false
}

// finish runs the checks kingpin did after reading all arguments, in the
// same order: bad defaults, bad values, missing subcommand, missing
// required values, missing command.
func (s *checker) finish() error {
	// After --help, kingpin did not enter default commands.
	for s.node.DefaultCmd != nil && !s.sawHelp {
		s.enterCommand(s.node.DefaultCmd)
	}
	if s.syntaxOnly {
		return nil
	}
	if err := s.checkDefaults(); err != nil {
		return err
	}
	if s.valueErr != nil || s.noRequired {
		return s.valueErr
	}
	commands := hasCommands(s.node)
	if commands && len(s.path) > 1 {
		return fmt.Errorf("must select a subcommand of '%s'", fullCommand(s.node))
	}
	if err := s.missingRequired(); err != nil {
		return err
	}
	if commands {
		return ErrCommandNotSpecified
	}
	return nil
}

// checkDefaults checks the defaults of what was not given, as kingpin did
// first, unless --help was given.
func (s *checker) checkDefaults() error {
	if s.syntaxOnly || s.sawHelp {
		return nil
	}
	for _, f := range s.flags {
		if !s.given[f.Value] {
			if err := checkDefault(f.Value); err != nil {
				return err
			}
		}
	}
	for _, a := range s.args {
		if !s.given[a] {
			if err := checkDefault(a); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *checker) missingRequired() error {
	var missing []string
	for _, f := range s.flags {
		if f.Required && !s.given[f.Value] && !envSet(f.Value) {
			if !v2From23() {
				return fmt.Errorf("required flag --%s not provided", f.Name)
			}
			missing = append(missing, "'--"+f.Name+"'")
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("required flag(s) %s not provided", strings.Join(missing, ", "))
	}
	for _, a := range s.args {
		if a.Required && !s.given[a] && !envSet(a) {
			return fmt.Errorf("required argument '%s' not provided", a.Name)
		}
	}
	return nil
}

// envSet reports whether one of a value's environment variables is set and
// not empty.
// kingpin treated an empty variable as unset.
func envSet(v *kong.Value) bool {
	_, ok := envValue(v)
	return ok
}

// envValue returns the first environment variable of v that is set and not
// empty.
func envValue(v *kong.Value) (string, bool) {
	for _, e := range v.Tag.Envs {
		if s := os.Getenv(e); s != "" {
			return s, true
		}
	}
	return "", false
}

// fullCommand returns the command path of a node, as kingpin's FullCommand
// did: names only, without the application name, aliases or arguments.
func fullCommand(n *kong.Node) string {
	var parts []string
	for ; n != nil && n.Type == kong.CommandNode; n = n.Parent {
		parts = append([]string{n.Name}, parts...)
	}
	return strings.Join(parts, " ")
}

// FullCommand returns the command a parse selected, as kingpin's
// FullCommand did, or "" when no command was selected.
func FullCommand(ctx *kong.Context) string {
	if ctx == nil {
		return ""
	}
	return fullCommand(ctx.Selected())
}

// valueless reports whether a flag takes no value: a boolean, or a counter,
// which kingpin read like one.
func valueless(f *kong.Flag) bool {
	return f.IsBool() || f.IsCounter()
}
