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
