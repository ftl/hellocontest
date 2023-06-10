package radio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/hamradio/cwclient"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/hamlib"
	"github.com/ftl/hellocontest/core/tci"
)

type Controller struct {
	radios    []core.Radio
	keyers    []core.Keyer
	bandplan  bandplan.Bandplan
	listeners []any

	activeRadio     radio
	activeKeyer     keyer
	activeKeyerName string
	radioAsKeyer    bool
	sendSpotsToTci  bool
}

type radio interface {
	keyer
	Disconnect()
	Active() bool
	SetFrequency(core.Frequency)
	SetBand(core.Band)
	SetMode(core.Mode)
	Refresh()
	Notify(any)
}

type keyer interface {
	Connect() error
	IsConnected() bool
	Speed(int)
	Send(text string)
	Abort()
}

func NewController(radios []core.Radio, keyers []core.Keyer, bandplan bandplan.Bandplan) *Controller {
	result := &Controller{
		radios:   radios,
		keyers:   keyers,
		bandplan: bandplan,
	}
	return result
}

func (c *Controller) Stop() {
	if c.activeRadio != nil {
		c.activeRadio.Disconnect()
		c.activeRadio = nil
	}
	if c.activeKeyer != nil {
		c.activeKeyer.Abort()
		c.activeKeyer = nil
		c.activeKeyerName = ""
	}
}

func (c *Controller) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Controller) emitRadioStatusChanged(available bool) {
	for _, listener := range c.listeners {
		if serviceStatusListener, ok := listener.(core.ServiceStatusListener); ok {
			serviceStatusListener.StatusChanged(core.RadioService, available)
		}
	}
}

func (c *Controller) emitKeyerStatusChanged(available bool) {
	for _, listener := range c.listeners {
		if serviceStatusListener, ok := listener.(core.ServiceStatusListener); ok {
			serviceStatusListener.StatusChanged(core.KeyerService, available)
		}
	}
}

func (c *Controller) emitRadioSelected(name string) {
	type listenerType interface {
		RadioSelected(string)
	}
	for _, listener := range c.listeners {
		if typedListener, ok := listener.(listenerType); ok {
			typedListener.RadioSelected(name)
		}
	}
}

func (c *Controller) emitKeyerSelected(name string) {
	type listenerType interface {
		KeyerSelected(string)
	}
	for _, listener := range c.listeners {
		if typedListener, ok := listener.(listenerType); ok {
			typedListener.KeyerSelected(name)
		}
	}
}

/* Radio */

func (c *Controller) SelectRadio(name string) error {
	config, ok := c.radioConfig(name)
	if !ok {
		return fmt.Errorf("cannot find radio %q", name)
	}

	if c.activeRadio != nil {
		c.activeRadio.Disconnect()
		c.activeRadio = nil
	}
	if c.activeKeyer != nil {
		c.activeKeyer.Abort()
		c.activeKeyer = nil
		c.activeKeyerName = ""
	}

	c.radioAsKeyer = normalizeName(config.Keyer) == core.RadioKeyer
	var radioKeyer keyer
	switch config.Type {
	case core.RadioTypeHamlib:
		hamlibClient := hamlib.New(config.Address, c.bandplan)
		hamlibClient.KeepOpen()
		c.activeRadio = hamlibClient
		radioKeyer = hamlibClient
	case core.RadioTypeTCI:
		tciClient, err := tci.NewClient(config.Address, c.bandplan)
		if err != nil {
			c.emitRadioSelected("")
			return err
		}
		tciClient.SetSendSpots(c.sendSpotsToTci)
		c.activeRadio = tciClient
		radioKeyer = tciClient
	default:
		c.emitRadioSelected("")
		return fmt.Errorf("unknown radio type %q", config.Type)
	}

	for _, listener := range c.listeners {
		c.activeRadio.Notify(listener)
	}
	c.activeRadio.Notify(connectionChangedFunc(c.onRadioConnectionChanged))
	c.emitRadioSelected(config.Name)

	if c.radioAsKeyer {
		c.activeKeyer = radioKeyer
		c.activeKeyerName = core.RadioKeyer
		c.emitKeyerSelected(core.RadioKeyer)
		return nil
	}
	return c.SelectKeyer(config.Keyer)
}

func (c *Controller) radioConfig(name string) (core.Radio, bool) {
	name = normalizeName(name)
	for _, config := range c.radios {
		if normalizeName(config.Name) == name {
			return config, true
		}
	}
	return core.Radio{}, false
}

func (c *Controller) onRadioConnectionChanged(connected bool) {
	c.emitRadioStatusChanged(connected)
	if c.radioAsKeyer {
		c.emitKeyerStatusChanged(connected)
	}
}

func (c *Controller) Active() bool {
	if c.activeRadio == nil {
		return false
	}
	return c.activeRadio.Active()
}

func (c *Controller) SetFrequency(f core.Frequency) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetFrequency(f)
}

func (c *Controller) SetBand(b core.Band) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetBand(b)
}

func (c *Controller) SetMode(m core.Mode) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetMode(m)
}

func (c *Controller) Refresh() {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.Refresh()
}

func (c *Controller) SetSendSpotsToTci(value bool) {
	c.sendSpotsToTci = value

	type tciRadioType interface {
		SetSendSpots(bool)
	}

	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.SetSendSpots(c.sendSpotsToTci)
}

/* Keyer */

func (c *Controller) SelectKeyer(name string) error {
	if normalizeName(c.activeKeyerName) == normalizeName(name) {
		return nil
	}
	radioAsKeyer := normalizeName(name) == core.RadioKeyer

	config, ok := c.keyerConfig(name)
	if !ok && !radioAsKeyer {
		return fmt.Errorf("cannot find keyer %q", name)
	}

	if c.activeKeyer != nil {
		c.activeKeyer.Abort()
		c.activeKeyer = nil
		c.activeKeyerName = ""
	}

	c.radioAsKeyer = radioAsKeyer
	if c.radioAsKeyer {
		c.activeKeyer = c.activeRadio
		c.activeKeyerName = core.RadioKeyer
		c.emitKeyerSelected(name)
		return nil
	}

	switch config.Type {
	case core.KeyerTypeCWDaemon:
		client, err := cwclient.New(toHostAndPort(config.Address, 6789))
		if err != nil {
			c.emitKeyerSelected("")
			return err
		}
		c.activeKeyer = client
		c.activeKeyerName = name
	default:
		c.emitKeyerSelected("")
		return fmt.Errorf("unknown keyer %q", name)
	}

	c.emitKeyerSelected(name)

	return nil
}

func toHostAndPort(address string, defaultPort int) (string, int) {
	address = strings.TrimSpace(address)
	if address == "" {
		return "localhost", defaultPort
	}

	parts := strings.Split(address, ":")
	if len(parts) == 1 {
		return parts[0], defaultPort
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return parts[0], defaultPort
	}

	return parts[0], port
}

func (c *Controller) keyerConfig(name string) (core.Keyer, bool) {
	name = normalizeName(name)
	for _, config := range c.keyers {
		if normalizeName(config.Name) == name {
			return config, true
		}
	}
	return core.Keyer{}, false
}

func (c *Controller) Connect() error {
	// TODO move the connection handling from Keyer to radio.Controller
	if c.activeKeyer == nil {
		return fmt.Errorf("no keyer configured")
	}
	return nil
}

func (c *Controller) IsConnected() bool {
	if c.activeKeyer == nil {
		return false
	}
	return c.activeKeyer.IsConnected()
}

func (c *Controller) Speed(speed int) {
	if c.activeKeyer == nil {
		return
	}
	c.activeKeyer.Speed(speed)
}

func (c *Controller) Send(text string) {
	if c.activeKeyer == nil {
		return
	}
	c.activeKeyer.Send(text)
}

func (c *Controller) Abort() {
	if c.activeKeyer == nil {
		return
	}
	c.activeKeyer.Abort()
}

/* Helpers */

func normalizeName(name string) string {
	return strings.TrimSpace(strings.ToLower(name))
}

type connectionChangedFunc func(bool)

func (f connectionChangedFunc) ConnectionChanged(connected bool) {
	f(connected)
}

type nullRadio struct{}

func (n *nullRadio) Disconnect()                 {}
func (n *nullRadio) Active() bool                { return false }
func (n *nullRadio) SetFrequency(core.Frequency) {}
func (n *nullRadio) SetBand(core.Band)           {}
func (n *nullRadio) SetMode(core.Mode)           {}
func (n *nullRadio) Refresh()                    {}
func (n *nullRadio) Notify(any)                  {}

type nullKeyer struct{}

func (n *nullKeyer) Connect() error    { return fmt.Errorf("no keyer configured") }
func (n *nullKeyer) IsConnected() bool { return false }
func (n *nullKeyer) Speed(int)         {}
func (n *nullKeyer) Send(text string)  {}
func (n *nullKeyer) Abort()            {}
func (n *nullKeyer) Notify(any)        {}
