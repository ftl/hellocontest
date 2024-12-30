package cabrillo

import (
	"fmt"
	"io"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/ftl/cabrillo"
	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
)

type View interface {
	Show() bool

	SetCategoryBand(string)
	SetCategoryMode(string)
	SetCategoryOperator(string)
	SetCategoryPower(string)
	SetCategoryAssisted(string)
	SetName(string)
	SetEmail(string)
	SetOpenAfterExport(bool)
}

type Controller struct {
	view View

	definition *conval.Definition
	qsoBand    cabrillo.CategoryBand
	qsoMode    cabrillo.CategoryMode

	category        cabrillo.Category
	name            string
	email           string
	openAfterExport bool
}

func NewController() *Controller {
	result := &Controller{
		openAfterExport: false,
	}

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

func (c *Controller) Run(settings core.Settings, claimedScore int, qsos []core.QSO) (*cabrillo.Log, bool, bool) {
	c.definition = settings.Contest().Definition
	c.qsoBand, c.qsoMode = findBandAndMode(qsos)
	c.category.Band = c.qsoBand
	c.category.Mode = c.qsoMode

	c.updateCategorySettings()
	c.view.SetName(c.name)
	c.view.SetEmail(c.email)
	c.view.SetOpenAfterExport(c.openAfterExport)
	accepted := c.view.Show()
	if !accepted {
		return nil, false, false
	}

	export := createCabrilloLog(settings, claimedScore, qsos)
	export.Category = c.category
	export.Name = c.name
	export.Email = c.email

	return export, c.openAfterExport, true
}

func (c *Controller) Categories() []string {
	if c.definition == nil {
		return nil
	}
	result := make([]string, len(c.definition.Categories))
	for i, category := range c.definition.Categories {
		result[i] = category.Name
	}
	return result
}

func (c *Controller) SetCategory(name string) {
	category, found := c.findCategory(name)
	if !found {
		log.Printf("no category with name %q found", name)
		return
	}

	c.category.Assisted = convalToCabrilloAssisted(category)
	c.category.Band = convalToCabrilloBand(category, c.definition.Bands, c.qsoBand)
	c.category.Mode = convalToCabrilloMode(category, c.definition.Modes, c.qsoMode)
	c.category.Operator = convalToCabrilloOperator(category)
	c.category.Power = convalToCabrilloPower(category)
	c.updateCategorySettings()
}

func (c *Controller) updateCategorySettings() {
	log.Printf("new category settings: %+v", c.category)
	c.view.SetCategoryAssisted(string(c.category.Assisted))
	c.view.SetCategoryBand(string(c.category.Band))
	c.view.SetCategoryMode(string(c.category.Mode))
	c.view.SetCategoryOperator(string(c.category.Operator))
	c.view.SetCategoryPower(string(c.category.Power))
}

func (c *Controller) findCategory(name string) (conval.Category, bool) {
	if c.definition == nil {
		return conval.Category{}, false
	}
	for _, category := range c.definition.Categories {
		if category.Name == name {
			return category, true
		}
	}
	return conval.Category{}, false
}

func (c *Controller) CategoryAssisted() []string {
	return []string{"", string(cabrillo.Assisted), string(cabrillo.NonAssisted)}
}

func (c *Controller) SetCategoryAssisted(assisted string) {
	c.category.Assisted = cabrillo.CategoryAssisted(assisted)
}

func (c *Controller) CategoryBands() []string {
	if c.definition.Bands == nil {
		return []string{
			"ALL", "160M", "80M", "40M", "20M", "15M", "10M", "6M", "4M", "2M", "222", "432",
			"902", "1.2G", "2.3G", "3.4G", "5.7G", "10G", "24G", "47G", "75G", "122G", "134G",
			"241G", "LIGHT", "VHF-3-BAND", "VHF-FM-ONLY",
		}
	}
	result := make([]string, len(c.definition.Bands)+1)
	result[0] = "ALL"
	for i, band := range c.definition.Bands {
		result[i+1] = string(convertBand(band))
	}
	return result
}

func (c *Controller) SetCategoryBand(band string) {
	c.category.Band = cabrillo.CategoryBand(strings.ToUpper(band))
}

func (c *Controller) CategoryModes() []string {
	if c.definition.Modes == nil {
		return []string{"CW", "PH", "RY", "DG", "FM", "MIXED"}
	}
	result := make([]string, len(c.definition.Modes))
	for i, mode := range c.definition.Modes {
		result[i] = string(convertMode(mode))
	}
	if len(result) > 1 {
		result = append(result, "MIXED")
	}
	return result
}

func (c *Controller) SetCategoryMode(mode string) {
	c.category.Mode = cabrillo.CategoryMode(strings.ToUpper(mode))
}

func (c *Controller) CategoryOperators() []string {
	return []string{"SINGLE-OP", "MULTI-OP", "CHECKLOG"}
}

func (c *Controller) SetCategoryOperator(operator string) {
	c.category.Operator = cabrillo.CategoryOperator(strings.ToUpper(operator))
}

func (c *Controller) CategoryPowers() []string {
	return []string{"QRP", "LOW", "HIGH"}
}

func (c *Controller) SetCategoryPower(power string) {
	c.category.Power = cabrillo.CategoryPower(strings.ToUpper(power))
}

func (c *Controller) SetName(name string) {
	c.name = name
}

func (c *Controller) SetEmail(email string) {
	c.email = email
}

func (c *Controller) SetOpenAfterExport(open bool) {
	c.openAfterExport = open
}

func createCabrilloLog(settings core.Settings, claimedScore int, qsos []core.QSO) *cabrillo.Log {
	export := cabrillo.NewLog()
	export.Callsign = settings.Station().Callsign
	export.CreatedBy = "Hello Contest"
	export.Contest = cabrillo.ContestIdentifier(settings.Contest().Definition.Identifier)
	export.Operators = []callsign.Callsign{settings.Station().Operator}
	export.GridLocator = settings.Station().Locator
	export.ClaimedScore = claimedScore
	export.Certificate = true

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

	return export
}

// Export writes the given QSOs to the given writer in the Cabrillo format.
// The header is very limited and needs to be completed manually after the log was written.
func Export(w io.Writer, export *cabrillo.Log) error {
	return cabrillo.Write(w, export, false)
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

func findBandAndMode(qsos []core.QSO) (band cabrillo.CategoryBand, mode cabrillo.CategoryMode) {
	band = ""
	mode = ""
	for _, qso := range qsos {
		qsoBand := cabrillo.CategoryBand(strings.ToUpper(string(qso.Band)))
		if band == "" {
			band = qsoBand
		} else if band != cabrillo.BandAll && band != qsoBand {
			band = cabrillo.BandAll
		}

		qsoMode := cabrillo.CategoryMode(strings.ToUpper(string(qso.Mode)))
		if mode == "" {
			mode = qsoMode
		} else if mode != cabrillo.ModeMIXED && mode != qsoMode {
			mode = cabrillo.ModeMIXED
		}
	}
	return band, mode
}
