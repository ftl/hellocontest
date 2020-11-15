package cfg

import (
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/cfg"
	"github.com/ftl/hamradio/locator"
)

const Filename = "hellocontest.json"

var Default = Data{
	MyCall:              "DL0ABC",
	MyLocator:           "AA00zz",
	EnterTheirNumber:    true,
	EnterTheirXchange:   true,
	AllowMultiBand:      true,
	AllowMultiMode:      true,
	CabrilloQSOTemplate: "{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}} {{.TheirCall}} {{.TheirReport}} {{.TheirNumber}} {{.TheirXchange}}",
	HamlibAddress:       "localhost:4532",
	KeyerHost:           "localhost",
	KeyerPort:           6789,
	KeyerWPM:            20,
	KeyerSPMacros: []string{
		"{{.MyCall}}",
		"rr {{.MyReport}} {{.MyNumber}} {{.MyXchange}}",
		"tu gl",
		"nr {{.MyNumber}} {{.MyXchange}} {{.MyNumber}} {{.MyXchange}}",
	},
	KeyerRunMacros: []string{
		"cq {{.MyCall}} test",
		"{{.TheirCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}}",
		"tu {{.MyCall}} test",
		"nr {{.MyNumber}} {{.MyXchange}} {{.MyNumber}} {{.MyXchange}}",
	},
	Score: Score{
		CountPerBand:            true,
		SameCountryPoints:       1,
		SameContinentPoints:     3,
		OtherPoints:             5,
		SpecificCountryPoints:   0,
		SpecificCountryPrefixes: []string{},
		Multis:                  []string{"CQ", "DXCC"},
	},
}

// Load loads the configuration from the default location (see github.com/ftl/cfg/LoadJSON()).
func Load() (*LoadedConfiguration, error) {
	if !cfg.Exists("", Filename) {
		cfg.PrepareDirectory("")
		err := cfg.SaveJSON("", Filename, Default)
		if err != nil {
			return nil, err
		}
	}

	var data Data
	err := cfg.LoadJSON("", Filename, &data)
	if err != nil {
		return nil, err
	}
	return &LoadedConfiguration{
		data: data,
	}, nil
}

// Static creates a static configuration instance with the given data.
func Static(myCall callsign.Callsign, myLocator locator.Locator) *StaticConfiguration {
	return &StaticConfiguration{
		myCall:    myCall,
		myLocator: myLocator,
	}
}

// Directory returns the configuration directory. It panics if the directory could not be determined.
func Directory() string {
	dir, err := cfg.Directory("")
	if err != nil {
		panic(errors.Wrap(err, "cannot determine configuration directory"))
	}
	return dir
}

func AbsoluteFilename() string {
	return filepath.Join(Directory(), Filename)
}

type Data struct {
	MyCall              string   `json:"mycall"`
	MyLocator           string   `json:"locator"`
	EnterTheirNumber    bool     `json:"enter_their_number"`
	EnterTheirXchange   bool     `json:"enter_their_xchange"`
	AllowMultiBand      bool     `json:"allow_multi_band"`
	AllowMultiMode      bool     `json:"allow_multi_mode"`
	CabrilloQSOTemplate string   `json:"cabrillo_qso"`
	KeyerHost           string   `json:"keyer_host"`
	KeyerPort           int      `json:"keyer_port"`
	KeyerWPM            int      `json:"keyer_wpm"`
	KeyerSPMacros       []string `json:"keyer_sp_macros"`
	KeyerRunMacros      []string `json:"keyer_run_macros"`
	HamlibAddress       string   `json:"hamlib_address"`
	Score               Score    `json:"score"`
}

type Score struct {
	CountPerBand bool `json:"count_per_band"`

	SameCountryPoints   int `json:"same_country_points"`
	SameContinentPoints int `json:"same_continent_points"`
	OtherPoints         int `json:"other_points"`

	SpecificCountryPoints   int      `json:"specific_country_points"`
	SpecificCountryPrefixes []string `json:"specific_country_prefixes"`

	Multis []string `json:"multis"`
}

type LoadedConfiguration struct {
	data Data
}

func (c *LoadedConfiguration) MyCall() callsign.Callsign {
	myCall, _ := callsign.Parse(c.data.MyCall)
	return myCall
}

func (c *LoadedConfiguration) MyLocator() locator.Locator {
	myLocator, _ := locator.Parse(c.data.MyLocator)
	return myLocator
}

func (c *LoadedConfiguration) EnterTheirNumber() bool {
	return c.data.EnterTheirNumber
}

func (c *LoadedConfiguration) EnterTheirXchange() bool {
	return c.data.EnterTheirXchange
}

func (c *LoadedConfiguration) CabrilloQSOTemplate() string {
	return c.data.CabrilloQSOTemplate
}

func (c *LoadedConfiguration) AllowMultiBand() bool {
	return c.data.AllowMultiBand
}

func (c *LoadedConfiguration) AllowMultiMode() bool {
	return c.data.AllowMultiMode
}

func (c *LoadedConfiguration) KeyerHost() string {
	return c.data.KeyerHost
}

func (c *LoadedConfiguration) KeyerPort() int {
	return c.data.KeyerPort
}

func (c *LoadedConfiguration) KeyerWPM() int {
	return c.data.KeyerWPM
}

func (c *LoadedConfiguration) KeyerSPMacros() []string {
	if c.data.KeyerSPMacros == nil {
		return []string{}
	}
	return c.data.KeyerSPMacros
}

func (c *LoadedConfiguration) KeyerRunMacros() []string {
	if c.data.KeyerRunMacros == nil {
		return []string{}
	}
	return c.data.KeyerRunMacros
}

func (c *LoadedConfiguration) HamlibAddress() string {
	return c.data.HamlibAddress
}

func (c *LoadedConfiguration) CountPerBand() bool {
	return c.data.Score.CountPerBand
}

func (c *LoadedConfiguration) SameCountryPoints() int {
	return c.data.Score.SameCountryPoints
}

func (c *LoadedConfiguration) SameContinentPoints() int {
	return c.data.Score.SameContinentPoints
}

func (c *LoadedConfiguration) OtherPoints() int {
	return c.data.Score.OtherPoints
}

func (c *LoadedConfiguration) SpecificCountryPoints() int {
	return c.data.Score.SpecificCountryPoints
}

func (c *LoadedConfiguration) SpecificCountryPrefixes() []string {
	return c.data.Score.SpecificCountryPrefixes
}

func (c *LoadedConfiguration) Multis() []string {
	return c.data.Score.Multis
}

type StaticConfiguration struct {
	myCall    callsign.Callsign
	myLocator locator.Locator
}

func (c *StaticConfiguration) MyCall() callsign.Callsign {
	return c.myCall
}

func (c *StaticConfiguration) MyLocator() locator.Locator {
	return c.myLocator
}

func (c *StaticConfiguration) EnterTheirNumber() bool {
	return true
}

func (c *StaticConfiguration) EnterTheirXchange() bool {
	return true
}

func (c *StaticConfiguration) CabrilloQSOTemplate() string {
	return "{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}} {{.TheirCall}} {{.TheirReport}} {{.TheirNumber}} {{.TheirXchange}}"
}

func (c *StaticConfiguration) AllowMultiBand() bool {
	return false
}

func (c *StaticConfiguration) AllowMultiMode() bool {
	return false
}

func (c *StaticConfiguration) KeyerHost() string {
	return ""
}

func (c *StaticConfiguration) KeyerPort() int {
	return 0
}

func (c *StaticConfiguration) KeyerWPM() int {
	return 25
}

func (c *StaticConfiguration) KeyerSPMacros() []string {
	return []string{}
}

func (c *StaticConfiguration) KeyerRunMacros() []string {
	return []string{}
}

func (c *StaticConfiguration) HamlibAddress() string {
	return ""
}

func (c *StaticConfiguration) CountPerBand() bool {
	return true
}

func (c *StaticConfiguration) SameCountryPoints() int {
	return 1
}

func (c *StaticConfiguration) SameContinentPoints() int {
	return 3
}

func (c *StaticConfiguration) OtherPoints() int {
	return 5
}

func (c *StaticConfiguration) SpecificCountryPoints() int {
	return 0
}

func (c *StaticConfiguration) SpecificCountryPrefixes() []string {
	return []string{}
}

func (c *StaticConfiguration) Multis() []string {
	return []string{}
}
