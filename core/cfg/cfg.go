package cfg

import (
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/cfg"
	"github.com/ftl/hamradio/locator"
	"github.com/ftl/hellocontest/core"
)

// Load loads the configuration from the default location (see github.com/ftl/cfg/LoadDefault())
func Load() (core.Configuration, error) {
	configuration, err := cfg.LoadDefault()
	if err != nil {
		return nil, err
	}
	return &loaded{
		configuration: configuration,
	}, nil
}

// Static creates a static configuration instance with the given data.
func Static(myCall callsign.Callsign, myLocator locator.Locator) core.Configuration {
	return &static{
		myCall:    myCall,
		myLocator: myLocator,
	}
}

const (
	enterTheirNumber    cfg.Key = "hellocontest.enter.theirNumber"
	enterTheirXchange   cfg.Key = "hellocontest.enter.theirXchange"
	allowMultiBand      cfg.Key = "hellocontest.enter.allowMultiBand"
	allowMultiMode      cfg.Key = "hellocontest.enter.allowMultiMode"
	cabrilloQSOTemplate cfg.Key = "hellocontest.cabrillo.qso"
	keyerHost           cfg.Key = "hellocontest.keyer.host"
	keyerPort           cfg.Key = "hellocontest.keyer.port"
	keyerSPPatterns     cfg.Key = "hellocontest.keyer.sp"
	keyerRunPatterns    cfg.Key = "hellocontest.keyer.run"
)

type loaded struct {
	configuration cfg.Configuration
}

func (l loaded) MyCall() callsign.Callsign {
	value := l.configuration.Get(cfg.MyCall, "").(string)
	myCall, _ := callsign.Parse(value)
	return myCall
}

func (l loaded) MyLocator() locator.Locator {
	value := l.configuration.Get(cfg.MyLocator, "").(string)
	myLocator, _ := locator.Parse(value)
	return myLocator
}

func (l loaded) EnterTheirNumber() bool {
	return l.configuration.Get(enterTheirNumber, true).(bool)
}

func (l loaded) EnterTheirXchange() bool {
	return l.configuration.Get(enterTheirXchange, true).(bool)
}

func (l loaded) CabrilloQSOTemplate() string {
	defaultTemplate := "{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}} {{.TheirCall}} {{.TheirReport}} {{.TheirNumber}} {{.TheirXchange}}"
	return l.configuration.Get(cabrilloQSOTemplate, defaultTemplate).(string)
}

func (l loaded) AllowMultiBand() bool {
	return l.configuration.Get(allowMultiBand, false).(bool)
}

func (l loaded) AllowMultiMode() bool {
	return l.configuration.Get(allowMultiMode, false).(bool)
}

func (l loaded) KeyerHost() string {
	return l.configuration.Get(keyerHost, "").(string)
}

func (l loaded) KeyerPort() int {
	return int(l.configuration.Get(keyerPort, 0.0).(float64))
}

func (l loaded) KeyerSPPatterns() []string {
	return l.configuration.GetStrings(keyerSPPatterns, []string{})
}

func (l loaded) KeyerRunPatterns() []string {
	return l.configuration.GetStrings(keyerRunPatterns, []string{})
}

type static struct {
	myCall    callsign.Callsign
	myLocator locator.Locator
}

func (s static) MyCall() callsign.Callsign {
	return s.myCall
}

func (s static) MyLocator() locator.Locator {
	return s.myLocator
}

func (s static) EnterTheirNumber() bool {
	return true
}

func (s static) EnterTheirXchange() bool {
	return true
}

func (s static) CabrilloQSOTemplate() string {
	return "{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}} {{.TheirCall}} {{.TheirReport}} {{.TheirNumber}} {{.TheirXchange}}"
}

func (s static) AllowMultiBand() bool {
	return false
}

func (s static) AllowMultiMode() bool {
	return false
}

func (s static) KeyerHost() string {
	return ""
}

func (s static) KeyerPort() int {
	return 0
}

func (s static) KeyerSPPatterns() []string {
	return []string{}
}

func (s static) KeyerRunPatterns() []string {
	return []string{}
}
