package kongcompat

import (
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
)

// Application stands in for kingpin's Application in code that still calls
// it at run time.
// It parses, prints usage and reports errors the way kingpin did, using a
// kong parser built from the program's grammar.
//
// New code can use Kong directly.
type Application struct {
	// Kong is the underlying parser.
	Kong *kong.Kong
	// Name and Help show in usage, and can still be changed after New.
	Name, Help string
	// HelpFlag is the --help flag.
	HelpFlag *FlagClause
	// Grammar is the struct given to New, which the parser fills in.
	Grammar any

	usage *UsageContext
	// parsed is set once the command line was looked at, which is when
	// kingpin added its help command.
	parsed bool
}

// New builds a kong parser for cli, the struct that replaced the program's
// kingpin.New call and flag declarations.
// As with kingpin, --help has no -h short form unless set through HelpFlag,
// and the program exits wherever kingpin would have exited it.
// Use Terminate to change that.
func New(cli any, name, help string, options ...kong.Option) *Application {
	opts := []kong.Option{kong.Name(name), kong.Description(help), NoHelpShortFlag(), CommandNames()}
	if version == 2 {
		opts = append(opts, UsageToStderr())
	}
	opts = append(opts, Values())
	opts = append(opts, newOptions...)
	opts = append(opts, CommandVars())
	k := kong.Must(cli, append(opts, options...)...)
	app := &Application{Kong: k, Name: name, Help: help, Grammar: cli}
	app.HelpFlag = app.GetFlag("help")
	return app
}

// Bind is kong.Bind for a parser that is already built.
// It gives the commands' hooks the values the program used to pass to the
// functions that registered its commands.
func (a *Application) Bind(values ...any) *Application {
	_ = kong.Bind(values...).Apply(a.Kong)
	return a
}

// BindTo is kong.BindTo for a parser that is already built: it binds value
// to the interface iface points to, such as (*io.Writer)(nil).
func (a *Application) BindTo(value, iface any) *Application {
	_ = kong.BindTo(value, iface).Apply(a.Kong)
	return a
}

// Writers sets where usage and errors go (kingpin v3).
func (a *Application) Writers(out, err io.Writer) *Application {
	a.Kong.Stdout, a.Kong.Stderr = out, err
	return a
}

// Writer sends both usage and errors to w (kingpin v2).
func (a *Application) Writer(w io.Writer) *Application {
	return a.Writers(w, w)
}

// UsageWriter sets where usage goes (kingpin v2).
func (a *Application) UsageWriter(w io.Writer) *Application {
	a.Kong.Stdout = w
	return a
}

// ErrorWriter sets where errors go (kingpin v2).
func (a *Application) ErrorWriter(w io.Writer) *Application {
	a.Kong.Stderr = w
	return a
}

// Terminate sets the function called to exit the program, as in kingpin.
// With nil, the program keeps running.
func (a *Application) Terminate(terminate func(int)) *Application {
	if terminate == nil {
		terminate = func(int) {}
	}
	a.Kong.Exit = terminate
	return a
}

// UsageContext sets the template, functions and variables of usage.
func (a *Application) UsageContext(uc *UsageContext) *Application {
	a.usage = uc
	return a
}

// UsageTemplate sets the usage template.
func (a *Application) UsageTemplate(template string) *Application {
	uc := UsageContext{}
	if a.usage != nil {
		uc = *a.usage
	}
	uc.Template = template
	a.usage = &uc
	return a
}

// Version sets the version --version prints.
// The flag itself comes from the grammar.
func (a *Application) Version(version string) *Application {
	_ = kong.Vars{"version": version}.Apply(a.Kong)
	return a
}

// Author sets the author shown by some templates.
func (a *Application) Author(author string) *Application {
	_ = kong.Vars{"author": author}.Apply(a.Kong)
	return a
}

// sync picks up changes the program made to Name and Help after New, since
// kingpin only read them when it printed usage.
func (a *Application) sync() {
	a.Kong.Model.Name, a.Kong.Model.Help = a.Name, a.Help
}

// GetFlag returns the application's flag with the given long name, or nil.
func (a *Application) GetFlag(name string) *FlagClause {
	for _, f := range a.Kong.Model.Flags {
		if f.Name == name {
			return newFlagClause(f)
		}
	}
	return nil
}

// Parse parses args as kingpin's Parse did and returns the selected command,
// such as "db migrate".
// See the Parse function for details.
func (a *Application) Parse(args []string) (string, error) {
	a.sync()
	a.parsed = true
	return Parse(a.Kong, args, a.usage)
}

// ParseContext walks args through the grammar without setting any value, as
// kingpin's ParseContext did.
// It returns what it found and the error kingpin would have returned.
func (a *Application) ParseContext(args []string) (*ParseContext, error) {
	a.parsed = true
	if version == 2 {
		expanded, err := ExpandArgs(args)
		if err != nil {
			return &ParseContext{Application: a}, err
		}
		args = expanded
	}
	s, err := scan(a.Kong, args, true)
	ctx := &ParseContext{Application: a, Elements: s.elements, path: s.path}
	if len(s.path) > 1 {
		ctx.SelectedCommand = newCmdClause(s.path[len(s.path)-1])
	}
	return ctx, err
}

// Model returns the data model usage templates see.
// As with kingpin, the help command only shows once the command line was
// parsed.
func (a *Application) Model() *ApplicationModel {
	a.sync()
	m := Model(a.Kong.Model)
	if !a.parsed {
		var cmds []*CmdModel
		for _, c := range m.Commands {
			if !isHelpCommand(c.Node) {
				cmds = append(cmds, c)
			}
		}
		m.Commands = cmds
	}
	m.Version, m.Author = kongVar(a.Kong, "version"), kongVar(a.Kong, "author")
	return m
}

// Usage prints the usage of the command args select.
func (a *Application) Usage(args []string) {
	a.sync()
	a.parsed = true
	PrintUsage(a.Kong, args, a.usage)
}

// UsageForContext prints the usage of the command ctx reached.
func (a *Application) UsageForContext(ctx *ParseContext) error {
	return a.usageFor(ctx, a.usage)
}

func (a *Application) usageFor(ctx *ParseContext, uc *UsageContext) error {
	a.sync()
	var path []*kong.Node
	if ctx != nil {
		path = ctx.path
	}
	return usageAt(a.Kong.Stdout, a.Kong, path, uc)
}

// Errorf prints "<name>: error: <message>" on the error writer.
func (a *Application) Errorf(format string, args ...any) {
	a.sync()
	Errors(a.Kong).Errorf(format, args...)
}

// Fatalf prints an error and terminates with status 1.
func (a *Application) Fatalf(format string, args ...any) {
	a.sync()
	Errors(a.Kong).Fatalf(format, args...)
}

// FatalIfError does nothing if err is nil.
// Otherwise it prints err, prefixed with the formatted message when format
// is not empty, and terminates with status 1.
func (a *Application) FatalIfError(err error, format string, args ...any) {
	a.sync()
	Errors(a.Kong).FatalIfError(err, format, args...)
}

// FatalUsage prints an error and the application's usage, and terminates
// with status 1.
func (a *Application) FatalUsage(format string, args ...any) {
	a.Errorf(format, args...)
	a.sync()
	// kingpin v2 printed this usage to the error writer, v3 to the usage one.
	w := a.Kong.Stdout
	if version == 2 {
		w = a.Kong.Stderr
	}
	_ = usageAt(w, a.Kong, nil, a.usage)
	a.Kong.Exit(1)
}

// FatalUsageContext prints an error and the usage of the command ctx
// reached, and terminates with status 1.
func (a *Application) FatalUsageContext(ctx *ParseContext, format string, args ...any) {
	a.Errorf(format, args...)
	_ = a.UsageForContext(ctx)
	a.Kong.Exit(1)
}

// ParseContext is what Application.ParseContext found on the command line.
type ParseContext struct {
	Application *Application
	// SelectedCommand is the deepest command reached, or nil.
	SelectedCommand *CmdClause
	// Elements are the flags, arguments and commands found, in order.
	Elements ParseElements

	path []*kong.Node
}

// ParseElement is a flag, argument or command found on the command line.
// The clause is in both Clause, as in kingpin v2, and OneOf, as in kingpin
// v3.
// Value is what was given to a flag or argument.
type ParseElement struct {
	Clause any
	OneOf  OneOfClause
	Value  *string
}

// OneOfClause holds the clause of a ParseElement, as kingpin v3 did.
type OneOfClause struct {
	Flag *FlagClause
	Arg  *ArgClause
	Cmd  *CmdClause
}

// ParseElements are the elements of a command line.
type ParseElements []*ParseElement

// FlagMap returns the flags found, by long name.
func (p ParseElements) FlagMap() map[string]*ParseElement {
	out := map[string]*ParseElement{}
	for _, e := range p {
		if f, ok := e.Clause.(*FlagClause); ok {
			out[f.flag.Name] = e
		}
	}
	return out
}

// ArgMap returns the arguments found, by name.
func (p ParseElements) ArgMap() map[string]*ParseElement {
	out := map[string]*ParseElement{}
	for _, e := range p {
		if a, ok := e.Clause.(*ArgClause); ok {
			out[a.arg.Name] = e
		}
	}
	return out
}

// CmdClause is a command of the grammar.
type CmdClause struct {
	node *kong.Node
	// cmdMixin.flagGroup.long is where kingpin kept a command's flags, for
	// code that reads them through reflection.
	cmdMixin cmdMixin
}

type cmdMixin struct{ *flagGroup }

type flagGroup struct{ long map[string]*FlagClause }

// Each kong flag, argument and command keeps one clause, as in kingpin, so
// code comparing them, as in element.Clause == app.HelpFlag, still works.
var flagClauses, argClauses, cmdClauses sync.Map

func newFlagClause(f *kong.Flag) *FlagClause {
	c, _ := flagClauses.LoadOrStore(f, &FlagClause{f})
	return c.(*FlagClause)
}

func newArgClause(a *kong.Positional) *ArgClause {
	c, _ := argClauses.LoadOrStore(a, &ArgClause{a})
	return c.(*ArgClause)
}

func newCmdClause(n *kong.Node) *CmdClause {
	if c, ok := cmdClauses.Load(n); ok {
		return c.(*CmdClause)
	}
	long := map[string]*FlagClause{}
	for _, f := range n.Flags {
		long[f.Name] = newFlagClause(f)
	}
	c, _ := cmdClauses.LoadOrStore(n, &CmdClause{node: n, cmdMixin: cmdMixin{&flagGroup{long}}})
	return c.(*CmdClause)
}

// CommandOf returns the command whose struct is cmd, or nil.
//
// Pass a pointer, such as the *BuildCommand a constructor returned, to tell
// apart commands that share a type.
// Otherwise the first command of that type is returned.
func (a *Application) CommandOf(cmd any) *CmdClause {
	v := reflect.ValueOf(cmd)
	t := reflect.TypeOf(cmd)
	if t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	var byType, same *kong.Node
	nodes(a.Kong.Model.Node, func(n *kong.Node) {
		if n.Type != kong.CommandNode || !n.Target.IsValid() || n.Target.Type() != t {
			return
		}
		if byType == nil {
			byType = n
		}
		if same == nil && v.Kind() == reflect.Pointer && n.Target.CanAddr() && n.Target.Addr().Pointer() == v.Pointer() {
			same = n
		}
	})
	if same == nil {
		same = byType
	}
	if same == nil {
		return nil
	}
	return newCmdClause(same)
}

// FullCommand returns the command's full name, as CommandName does.
func (c *CmdClause) FullCommand() string { return CommandName(c.node) }

func (c *CmdClause) model() *CmdModel {
	var parent *CmdModel
	if c.node.Parent != nil && c.node.Parent.Type == kong.CommandNode {
		parent = newCmdClause(c.node.Parent).model()
	}
	return cmdModel(c.node, parent, depth(c.node))
}

// FlagClause is a flag of the grammar.
type FlagClause struct{ flag *kong.Flag }

// Short sets the flag's short form.
func (f *FlagClause) Short(short rune) *FlagClause {
	f.flag.Short = short
	return f
}

// Hidden hides the flag from usage.
func (f *FlagClause) Hidden() *FlagClause {
	f.flag.Hidden = true
	return f
}

// Model returns the flag's data model.
func (f *FlagClause) Model() *ClauseModel {
	return clause(f.flag.Value, f.flag.Short, f.flag.PlaceHolder, f.flag.Hidden)
}

// ArgClause is a positional argument of the grammar.
type ArgClause struct{ arg *kong.Positional }

// Model returns the argument's data model.
func (a *ArgClause) Model() *ClauseModel {
	return argClause(a.arg)
}

func depth(n *kong.Node) int {
	d := 0
	for p := n; p != nil && p.Type == kong.CommandNode; p = p.Parent {
		d++
	}
	return d
}

// CommandName returns a command's full name as kingpin's FullCommand did.
// That is the command path, plus in kingpin v3 the arguments of every
// level, as in "auth add [<name>]".
func CommandName(n *kong.Node) string {
	if version == 2 {
		return fullCommand(n)
	}
	var levels []*kong.Node
	for p := n; p != nil && p.Type == kong.CommandNode; p = p.Parent {
		levels = append([]*kong.Node{p}, levels...)
	}
	var out []string
	for _, l := range levels {
		out = append(out, l.Name)
		for _, arg := range l.Positional {
			text := "<" + arg.Name + ">"
			if cumulative(arg) {
				text += " ..."
			}
			if !arg.Required {
				text = "[" + text + "]"
			}
			out = append(out, text)
		}
	}
	return strings.Join(out, " ")
}

// CommandNames sets the string fields tagged kongcompat:"command-name" in
// every command struct to the command's full name (see CommandName).
// They are filled when the parser is built, because kingpin's FullCommand
// worked at any time, even before parsing.
func CommandNames() kong.Option {
	return kong.PostBuild(func(k *kong.Kong) error {
		nodes(k.Model.Node, func(n *kong.Node) {
			if n.Type == kong.CommandNode && n.Target.IsValid() {
				setCommandName(n.Target, CommandName(n))
			}
		})
		return nil
	})
}

func setCommandName(v reflect.Value, name string) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		switch {
		case f.Tag.Get("kongcompat") == "command-name" && f.Type.Kind() == reflect.String && v.Field(i).CanSet():
			v.Field(i).SetString(name)
		case f.Anonymous:
			// An embedded struct is part of the same command.
			setCommandName(v.Field(i), name)
		}
	}
}

// Commands returns the command structs that are a T, in grammar order.
// T is usually an interface the commands implement.
func Commands[T any](a *Application) []T {
	var out []T
	nodes(a.Kong.Model.Node, func(n *kong.Node) {
		if n.Type == kong.CommandNode && n.Target.IsValid() && n.Target.CanAddr() {
			if c, ok := n.Target.Addr().Interface().(T); ok {
				out = append(out, c)
			}
		}
	})
	return out
}
