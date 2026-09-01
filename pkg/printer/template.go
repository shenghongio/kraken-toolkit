package printer

var HelpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}
{{end}}
{{- if .HasAvailableLocalFlags}}
Flags:
	{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}
{{- if .HasAvailableInheritedFlags}}
Global Flags:
	{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}
{{- if .Runnable}}
Command: {{.UseLine}}
{{end}}
{{- if .HasAvailableSubCommands}}
{{.CommandsHelpString}}
{{end}}`
