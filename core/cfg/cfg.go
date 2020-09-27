package cfg

import (
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/cfg"
	"github.com/ftl/hamradio/locator"
	"github.com/pkg/errors"
)

// Load loads the configuration from the default location (see github.com/ftl/cfg/LoadDefault())
func Load() (*LoadedConfiguration, error) {
	configuration, err := cfg.LoadDefault()
	if err != nil {
		return nil, err
	}
	return &LoadedConfiguration{
		configuration: configuration,
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

const (
	enterTheirNumber    cfg.Key = "hellocontest.enter.theirNumber"
	enterTheirXchange   cfg.Key = "hellocontest.enter.theirXchange"
	allowMultiBand      cfg.Key = "hellocontest.enter.allowMultiBand"
	allowMultiMode      cfg.Key = "hellocontest.enter.allowMultiMode"
	cabrilloQSOTemplate cfg.Key = "hellocontest.cabrillo.qso"
	keyerHost           cfg.Key = "hellocontest.keyer.host"
	keyerPort           cfg.Key = "hellocontest.keyer.port"
	keyerWPM            cfg.Key = "hellocontest.keyer.wpm"
	keyerSPPatterns     cfg.Key = "hellocontest.keyer.sp"
	keyerRunPatterns    cfg.Key = "hellocontest.keyer.run"
)

type LoadedConfiguration struct {
	configuration cfg.Configuration
}

func (l *LoadedConfiguration) MyCall() callsign.Callsign {
	value := l.configuration.Get(cfg.MyCall, "").(string)
	myCall, _ := callsign.Parse(value)
	return myCall
}

func (l *LoadedConfiguration) MyLocator() locator.Locator {
	value := l.configuration.Get(cfg.MyLocator, "").(string)
	myLocator, _ := locator.Parse(value)
	return myLocator
}

func (l *LoadedConfiguration) EnterTheirNumber() bool {
	return l.configuration.Get(enterTheirNumber, true).(bool)
}

func (l *LoadedConfiguration) EnterTheirXchange() bool {
	return l.configuration.Get(enterTheirXchange, true).(bool)
}

func (l *LoadedConfiguration) CabrilloQSOTemplate() string {
	defaultTemplate := "{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}} {{.TheirCall}} {{.TheirReport}} {{.TheirNumber}} {{.TheirXchange}}"
	return l.configuration.Get(cabrilloQSOTemplate, defaultTemplate).(string)
}

func (l *LoadedConfiguration) AllowMultiBand() bool {
	return l.configuration.Get(allowMultiBand, false).(bool)
}

func (l *LoadedConfiguration) AllowMultiMode() bool {
	return l.configuration.Get(allowMultiMode, false).(bool)
}

func (l *LoadedConfiguration) KeyerHost() string {
	return l.configuration.Get(keyerHost, "").(string)
}

func (l *LoadedConfiguration) KeyerPort() int {
	return int(l.configuration.Get(keyerPort, 0.0).(float64))
}

func (l *LoadedConfiguration) KeyerWPM() int {
	return int(l.configuration.Get(keyerWPM, 25.0).(float64))
}

func (l *LoadedConfiguration) KeyerSPPatterns() []string {
	return l.configuration.GetStrings(keyerSPPatterns, []string{})
}

func (l *LoadedConfiguration) KeyerRunPatterns() []string {
	return l.configuration.GetStrings(keyerRunPatterns, []string{})
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

func (s *StaticConfiguration) KeyerSPPatterns() []string {
	return []string{}
}

func (s *StaticConfiguration) KeyerRunPatterns() []string {
	return []string{}
}
