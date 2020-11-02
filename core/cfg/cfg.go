package cfg

import (
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
}

type LoadedConfiguration struct {
	data Data
}

func (l *LoadedConfiguration) MyCall() callsign.Callsign {
	myCall, _ := callsign.Parse(l.data.MyCall)
	return myCall
}

func (l *LoadedConfiguration) MyLocator() locator.Locator {
	myLocator, _ := locator.Parse(l.data.MyLocator)
	return myLocator
}

func (l *LoadedConfiguration) EnterTheirNumber() bool {
	return l.data.EnterTheirNumber
}

func (l *LoadedConfiguration) EnterTheirXchange() bool {
	return l.data.EnterTheirXchange
}

func (l *LoadedConfiguration) CabrilloQSOTemplate() string {
	return l.data.CabrilloQSOTemplate
}

func (l *LoadedConfiguration) AllowMultiBand() bool {
	return l.data.AllowMultiBand
}

func (l *LoadedConfiguration) AllowMultiMode() bool {
	return l.data.AllowMultiMode
}

func (l *LoadedConfiguration) KeyerHost() string {
	return l.data.KeyerHost
}

func (l *LoadedConfiguration) KeyerPort() int {
	return l.data.KeyerPort
}

func (l *LoadedConfiguration) KeyerWPM() int {
	return l.data.KeyerWPM
}

func (l *LoadedConfiguration) KeyerSPMacros() []string {
	if l.data.KeyerSPMacros == nil {
		return []string{}
	}
	return l.data.KeyerSPMacros
}

func (l *LoadedConfiguration) KeyerRunMacros() []string {
	if l.data.KeyerRunMacros == nil {
		return []string{}
	}
	return l.data.KeyerRunMacros
}

func (l *LoadedConfiguration) HamlibAddress() string {
	return l.data.HamlibAddress
}

type StaticConfiguration struct {
	myCall    callsign.Callsign
	myLocator locator.Locator
}

func (s *StaticConfiguration) MyCall() callsign.Callsign {
	return s.myCall
}

func (s *StaticConfiguration) MyLocator() locator.Locator {
	return s.myLocator
}

func (s *StaticConfiguration) EnterTheirNumber() bool {
	return true
}

func (s *StaticConfiguration) EnterTheirXchange() bool {
	return true
}

func (s *StaticConfiguration) CabrilloQSOTemplate() string {
	return "{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}} {{.TheirCall}} {{.TheirReport}} {{.TheirNumber}} {{.TheirXchange}}"
}

func (s *StaticConfiguration) AllowMultiBand() bool {
	return false
}

func (s *StaticConfiguration) AllowMultiMode() bool {
	return false
}

func (s *StaticConfiguration) KeyerHost() string {
	return ""
}

func (s *StaticConfiguration) KeyerPort() int {
	return 0
}

func (s *StaticConfiguration) KeyerWPM() int {
	return 25
}

func (s *StaticConfiguration) KeyerSPMacros() []string {
	return []string{}
}

func (s *StaticConfiguration) KeyerRunMacros() []string {
	return []string{}
}

func (s *StaticConfiguration) HamlibAddress() string {
	return ""
}
