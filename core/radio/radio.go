package radio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ftl/hamradio/bandplan"

	"github.com/ftl/hellocontest/core/cwdaemon"
	"github.com/ftl/hellocontest/core/demoradio"
	"github.com/ftl/hellocontest/core/hamlib"
	"github.com/ftl/hellocontest/core/model"
	"github.com/ftl/hellocontest/core/tci"
)

const (
	hamlibVFO1Option = "vfo1"
	hamlibVFO2Option = "vfo2"
	tciTRXOption     = "trx"
	tciTRX2Option    = "trx2"

	incrementalTuningPerVFOOption = "rit_xit_per_vfo"
)

type View interface {
	AddRadio(name string)
	AddKeyer(name string)

	SetRadioSelected(name string)
	SetKeyerSelected(name string)
}

type Controller struct {
	radios      []model.Radio
	keyers      []model.Keyer
	bandplan    bandplan.Bandplan
	listeners   []any
	asyncRunner model.AsyncRunner

	view          View
	ignoreUpdates bool

	activeRadio     radio
	activeRadioName string
	activeKeyer     keyer
	activeKeyerName string
	radioAsKeyer    bool
	sendSpotsToTci  bool

	incrementalTuningPerVFO bool
}

type radio interface {
	keyer
	Disconnect()
	Active() bool
	SingleVFO() bool
	SetCurrentVFO(model.VFOID)
	SetTXVFO(model.VFOID)
	SetFrequency(model.VFOID, model.Frequency)
	SetBand(model.VFOID, model.Band)
	SetMode(model.VFOID, model.Mode)
	SetIncrementalTuning(model.VFOID, model.IncrementalTuningKind, bool, model.Frequency)
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

func NewController(radios []model.Radio, keyers []model.Keyer, bandplan bandplan.Bandplan, asyncRunner model.AsyncRunner) *Controller {
	result := &Controller{
		radios:      radios,
		keyers:      keyers,
		bandplan:    bandplan,
		asyncRunner: asyncRunner,
	}
	return result
}

// bandplanAware is implemented by radio clients that can swap their bandplan
// while connected.
type bandplanAware interface {
	SetBandplan(bandplan.Bandplan)
}

// SetBandplan swaps the bandplan used for band lookups and pushes it into the
// currently-connected radio client, if any. New connections read the updated
// bandplan when they are created.
func (c *Controller) SetBandplan(bandplan bandplan.Bandplan) {
	c.bandplan = bandplan
	if r, ok := c.activeRadio.(bandplanAware); ok {
		r.SetBandplan(bandplan)
	}
}

func (c *Controller) SetView(view View) {
	c.view = view
	c.doIgnoreUpdates(func() {
		for _, radio := range c.radios {
			view.AddRadio(radio.Name)
		}
		view.AddKeyer(model.RadioKeyer)
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
	model.Emit(c.listeners, func(listener model.ServiceStatusListener) {
		listener.StatusChanged(model.RadioService, available)
	})
}

func (c *Controller) emitKeyerStatusChanged(available bool) {
	model.Emit(c.listeners, func(listener model.ServiceStatusListener) {
		listener.StatusChanged(model.KeyerService, available)
	})
}

func (c *Controller) emitRadioChanged(name string, singleVFO bool) {
	model.Emit(c.listeners, func(listener model.RadioChangedListener) {
		listener.RadioChanged(name, singleVFO)
	})
	c.incrementalTuningPerVFO = false
	if config, ok := c.radioConfig(name); ok {
		c.incrementalTuningPerVFO = config.Options[incrementalTuningPerVFOOption] == "true"
	}
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
	model.Emit(c.listeners, func(listener listenerType) {
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

	c.radioAsKeyer = normalizeName(config.Keyer) == model.RadioKeyer
	var radio radio
	var err error
	switch config.Type {
	case model.RadioTypeHamlib:
		radio = c.newHamlibClient(config)
	case model.RadioTypeTCI:
		radio, err = c.newTCIClient(config)
	case model.RadioTypeDemo:
		radio = demoradio.New(c.bandplan, config.Options[model.DemoSingleVFOOption] != "false", c.asyncRunner)
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
	c.activeRadio.Notify(model.ConnectionChangedFunc(c.onRadioConnectionChanged))
	c.emitRadioChanged(config.Name, c.activeRadio.SingleVFO())
	c.onRadioConnectionChanged(c.activeRadio.IsConnected())

	if c.radioAsKeyer {
		c.activeKeyer = c.activeRadio
		c.activeKeyerName = model.RadioKeyer
		c.emitKeyerSelected(model.RadioKeyer)
		c.onKeyerConnectionChanged(c.activeKeyer.IsConnected())
		return nil
	}
	return c.SelectKeyer(config.Keyer)
}

func (c *Controller) radioConfig(name string) (model.Radio, bool) {
	name = normalizeName(name)
	for _, config := range c.radios {
		if normalizeName(config.Name) == name {
			return config, true
		}
	}
	return model.Radio{}, false
}

func (c *Controller) onRadioConnectionChanged(connected bool) {
	c.emitRadioStatusChanged(connected)
	if c.radioAsKeyer {
		c.emitKeyerStatusChanged(connected)
	}
}

func (c *Controller) newHamlibClient(config model.Radio) radio {
	vfo1, ok := config.Options[hamlibVFO1Option]
	if !ok {
		vfo1 = ""
	}
	vfo2, ok := config.Options[hamlibVFO2Option]
	if !ok {
		vfo2 = ""
	}
	hamlibClient := hamlib.New(config.Address, c.bandplan, vfo1, vfo2, c.asyncRunner)
	hamlibClient.KeepOpen()
	return hamlibClient
}

func (c *Controller) newTCIClient(config model.Radio) (radio, error) {
	var err error
	trx := 0
	trxStr, ok := config.Options[tciTRXOption]
	if ok {
		trx, err = strconv.Atoi(trxStr)
		if err != nil {
			return nil, fmt.Errorf("invalid trx option: %v", err)
		}
	}
	trx2 := -1
	trx2Str, ok := config.Options[tciTRX2Option]
	if ok {
		trx2, err = strconv.Atoi(trx2Str)
		if err != nil {
			return nil, fmt.Errorf("invalid trx2 option: %v", err)
		}
	}
	tciClient, err := tci.NewClient(config.Address, trx, trx2, c.bandplan, c.asyncRunner)
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

func (c *Controller) SetCurrentVFO(vfo model.VFOID) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetCurrentVFO(vfo)
}

func (c *Controller) SetTXVFO(vfo model.VFOID) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetTXVFO(vfo)
}

func (c *Controller) SetFrequency(vfo model.VFOID, frequency model.Frequency) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetFrequency(vfo, frequency)
}

func (c *Controller) SetBand(vfo model.VFOID, band model.Band) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetBand(vfo, band)
}

func (c *Controller) SetMode(vfo model.VFOID, mode model.Mode) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetMode(vfo, mode)
}

func (c *Controller) IncrementalTuningPerVFO() bool {
	return c.incrementalTuningPerVFO
}

func (c *Controller) SetIncrementalTuning(vfo model.VFOID, kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	if c.activeRadio == nil {
		return
	}
	c.activeRadio.SetIncrementalTuning(vfo, kind, active, offset)
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
	EntryAdded(model.BandmapEntry)
	EntryUpdated(model.BandmapEntry)
	EntryRemoved(model.BandmapEntry)
	MarkersChanged([]model.BandmapMarker)
}

func (c *Controller) SetSendSpotsToTci(value bool) {
	c.sendSpotsToTci = value

	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.SetSendSpots(c.sendSpotsToTci)
}

func (c *Controller) EntryAdded(entry model.BandmapEntry) {
	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.EntryAdded(entry)
}

func (c *Controller) EntryUpdated(entry model.BandmapEntry) {
	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.EntryUpdated(entry)
}

func (c *Controller) EntryRemoved(entry model.BandmapEntry) {
	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.EntryRemoved(entry)
}

func (c *Controller) MarkersChanged(markers []model.BandmapMarker) {
	tciRadio, ok := c.activeRadio.(tciRadioType)
	if !ok {
		return
	}
	tciRadio.MarkersChanged(markers)
}

/* Keyer */

func (c *Controller) Keyer() string {
	return c.activeKeyerName
}

func (c *Controller) SelectKeyer(name string) error {
	if normalizeName(c.activeKeyerName) == normalizeName(name) {
		return nil
	}
	radioAsKeyer := normalizeName(name) == model.RadioKeyer

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
		c.activeKeyerName = model.RadioKeyer
		c.emitKeyerSelected(name)
		return nil
	}

	switch config.Type {
	case model.KeyerTypeCWDaemon:
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

	c.activeKeyer.Notify(model.ConnectionChangedFunc(c.onKeyerConnectionChanged))
	c.emitKeyerSelected(name)
	c.emitKeyerStatusChanged(c.activeKeyer.IsConnected())

	return nil
}

func (c *Controller) keyerConfig(name string) (model.Keyer, bool) {
	name = normalizeName(name)
	for _, config := range c.keyers {
		if normalizeName(config.Name) == name {
			return config, true
		}
	}
	return model.Keyer{}, false
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
