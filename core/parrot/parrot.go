package parrot

import (
	"time"

	"github.com/ftl/hellocontest/core"
)

const DefaultInterval = 10 * time.Second
const CQMessageIndex = 0

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

type View interface {
	SetParrotActive(active bool)
}

func New(workmodeController WorkmodeController, keyer Keyer) *Parrot {
	ticker := time.NewTicker(DefaultInterval)
	ticker.Stop()
	result := &Parrot{
		active:             false,
		interval:           DefaultInterval,
		ticker:             ticker,
		workmodeController: workmodeController,
		keyer:              keyer,
	}

	go result.run()

	return result
}

// Parrot is a CQ 🦜
type Parrot struct {
	view               View
	active             bool
	interval           time.Duration
	ticker             *time.Ticker
	workmodeController WorkmodeController
	keyer              Keyer

	listeners []any
}

func (p *Parrot) run() {
	for range p.ticker.C {
		p.keyer.Send(CQMessageIndex)
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
	p.interval = interval
	if p.active {
		p.ticker.Reset(p.interval)
	}
}

func (p *Parrot) Start() {
	if p.active {
		return
	}
	p.setActive(true)

	p.workmodeController.SetWorkmode(core.Run)
	p.keyer.Send(CQMessageIndex)
	p.ticker.Reset(p.interval)
}

func (p *Parrot) Stop() {
	if !p.active {
		return
	}
	p.setActive(false)

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
