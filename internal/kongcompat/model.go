// Package kongcompat makes a program that now uses kong look and behave like
// it did with kingpin: same usage output, same error messages, same help
// command and shell completion.
//
// kong does the parsing.
// This package only reproduces what users saw and what scripts relied on.
// You can drop it once the program uses kong's own help and errors.
package kongcompat

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
)

// FlagGroupModel lists flags in declaration order.
type FlagGroupModel struct {
	Flags []*ClauseModel
}

// FlagByName returns the flag with the given long name.
func (f *FlagGroupModel) FlagByName(name string) *ClauseModel {
	for _, flag := range f.Flags {
		if flag.Name == name {
			return flag
		}
	}
	return nil
}

// FlagSummary lists the required flags, followed by [<flags>] when there are
// optional ones.
func (f *FlagGroupModel) FlagSummary() string {
	out := []string{}
	count := 0
	for _, flag := range f.Flags {
		// kingpin v2.3.0 and later left all their built-in flags out of the
		// count, other releases only --help.
		if v2From23() && !v2BuiltinFlags[flag.Name] || !v2From23() && flag.Name != "help" {
			count++
		}
		if flag.Required {
			if flag.IsBoolFlag() {
				if flag.IsNegatable() {
					out = append(out, fmt.Sprintf("--[no-]%s", flag.Name))
				} else {
					out = append(out, fmt.Sprintf("--%s", flag.Name))
				}
			} else {
				out = append(out, fmt.Sprintf("--%s=%s", flag.Name, flag.FormatPlaceHolder()))
			}
		}
	}
	if count != len(out) {
		out = append(out, "[<flags>]")
	}
	return strings.Join(out, " ")
}

// ClauseModel describes a flag or a positional argument.
type ClauseModel struct {
	Name        string
	Help        string
	Short       rune
	Default     []string
	PlaceHolder string
	Required    bool
	Hidden      bool
	// Value is the kong value behind the flag or argument.
	Value      *kong.Value
	Cumulative bool
	Envar      string

	boolFlag bool
	// quoted is set for a value kingpin knew was a string but kong does
	// not, such as a was-set wrapper.
	quoted bool
}

// String returns the current value.
func (c *ClauseModel) String() string {
	if c.Value == nil || !c.Value.Target.IsValid() {
		return ""
	}
	return fmt.Sprint(c.Value.Target.Interface())
}

// IsBoolFlag reports whether the flag takes no value, like booleans and
// counters.
func (c *ClauseModel) IsBoolFlag() bool {
	return c.boolFlag || c.Value != nil && (c.Value.IsBool() || c.Value.IsCounter())
}

// IsNegatable reports whether the flag also has a --no- form, which every
// boolean flag had in kingpin v2.
func (c *ClauseModel) IsNegatable() bool {
	if version == 2 {
		return !legacy && c.IsBoolFlag()
	}
	return c.Value != nil && c.Value.Flag != nil && c.Value.Flag.Tag.Negatable != ""
}

// HelpWithEnvar returns the help followed by the environment variable, as
// kingpin v2's templates show it.
func (c *ClauseModel) HelpWithEnvar() string {
	if c.Envar == "" {
		return c.Help
	}
	return fmt.Sprintf("%s ($%s)", c.Help, c.Envar)
}

// FormatPlaceHolder returns the text shown for the flag's value.
func (c *ClauseModel) FormatPlaceHolder() string {
	if c.PlaceHolder != "" {
		return c.PlaceHolder
	}
	switch {
	case len(c.Default) > 0 && c.isString():
		return quote(c.Default)
	case len(c.Default) > 1:
		return c.Default[0] + "..."
	case len(c.Default) > 0:
		return c.Default[0]
	}
	return strings.ToUpper(c.Name)
}

// isString reports whether the value is a plain string, which kingpin quoted
// in placeholders.
// Enums, paths and custom types were not.
func (c *ClauseModel) isString() bool {
	if c.quoted {
		return true
	}
	if c.Value == nil || c.Value.Enum != "" || isPath(c.Value) || !c.Value.Target.IsValid() {
		return false
	}
	return c.Value.Target.Type() == reflect.TypeOf("")
}

// quote shows a string default as kingpin did.
func quote(defaults []string) string {
	s := strconv.Quote(defaults[0])
	if len(defaults) > 1 {
		s += "..."
	}
	return s
}

// ArgGroupModel lists positional arguments.
type ArgGroupModel struct {
	Args []*ClauseModel
}

// ArgSummary renders the arguments as in "<a> [<b>]".
func (a *ArgGroupModel) ArgSummary() string {
	depth := 0
	out := []string{}
	for _, arg := range a.Args {
		h := "<" + arg.Name + ">"
		if version == 2 {
			if v2From23() && arg.PlaceHolder != "" {
				h = arg.PlaceHolder
			}
		} else if arg.Cumulative {
			h += " ..."
		}
		if !arg.Required {
			h = "[" + h
			depth++
		}
		out = append(out, h)
	}
	if len(out) == 0 {
		return ""
	}
	out[len(out)-1] = out[len(out)-1] + strings.Repeat("]", depth)
	return strings.Join(out, " ")
}

// CmdGroupModel lists subcommands.
type CmdGroupModel struct {
	Commands []*CmdModel
}

// FlattenedCommands returns the runnable commands below the group, depth
// first.
func (c *CmdGroupModel) FlattenedCommands() (out []*CmdModel) {
	for _, cmd := range c.Commands {
		if len(cmd.Commands) == 0 {
			out = append(out, cmd)
		}
		out = append(out, cmd.FlattenedCommands()...)
	}
	return
}

// CmdModel describes a command.
type CmdModel struct {
	Name    string
	Aliases []string
	Help    string
	Depth   int
	Hidden  bool
	Default bool
	Parent  *CmdModel
	// Node is the kong node behind the command.
	Node *kong.Node
	*FlagGroupModel
	*ArgGroupModel
	*CmdGroupModel
}

// String is how templates print a command: kingpin v3 showed the flags and
// arguments of every level, v2 only the command path.
func (c *CmdModel) String() string {
	if version == 2 {
		return c.FullCommand()
	}
	return c.CmdSummary()
}

// CmdSummary renders the command path with each level's flags and arguments.
func (c *CmdModel) CmdSummary() string {
	out := []string{}
	for cursor := c; cursor != nil; cursor = cursor.Parent {
		text := cursor.Name
		if cursor.Default {
			text = "*" + text
		}
		if flags := cursor.FlagSummary(); flags != "" {
			text += " " + flags
		}
		if args := cursor.ArgSummary(); args != "" {
			text += " " + args
		}
		out = append([]string{text}, out...)
	}
	return strings.Join(out, " ")
}

// FullCommand is the command path, without arguments or flags.
func (c *CmdModel) FullCommand() string {
	out := []string{}
	for i := c; i != nil; i = i.Parent {
		out = append([]string{i.Name}, out...)
	}
	return strings.Join(out, " ")
}

// ApplicationModel describes the whole command line.
type ApplicationModel struct {
	Name    string
	Help    string
	Version string
	Author  string
	*ArgGroupModel
	*CmdGroupModel
	*FlagGroupModel
}

// AppSummary renders the application usage line.
func (a *ApplicationModel) AppSummary() string {
	summary := a.Name
	if flags := a.FlagSummary(); flags != "" {
		summary += " " + flags
	}
	if args := a.ArgSummary(); args != "" {
		summary += " " + args
	}
	if len(a.Commands) > 0 {
		summary += " <command>"
	}
	return summary
}

// FindCommand returns the model of the command at the given kong node.
func (a *ApplicationModel) FindCommand(n *kong.Node) *CmdModel {
	var find func(cmds []*CmdModel) *CmdModel
	find = func(cmds []*CmdModel) *CmdModel {
		for _, c := range cmds {
			if c.Node == n {
				return c
			}
			if found := find(c.Commands); found != nil {
				return found
			}
		}
		return nil
	}
	return find(a.Commands)
}

// Model builds the usage model of a kong application.
func Model(app *kong.Application) *ApplicationModel {
	return &ApplicationModel{
		Name:           app.Name,
		Help:           app.Help,
		FlagGroupModel: appFlags(app),
		ArgGroupModel:  argGroup(app.Positional),
		CmdGroupModel:  cmdGroup(app.Node, nil, 0),
	}
}

// builtinFlags are the hidden flags kingpin gave every application, for shell
// completion and, in v2, long help and man pages.
// They never show in help.
// Except in kingpin v2.3.0 and later, they still add "[<flags>]" to the
// usage line.
func builtinFlags() []string {
	if version == 2 {
		return []string{"help-long", "help-man", "completion-bash", "completion-script-bash", "completion-script-zsh"}
	}
	return []string{"completion-bash", "completion-script-bash", "completion-script-zsh"}
}

var v2BuiltinFlags = map[string]bool{
	"help": true, "help-long": true, "help-man": true,
	"completion-bash": true, "completion-script-bash": true, "completion-script-zsh": true,
}

// appFlags lists the application's flags as kingpin did: help first, then
// the hidden built-in flags, then the application's own.
func appFlags(app *kong.Application) *FlagGroupModel {
	m := flagGroupModel(commandFlags(app.Node))
	var out []*ClauseModel
	for i, f := range m.Flags {
		out = append(out, f)
		if i == 0 && f.Name == "help" {
			if env, ok := helpEnvars.Load(app); ok {
				f.Envar = env.(string)
			}
			if version == 2 {
				f.Help = "Show context-sensitive help (also try --help-long and --help-man)."
			}
			for _, name := range builtinFlags() {
				out = append(out, &ClauseModel{Name: name, Hidden: true, boolFlag: true})
			}
		}
	}
	m.Flags = out
	return m
}

func flagGroupModel(flags []*kong.Flag) *FlagGroupModel {
	m := &FlagGroupModel{}
	for _, f := range flags {
		m.Flags = append(m.Flags, clause(f.Value, f.Short, f.PlaceHolder, f.Hidden))
	}
	return m
}

func argGroup(args []*kong.Positional) *ArgGroupModel {
	m := &ArgGroupModel{}
	for _, a := range args {
		m.Args = append(m.Args, argClause(a))
	}
	return m
}

func argClause(a *kong.Positional) *ClauseModel {
	return clause(a, 0, a.Tag.PlaceHolder, a.Tag.Hidden)
}

func clause(v *kong.Value, short rune, placeholder string, hidden bool) *ClauseModel {
	c := &ClauseModel{
		Name:        v.Name,
		Help:        v.Help,
		Short:       short,
		PlaceHolder: placeholder,
		Required:    v.Required,
		Hidden:      hidden,
		Value:       v,
		Cumulative:  cumulative(v),
	}
	if v.HasDefault && !emptyEnumDefault(v) {
		c.Default = valueParts(v, v.Default)
	}
	if len(v.Tag.Envs) > 0 {
		c.Envar = v.Tag.Envs[0]
	}
	// The tool writes kingpin's quoted default as the placeholder of such
	// a string value, to show it as kingpin did.
	// The program did not set that placeholder.
	if len(c.Default) > 0 && placeholder == quote(c.Default) {
		c.PlaceHolder, c.quoted = "", true
	}
	return c
}

// valueParts splits a default or an environment variable into the values
// of a list or a map.
func valueParts(v *kong.Value, text string) []string {
	switch {
	case v.IsMap() && v.Tag.MapSep != -1:
		return kong.SplitEscaped(text, v.Tag.MapSep)
	case cumulative(v) && v.Tag.Sep != -1:
		return kong.SplitEscaped(text, v.Tag.Sep)
	}
	return []string{text}
}

// emptyEnumDefault reports whether v is an optional enum with an empty
// default.
// kong needs that default but kingpin had none, so usage leaves it out.
func emptyEnumDefault(v *kong.Value) bool {
	if v.Default != "" || v.Enum == "" {
		return false
	}
	_, ok := v.EnumMap()[""]
	return ok
}

func cmdGroup(n *kong.Node, parent *CmdModel, depth int) *CmdGroupModel {
	m := &CmdGroupModel{}
	for _, child := range n.Children {
		if child.Type != kong.CommandNode {
			continue
		}
		m.Commands = append(m.Commands, cmdModel(child, parent, depth+1))
	}
	return m
}

func cmdModel(n *kong.Node, parent *CmdModel, depth int) *CmdModel {
	cmd := &CmdModel{
		Name:           n.Name,
		Aliases:        n.Aliases,
		Help:           n.Help,
		Depth:          depth,
		Hidden:         n.Hidden,
		Default:        n.Parent != nil && n.Parent.DefaultCmd == n,
		Parent:         parent,
		Node:           n,
		FlagGroupModel: flagGroupModel(commandFlags(n)),
		ArgGroupModel:  argGroup(n.Positional),
	}
	cmd.CmdGroupModel = cmdGroup(n, cmd, depth)
	return cmd
}

// ContextModel is what a usage template sees of the parsed command line: the
// command it reached, and the flags and arguments in scope there.
type ContextModel struct {
	SelectedCommand *CmdModel
	*FlagGroupModel
	*ArgGroupModel
}

// contextModel builds the template context for the last command of path, or
// for the application itself when path holds no command.
func contextModel(app *ApplicationModel, root *kong.Application, path []*kong.Node) *ContextModel {
	cm := &ContextModel{FlagGroupModel: appFlags(root), ArgGroupModel: argGroup(root.Positional)}
	for _, n := range path {
		if n.Type != kong.CommandNode {
			continue
		}
		cm.SelectedCommand = app.FindCommand(n)
		cm.Flags = append(cm.Flags, flagGroupModel(commandFlags(n)).Flags...)
		cm.Args = append(cm.Args, argGroup(n.Positional).Args...)
	}
	return cm
}

// commandPath returns the application node and the commands a parse went
// through.
func commandPath(ctx *kong.Context) []*kong.Node {
	if ctx == nil {
		return nil
	}
	var out []*kong.Node
	for _, p := range ctx.Path {
		if p.Command != nil {
			out = append(out, p.Command)
		}
	}
	return out
}

// cumulative reports whether a value can be given several times: a slice, a
// map, a counter, or a type whose IsCumulative method says so.
func cumulative(v *kong.Value) bool {
	if v.IsCounter() {
		return true
	}
	if v.Target.IsValid() && v.Target.CanAddr() {
		switch c := v.Target.Addr().Interface().(type) {
		case interface{ IsCumulative() bool }:
			return c.IsCumulative()
		case kong.MapperValue:
			// A custom value took one value in kingpin, unless it said
			// otherwise, even when kong sees a map or a slice.
			return false
		}
	}
	return v.IsCumulative()
}

// commandFlags are the flags of a command, in the order its FlagOrder
// method lists them when it has one.
// The converter writes that method when the fields cannot follow the order
// kingpin showed the flags in.
func commandFlags(n *kong.Node) []*kong.Flag {
	if !n.Target.IsValid() || !n.Target.CanAddr() {
		return n.Flags
	}
	o, ok := n.Target.Addr().Interface().(interface{ FlagOrder() []string })
	if !ok {
		return n.Flags
	}
	rank := map[string]int{}
	for i, name := range o.FlagOrder() {
		rank[name] = i + 1
	}
	// The flags it does not list, such as --help, keep their place.
	flags := append([]*kong.Flag(nil), n.Flags...)
	var slots []int
	var listed []*kong.Flag
	for i, f := range flags {
		if _, ok := rank[f.Name]; ok {
			slots = append(slots, i)
			listed = append(listed, f)
		}
	}
	sort.SliceStable(listed, func(i, j int) bool { return rank[listed[i].Name] < rank[listed[j].Name] })
	for i, at := range slots {
		flags[at] = listed[i]
	}
	return flags
}
