package kongcompat

import (
	"io"
	"strings"
	"text/template"

	"github.com/alecthomas/kong"
)

// CompletionScript writes the bash or zsh script that wires the shell's
// completion to the program's --completion-bash flag.
func CompletionScript(w io.Writer, appName, shell string) error {
	tmpl := pick(v2LegacyBashCompletionTemplate, v2BashCompletionTemplate, v3BashCompletionTemplate)
	if shell == "zsh" {
		tmpl = pick(v2LegacyZshCompletionTemplate, v2ZshCompletionTemplate, v3ZshCompletionTemplate)
	}
	tmpl = withContinuations(tmpl)
	t, err := template.New("completion").Parse(tmpl)
	if err != nil {
		return err
	}
	return t.Execute(w, map[string]any{"App": map[string]string{"Name": appName}})
}

// Complete returns the completion candidates for a partial command line,
// following kingpin's --completion-bash rules.
// It offers flags when the last word starts with "--", the next argument's
// options when one is expected, and otherwise the subcommands.
func Complete(k *kong.Kong, args []string) []string {
	ctx, _ := Trace(k, args)
	scanned, _ := scan(k, args, true)
	node := k.Model.Node
	if ctx != nil {
		// kingpin's completion functions could read the values given so
		// far, with their defaults and environment variables, which are
		// read here the way Parse reads them.
		// A bad one doesn't stop the others.
		restore := readLikeKingpin(k, scanned.given)
		_ = ctx.Reset()
		_, _ = ctx.Apply()
		restore()
		fillMaps(k)
		node = ctx.Selected()
		if node == nil {
			node = k.Model.Node
		}
	}

	var currArg, prevArg string
	if len(args) > 0 {
		currArg = args[len(args)-1]
	}
	if len(args) > 1 {
		prevArg = args[len(args)-2]
	}

	if strings.HasPrefix(currArg, "--") || strings.HasPrefix(prevArg, "--") {
		// kingpin v2.3.0 and later took everything after "--" as an
		// argument, so it completed no flag there.
		if v2From23() && len(args) > 1 {
			for _, a := range args[:len(args)-1] {
				if a == "--" {
					return nil
				}
			}
		}
		var flagName, flagValue string
		if strings.HasPrefix(prevArg, "--") && !strings.HasPrefix(currArg, "--") {
			flagName, flagValue = prevArg[2:], currArg
		} else if strings.HasPrefix(currArg, "--") {
			flagName = currArg[2:]
		}
		options, flagMatched, valueMatched := flagCompletion(commandFlags(node), flagName, flagValue, hintsOf(node))
		if valueMatched {
			return cmdCompletion(node, ctx, args, scanned.elements)
		}
		if node != k.Model.Node && !flagMatched {
			top, topMatched, topValueMatched := flagCompletion(k.Model.Flags, flagName, flagValue, hintsOf(k.Model.Node))
			if topValueMatched {
				return cmdCompletion(node, ctx, args, scanned.elements)
			}
			if topMatched {
				options = top
			} else {
				options = append(options, top...)
			}
		}
		return options
	}
	return cmdCompletion(node, ctx, args, scanned.elements)
}

func flagCompletion(flags []*kong.Flag, name, value string, hints map[string]func() []string) (options []string, flagMatch, valueMatch bool) {
	for _, f := range flags {
		if f.Name == name {
			opts := enumOptions(f.Value)
			if h, ok := hints[f.Name]; ok {
				// HintOptions and HintAction added up in kingpin.
				opts = append(tagHints(f.Value), h()...)
			}
			if len(opts) == 0 {
				return opts, true, true
			}
			isPrefix, matched := false, false
			for _, o := range opts {
				if value == o {
					matched = true
				} else if strings.HasPrefix(o, value) {
					isPrefix = true
				}
			}
			return opts, true, !isPrefix && matched
		}
		if !f.Hidden {
			options = append(options, "--"+f.Name)
		}
	}
	return options, false, false
}

// hinter is a command struct that gives completion hints only known at run
// time, by flag or argument name.
type hinter interface {
	Hints() map[string][]string
}

// actionHinter gives functions computing them, as kingpin's HintAction
// did, only called when needed.
type actionHinter interface {
	HintActions() map[string]func() []string
}

func hintsOf(n *kong.Node) map[string]func() []string {
	if n == nil || !n.Target.IsValid() {
		return nil
	}
	v := n.Target
	if v.CanAddr() {
		v = v.Addr()
	}
	out := map[string]func() []string{}
	if h, ok := v.Interface().(hinter); ok {
		for name, hints := range h.Hints() {
			hints := hints
			out[name] = func() []string { return hints }
		}
	}
	if h, ok := v.Interface().(actionHinter); ok {
		for name, f := range h.HintActions() {
			out[name] = f
		}
	}
	return out
}

func named(cmd *kong.Node, args []string) bool {
	for _, a := range args {
		if a == cmd.Name {
			return true
		}
		for _, alias := range cmd.Aliases {
			if a == alias {
				return true
			}
		}
	}
	return false
}

// tagHints returns the constant HintOptions kept in v's hints tag.
func tagHints(v *kong.Value) []string {
	if hints := v.Tag.Get("hints"); hints != "" {
		return strings.Split(hints, ",")
	}
	return nil
}

// enumOptions returns the completions of a value: its hints if the program
// gave some, which kingpin preferred over the enum's options, or else the
// options of a single-valued enum.
// kingpin v2 only offered enum options for flags.
func enumOptions(v *kong.Value) []string {
	if hints := tagHints(v); hints != nil {
		return hints
	}
	if v.Enum == "" || v.IsSlice() || version == 2 && v.Flag == nil {
		return nil
	}
	var out []string
	for _, o := range v.EnumSlice() {
		if o != "" {
			out = append(out, o)
		}
	}
	return out
}

func cmdCompletion(n *kong.Node, ctx *kong.Context, args []string, elements ParseElements) []string {
	var options []string
	// Count the arguments given, except the empty word being completed
	// and an argument that takes more values.
	satisfied := 0
	for _, el := range elements {
		ac, ok := el.Clause.(*ArgClause)
		if !ok || el.Value == nil || *el.Value == "" || cumulative(ac.arg) {
			continue
		}
		for _, p := range n.Positional {
			if p == ac.arg {
				satisfied++
			}
		}
	}
	if ctx != nil {
		for _, p := range ctx.Path {
			// A default command entered without being named also offers
			// every command at its level.
			if p.Command != nil && p.Command.Parent != nil && !named(p.Command, args) {
				for _, c := range p.Command.Parent.Children {
					if c.Type == kong.CommandNode {
						options = append(options, c.Name)
					}
				}
			}
		}
	}
	if satisfied < len(n.Positional) {
		arg := n.Positional[satisfied]
		if h, ok := hintsOf(n)[arg.Name]; ok {
			return append(append(options, tagHints(arg)...), h()...)
		}
		return append(options, enumOptions(arg)...)
	}
	for _, c := range n.Children {
		if c.Type == kong.CommandNode && !c.Hidden {
			options = append(options, c.Name)
		}
	}
	return options
}
