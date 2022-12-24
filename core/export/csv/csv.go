package csv

import (
	"fmt"
	"io"
	"strconv"
	"strings"
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
	`{{.Band}};{{.Frequency}};{{.Mode}};{{.Date}};{{.Time}};{{.MyCall}};{{.MyReport}};{{.MyNumber}};"{{.MyExchange}}";{{.TheirCall}};{{.TheirReport}};{{.TheirNumber}};"{{.TheirExchange}}";"{{.TheirPrefix}}";"{{.TheirContinent}}";"{{.TheirITUZone}}";"{{.TheirCQZone}}";{{.Points}};"{{.Duplicate}}"`))

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
		"MyExchange":     strings.Join(qso.MyExchange, " "),
		"TheirCall":      qso.Callsign.String(),
		"TheirReport":    qso.TheirReport.String(),
		"TheirNumber":    qso.TheirNumber.String(),
		"TheirXchange":   strings.Join(qso.TheirExchange, " "),
		"TheirPrefix":    qso.DXCC.PrimaryPrefix,
		"TheirContinent": qso.DXCC.Continent,
		"TheirITUZone":   fmt.Sprintf("%d", qso.DXCC.ITUZone),
		"TheirCQZone":    fmt.Sprintf("%d", qso.DXCC.CQZone),
		"Points":         strconv.Itoa(qso.Points),
		"Duplicate":      "",
	}
	if qso.Duplicate {
		fillins["Duplicate"] = "X"
		fillins["Points"] = "0"
	}

	err := csvTemplate.Execute(w, fillins)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
}
