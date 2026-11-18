package summary

import (
	"html/template"
	"io"
	"strings"

	"github.com/ftl/hellocontest/core/model"
)

func Export(w io.Writer, summary model.Summary) error {
	return summaryTemplate.Execute(w, summary)
}

var summaryTemplate = template.Must(template.New("summary").
	Funcs(template.FuncMap{
		"join":     strings.Join,
		"datetime": model.FormatTimestamp,
		"duration": model.FormatDuration,
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
{{.ScoreTable}}

created with Hello Contest - https://github.com/ftl/hellocontest
`))
