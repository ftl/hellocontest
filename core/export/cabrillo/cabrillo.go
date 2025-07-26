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
	SetCategoryStation(string)
	SetCategoryTransmitter(string)
	SetCategoryOverlay(string)
	SetCategoryTime(string)

	SetName(string)
	SetEmail(string)
	SetLocation(string)
	SetAddressText(string)
	SetAddressCity(string)
	SetAddressPostalCode(string)
	SetAddressStateProvince(string)
	SetAddressCountry(string)
	SetClub(string)
	SetSpecific(string)

	SetCertificate(bool)
	SetSoapBox(string)

	SetOpenUploadAfterExport(bool)
	SetOpenAfterExport(bool)
}

type Controller struct {
	view View

	definition *conval.Definition
	qsoBand    cabrillo.CategoryBand
	qsoMode    cabrillo.CategoryMode

	category             cabrillo.Category
	name                 string
	email                string
	location             string
	addressText          string
	addressCity          string
	addressPostalCode    string
	addressStateProvince string
	addressCountry       string
	club                 string
	specific             string
	certificate          bool
	soapBox              string

	openUploadAfterExport bool
	openAfterExport       bool
}

type Result struct {
	Export                *cabrillo.Log
	OpenUploadAfterExport bool
	OpenAfterExport       bool
}

func NewController() *Controller {
	result := &Controller{
		openUploadAfterExport: false,
		openAfterExport:       false,
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

func (c *Controller) Run(settings core.Settings, claimedScore int, qsos []core.QSO) (Result, bool) {
	c.definition = settings.Contest().Definition
	c.qsoBand, c.qsoMode = findBandAndMode(qsos)
	c.category.Band = c.qsoBand
	c.category.Mode = c.qsoMode

	c.updateCategorySettings()
	c.view.SetName(c.name)
	c.view.SetEmail(c.email)
	c.view.SetLocation(c.location)
	c.view.SetAddressText(c.addressText)
	c.view.SetAddressCity(c.addressCity)
	c.view.SetAddressPostalCode(c.addressPostalCode)
	c.view.SetAddressStateProvince(c.addressStateProvince)
	c.view.SetAddressCountry(c.addressCountry)
	c.view.SetClub(c.club)
	c.view.SetSpecific(c.specific)
	c.view.SetCertificate(c.certificate)
	c.view.SetSoapBox(c.soapBox)
	c.view.SetOpenUploadAfterExport(c.openUploadAfterExport)
	c.view.SetOpenAfterExport(c.openAfterExport)

	accepted := c.view.Show()
	if !accepted {
		return Result{nil, false, false}, false
	}

	export := createCabrilloLog(settings, claimedScore, qsos)
	export.Category = c.category
	export.Name = c.name
	export.Email = c.email
	export.Location = c.location
	export.Address.Text = c.addressText
	export.Address.City = c.addressCity
	export.Address.Postalcode = c.addressPostalCode
	export.Address.StateProvince = c.addressStateProvince
	export.Address.Country = c.addressCountry
	export.Club = c.club
	export.Custom[cabrillo.SpecificTag] = c.specific
	export.Certificate = c.certificate
	export.Soapbox = c.soapBox

	return Result{export, c.openUploadAfterExport, c.openAfterExport}, true
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
	c.category.Transmitter = convalToCabrilloTransmitter(category)
	c.category.Overlay = convertOverlay(category.Overlay)
	c.updateCategorySettings()
}

func (c *Controller) updateCategorySettings() {
	log.Printf("new category settings: %+v", c.category)
	c.view.SetCategoryAssisted(string(c.category.Assisted))
	c.view.SetCategoryBand(string(c.category.Band))
	c.view.SetCategoryMode(string(c.category.Mode))
	c.view.SetCategoryOperator(string(c.category.Operator))
	c.view.SetCategoryPower(string(c.category.Power))
	c.view.SetCategoryStation(string(c.category.Station))
	c.view.SetCategoryTransmitter(string(c.category.Transmitter))
	c.view.SetCategoryOverlay(string(c.category.Overlay))
	c.view.SetCategoryTime(string(c.category.Time))
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

func (c *Controller) CategoryBands() []string {
	if c.definition.Bands == nil {
		return []string{
			string(cabrillo.BandAll),
			string(cabrillo.Band160m),
			string(cabrillo.Band80m),
			string(cabrillo.Band40m),
			string(cabrillo.Band20m),
			string(cabrillo.Band15m),
			string(cabrillo.Band10m),
			string(cabrillo.Band6m),
			string(cabrillo.Band4m),
			string(cabrillo.Band2m),
			string(cabrillo.Band222),
			string(cabrillo.Band432),
			string(cabrillo.Band902),
			string(cabrillo.Band1_2G),
			string(cabrillo.Band2_3G),
			string(cabrillo.Band3_4G),
			string(cabrillo.Band5_6G),
			string(cabrillo.Band10G),
			string(cabrillo.Band24G),
			string(cabrillo.Band47G),
			string(cabrillo.Band75G),
			string(cabrillo.Band122G),
			string(cabrillo.Band134G),
			string(cabrillo.Band241G),
			string(cabrillo.BandLight),
			string(cabrillo.BandVHF_3Band),
			string(cabrillo.BandVHF_FMOnly),
		}
	}
	result := make([]string, len(c.definition.Bands)+1)
	result[0] = string(cabrillo.BandAll)
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
		return []string{
			string(cabrillo.ModeCW),
			string(cabrillo.ModeSSB),
			string(cabrillo.ModeRTTY),
			string(cabrillo.ModeDIGI),
			string(cabrillo.ModeMIXED),
		}
	}
	result := make([]string, len(c.definition.Modes))
	for i, mode := range c.definition.Modes {
		result[i] = string(convertMode(mode))
	}
	if len(result) > 1 {
		result = append(result, string(cabrillo.ModeMIXED))
	}
	return result
}

func (c *Controller) SetCategoryMode(mode string) {
	c.category.Mode = cabrillo.CategoryMode(strings.ToUpper(mode))
}

func (c *Controller) CategoryOperators() []string {
	return []string{
		string(cabrillo.SingleOperator),
		string(cabrillo.MultiOperator),
		string(cabrillo.Checklog),
	}
}

func (c *Controller) SetCategoryOperator(operator string) {
	c.category.Operator = cabrillo.CategoryOperator(strings.ToUpper(operator))
}

func (c *Controller) CategoryPowers() []string {
	return []string{
		string(cabrillo.QRP),
		string(cabrillo.LowPower),
		string(cabrillo.HighPower),
	}
}

func (c *Controller) SetCategoryPower(power string) {
	c.category.Power = cabrillo.CategoryPower(strings.ToUpper(power))
}

func (c *Controller) CategoryAssisted() []string {
	return []string{string(cabrillo.Assisted), string(cabrillo.NonAssisted)}
}

func (c *Controller) SetCategoryAssisted(assisted string) {
	c.category.Assisted = cabrillo.CategoryAssisted(assisted)
}

func (c *Controller) CategoryStations() []string {
	return []string{
		string(cabrillo.DistributedStation),
		string(cabrillo.FixedStation),
		string(cabrillo.MobileStation),
		string(cabrillo.PortableStation),
		string(cabrillo.RoverStation),
		string(cabrillo.RoverLimitedStation),
		string(cabrillo.RoverUnlimitedStation),
		string(cabrillo.ExpeditionStation),
		string(cabrillo.HQStation),
		string(cabrillo.SchoolStation),
		string(cabrillo.ExplorerStation),
	}
}

func (c *Controller) SetCategoryStation(station string) {
	c.category.Station = cabrillo.CategoryStation(station)
}

func (c *Controller) CategoryTransmitters() []string {
	return []string{
		string(cabrillo.OneTransmitter),
		string(cabrillo.TwoTransmitter),
		string(cabrillo.LimitedTransmitter),
		string(cabrillo.UnlimitedTransmitter),
		string(cabrillo.SWL),
	}
}

func (c *Controller) SetCategoryTransmitter(transmitter string) {
	c.category.Transmitter = cabrillo.CategoryTransmitter(transmitter)
}

func (c *Controller) CategoryOverlays() []string {
	if len(c.definition.Overlays) == 0 {
		return []string{
			string(cabrillo.ClassicOverlay),
			string(cabrillo.RookieOverlay),
			string(cabrillo.TBWiresOverlay),
			string(cabrillo.YouthOverlay),
			string(cabrillo.NoviceTechOverlay),
			string(cabrillo.Over50Overlay),
			string(cabrillo.YLOverlay),
		}
	}

	result := make([]string, len(c.definition.Overlays))
	for i, overlay := range c.definition.Overlays {
		result[i] = string(convertOverlay(overlay))
	}
	return result
}

func (c *Controller) SetCategoryOverlay(overlay string) {
	c.category.Overlay = cabrillo.CategoryOverlay(overlay)
}

func (c *Controller) CategoryTimes() []string {
	return []string{
		string(cabrillo.Hours6),
		string(cabrillo.Hours8),
		string(cabrillo.Hours12),
		string(cabrillo.Hours24),
	}
}

func (c *Controller) SetCategoryTime(time string) {
	c.category.Time = cabrillo.CategoryTime(time)
}

func (c *Controller) SetName(name string) {
	c.name = name
}

func (c *Controller) SetEmail(email string) {
	c.email = email
}

func (c *Controller) SetLocation(location string) {
	c.location = location
}

func (c *Controller) SetAddressText(addressText string) {
	c.addressText = addressText
}

func (c *Controller) SetAddressCity(addressCity string) {
	c.addressCity = addressCity
}

func (c *Controller) SetAddressPostalCode(addressPostalCode string) {
	c.addressPostalCode = addressPostalCode
}

func (c *Controller) SetAddressStateProvince(addressStateProvince string) {
	c.addressStateProvince = addressStateProvince
}

func (c *Controller) SetAddressCountry(addressCountry string) {
	c.addressCountry = addressCountry
}

func (c *Controller) SetClub(club string) {
	c.club = club
}

func (c *Controller) SetSpecific(specific string) {
	c.specific = specific
}

func (c *Controller) SetCertificate(certificate bool) {
	c.certificate = certificate
}

func (c *Controller) SetSoapBox(soapBox string) {
	c.soapBox = soapBox
}

func (c *Controller) SetOpenUploadAfterExport(open bool) {
	c.openUploadAfterExport = open
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
