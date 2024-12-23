package cabrillo

import (
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/ftl/cabrillo"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

type View interface {
	Show() bool
}

type Controller struct {
	view View
}

func NewController() *Controller {
	result := &Controller{}

	return result
}

func (c *Controller) SetView(view View) {
	if view == nil {
		panic("cabrillo.Controller.SetView must not be called with nil")
	}
	if c.view != nil {
		panic("cabrillo.Controller.SetView was already called")
	}

	c.view = view
}

func (c *Controller) Run() bool {
	accepted := c.view.Show()
	if !accepted {
		return false
	}

	return accepted
}

// Export writes the given QSOs to the given writer in the Cabrillo format.
// The header is very limited and needs to be completed manually after the log was written.
func Export(w io.Writer, settings core.Settings, claimedScore int, qsos ...core.QSO) error {
	export := cabrillo.NewLog()
	export.Callsign = settings.Station().Callsign
	export.CreatedBy = "Hello Contest"
	export.Contest = cabrillo.ContestIdentifier(settings.Contest().Name)
	export.Operators = []callsign.Callsign{settings.Station().Operator}
	export.GridLocator = settings.Station().Locator
	export.ClaimedScore = claimedScore

	qsoData := make([]cabrillo.QSO, 0, len(qsos))
	ignoredQSOs := make([]cabrillo.QSO, 0, len(qsos))
	for _, qso := range qsos {
		exportedQSO := toQSO(qso, settings.Station().Callsign)
		if qso.Duplicate {
			ignoredQSOs = append(ignoredQSOs, exportedQSO)
		} else {
			qsoData = append(qsoData, exportedQSO)
		}
	}
	export.QSOData = qsoData
	export.IgnoredQSOs = ignoredQSOs

	return cabrillo.WriteWithTags(w, export, false, false, cabrillo.CreatedByTag, cabrillo.ContestTag,
		cabrillo.CallsignTag, cabrillo.OperatorsTag, cabrillo.GridLocatorTag, cabrillo.ClaimedScoreTag,
		cabrillo.Tag("SPECIFIC"), cabrillo.CategoryAssistedTag, cabrillo.CategoryBandTag, cabrillo.CategoryModeTag,
		cabrillo.CategoryOperatorTag, cabrillo.CategoryPowerTag, cabrillo.ClubTag, cabrillo.NameTag,
		cabrillo.EmailTag)
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

var mode = map[core.Mode]cabrillo.QSOMode{
	core.NoMode:      "",
	core.ModeCW:      cabrillo.QSOModeCW,
	core.ModeSSB:     cabrillo.QSOModePhone,
	core.ModeFM:      cabrillo.QSOModeFM,
	core.ModeRTTY:    cabrillo.QSOModeRTTY,
	core.ModeDigital: cabrillo.QSOModeDigi,
}

func writeQSO(w io.Writer, t *template.Template, mycall callsign.Callsign, qso core.QSO) error {
	var frequency string
	if qso.Frequency == 0 {
		frequency = qrg[qso.Band]
	} else {
		frequency = fmt.Sprintf("%5.0f", qso.Frequency/1000.0)
	}
	fillins := map[string]string{
		"QRG":           frequency,
		"Mode":          string(mode[qso.Mode]),
		"Date":          qso.Time.In(time.UTC).Format("2006-01-02"),
		"Time":          qso.Time.In(time.UTC).Format("1504"),
		"MyCall":        mycall.String(),
		"MyReport":      qso.MyReport.String(),
		"MyNumber":      qso.MyNumber.String(),
		"MyExchange":    strings.Join(qso.MyExchange, " "),
		"TheirCall":     qso.Callsign.String(),
		"TheirReport":   qso.TheirReport.String(),
		"TheirNumber":   qso.TheirNumber.String(),
		"TheirExchange": strings.Join(qso.TheirExchange, " "),
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

func toQSO(qso core.QSO, mycall callsign.Callsign) cabrillo.QSO {
	var frequency string
	if qso.Frequency == 0 {
		frequency = qrg[qso.Band]
	} else {
		frequency = fmt.Sprintf("%5.0f", qso.Frequency/1000.0)
	}

	return cabrillo.QSO{
		Frequency: cabrillo.QSOFrequency(frequency),
		Mode:      mode[qso.Mode],
		Timestamp: qso.Time,
		Sent: cabrillo.QSOInfo{
			Call:     mycall,
			Exchange: qso.MyExchange,
		},
		Received: cabrillo.QSOInfo{
			Call:     qso.Callsign,
			Exchange: qso.TheirExchange,
		},
		Transmitter: 0,
	}
}
