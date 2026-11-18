package hamlib

import (
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/hl-go"

	"github.com/ftl/hellocontest/core/model"
)

const (
	doBufferSize = 10
	pttThreshold = 4 // number of consecutive PTT-off polls before emitting PTT-off
)

type Client struct {
	client *hl.RigClient

	bandplan    bandplan.Bandplan
	vfos        []hl.VFO
	asyncRunner model.AsyncRunner

	listeners []any

	pollingInterval time.Duration
	requestTimeout  time.Duration
	do              chan func()
	done            chan struct{}
	loopRunning     *atomic.Bool
	loopStopped     chan struct{}

	currentVFO hl.VFO
	singleVFO  bool
	lastState  []vfoState
	audioLevel []float64

	ptt         bool
	pttOffCount int

	frequencyLock      sync.Mutex
	pendingFrequency   []model.Frequency
	frequencyScheduled []bool
}

type vfoState struct {
	vfo       hl.VFO
	frequency model.Frequency
	band      model.Band
	mode      model.Mode
	txVFO     bool
	xitActive bool
	xitOffset model.Frequency
	ritActive bool
	ritOffset model.Frequency
}

func New(address string, bandplan bandplan.Bandplan, vfo1, vfo2 string, asyncRunner model.AsyncRunner) *Client {
	if asyncRunner == nil {
		asyncRunner = func(f func()) { f() }
	}
	vfos, singleVFO := sanitizeVFOs(vfo1, vfo2)
	result := &Client{
		client:          hl.NewRigClient(address),
		bandplan:        bandplan,
		vfos:            vfos,
		asyncRunner:     asyncRunner,
		pollingInterval: 500 * time.Millisecond,
		requestTimeout:  500 * time.Millisecond,
		do:              make(chan func(), doBufferSize),
		done:            make(chan struct{}),
		loopRunning:     new(atomic.Bool),
		loopStopped:     make(chan struct{}),
		currentVFO:      hl.CurrVFO,
		singleVFO:       singleVFO,
		lastState:       make([]vfoState, int(model.VFOCount)),
		audioLevel:      make([]float64, int(model.VFOCount)),

		pendingFrequency:   make([]model.Frequency, int(model.VFOCount)),
		frequencyScheduled: make([]bool, int(model.VFOCount)),
	}
	result.client.Notify(result)
	if singleVFO {
		log.Printf("hamlib: using SINGLE VFO: %s", result.vfos[model.VFO1])
	} else {
		log.Printf("hamlib: using VFOs: %v", result.vfos)
	}
	return result
}

func sanitizeVFOs(vfo1, vfo2 string) ([]hl.VFO, bool) {
	result := []hl.VFO{sanitizeHamlibVFO(vfo1), sanitizeHamlibVFO(vfo2)}
	singleVFO := false
	if result[model.VFO1] == "" {
		result[model.VFO1] = hl.CurrVFO
		singleVFO = true
	}
	return result, singleVFO
}

func sanitizeHamlibVFO(s string) hl.VFO {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" { // shortcut
		return ""
	}

	validVFOs := []hl.VFO{hl.MainVFO, hl.SubVFO, hl.VFOA, hl.VFOB, hl.VFOC, hl.MainAVFO, hl.SubAVFO, hl.MainBVFO, hl.SubBVFO}
	for _, vfo := range validVFOs {
		if strings.ToLower(string(vfo)) == s {
			return vfo
		}
	}

	log.Printf("hamlib: invalid VFO: %s", s)
	return ""
}

func (c *Client) toVFOID(vfo hl.VFO) (model.VFOID, bool) {
	normalizedVFO := strings.ToLower(string(vfo))
	for id := range c.vfos {
		if strings.ToLower(string(c.vfos[id])) == normalizedVFO {
			return model.VFOID(id), true
		}
	}
	return -1, false
}

func (c *Client) run() {
	go func() {
		defer close(c.loopStopped)
		defer c.loopRunning.Store(false)
		c.loopRunning.Store(true)
		currentState := make([]vfoState, model.VFOCount)
		for {
			currentVFO, currentState, err := c.poll(currentState)
			if err != nil {
				log.Printf("hamlib: cannot poll: %v", err)
			} else {
				lastVFO := c.currentVFO
				c.currentVFO = currentVFO
				c.emitCurrentVFOChanged(lastVFO, currentVFO)
				c.updateTXVFO(currentState)

				for vfo := range model.VFOCount {
					c.emitChangeNotifications(vfo, c.lastState[vfo], currentState[vfo])
					c.lastState[vfo] = currentState[vfo]
				}

				c.pollPTT(currentState)
			}

			select {
			case f := <-c.do:
				f()
			case <-time.After(c.pollingInterval):
			case <-c.done:
				return
			}
		}
	}()
}

// pollPTT reads the radio-wide PTT state and applies debounce.
// PTT-on is emitted immediately; PTT-off is emitted only after
// pttThreshold consecutive polls read PTT=off.
func (c *Client) pollPTT(currentState []vfoState) {
	pttStatus, err := c.client.GetPTT(hl.CurrVFO)
	if err != nil {
		return
	}
	rawPTT := pttStatus != hl.PTTOff

	if rawPTT {
		c.pttOffCount = 0
		if !c.ptt {
			c.ptt = true
			c.emitPTTChanged(c.txVFOID(currentState), true)
		}
		return
	}

	// rawPTT is false
	if !c.ptt {
		return
	}
	c.pttOffCount++
	if c.pttOffCount >= pttThreshold {
		c.ptt = false
		c.pttOffCount = 0
		c.emitPTTChanged(c.txVFOID(currentState), false)
	}
}

// txVFOID returns the VFOID of the VFO currently designated for transmit.
func (c *Client) txVFOID(state []vfoState) model.VFOID {
	for vfo := range model.VFOCount {
		if state[vfo].txVFO {
			return model.VFOID(vfo)
		}
	}
	return model.VFO1
}

// updateTXVFO stores the TX VFO derived from the freshly polled state and emits
// a change notification when it differs from the previously stored value.
// Must only be called from the run goroutine.
func (c *Client) updateTXVFO(state []vfoState) {
	lastTXVFO := c.txVFOID(c.lastState)
	currentTXVFO := c.txVFOID(state)

	if lastTXVFO != currentTXVFO {
		c.emitTXVFOChanged(currentTXVFO)
	}
}

// poll must only be called from the run goroutine!
func (c *Client) poll(state []vfoState) (hl.VFO, []vfoState, error) {
	if c.singleVFO {
		vfoState, err := c.pollSingleVFO(hl.CurrVFO)
		if err != nil {
			return hl.CurrVFO, state, err
		}
		state[model.VFO1] = vfoState
		return hl.CurrVFO, state, nil
	}

	return c.pollDualVFO(state)
}

// pollSingleVFO must only be called from the run goroutine!
func (c *Client) pollSingleVFO(vfo hl.VFO) (vfoState, error) {
	if vfo == "" {
		return vfoState{}, nil
	}

	frequency, mode, _, _, _, err := c.client.GetVFOInfo(vfo)
	if err != nil {
		return vfoState{}, err
	}

	result := vfoState{
		vfo:       vfo,
		frequency: model.Frequency(frequency),
		band:      toCoreBand(c.bandplan.ByFrequency(hamradio.Frequency(frequency)).Name),
		mode:      toCoreMode(mode),
		txVFO:     true,
	}
	result = c.pollVFOAdditional(vfo, result)

	return result, nil
}

// pollDualVFO must only be called from the run goroutine!
func (c *Client) pollDualVFO(state []vfoState) (hl.VFO, []vfoState, error) {
	rigInfo, err := c.client.GetRigInfo()
	if err != nil {
		// TODO: be a bit smarter than just giving up
		log.Printf("hamlib: cannot retrieve RigInfo: %v", err)
		return c.currentVFO, state, err
	}
	// log.Printf("hamlib: RigInfo: %v", rigInfo)

	for _, vfoInfo := range rigInfo.VFOs {
		vfoID, ok := c.toVFOID(vfoInfo.VFO)
		if !ok {
			continue
		}
		vfoState := state[vfoID]
		vfoState.vfo = vfoInfo.VFO
		vfoState.frequency = model.Frequency(vfoInfo.Frequency)
		vfoState.band = toCoreBand(c.bandplan.ByFrequency(hamradio.Frequency(vfoInfo.Frequency)).Name)
		vfoState.mode = toCoreMode(vfoInfo.Mode)
		vfoState.txVFO = vfoInfo.TXActive
		state[vfoID] = vfoState
	}

	var currentVFOID model.VFOID
	var currentVFOOK bool
	currentVFO, err := c.client.GetVFO()
	if err != nil {
		// GetVFO is not supported on every radio, fallback to VFO1
		// log.Printf("hamlib: get_vfo not supported, using VFO %s", c.currentVFO)
		currentVFO = c.currentVFO
		currentVFOID = model.VFO1
		currentVFOOK = true
	} else {
		currentVFOID, currentVFOOK = c.toVFOID(currentVFO)
	}
	if currentVFOOK {
		vfoState := state[currentVFOID]
		vfoState = c.pollVFOAdditional(vfoState.vfo, vfoState)
		state[currentVFOID] = vfoState
	}

	return currentVFO, state, nil
}

// pollVFOAdditional must only be called from the run goroutine!
func (c *Client) pollVFOAdditional(vfo hl.VFO, state vfoState) vfoState {
	result := state
	result.xitActive, result.xitOffset = c.pollIncrementalTuning(vfo, hl.XITFunction, c.client.GetXIT)
	result.ritActive, result.ritOffset = c.pollIncrementalTuning(vfo, hl.RITFunction, c.client.GetRIT)
	return result
}

func (c *Client) pollIncrementalTuning(vfo hl.VFO, function hl.Function, getOffset func(hl.VFO) (hl.Frequency, error)) (bool, model.Frequency) {
	active, err := c.client.GetFunc(vfo, function)
	if err != nil {
		return false, 0
	}
	if !active {
		return false, 0
	}
	offset, err := getOffset(vfo)
	if err != nil {
		return false, 0
	}
	return true, model.Frequency(offset)
}

// doInLoop hands the given function over to the rig's command loop. It returns
// false if the loop is not running and the function will not be executed.
func (c *Client) doInLoop(f func()) bool {
	loopRunning := c.loopRunning.Load()
	if !loopRunning {
		return false
	}
	c.do <- f
	return true
}

func (c *Client) KeepOpen() {
	err := c.connectAndRun(true)
	if err != nil {
		log.Printf("hamlib: connection error: %v", err)
	}
}

func (c *Client) Connect() error {
	return c.connectAndRun(false)
}

func (c *Client) connectAndRun(automaticReconnect bool) error {
	err := c.client.Open(automaticReconnect)
	if err != nil && !automaticReconnect {
		return err
	}
	c.run()
	return nil
}

func (c *Client) Disconnect() {
	select {
	case <-c.done:
		return
	default:
	}
	close(c.done)
	<-c.loopStopped
	c.client.Close()
}

func (c *Client) RigConnected(connected bool) {
	c.emitConnectionChanged(connected)
}

func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

func (c *Client) Active() bool {
	return c.IsConnected()
}

func (c *Client) SingleVFO() bool {
	return c.singleVFO
}

func (c *Client) SetCurrentVFO(vfo model.VFOID) {
	if c.singleVFO {
		return
	}
	c.doInLoop(func() {
		err := c.client.SetVFO(c.vfos[vfo])
		if err != nil {
			log.Printf("hamlib: cannot set current VFO: %v", err)
		}
	})
}

func (c *Client) SetTXVFO(vfo model.VFOID) {
	if c.singleVFO {
		return
	}

	/*
		This is how switching on SPLIT without changing the current VFO works with the IC-7610, YMMV:
		Sub 1 Sub -> Main focus, Split on
		Main 0 Main -> Main focus, Split off
		Main 1 Main -> Sub focus, Split on
		Sub 0 Sub -> Sub focus, Split off
	*/

	c.doInLoop(func() {
		hlVFO := c.vfos[vfo]
		enableSplit := vfo == model.VFO2
		switch c.currentVFO {
		case c.vfos[model.VFO1]:
			if enableSplit {
				hlVFO = c.vfos[model.VFO2]
			} else {
				hlVFO = c.vfos[model.VFO1]
			}
		case c.vfos[model.VFO2]:
			if enableSplit {
				hlVFO = c.vfos[model.VFO1]
			} else {
				hlVFO = c.vfos[model.VFO2]
			}
		}
		err := c.client.SetSplitVFO(hlVFO, enableSplit, hlVFO)
		if err != nil {
			log.Printf("hamlib: cannot set TX VFO: %v", err)
		}
	})
}

// SetFrequency sets the frequency of the given VFO. Only one frequency command
// per VFO is queued at a time, and it always uses the most recent frequency. A
// rig that cannot keep up would otherwise work off a backlog of outdated
// frequencies long after the tuning stopped.
func (c *Client) SetFrequency(vfo model.VFOID, frequency model.Frequency) {
	c.frequencyLock.Lock()
	c.pendingFrequency[vfo] = frequency
	alreadyScheduled := c.frequencyScheduled[vfo]
	c.frequencyScheduled[vfo] = true
	c.frequencyLock.Unlock()

	if alreadyScheduled {
		return
	}

	scheduled := c.doInLoop(func() {
		err := c.client.SetFrequency(c.vfos[vfo], hl.Frequency(c.takePendingFrequency(vfo)))
		if err != nil {
			log.Printf("hamlib: cannot set frequency: %v", err)
		}
	})
	if !scheduled {
		c.takePendingFrequency(vfo)
	}
}

// takePendingFrequency returns the most recent frequency for the given VFO and
// allows a new frequency command to be queued.
func (c *Client) takePendingFrequency(vfo model.VFOID) model.Frequency {
	c.frequencyLock.Lock()
	defer c.frequencyLock.Unlock()

	c.frequencyScheduled[vfo] = false
	return c.pendingFrequency[vfo]
}

func (c *Client) SetBand(vfo model.VFOID, band model.Band) {
	outgoingBand, ok := c.bandplan[toBandplanBandName(band)]
	if !ok {
		log.Printf("hamlib: unknown band %v", band)
		return
	}

	c.doInLoop(func() {
		frequency := findModePortionCenter(c.bandplan, int(outgoingBand.Center()), toBandplanMode(c.lastState[vfo].mode))
		log.Printf("hamlib: switching %s to %s at %v", c.vfos[vfo], band, frequency)
		err := c.client.SetFrequency(c.vfos[vfo], hl.Frequency(frequency))
		if err != nil {
			log.Printf("hamlib: cannot switch to band: %v", err)
			return
		}
	})
}

func (c *Client) SetMode(vfo model.VFOID, mode model.Mode) {
	c.doInLoop(func() {
		err := c.client.SetMode(c.vfos[vfo], toClientMode(mode), 0)
		if err != nil {
			log.Printf("hamlib: cannot switch to mode: %v", err)
			return
		}
	})
}

// SetBandplan swaps the bandplan used for band lookups. Applied inside the
// client loop, matching how the loop reads the bandplan.
func (c *Client) SetBandplan(bandplan bandplan.Bandplan) {
	c.doInLoop(func() {
		c.bandplan = bandplan
	})
}

func (c *Client) SetIncrementalTuning(vfo model.VFOID, kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	function := hl.XITFunction
	setOffset := c.client.SetXIT
	if kind == model.RIT {
		function = hl.RITFunction
		setOffset = c.client.SetRIT
	}
	c.doInLoop(func() {
		lastActive, lastOffset := c.lastState[vfo].xitActive, c.lastState[vfo].xitOffset
		if kind == model.RIT {
			lastActive, lastOffset = c.lastState[vfo].ritActive, c.lastState[vfo].ritOffset
		}
		if active == lastActive && offset == lastOffset {
			return
		}

		if active != lastActive {
			err := c.client.SetFunc(c.vfos[vfo], function, active)
			if err != nil {
				log.Printf("hamlib: cannot set %s function: %v", kind, err)
				return
			}
		}

		if active && (offset != lastOffset) {
			err := setOffset(c.vfos[vfo], hl.Frequency(offset))
			if err != nil {
				log.Printf("hamlib: cannot set %s offset: %v", kind, err)
				return
			}
		}
	})
}

func (c *Client) Refresh() {
	c.doInLoop(func() {
		for vfo := range model.VFOCount {
			c.emitChangeNotifications(vfo, vfoState{}, c.lastState[vfo])
		}
	})
}

func (c *Client) Speed(speed int) {
	c.doInLoop(func() {
		err := c.client.SetLevel(hl.CurrVFO, hl.KeyerSpeedLevel, float64(speed))
		if err != nil {
			log.Printf("hamlib: cannot set morse speed: %v", err)
		}
	})
}

func (c *Client) Send(text string) {
	c.doInLoop(func() {
		err := c.client.SendMorse(text)
		if err != nil {
			log.Printf("hamlib: cannot send morse text: %v", err)
		}
	})
}

func (c *Client) Abort() {
	c.doInLoop(func() {
		err := c.client.StopMorse()
		if err != nil {
			log.Printf("hamlib: cannot stop morse transmission: %v", err)
		}
	})
}

func (c *Client) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) emitConnectionChanged(connected bool) {
	model.Emit(c.listeners, func(listener model.ConnectionChangedListener) {
		listener.ConnectionChanged(connected)
	})
}

func (c *Client) emitCurrentVFOChanged(last, current hl.VFO) {
	if last == current {
		return
	}

	vfoID, ok := c.toVFOID(current)
	if !ok {
		log.Printf("VFO change ignored: %s -> %s", last, current)
		return
	}
	log.Printf("current vfo changed: %d = %s", vfoID, c.vfos[vfoID])

	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.CurrentVFOListener) {
			listener.CurrentVFOChanged(vfoID)
		})
	})
}

func (c *Client) emitChangeNotifications(vfo model.VFOID, last, current vfoState) {
	c.asyncRunner(func() {
		if last.frequency != current.frequency {
			c.emitFrequencyChanged(vfo, current.frequency)
		}
		if last.band != current.band {
			c.emitBandChanged(vfo, current.band)
		}
		if last.mode != current.mode {
			c.emitModeChanged(vfo, current.mode)
		}
		c.emitIncrementalTuningChanges(vfo, model.XIT, last.xitActive, last.xitOffset, current.xitActive, current.xitOffset)
		c.emitIncrementalTuningChanges(vfo, model.RIT, last.ritActive, last.ritOffset, current.ritActive, current.ritOffset)
	})
}

func (c *Client) emitIncrementalTuningChanges(vfo model.VFOID, kind model.IncrementalTuningKind, lastActive bool, lastOffset model.Frequency, currentActive bool, currentOffset model.Frequency) {
	if (lastActive != currentActive) || (currentActive && (lastOffset != currentOffset)) {
		c.emitIncrementalTuningChanged(vfo, kind, currentActive, currentOffset)
	}
}

func (c *Client) emitFrequencyChanged(vfo model.VFOID, frequency model.Frequency) {
	model.Emit(c.listeners, func(listener model.VFOFrequencyListener) {
		listener.VFOFrequencyChanged(vfo, frequency)
	})
}

func (c *Client) emitBandChanged(vfo model.VFOID, band model.Band) {
	model.Emit(c.listeners, func(listener model.VFOBandListener) {
		listener.VFOBandChanged(vfo, band)
	})
}

func (c *Client) emitModeChanged(vfo model.VFOID, mode model.Mode) {
	model.Emit(c.listeners, func(listener model.VFOModeListener) {
		listener.VFOModeChanged(vfo, mode)
	})
}

func (c *Client) emitIncrementalTuningChanged(vfo model.VFOID, kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	model.Emit(c.listeners, func(listener model.VFOIncrementalTuningListener) {
		listener.VFOIncrementalTuningChanged(vfo, kind, active, offset)
	})
}

func (c *Client) emitPTTChanged(vfo model.VFOID, active bool) {
	model.Emit(c.listeners, func(listener model.VFOPTTListener) {
		listener.VFOPTTChanged(vfo, active)
	})
}

func (c *Client) emitTXVFOChanged(vfo model.VFOID) {
	c.asyncRunner(func() {
		model.Emit(c.listeners, func(listener model.TXVFOListener) {
			listener.TXVFOChanged(vfo)
		})
	})
}

func toCoreBand(bandName bandplan.BandName) model.Band {
	if bandName == bandplan.BandUnknown {
		return model.NoBand
	}
	return model.Band(bandName)
}

func toCoreMode(mode hl.Mode) model.Mode {
	switch mode {
	case hl.ModeUSB, hl.ModeLSB:
		return model.ModeSSB
	case hl.ModeCW, hl.ModeCWR:
		return model.ModeCW
	case hl.ModeRTTY, hl.ModeRTTYR:
		return model.ModeRTTY
	case hl.ModeFM, hl.ModeWFM:
		return model.ModeFM
	case hl.ModePKTLSB, hl.ModePKTUSB, hl.ModePKTFM, hl.ModeECSSLSB, hl.ModeECSSUSB, hl.ModeSAM, hl.ModeSAL, hl.ModeSAH:
		return model.ModeDigital
	default:
		return model.NoMode
	}
}

func toClientMode(mode model.Mode) hl.Mode {
	switch mode {
	case model.ModeCW:
		return hl.ModeCW
	case model.ModeSSB:
		return hl.ModeUSB // TODO make this dependent of the current frequency either LSB or USB
	case model.ModeFM:
		return hl.ModeFM
	case model.ModeRTTY:
		return hl.ModeRTTY
	case model.ModeDigital:
		return hl.ModePKTUSB
	default:
		return ""
	}
}

func findModePortionCenter(bp bandplan.Bandplan, f int, mode bandplan.Mode) model.Frequency {
	log.Printf("find mode portion center: %d %s", f, mode)
	frequency := hamradio.Frequency(f)
	band := bp.ByFrequency(frequency)
	var modePortion bandplan.Portion
	var currentPortion bandplan.Portion
	for _, portion := range band.Portions {
		if (portion.Mode == mode && portion.From < frequency) || modePortion.Mode != mode {
			modePortion = portion
		}
		if portion.Contains(frequency) {
			currentPortion = portion
		}
		if modePortion.Mode == mode && currentPortion.Mode != "" {
			break
		}
	}
	if currentPortion.Mode == mode {
		return model.Frequency(currentPortion.Center())
	}
	if modePortion.Mode == mode {
		return model.Frequency(modePortion.Center())
	}
	return model.Frequency(band.Center())
}

func toBandplanMode(mode model.Mode) bandplan.Mode {
	switch mode {
	case model.ModeCW:
		return bandplan.ModeCW
	case model.ModeSSB, model.ModeFM:
		return bandplan.ModePhone
	case model.ModeDigital, model.ModeRTTY:
		return bandplan.ModeDigital
	default:
		return bandplan.ModeDigital
	}
}

func toBandplanBandName(band model.Band) bandplan.BandName {
	if band == model.NoBand {
		return bandplan.BandUnknown
	}
	return bandplan.BandName(band)
}
