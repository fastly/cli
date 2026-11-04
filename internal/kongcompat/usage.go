package kongcompat

import (
	"bytes"
	"fmt"
	"go/doc"
	"io"
	"reflect"
	"strings"
	"text/template"

	"github.com/alecthomas/kong"
)

// UsageContext configures how usage is rendered, as kingpin's did.
type UsageContext struct {
	// Template is a text/template, DefaultUsageTemplate if empty.
	Template string
	// Indent is the indentation multiplier, 2 if zero.
	Indent int
	// Width is where text wraps, the terminal width if zero.
	Width int
	// Funcs are extra template functions.
	Funcs template.FuncMap
	// Vars are extra values passed to the template.
	Vars map[string]any
}

const preIndent = "  "

// Usage prints the usage of the command ctx reached, or of the whole
// application when ctx is nil.
// The template gets the same data kingpin gave it: .App for the application
// model, .Context for the selected command and the flags and arguments in
// scope, and .Width.
func Usage(w io.Writer, k *kong.Kong, ctx *kong.Context, uc *UsageContext) error {
	return usageAt(w, k, commandPath(ctx), uc)
}

// usageAt prints the usage of the last command of path.
func usageAt(w io.Writer, k *kong.Kong, path []*kong.Node, uc *UsageContext) error {
	if uc == nil {
		uc = &UsageContext{}
	}
	indent := uc.Indent
	if indent == 0 {
		indent = 2
	}
	width := uc.Width
	if width == 0 {
		width = guessWidth(w)
	}
	tmpl := uc.Template
	if tmpl == "" {
		tmpl = pick(v2LegacyDefaultUsageTemplate, v2DefaultUsageTemplate, v3DefaultUsageTemplate)
	}
	tmpl = withContinuations(tmpl)
	funcs := templateFuncs(indent, width)
	for name, fn := range uc.Funcs {
		funcs[name] = fn
	}
	t, err := template.New("usage").Funcs(funcs).Parse(tmpl)
	if err != nil {
		return err
	}
	app := Model(k.Model)
	app.Version, app.Author = kongVar(k, "version"), kongVar(k, "author")
	data := map[string]any{
		"App":     app,
		"Width":   width,
		"Context": contextModel(app, k.Model, path),
	}
	for name, v := range uc.Vars {
		data[name] = v
	}
	return t.Execute(w, data)
}

// kongVar returns a variable set with kong.Vars, for the "version" and
// "author" that kingpin's man pages show.
// kong keeps its variables private, so they are read through reflection,
// and the result is empty if kong's internals change.
func kongVar(k *kong.Kong, name string) string {
	vars := reflect.ValueOf(k).Elem().FieldByName("vars")
	if vars.Kind() != reflect.Map || vars.Type().Key().Kind() != reflect.String {
		return ""
	}
	v := vars.MapIndex(reflect.ValueOf(name))
	if !v.IsValid() || v.Kind() != reflect.String {
		return ""
	}
	return v.String()
}

func templateFuncs(indent, width int) template.FuncMap {
	return template.FuncMap{
		"T": func(s string) string { return s },
		"Indent": func(level int) string {
			return strings.Repeat(" ", level*indent)
		},
		"Wrap": func(indent int, s string) string {
			buf := bytes.NewBuffer(nil)
			indentText := strings.Repeat(" ", indent)
			doc.ToText(buf, s, indentText, "  "+indentText, width-indent)
			return buf.String()
		},
		"FormatFlag": formatFlag,
		"FlagsToTwoColumns": func(f []*ClauseModel) [][2]string {
			rows := [][2]string{}
			haveShort := false
			for _, flag := range f {
				if flag.Short != 0 {
					haveShort = true
					break
				}
			}
			for _, flag := range f {
				if !flag.Hidden {
					help := flag.Help
					if v2From23() {
						help = flag.HelpWithEnvar()
					}
					rows = append(rows, [2]string{formatFlag(haveShort, flag), help})
				}
			}
			return rows
		},
		"RequiredFlags": func(f []*ClauseModel) []*ClauseModel {
			out := []*ClauseModel{}
			for _, flag := range f {
				if flag.Required {
					out = append(out, flag)
				}
			}
			return out
		},
		"OptionalFlags": func(f []*ClauseModel) []*ClauseModel {
			out := []*ClauseModel{}
			for _, flag := range f {
				if !flag.Required {
					out = append(out, flag)
				}
			}
			return out
		},
		"ArgsToTwoColumns": func(a []*ClauseModel) [][2]string {
			rows := [][2]string{}
			for _, arg := range a {
				if v2From23() && arg.Hidden {
					continue
				}
				s := "<" + arg.Name + ">"
				if v2From23() && arg.PlaceHolder != "" {
					s = arg.PlaceHolder
				}
				if !arg.Required {
					s = "[" + s + "]"
				}
				help := arg.Help
				if v2From23() {
					help = arg.HelpWithEnvar()
				}
				rows = append(rows, [2]string{s, help})
			}
			return rows
		},
		"CommandsToTwoColumns": func(c []*CmdModel) [][2]string {
			return commandsToColumns(indent, c)
		},
		"FormatTwoColumns": func(rows [][2]string) string {
			buf := bytes.NewBuffer(nil)
			formatTwoColumns(buf, indent, indent, width, rows)
			return buf.String()
		},
		"FormatTwoColumnsWithIndent": func(rows [][2]string, indent, padding int) string {
			buf := bytes.NewBuffer(nil)
			formatTwoColumns(buf, indent, padding, width, rows)
			return buf.String()
		},
		"FormatAppUsage":     formatAppUsage,
		"FormatCommandUsage": formatCmdUsage,
		"IsCumulative": func(v *kong.Value) bool {
			return v != nil && cumulative(v)
		},
		"Char": func(c rune) string {
			return string(c)
		},
	}
}

func formatTwoColumns(w io.Writer, indent, padding, width int, rows [][2]string) {
	// The first column is as wide as its longest entry, up to 30 characters.
	s := 0
	for _, row := range rows {
		if c := len(row[0]); c > s && c < 30 {
			s = c
		}
	}
	indentStr := strings.Repeat(" ", indent)
	offsetStr := strings.Repeat(" ", s+padding)
	for _, row := range rows {
		buf := bytes.NewBuffer(nil)
		doc.ToText(buf, row[1], "", preIndent, width-s-padding-indent)
		lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
		fmt.Fprintf(w, "%s%-*s%*s", indentStr, s, row[0], padding, "")
		if len(row[0]) >= 30 {
			fmt.Fprintf(w, "\n%s%s", indentStr, offsetStr)
		}
		fmt.Fprintf(w, "%s\n", lines[0])
		for _, line := range lines[1:] {
			fmt.Fprintf(w, "%s%s%s\n", indentStr, offsetStr, line)
		}
	}
}

func formatAppUsage(app *ApplicationModel) string {
	s := []string{app.Name}
	if len(app.Flags) > 0 {
		s = append(s, app.FlagSummary())
	}
	if len(app.Args) > 0 {
		s = append(s, app.ArgSummary())
	}
	return strings.Join(s, " ")
}

// formatCmdUsage returns the usage line of a command.
// kingpin v2.3.0 and later summarized the command's flags and arguments
// there, while v3 and older v2 releases summarized the application's.
func formatCmdUsage(app *ApplicationModel, cmd *CmdModel) string {
	flags, args := app.FlagGroupModel, app.ArgGroupModel
	if v2From23() {
		flags, args = cmd.FlagGroupModel, cmd.ArgGroupModel
	}
	s := []string{app.Name, cmd.String()}
	if len(flags.Flags) > 0 {
		s = append(s, flags.FlagSummary())
	}
	if len(args.Args) > 0 {
		s = append(s, args.ArgSummary())
	}
	return strings.Join(s, " ")
}

// formatFlag renders a flag as in "-n, --name=NAME".
// haveShort pads flags without a short form so they line up with the others.
func formatFlag(haveShort bool, flag *ClauseModel) string {
	flagString := ""
	name := flag.Name
	if flag.IsNegatable() {
		name = "[no-]" + name
	}
	if flag.Short != 0 {
		flagString += fmt.Sprintf("-%c, --%s", flag.Short, name)
	} else if haveShort {
		flagString += fmt.Sprintf("    --%s", name)
	} else {
		flagString += fmt.Sprintf("--%s", name)
	}
	if !flag.IsBoolFlag() {
		flagString += fmt.Sprintf("=%s", flag.FormatPlaceHolder())
	}
	if flag.Cumulative {
		flagString += " ..."
	}
	return flagString
}

func commandsToColumns(indent int, cmds []*CmdModel) [][2]string {
	out := [][2]string{}
	for _, cmd := range cmds {
		if cmd.Hidden {
			continue
		}
		left := cmd.Name
		if cmd.FlagSummary() != "" {
			left += " " + cmd.FlagSummary()
		}
		args := []string{}
		for _, arg := range cmd.Args {
			if arg.Required {
				argText := "<" + arg.Name + ">"
				if arg.Cumulative {
					argText += " ..."
				}
				args = append(args, argText)
			}
		}
		if len(args) != 0 {
			left += " " + strings.Join(args, " ")
		}
		out = append(out, [2]string{strings.Repeat(" ", cmd.Depth*indent-1) + left, cmd.Help})
		out = append(out, commandsToColumns(indent, cmd.Commands)...)
	}
	return out
}
