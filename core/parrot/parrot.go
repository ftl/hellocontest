package parrot

import (
	"sync"
	"time"

	"github.com/ftl/hellocontest/core"
)

const (
	DefaultInterval = 10 * time.Second
	tickInterval    = 1 * time.Second

	CQMessageIndex = 0
)

type WorkmodeController interface {
	SetWorkmode(workmode core.Workmode)
}

type Keyer interface {
	Send(index int)
	Stop()
}

type ParrotActiveListener interface {
	ParrotActive(active bool)
}

type ParrotTimeListener interface {
	ParrotTimeLeft(timeLeft time.Duration)
}

type View interface {
	SetParrotActive(active bool)
}

func New(workmodeController WorkmodeController, keyer Keyer, runAsync core.AsyncRunner) *Parrot {
	ticker := time.NewTicker(tickInterval)
	ticker.Stop()
	result := &Parrot{
		workmodeController: workmodeController,
		keyer:              keyer,
		runAsync:           runAsync,
		active:             false,
		interval:           DefaultInterval,
		ticker:             ticker,
		tickLock:           &sync.Mutex{},
	}

	go result.run()

	return result
}

// Parrot is a CQ 🦜
type Parrot struct {
	view               View
	workmodeController WorkmodeController
	keyer              Keyer
	runAsync           core.AsyncRunner

	active    bool
	interval  time.Duration
	remaining time.Duration
	ticker    *time.Ticker
	tickLock  *sync.Mutex // log for interval and remaining

	listeners []any
}

func (p *Parrot) run() {
	for range p.ticker.C {
		p.tickLock.Lock()
		p.remaining -= tickInterval
		remaining := p.remaining
		if p.remaining <= 0 {
			p.remaining = p.interval
		}
		p.tickLock.Unlock()

		p.runAsync(func() {
			if remaining >= 0 {
				p.emitParrotTimeLeft(remaining)
			}
			if remaining <= 0 {
				p.keyer.Send(CQMessageIndex)
			}
		})
	}
}

func (p *Parrot) setActive(active bool) {
	p.active = active
	if p.view != nil {
		p.view.SetParrotActive(p.active)
	}
	p.emitParrotActive(p.active)
}

func (p *Parrot) Notify(listener any) {
	p.listeners = append(p.listeners, listener)
}

func (p *Parrot) SetView(view View) {
	if view == nil {
		panic("parrot.Parrot.SetView must not be called with nil")
	}
	if p.view != nil {
		panic("parrot.Parrot.SetView was already called")
	}

	p.view = view
	view.SetParrotActive(p.active)
}

func (p *Parrot) SetInterval(interval time.Duration) {
	p.tickLock.Lock()
	defer p.tickLock.Unlock()
	p.interval = interval
}

func (p *Parrot) resetTick() {
	p.tickLock.Lock()
	defer p.tickLock.Unlock()
	p.remaining = 0
	p.emitParrotTimeLeft(0)
}

func (p *Parrot) Start() {
	if p.active {
		return
	}
	p.setActive(true)

	p.workmodeController.SetWorkmode(core.Run)

	p.resetTick()
	p.ticker.Reset(tickInterval)
}

func (p *Parrot) Stop() {
	if !p.active {
		return
	}
	p.setActive(false)

	p.resetTick()
	p.ticker.Stop()
}

func (p *Parrot) CallsignEntered(call string) {
	if !p.active {
		return
	}

	p.keyer.Stop()
}

func (p *Parrot) KeyerStopped() {
	if !p.active {
		return
	}

	p.Stop()
}

func (p *Parrot) WorkmodeChanged(workmode core.Workmode) {
	if p.active && workmode != core.Run {
		p.keyer.Stop()
	}
}

func (p *Parrot) emitParrotActive(active bool) {
	for _, listener := range p.listeners {
		if parrotActiveListener, ok := listener.(ParrotActiveListener); ok {
			parrotActiveListener.ParrotActive(active)
		}
	}
}

func (p *Parrot) emitParrotTimeLeft(timeLeft time.Duration) {
	for _, listener := range p.listeners {
		if l, ok := listener.(ParrotTimeListener); ok {
			l.ParrotTimeLeft(timeLeft)
		}
	}
}
