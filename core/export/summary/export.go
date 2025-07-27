package summary

import (
	"html/template"
	"io"
	"strings"

	"github.com/ftl/hellocontest/core"
)

func Export(w io.Writer, summary core.Summary) error {
	return summaryTemplate.Execute(w, summary)
}

var summaryTemplate = template.Must(template.New("summary").
	Funcs(template.FuncMap{
		"join":     strings.Join,
		"datetime": core.FormatTimestamp,
		"duration": core.FormatDuration,
	}).
	Parse(`Contest:    {{.ContestName}}
Cabrillo:   {{.CabrilloName}}
Start Time: {{datetime .StartTime}}
Callsign:   {{.Callsign}}
{{- if .MyExchanges}}
Exchange:   {{.MyExchanges}}
{{- end}}

{{if .WorkingConditions -}}
Working Conditions: {{.WorkingConditions}}
{{end -}}
Worked Modes:       {{join .WorkedModes ", "}}
Worked Bands:       {{join .WorkedBands ", "}}
Operating Time:     {{duration .TimeReport.OperationTime}}

Claimed Score:
{{.Score}}

created with Hello Contest - https://github.com/ftl/hellocontest
`))
