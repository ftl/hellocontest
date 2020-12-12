package csv

import (
	"fmt"
	"io"
	"strconv"
	"text/template"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"
	"github.com/ftl/hellocontest/core"
)

// DXCCFinder returns a list of matching prefixes for the given string and indicates if there was a match at all.
type DXCCFinder interface {
	Find(string) (dxcc.Prefix, bool)
}

// Export writes the given QSOs to the given writer in the CSV format.
// The header is very limited and needs to be completed manually after the log was written.
func Export(w io.Writer, mycall callsign.Callsign, qsos ...core.QSO) error {
	for _, qso := range qsos {
		if err := writeQSO(w, mycall, qso); err != nil {
			return err
		}
	}
	return nil
}

var csvTemplate = template.Must(template.New("").Parse(
	`{{.Band}};{{.Frequency}};{{.Mode}};{{.Date}};{{.Time}};{{.MyCall}};{{.MyReport}};{{.MyNumber}};"{{.MyXchange}}";{{.TheirCall}};{{.TheirReport}};{{.TheirNumber}};"{{.TheirXchange}}";"{{.TheirPrefix}}";"{{.TheirContinent}}";"{{.TheirITUZone}}";"{{.TheirCQZone}}";{{.Points}};{{.Multis}}`))

func writeQSO(w io.Writer, mycall callsign.Callsign, qso core.QSO) error {
	fillins := map[string]string{
		"Band":           qso.Band.String(),
		"Frequency":      fmt.Sprintf("%5.3f", float64(qso.Frequency/1000000.0)),
		"Mode":           qso.Mode.String(),
		"Date":           qso.Time.In(time.UTC).Format("2006-01-02"),
		"Time":           qso.Time.In(time.UTC).Format("1504"),
		"MyCall":         mycall.String(),
		"MyReport":       qso.MyReport.String(),
		"MyNumber":       qso.MyNumber.String(),
		"MyXchange":      qso.MyXchange,
		"TheirCall":      qso.Callsign.String(),
		"TheirReport":    qso.TheirReport.String(),
		"TheirNumber":    qso.TheirNumber.String(),
		"TheirXchange":   qso.TheirXchange,
		"TheirPrefix":    qso.DXCC.PrimaryPrefix,
		"TheirContinent": qso.DXCC.Continent,
		"TheirITUZone":   fmt.Sprintf("%d", qso.DXCC.ITUZone),
		"TheirCQZone":    fmt.Sprintf("%d", qso.DXCC.CQZone),
		"Points":         strconv.Itoa(qso.Points),
		"Multis":         strconv.Itoa(qso.Multis),
	}

	err := csvTemplate.Execute(w, fillins)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
}
