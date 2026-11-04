package kongcompat

import "regexp"

// The templates are kingpin's, unchanged, so programs keep the help output
// and completion scripts they had.
// The exported ones are those of the kingpin version the program used.

var (
	// DefaultUsageTemplate is kingpin's default usage template.
	DefaultUsageTemplate = pick(v2LegacyDefaultUsageTemplate, v2DefaultUsageTemplate, v3DefaultUsageTemplate)
	// SeparateOptionalFlagsUsageTemplate lists required and optional flags
	// separately (kingpin v2).
	SeparateOptionalFlagsUsageTemplate = pick(v2LegacySeparateOptionalFlagsUsageTemplate, v2SeparateOptionalFlagsUsageTemplate, v2SeparateOptionalFlagsUsageTemplate)
	// CompactUsageTemplate formats commands compactly, for large command trees.
	CompactUsageTemplate = pick(v2LegacyCompactUsageTemplate, v2CompactUsageTemplate, v3CompactUsageTemplate)
	// BashCompletionTemplate renders the bash completion script.
	BashCompletionTemplate = pick(v2LegacyBashCompletionTemplate, v2BashCompletionTemplate, v3BashCompletionTemplate)
	// ZshCompletionTemplate renders the zsh completion script.
	ZshCompletionTemplate = pick(v2LegacyZshCompletionTemplate, v2ZshCompletionTemplate, v3ZshCompletionTemplate)
	// LongHelpTemplate is kingpin v2's --help-long template.
	LongHelpTemplate = pick(v2LegacyLongHelpTemplate, v2LongHelpTemplate, v2LongHelpTemplate)
	// ManPageTemplate is the --help-man template.
	ManPageTemplate = pick(v2LegacyManPageTemplate, v2ManPageTemplate, v3ManPageTemplate)
)

// pick returns the template of the kingpin release the program came from.
func pick(v2Legacy, v2, v3 string) string {
	switch {
	case version == 2 && legacy:
		return v2Legacy
	case version == 2:
		return v2
	}
	return v3
}

// withContinuations converts the templates of kingpin v2 before v2.3.0 for
// text/template.
// They were written for github.com/alecthomas/template, where a backslash
// right after an action drops the newlines that follow it.
// Other templates are returned as is.
func withContinuations(tmpl string) string {
	if version != 2 || !legacy {
		return tmpl
	}
	return lineJoin.ReplaceAllString(tmpl, "}}")
}

var lineJoin = regexp.MustCompile(`}}\\[\r\n]*`)

var v3DefaultUsageTemplate = `{{define "FormatCommands" -}}
{{range .FlattenedCommands -}}
{{if not .Hidden}}
  {{.CmdSummary}}
{{.Help|Wrap 4}}
{{if .Flags -}}
{{with .Flags|FlagsToTwoColumns}}{{FormatTwoColumnsWithIndent . 4 2}}{{end}}
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
{{T "usage:"}} {{.App.Name}} {{.App.FlagSummary}} {{.Context.SelectedCommand.CmdSummary}}
{{else}}
{{T "usage:"}} {{template "FormatUsage" .App}}
{{end}}
{{if .Context.Flags -}}
{{T "Flags:"}}
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.Args -}}
{{T "Args:"}}
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.SelectedCommand -}}
{{if len .Context.SelectedCommand.Commands -}}
{{T "Subcommands:"}}
{{template "FormatCommands" .Context.SelectedCommand}}
{{end -}}
{{else if .App.Commands -}}
{{T "Commands:" -}}
{{template "FormatCommands" .App}}
{{end -}}
`

var v3CompactUsageTemplate = `{{define "FormatCommand" -}}
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

{{if .Context.SelectedCommand -}}
{{T "usage:"}} {{.App.Name}} {{template "FormatUsage" .Context.SelectedCommand}}
{{else -}}
{{T "usage:"}} {{.App.Name}}{{template "FormatUsage" .App}}
{{end -}}
{{if .Context.Flags -}}
{{T "Flags:"}}
{{.Context.Flags|FlagsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.Args -}}
{{T "Args:"}}
{{.Context.Args|ArgsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{if .Context.SelectedCommand -}}
{{if .Context.SelectedCommand.Commands -}}
{{T "Commands:"}}
  {{.Context.SelectedCommand}}
{{.Context.SelectedCommand.Commands|CommandsToTwoColumns|FormatTwoColumns}}
{{end -}}
{{else if .App.Commands -}}
{{T "Commands:"}}
{{.App.Commands|CommandsToTwoColumns|FormatTwoColumns}}
{{end -}}
`

var v3BashCompletionTemplate = `
_{{.App.Name}}_bash_autocomplete() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[0]} --completion-bash ${COMP_WORDS[@]:1:$COMP_CWORD} )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}
complete -F _{{.App.Name}}_bash_autocomplete {{.App.Name}}

`

var v3ZshCompletionTemplate = `
#compdef {{.App.Name}}
autoload -U compinit && compinit
autoload -U bashcompinit && bashcompinit

_{{.App.Name}}_bash_autocomplete() {
    local cur prev opts base
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    opts=$( ${COMP_WORDS[0]} --completion-bash ${COMP_WORDS[@]:1:$COMP_CWORD} )
    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    [[ $COMPREPLY ]] && return
    compgen -f
    return 0
}
complete -F _{{.App.Name}}_bash_autocomplete {{.App.Name}}
`

var v3ManPageTemplate = `{{define "FormatFlags" -}}
{{range .Flags -}}
{{if not .Hidden -}}
.TP
\fB{{if .Short}}-{{.Short|Char}}, {{end}}--{{.Name}}{{if not .IsBoolFlag}}={{.FormatPlaceHolder}}{{end}}\fR
{{.Help}}
{{end -}}
{{end -}}
{{end -}}

{{define "FormatCommand" -}}
{{end -}}

{{define "FormatCommands" -}}
{{range .FlattenedCommands -}}
{{if not .Hidden -}}
.SS
\fB{{.CmdSummary}}\fR
.PP
{{.Help}}
{{template "FormatFlags" . -}}
{{end -}}
{{end -}}
{{end -}}

{{define "FormatUsage" -}}
{{if .FlagSummary}} {{.FlagSummary}}{{end -}}
{{if .Commands}} <command> [<args> ...]{{end}}\fR
{{end -}}

.TH {{.App.Name}} 1 {{.App.Version}} "{{.App.Author}}"
.SH "NAME"
{{.App.Name}}
.SH "SYNOPSIS"
.TP
\fB{{.App.Name}}{{template "FormatUsage" .App}}
.SH "DESCRIPTION"
{{.App.Help}}
.SH "OPTIONS"
{{template "FormatFlags" .App -}}
{{if .App.Commands -}}
.SH "COMMANDS"
{{template "FormatCommands" .App -}}
{{end -}}
`
