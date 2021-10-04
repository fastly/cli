package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

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
