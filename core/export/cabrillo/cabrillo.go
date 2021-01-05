package cabrillo

import (
	"fmt"
	"io"
	"text/template"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

// Export writes the given QSOs to the given writer in the Cabrillo format.
// The header is very limited and needs to be completed manually after the log was written.
func Export(w io.Writer, t *template.Template, settings core.Settings, claimedScore int, qsos ...core.QSO) error {
	stationCallsign := settings.Station().Callsign
	head := []string{
		"START-OF-LOG: 3.0",
		"CREATED-BY: Hello Contest",
		fmt.Sprintf("CONTEST: %s", settings.Contest().Name),
		fmt.Sprintf("CALLSIGN: %s", settings.Station().Callsign),
		fmt.Sprintf("OPERATORS: %s", settings.Station().Operator),
		fmt.Sprintf("GRID-LOCATOR: %s", settings.Station().Locator),
		fmt.Sprintf("CLAIMED-SCORE: %d", claimedScore),
		"SPECIFIC:",
		"CATEGORY-ASSISTED: (ASSISTED|NON-ASSISTED)",
		"CATEGORY-BAND: ALL",
		"CATEGORY-MODE: (CW|DIGI|FM|RTTY|SSB|MIXED)",
		"CATEGORY-OPERATOR: (SINGLE-OP|MULTI-OP|CHECKLOG)",
		"CATEGORY-POWER: (HIGH|LOW|QRP)",
		"CLUB:",
		"NAME:",
		"EMAIL:",
	}
	tail := []string{
		"END-OF-LOG:",
	}

	for _, line := range head {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	for _, qso := range qsos {
		if err := writeQSO(w, t, stationCallsign, qso); err != nil {
			return err
		}
	}

	for _, line := range tail {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	return nil
}

var qrg = map[core.Band]string{
	core.NoBand:   "",
	core.Band160m: "1800",
	core.Band80m:  "3500",
	core.Band60m:  "5351",
	core.Band40m:  "7000",
	core.Band30m:  "10100",
	core.Band20m:  "14000",
	core.Band17m:  "18100",
	core.Band15m:  "21000",
	core.Band12m:  "24890",
	core.Band10m:  "28000",
}

var mode = map[core.Mode]string{
	core.NoMode:      "",
	core.ModeCW:      "CW",
	core.ModeSSB:     "PH",
	core.ModeFM:      "FM",
	core.ModeRTTY:    "RY",
	core.ModeDigital: "DG",
}

func writeQSO(w io.Writer, t *template.Template, mycall callsign.Callsign, qso core.QSO) error {
	var frequency string
	if qso.Frequency == 0 {
		frequency = qrg[qso.Band]
	} else {
		frequency = fmt.Sprintf("%5.0f", qso.Frequency/1000.0)
	}
	fillins := map[string]string{
		"QRG":          frequency,
		"Mode":         mode[qso.Mode],
		"Date":         qso.Time.In(time.UTC).Format("2006-01-02"),
		"Time":         qso.Time.In(time.UTC).Format("1504"),
		"MyCall":       mycall.String(),
		"MyReport":     qso.MyReport.String(),
		"MyNumber":     qso.MyNumber.String(),
		"MyXchange":    qso.MyXchange,
		"TheirCall":    qso.Callsign.String(),
		"TheirReport":  qso.TheirReport.String(),
		"TheirNumber":  qso.TheirNumber.String(),
		"TheirXchange": qso.TheirXchange,
	}

	_, err := fmt.Fprintf(w, "QSO: ")
	if err != nil {
		return err
	}
	err = t.Execute(w, fillins)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w)
	return err
}
