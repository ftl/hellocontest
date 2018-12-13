package cabrillo

import (
	"fmt"
	"io"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

// Export writes the given QSOs to the given writer in the Cabrillo format.
// The header is very limited and needs to be completed manually after the log was written.
func Export(w io.Writer, mycall callsign.Callsign, myExchange core.Exchanger, theirExchange core.Exchanger, qsos ...core.QSO) error {
	head := []string{
		"START-OF-LOG: 3.0",
		"CREATED-BY: Hello Contest",
		fmt.Sprintf("CALLSIGN: %s", mycall),
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
		if _, err := fmt.Fprintln(w, qsoLine(mycall, myExchange, theirExchange, qso)); err != nil {
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

func qsoLine(mycall callsign.Callsign, myExchange core.Exchanger, theirExchange core.Exchanger, qso core.QSO) string {
	timestamp := qso.Time.In(time.UTC).Format("2006-01-02 1504")
	return fmt.Sprintf("QSO: %s %s %s %s %s %s %s %s %s",
		qrg[qso.Band],
		mode[qso.Mode],
		timestamp,
		mycall,
		qso.MyReport,
		myExchange(qso),
		qso.Callsign,
		qso.TheirReport,
		theirExchange(qso),
	)
}
