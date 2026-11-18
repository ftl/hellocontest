package vfo

import (
	"log"
	"sync"
	"time"

	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/hellocontest/core/model"
)

// pendingFrequencyTimeout defines how long a commanded frequency is used as the
// base for subsequent frequency changes while the rig has not yet confirmed it.
// After this time without a new command, the rig is trusted again.
const pendingFrequencyTimeout = 1 * time.Second

type Client interface {
	Notify(any)
	Active() bool
	Refresh()
	SetCurrentVFO(model.VFOID)
	SetTXVFO(model.VFOID)
	SetFrequency(model.VFOID, model.Frequency)
	SetBand(model.VFOID, model.Band)
	SetMode(model.VFOID, model.Mode)
	IncrementalTuningPerVFO() bool
	SetIncrementalTuning(model.VFOID, model.IncrementalTuningKind, bool, model.Frequency)
}

type Logbook interface {
	LastBand() model.Band
	LastMode() model.Mode
}

type VFO struct {
	id   model.VFOID
	name string

	clock         model.Clock
	bandplan      bandplan.Bandplan
	logbook       Logbook
	client        Client
	offlineClient *offlineClient
	refreshing    bool
	asyncRunner   model.AsyncRunner

	state    [2]incrementalTuningState
	workmode model.Workmode

	pendingFrequency model.Frequency
	pendingSince     time.Time
	staleFrequency   model.Frequency
	commandedSince   time.Time
	awaitingSettle   bool

	listeners []any
}

func NewVFO(id model.VFOID, name string, clock model.Clock, bandplan bandplan.Bandplan, logbook Logbook, asyncRunner model.AsyncRunner) *VFO {
	result := &VFO{
		id:          id,
		name:        name,
		clock:       clock,
		bandplan:    bandplan,
		logbook:     logbook,
		asyncRunner: asyncRunner,
	}
	result.offlineClient = newOfflineClient(result)
	result.SetClient(nil)

	return result
}

// SetBandplan swaps the bandplan used for band lookups. Subsequent lookups use
// the new plan.
func (v *VFO) SetBandplan(bandplan bandplan.Bandplan) {
	v.bandplan = bandplan
}

func (v *VFO) SetClient(client Client) {
	v.client = client
	if client != nil {
		v.client.Notify(v)
	}
}

func (v *VFO) Name() string {
	return v.name
}

func (v *VFO) Notify(listener any) {
	v.listeners = append(v.listeners, listener)
}

func (v *VFO) online() bool {
	return v.client != nil && v.client.Active()
}

func (v *VFO) Refresh() {
	v.refreshing = true
	defer func() { v.refreshing = false }()

	if !v.online() {
		v.offlineClient.Refresh()
		return
	}
	v.client.Refresh()
	v.emitIncrementalTuningVisibilityChanged(model.RIT)
	v.emitIncrementalTuningVisibilityChanged(model.XIT)
}

func (v *VFO) Frequency() model.Frequency {
	// a command that the rig did not confirm yet already names the frequency of the
	// operator, therefore the base frequency answers this question
	return v.baseFrequency(v.clock.Now())
}

func (v *VFO) SetFrequency(frequency model.Frequency) {
	v.setFrequency(frequency, v.clock.Now())
}

func (v *VFO) setFrequency(frequency model.Frequency, now time.Time) {
	// the application jumps to another frequency, therefore the reports of the rig that
	// follow belong to this command and are no tuning of the operator
	v.staleFrequency = v.currentFrequency()
	v.commandedSince = now
	v.awaitingSettle = true

	v.sendFrequency(frequency, now)
}

func (v *VFO) sendFrequency(frequency model.Frequency, now time.Time) {
	v.pendingFrequency = frequency
	v.pendingSince = now

	if v.online() {
		v.client.SetFrequency(v.id, frequency)
	} else {
		v.offlineClient.SetFrequency(frequency)
	}
}

// ShiftFrequency changes the current frequency by delta (may be negative).
func (v *VFO) ShiftFrequency(delta model.Frequency) {
	v.shiftFrequency(delta, v.clock.Now())
}

func (v *VFO) shiftFrequency(delta model.Frequency, now time.Time) {
	// a dial or the keyboard shifts the frequency, which is a tuning of the operator like
	// turning the knob of the rig. It also ends a command that is still in flight.
	v.commandedSince = time.Time{}
	v.awaitingSettle = false

	v.sendFrequency(v.baseFrequency(now)+delta, now)
}

// baseFrequency returns the frequency that subsequent frequency changes are
// based on. As long as the rig has not confirmed the last commanded frequency,
// that frequency is used instead of the last frequency reported by the rig.
// Otherwise a rig that cannot keep up with fast changes would make the
// frequency jump back and forth.
func (v *VFO) baseFrequency(now time.Time) model.Frequency {
	if v.frequencyCommandPending(now) {
		return v.pendingFrequency
	}
	return v.currentFrequency()
}

// clearPendingFrequency makes the next frequency change use the frequency
// reported by the rig again.
func (v *VFO) clearPendingFrequency() {
	v.pendingSince = time.Time{}
}

// currentFrequency returns the last known frequency of this VFO. The offline
// client is kept in sync with the rig via VFOFrequencyChanged, so it holds the
// current frequency in both online and offline mode.
func (v *VFO) currentFrequency() model.Frequency {
	c := v.offlineClient
	c.stateLock.RLock()
	defer c.stateLock.RUnlock()
	return c.lastState(c.currentBand).frequency
}

func (v *VFO) SetBand(band model.Band) {
	v.clearPendingFrequency()
	if v.online() {
		v.client.SetBand(v.id, band)
	} else {
		v.offlineClient.SetBand(band)
	}
}

func (v *VFO) SetMode(mode model.Mode) {
	if v.online() {
		v.client.SetMode(v.id, mode)
	} else {
		v.offlineClient.SetMode(mode)
	}
}

func (v *VFO) LogbookLoaded() {
	log.Printf("VFO logbook changed")

	// only the offline client follows the logbook, a connected rig keeps its own band and mode
	if !v.online() {
		lastBand := v.logbook.LastBand()
		if lastBand != model.NoBand {
			v.offlineClient.SetBand(lastBand)
		}

		lastMode := v.logbook.LastMode()
		if lastMode != model.NoMode {
			v.offlineClient.SetMode(lastMode)
		}
	}

	// the listeners reset their state when the logbook changes, they need the current state again
	v.Refresh()
}

func (v *VFO) VFOFrequencyChanged(vfo model.VFOID, frequency model.Frequency) {
	v.vfoFrequencyChanged(vfo, frequency, v.clock.Now())
}

func (v *VFO) vfoFrequencyChanged(vfo model.VFOID, frequency model.Frequency, now time.Time) {
	if vfo != v.id {
		return
	}

	if v.staleReport(frequency, now) {
		// a poll of the rig that started before the command. Reporting it would look
		// like the operator tuned back to the old frequency.
		return
	}

	if frequency == v.pendingFrequency {
		// the rig caught up with the frequency that the application sent
		v.clearPendingFrequency()
	}

	if v.awaitingCommandedChange(now) {
		// the first report after the command is the rig that follows it. It may land next
		// to the commanded frequency, because a rig rounds to its own raster. Every
		// report after this one belongs to the operator again.
		v.awaitingSettle = false
		v.offlineClient.SetFrequency(frequency)
		return
	}

	// the application commanded no change, therefore the operator tuned the VFO. The
	// listeners learn about the tuning before they learn about the new frequency, so
	// that they can still compare it with the frequency they know.
	v.emitTuned(frequency)
	v.offlineClient.SetFrequency(frequency)
}

func (v *VFO) staleReport(frequency model.Frequency, now time.Time) bool {
	return frequency == v.staleFrequency && v.withinCommandTimeout(now)
}

func (v *VFO) awaitingCommandedChange(now time.Time) bool {
	return v.awaitingSettle && v.withinCommandTimeout(now)
}

func (v *VFO) withinCommandTimeout(now time.Time) bool {
	return !v.commandedSince.IsZero() && now.Sub(v.commandedSince) < pendingFrequencyTimeout
}

func (v *VFO) frequencyCommandPending(now time.Time) bool {
	return !v.pendingSince.IsZero() && now.Sub(v.pendingSince) < pendingFrequencyTimeout
}

func (v *VFO) VFOBandChanged(vfo model.VFOID, band model.Band) {
	if vfo != v.id {
		return
	}
	v.offlineClient.SetBand(band)
}

func (v *VFO) VFOModeChanged(vfo model.VFOID, mode model.Mode) {
	if vfo != v.id {
		return
	}
	v.offlineClient.SetMode(mode)
}

func (v *VFO) VFOPTTChanged(vfo model.VFOID, active bool) {
	if vfo != v.id {
		return
	}
	v.offlineClient.SetPTT(active)
}

func (v *VFO) emitFrequencyChanged(frequency model.Frequency) {
	model.Emit(v.listeners, func(listener model.VFOFrequencyListener) {
		v.asyncRunner(func() {
			listener.VFOFrequencyChanged(v.id, frequency)
		})
	})
}

func (v *VFO) emitTuned(frequency model.Frequency) {
	model.Emit(v.listeners, func(listener model.VFOTunedListener) {
		v.asyncRunner(func() {
			listener.VFOTuned(v.id, frequency)
		})
	})
}

func (v *VFO) emitBandChanged(band model.Band) {
	model.Emit(v.listeners, func(listener model.VFOBandListener) {
		v.asyncRunner(func() {
			listener.VFOBandChanged(v.id, band)
		})
	})
}

func (v *VFO) emitModeChanged(mode model.Mode) {
	model.Emit(v.listeners, func(listener model.VFOModeListener) {
		v.asyncRunner(func() {
			listener.VFOModeChanged(v.id, mode)
		})
	})
}

func (v *VFO) emitPTTChanged(active bool) {
	model.Emit(v.listeners, func(listener model.VFOPTTListener) {
		v.asyncRunner(func() {
			listener.VFOPTTChanged(v.id, active)
		})
	})
}

type bandState struct {
	frequency model.Frequency
	mode      model.Mode
	itActive  [2]bool
	itOffset  [2]model.Frequency
}

type offlineClient struct {
	vfo         *VFO
	currentBand model.Band
	stateLock   *sync.RWMutex
	lastStates  map[model.Band]bandState
}

func newOfflineClient(vfo *VFO) *offlineClient {
	result := &offlineClient{
		vfo:         vfo,
		currentBand: model.Band160m,
		stateLock:   &sync.RWMutex{},
		lastStates:  make(map[model.Band]bandState),
	}
	_ = result.lastState(result.currentBand)
	return result
}

func (c *offlineClient) lastState(band model.Band) bandState {
	result, ok := c.lastStates[band]
	if ok {
		return result
	}

	plan, ok := c.vfo.bandplan[bandplan.BandName(band)]
	if !ok {
		log.Printf("Band %s not found in bandplan! (1)", band)
		return bandState{}
	}

	result = bandState{
		frequency: model.Frequency(plan.From),
		mode:      model.ModeCW,
	}
	c.lastStates[band] = result

	return result
}

func (c *offlineClient) Active() bool {
	return true
}

func (c *offlineClient) Refresh() {
	c.stateLock.RLock()
	state := c.lastState(c.currentBand)
	c.stateLock.RUnlock()

	c.vfo.emitFrequencyChanged(state.frequency)
	c.vfo.emitBandChanged(c.currentBand)
	c.vfo.emitModeChanged(state.mode)
}

func (c *offlineClient) SetFrequency(frequency model.Frequency) {
	planband := c.vfo.bandplan.ByFrequency(hamradio.Frequency(frequency))
	newBand := model.Band(planband.Name)

	c.stateLock.Lock()
	state := c.lastState(newBand)
	state.frequency = frequency
	c.lastStates[newBand] = state
	c.stateLock.Unlock()

	c.vfo.emitFrequencyChanged(frequency)

	if newBand == c.currentBand {
		return
	}
	c.currentBand = newBand
	c.vfo.emitBandChanged(c.currentBand)
}

func (c *offlineClient) SetBand(band model.Band) {
	plan, ok := c.vfo.bandplan[bandplan.BandName(band)]
	if !ok {
		log.Printf("Band %s not found in bandplan (2)", band)
		return
	}
	newBand := model.Band(plan.Name)
	if newBand == c.currentBand && !c.vfo.refreshing {
		return
	}

	c.stateLock.RLock()
	state := c.lastState(newBand)
	c.stateLock.RUnlock()

	c.vfo.emitFrequencyChanged(state.frequency)

	c.currentBand = newBand
	c.vfo.emitBandChanged(c.currentBand)
}

func (c *offlineClient) SetMode(mode model.Mode) {
	c.stateLock.Lock()
	state := c.lastState(c.currentBand)
	state.mode = mode
	c.lastStates[c.currentBand] = state
	c.stateLock.Unlock()

	c.vfo.emitModeChanged(mode)
}

func (c *offlineClient) SetIncrementalTuning(kind model.IncrementalTuningKind, active bool, offset model.Frequency) {
	c.stateLock.Lock()
	state := c.lastState(c.currentBand)
	state.itActive[kind] = active
	state.itOffset[kind] = offset
	c.lastStates[c.currentBand] = state
	c.stateLock.Unlock()

	c.vfo.emitIncrementalTuningChanged(kind, state.itActive[kind], state.itOffset[kind])
}

func (c *offlineClient) SetPTT(active bool) {
	c.vfo.emitPTTChanged(active)
}

func (c *offlineClient) IncrementalTuningPerVFO() bool {
	return false
}
