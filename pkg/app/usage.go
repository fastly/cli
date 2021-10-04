package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/kingpin"
)

type usageJSON struct {
	GlobalFlags []flagJSON    `json:"globalFlags"`
	Commands    []commandJSON `json:"commands"`
}

type flagJSON struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
	Required    bool   `json:"required"`
	Default     string `json:"default"`
	IsBool      bool   `json:"isBool"`
}

type commandJSON struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Flags       []flagJSON    `json:"flags"`
	Children    []commandJSON `json:"children"`
}

func getFlagJSON(models []*kingpin.ClauseModel) []flagJSON {
	var flags []flagJSON
	for _, f := range models {
		var flag flagJSON
		flag.Name = f.Name
		flag.Description = f.Help
		flag.Placeholder = f.PlaceHolder
		flag.Required = f.Required
		flag.Default = strings.Join(f.Default, ",")
		flag.IsBool = f.IsBoolFlag()
		flags = append(flags, flag)
	}
	return flags
}

func getGlobalFlagJSON(models []*kingpin.ClauseModel) []flagJSON {
	var globalFlags []*kingpin.ClauseModel
	for _, f := range models {
		if !f.Hidden {
			globalFlags = append(globalFlags, f)
		}
	}
	return getFlagJSON(globalFlags)
}

func getCommandJSON(models []*kingpin.CmdModel) []commandJSON {
	var commands []commandJSON
	for _, c := range models {
		var cmd commandJSON
		cmd.Name = c.Name
		cmd.Description = c.Help
		cmd.Flags = getFlagJSON(c.Flags)
		cmd.Children = getCommandJSON(c.Commands)
		commands = append(commands, cmd)
	}
	return commands
}

// UsageJSON returns a structured representation of the application usage
// documentation in JSON format. This is useful for machine consumtion.
func UsageJSON(app *kingpin.Application) (string, error) {
	usage := &usageJSON{
		GlobalFlags: getGlobalFlagJSON(app.Model().Flags),
		Commands:    getCommandJSON(app.Model().Commands),
	}

	j, err := json.Marshal(usage)
	if err != nil {
		return "", err
	}

	return string(j), nil
}

// Usage returns a contextual usage string for the application. In order to deal
// with Kingpin's annoying love of side effects, we have to swap the app.Writers
// to capture output; the out and err parameters, therefore, are the io.Writers
// re-assigned to the app via app.Writers after calling Usage.
func Usage(args []string, app *kingpin.Application, out, err io.Writer, vars map[string]interface{}) string {
	var buf bytes.Buffer
	app.Writers(&buf, io.Discard)
	app.UsageContext(&kingpin.UsageContext{
		Template: CompactUsageTemplate,
		Funcs:    UsageTemplateFuncs,
		Vars:     vars,
	})
	app.Usage(args)
	app.Writers(out, err)
	return buf.String()
}

// WARNING: kingpin has no way of decorating flags as being "global" therefore
// if you add/remove a global flag you will also need to update flag binding in
// pkg/app/app.go.
var globalFlags = map[string]bool{
	"help":    true,
	"token":   true,
	"verbose": true,
}

// UsageTemplateFuncs is a map of template functions which get passed to the
// usage template renderer.
var UsageTemplateFuncs = template.FuncMap{
	"CommandsToTwoColumns": func(c []*kingpin.CmdModel) [][2]string {
		rows := [][2]string{}
		for _, cmd := range c {
			if !cmd.Hidden {
				rows = append(rows, [2]string{cmd.Name, cmd.Help})
			}
		}
		return rows
	},
	"GlobalFlags": func(f []*kingpin.ClauseModel) []*kingpin.ClauseModel {
		flags := []*kingpin.ClauseModel{}
		for _, flag := range f {
			if globalFlags[flag.Name] {
				flags = append(flags, flag)
			}
		}
		return flags
	},
	"OptionalFlags": func(f []*kingpin.ClauseModel) []*kingpin.ClauseModel {
		optionalFlags := []*kingpin.ClauseModel{}
		for _, flag := range f {
			if !flag.Required && !flag.Hidden && !globalFlags[flag.Name] {
				optionalFlags = append(optionalFlags, flag)
			}
		}
		return optionalFlags
	},
	"Bold": func(s string) string {
		return text.Bold(s)
	},
}

// CompactUsageTemplate is the default usage template, rendered when users type
// e.g. just `fastly`, or use the `-h, --help` flag.
var CompactUsageTemplate = `{{define "FormatCommand" -}}
{{if .FlagSummary}} {{.FlagSummary}}{{end -}}
{{range .Args}} {{if not .Required}}[{{end}}<{{.Name}}>{{if .Value|IsCumulative}} ...{{end}}{{if not .Required}}]{{end}}{{end -}}
{{end -}}
{{define "FormatCommandList" -}}
{{range . -}}
{{if not .Hidden -}}
{{.Depth|Indent}}{{.Name}}{{if .Default}}*{{end}}{{template "FormatCommand" .}}
{{end -}}
{{template "FormatCommandList" .Commands -}}
{{end -}}
{{end -}}
{{define "FormatUsage" -}}
{{template "FormatCommand" .}}{{if .Commands}} <command> [<args> ...]{{end}}
{{if .Help}}
{{.Help|Wrap 0 -}}
{{end -}}
{{end -}}
{{define "FormatCommandName" -}}
{{if .Parent}}{{if .Parent.Parent}}{{.Parent.Parent.Name}} {{end -}}{{.Parent.Name}} {{end -}}{{.Name -}}
{{end -}}
{{if .Context.SelectedCommand -}}
{{T "USAGE"|Bold}}
  {{.App.Name}} {{template "FormatCommandName" .Context.SelectedCommand}}{{ template "FormatUsage" .Context.SelectedCommand}}
{{else -}}
{{T "USAGE"|Bold}}
  {{.App.Name}}{{template "FormatUsage" .App}}
{{end -}}
{{if .Context.Flags|RequiredFlags -}}
{{T "REQUIRED FLAGS"|Bold}}
{{.Context.Flags|RequiredFlags|FlagsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.Flags|OptionalFlags -}}
{{T "OPTIONAL FLAGS"|Bold}}
{{.Context.Flags|OptionalFlags|FlagsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.Flags|GlobalFlags -}}
{{T "GLOBAL FLAGS"|Bold}}
{{.Context.Flags|GlobalFlags|FlagsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.Args -}}
{{T "ARGS"|Bold}}
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.SelectedCommand -}}
{{if .Context.SelectedCommand.Commands -}}
{{T "COMMANDS"|Bold}}
  {{.Context.SelectedCommand}}
{{.Context.SelectedCommand.Commands|CommandsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{else if .App.Commands -}}
{{T "COMMANDS"|Bold}}
{{.App.Commands|CommandsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Notes}}
{{T "NOTES"|Bold}}
{{.Notes}}
{{end -}}
`

// VerboseUsageTemplate is the full-fat usage template, rendered when users type
// the long-form e.g. `fastly help service`.
const VerboseUsageTemplate = `{{define "FormatCommands" -}}
{{range .FlattenedCommands -}}
{{ if not .Hidden }}
  {{.CmdSummary|Bold }}
{{.Help|Wrap 4 }}
{{if .Flags -}}
{{with .Flags|FlagsToTwoColumns}}{{FormatTwoColumnsWithIndent . 4 2}}{{end -}}
{{end -}}
{{end -}}
{{end -}}
{{end -}}
{{define "FormatUsage" -}}
{{.AppSummary}}
{{if .Help}}
{{.Help|Wrap 0 -}}
{{end -}}
{{end -}}
{{if .Context.SelectedCommand -}}
{{T "USAGE"|Bold}}
  {{.App.Name}} {{.App.FlagSummary}} {{.Context.SelectedCommand.CmdSummary}}
{{else}}
{{- T "USAGE"|Bold}}
  {{template "FormatUsage" .App -}}
{{end -}}
{{if .Context.Flags|GlobalFlags }}
{{T "GLOBAL FLAGS"|Bold}}
{{.Context.Flags|GlobalFlags|FlagsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.Args -}}
{{T "ARGS"|Bold}}
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.SelectedCommand -}}
{{if len .Context.SelectedCommand.Commands -}}
{{T "SUBCOMMANDS\n"|Bold -}}
  {{ template "FormatCommands" .Context.SelectedCommand}}
{{end -}}
{{else if .App.Commands -}}
{{T "COMMANDS"|Bold -}}
  {{template "FormatCommands" .App}}
{{end -}}
`

// displayHelp returns a function that prints the help output for a command or
// command set.
//
// NOTE: This function is called multiple times within app.Run() and so we use
// a closure to prevent having to pass the same unchanging arguments each time.
func displayHelp(
	errLog errors.LogInterface,
	args []string,
	app *kingpin.Application,
	stdout, stderr io.Writer) func(vars map[string]interface{}, err error) error {

	return func(vars map[string]interface{}, err error) error {
		usage := Usage(args, app, stdout, stderr, vars)
		remediation := errors.RemediationError{Prefix: usage}
		if err != nil {
			errLog.Add(err)
			remediation.Inner = fmt.Errorf("error parsing arguments: %w", err)
		}
		return remediation
	}
}

// processCommandInput groups together all the logic related to parsing and
// processing the incoming command request from the user, as well as handling
// the various places where help output can be displayed.
func processCommandInput(
	opts RunOpts,
	app *kingpin.Application,
	globals *config.Data,
	commands []cmd.Command) (command cmd.Command, cmdName string, err error) {
	// As the `help` command model gets privately added as a side-effect of
	// kingping.Parse, we cannot add the `--format json` flag to the model.
	// Therefore, we have to manually parse the args slice here to check for the
	// existence of `help --format json`, if present we print usage JSON and
	// exit early.
	if cmd.ArgsIsHelpJSON(opts.Args) {
		json, err := UsageJSON(app)
		if err != nil {
			globals.ErrLog.Add(err)
			return command, cmdName, err
		}
		fmt.Fprintf(opts.Stdout, "%s", json)
		return command, cmdName, nil
	}

	// Use partial application to generate help output function.
	help := displayHelp(globals.ErrLog, opts.Args, app, opts.Stdout, io.Discard)

	// Handle parse errors and display contextual usage if possible. Due to bugs
	// and an obsession for lots of output side-effects in the kingpin.Parse
	// logic, we suppress it from writing any usage or errors to the writer by
	// swapping the writer with a no-op and then restoring the real writer
	// afterwards. This ensures usage text is only written once to the writer
	// and gives us greater control over our error formatting.
	app.Writers(io.Discard, io.Discard)

	// The `vars` variable is passed into our CLI's Usage() function and exposes
	// variables to the template used to generate help output.
	//
	// NOTE: The zero value of a map is nil.
	// A nil map has no keys, nor can keys be added until initialised.
	// We only initialise the map if a command has .Notes() implemented.
	var vars map[string]interface{}

	// NOTE: We call two similar methods below: ParseContext() and Parse().
	//
	// We call Parse() because we want the high-level side effect of processing
	// the command information, but we call ParseContext() because we require a
	// context object separately to identify if the --help flag was passed (this
	// isn't possible to do with the Parse() method).
	//
	// Internally Parse() calls ParseContext(), to help it handle specific
	// behaviours such as configuring pre and post conditional behaviours, as well
	// as other related settings.
	//
	// Ultimately this means that Parse() might fail because ParseContext()
	// failed, which happens if the given command or one of its sub commands are
	// unrecognised or if an unrecognised flag is provided, while Parse() can also
	// fail if a 'required' flag is missing.
	noargs := len(opts.Args) == 0
	ctx, err := app.ParseContext(opts.Args)
	if err != nil || noargs {
		if noargs {
			err = fmt.Errorf("command not specified")
		}
		return command, cmdName, help(vars, err)
	}

	// NOTE: The `fastly help` and `fastly --help` behaviours need to avoid
	// trying to call cmd.Select() as the context object will not return a useful
	// value for FullCommand(). The former will fail to find a match as it will
	// be set to `help [<command>...]` as it's a built-in command that we don't
	// control, and the latter --help flag variation will be an empty string as
	// there were no actual 'command' specified.
	var found bool
	if !cmd.IsHelpOnly(opts.Args) && !cmd.IsHelpFlagOnly(opts.Args) {
		command, found = cmd.Select(ctx.SelectedCommand.FullCommand(), commands)
		if !found {
			return command, cmdName, help(vars, err)
		}
	}

	// NOTE: Neither `fastly help` nor `fastly --help` have a .Notes() method.
	if cmd.ContextHasHelpFlag(ctx) && !cmd.IsHelpFlagOnly(opts.Args) {
		notes := command.Notes()
		if notes != "" {
			vars = make(map[string]interface{})
			vars["Notes"] = command.Notes()
		}
		return command, cmdName, help(vars, nil)
	}

	cmdName, err = app.Parse(opts.Args)
	if err != nil {
		return command, cmdName, help(vars, err)
	}

	// Restore output writers
	app.Writers(opts.Stdout, io.Discard)

	// A side-effect of suppressing app.Parse from writing output is the usage
	// isn't printed for the default `help` command. Therefore we capture it
	// here by calling Parse, again swapping the Writers. This also ensures the
	// larger and more verbose help formatting is used.
	if cmdName == "help" {
		var buf bytes.Buffer
		app.Writers(&buf, io.Discard)
		app.Parse(opts.Args)
		app.Writers(opts.Stdout, io.Discard)

		// The full-fat output of `fastly help` should have a hint at the bottom
		// for more specific help. Unfortunately I don't know of a better way to
		// distinguish `fastly help` from e.g. `fastly help configure` than this
		// check.
		if len(opts.Args) > 0 && opts.Args[len(opts.Args)-1] == "help" {
			fmt.Fprintln(&buf, "For help on a specific command, try e.g.")
			fmt.Fprintln(&buf, "")
			fmt.Fprintln(&buf, "\tfastly help configure")
			fmt.Fprintln(&buf, "\tfastly configure --help")
			fmt.Fprintln(&buf, "")
		}

		return command, cmdName, errors.RemediationError{Prefix: buf.String()}
	}

	return command, cmdName, nil
}
