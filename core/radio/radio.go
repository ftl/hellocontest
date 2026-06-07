package radio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ftl/hamradio/bandplan"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/cwdaemon"
	"github.com/ftl/hellocontest/core/hamlib"
	"github.com/ftl/hellocontest/core/tci"
)

const (
	hamlibVFO1Option   = "vfo1"
	hamlibVFO2Option   = "vfo2"
	tciTRXOption       = "trx"
	tciSingleVFOOption = "single_vfo"
)

type View interface {
	AddRadio(name string)
	AddKeyer(name string)

	SetRadioSelected(name string)
	SetKeyerSelected(name string)
}

type Controller struct {
	radios    []core.Radio
	keyers    []core.Keyer
	bandplan  bandplan.Bandplan
	listeners []any

	view          View
	ignoreUpdates bool

	activeRadio     radio
	activeRadioName string
	activeKeyer     keyer
	activeKeyerName string
	radioAsKeyer    bool
	sendSpotsToTci  bool
}

type radio interface {
	keyer
	Disconnect()
	Active() bool
	SingleVFO() bool
	SetCurrentVFO(core.VFOID)
	SetTXVFO(core.VFOID)
	SetFrequency(core.VFOID, core.Frequency)
	SetBand(core.VFOID, core.Band)
	SetMode(core.VFOID, core.Mode)
	SetXIT(bool, core.Frequency)
	MuteAudio(core.VFOID)
	UnmuteAudio(core.VFOID)
	ToggleAudio(core.VFOID)
	Refresh()
	Notify(any)
}

type keyer interface {
	IsConnected() bool
	Speed(int)
	Send(text string)
	Abort()
	Notify(any)
}

func NewController(radios []core.Radio, keyers []core.Keyer, bandplan bandplan.Bandplan) *Controller {
	result := &Controller{
		radios:   radios,
		keyers:   keyers,
		bandplan: bandplan,
	}
	return result
}

func (c *Controller) SetView(view View) {
	c.view = view
	c.doIgnoreUpdates(func() {
		for _, radio := range c.radios {
			view.AddRadio(radio.Name)
		}
		view.AddKeyer(core.RadioKeyer)
		for _, keyer := range c.keyers {
			view.AddKeyer(keyer.Name)
		}

		if c.activeRadio != nil {
			view.SetRadioSelected(c.activeRadioName)
		}
		if c.activeKeyer != nil {
			view.SetKeyerSelected(c.activeKeyerName)
		}
	})
}

func (c *Controller) doIgnoreUpdates(f func()) {
	c.ignoreUpdates = true
	defer func() {
		c.ignoreUpdates = false
	}()
	f()
}

func (c *Controller) Stop() {
	if c.activeRadio != nil {
		c.activeRadio.Disconnect()
		c.activeRadio = nil
		c.activeRadioName = ""
		c.emitRadioChanged("", true)
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
	core.Emit(c.listeners, func(listener core.ServiceStatusListener) {
		listener.StatusChanged(core.RadioService, available)
	})
}

func (c *Controller) emitKeyerStatusChanged(available bool) {
	core.Emit(c.listeners, func(listener core.ServiceStatusListener) {
		listener.StatusChanged(core.KeyerService, available)
	})
}

func (c *Controller) emitRadioChanged(name string, singleVFO bool) {
	core.Emit(c.listeners, func(listener core.RadioChangedListener) {
		listener.RadioChanged(name, singleVFO)
	})
	if c.view != nil {
		c.doIgnoreUpdates(func() {
			c.view.SetRadioSelected(name)
		})
	}
}

func (c *Controller) emitKeyerSelected(name string) {
	type listenerType interface {
		KeyerSelected(string)
	}
	core.Emit(c.listeners, func(listener listenerType) {
		listener.KeyerSelected(name)
	})
	if c.view != nil {
		c.doIgnoreUpdates(func() {
			c.view.SetKeyerSelected(name)
		})
	}
}

/* Radio */

func (c *Controller) Radio() string {
	return c.activeRadioName
}

func (c *Controller) SelectRadio(name string) error {
	config, ok := c.radioConfig(name)
	if !ok {
		return fmt.Errorf("cannot find radio %q", name)
	}

	if c.activeRadio != nil {
		c.activeRadio.Disconnect()
		c.activeRadio = nil
		c.activeRadioName = ""
	}
	if c.activeKeyer != nil {
		c.activeKeyer.Abort()
		c.activeKeyer = nil
		c.activeKeyerName = ""
	}

	c.radioAsKeyer = normalizeName(config.Keyer) == core.RadioKeyer
	var radio radio
	var err error
	switch config.Type {
	case core.RadioTypeHamlib:
		radio = c.newHamlibClient(config)
	case core.RadioTypeTCI:
		radio, err = c.newTCIClient(config)
	default:
		err = fmt.Errorf("unknown radio type %q", config.Type)
	}

	if err != nil {
		c.emitRadioChanged("", true)
		return err
	}
	c.activeRadio = radio
	c.activeRadioName = name

	for _, listener := range c.listeners {
		c.activeRadio.Notify(listener)
	}
	c.activeRadio.Notify(core.ConnectionChangedFunc(c.onRadioConnectionChanged))
	c.emitRadioChanged(config.Name, c.activeRadio.SingleVFO())
	c.onRadioConnectionChanged(c.activeRadio.IsConnected())

	if c.radioAsKeyer {
		c.activeKeyer = c.activeRadio
		c.activeKeyerName = core.RadioKeyer
		c.emitKeyerSelected(core.RadioKeyer)
		c.onKeyerConnectionChanged(c.activeKeyer.IsConnected())
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

func (c *Controller) newHamlibClient(config core.Radio) radio {
	vfo1, ok := config.Options[hamlibVFO1Option]
	if !ok {
		vfo1 = ""
	}
	vfo2, ok := config.Options[hamlibVFO2Option]
	if !ok {
		vfo2 = ""
	}
	hamlibClient := hamlib.New(config.Address, c.bandplan, vfo1, vfo2)
	hamlibClient.KeepOpen()
	return hamlibClient
}

func (c *Controller) newTCIClient(config core.Radio) (radio, error) {
	var err error
	trx := 0
	trxStr, ok := config.Options[tciTRXOption]
	if ok {
		trx, err = strconv.Atoi(trxStr)
		if err != nil {
			return nil, fmt.Errorf("invalid trx option: %v", err)
		}
	}
	singleVFO := false
	singleVFOStr, ok := config.Options[tciSingleVFOOption]
	if ok {
		singleVFO, err = strconv.ParseBool(singleVFOStr)
		if err != nil {
			return nil, fmt.Errorf("invalid single_vfo option: %v", err)
		}
	}
	tciClient, err := tci.NewClient(config.Address, trx, singleVFO, c.bandplan)
	if err != nil {
		return nil, err
	}
	tciClient.SetSendSpots(c.sendSpotsToTci)
	return tciClient, nil
}

func (c *Controller) Active() bool {
	if c.activeRadio == nil {
		return false
	}
	return c.activeRadio.Active()
}

func (c *Controller) SingleVFO() bool {
	if c.activeRadio == nil {
		return true
	}
	return c.activeRadio.SingleVFO()
}

func (c *Controller) SetCurrentVFO(vfo core.VFOID) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetCurrentVFO(vfo)
}

func (c *Controller) SetTXVFO(vfo core.VFOID) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetTXVFO(vfo)
}

func (c *Controller) SetFrequency(vfo core.VFOID, frequency core.Frequency) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetFrequency(vfo, frequency)
}

func (c *Controller) SetBand(vfo core.VFOID, band core.Band) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetBand(vfo, band)
}

func (c *Controller) SetMode(vfo core.VFOID, mode core.Mode) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetMode(vfo, mode)
}

func (c *Controller) SetXIT(active bool, offset core.Frequency) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetXIT(active, offset)
}

func (c *Controller) MuteAudio(vfo core.VFOID) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.MuteAudio(vfo)
}

func (c *Controller) UnmuteAudio(vfo core.VFOID) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.UnmuteAudio(vfo)
}

func (c *Controller) ToggleAudio(vfo core.VFOID) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.ToggleAudio(vfo)
}

func (c *Controller) Refresh() {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.Refresh()
}

/* TCI specific */

type tciRadioType interface {
	SetSendSpots(bool)
	EntryAdded(core.BandmapEntry)
	EntryUpdated(core.BandmapEntry)
	EntryRemoved(core.BandmapEntry)
}

func (c *Controller) SetSendSpotsToTci(value bool) {
	c.sendSpotsToTci = value

	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.SetSendSpots(c.sendSpotsToTci)
}

func (c *Controller) EntryAdded(entry core.BandmapEntry) {
	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.EntryAdded(entry)
}

func (c *Controller) EntryUpdated(entry core.BandmapEntry) {
	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.EntryUpdated(entry)
}

func (c *Controller) EntryRemoved(entry core.BandmapEntry) {
	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.EntryRemoved(entry)
}

/* Keyer */

func (c *Controller) Keyer() string {
	return c.activeKeyerName
}

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
		client, err := cwdaemon.NewClient(config.Address)
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

	c.activeKeyer.Notify(core.ConnectionChangedFunc(c.onKeyerConnectionChanged))
	c.emitKeyerSelected(name)
	c.emitKeyerStatusChanged(c.activeKeyer.IsConnected())

	return nil
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

func (c *Controller) onKeyerConnectionChanged(connected bool) {
	c.emitKeyerStatusChanged(connected)
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
