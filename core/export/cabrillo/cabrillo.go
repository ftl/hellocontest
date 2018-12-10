package cabrillo

import (
	"fmt"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

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
	timestamp := qso.Time.Format("2006-01-02 1504")
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
