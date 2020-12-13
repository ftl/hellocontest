package rate

import (
	"fmt"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/ticker"
)

type RateUpdatedListener interface {
	RateUpdated(core.QSORate)
}

type RateUpdatedListenerFunc func(core.QSORate)

func (f RateUpdatedListenerFunc) RateUpdated(Score core.QSORate) {
	f(Score)
}

func NewCounter() *Counter {
	result := &Counter{
		QSORate: core.QSORate{
			QSOsPerHours: make(core.QSOsPerHours),
		},
		view: new(nullView),
	}
	result.refreshTicker = ticker.New(result.Refresh)
	result.refreshTicker.Start()
	return result
}

type Counter struct {
	core.QSORate
	view View

	listeners []interface{}

	lastHourQSOs  qsoList
	lastQSOTime   time.Time
	refreshTicker *ticker.Ticker
}

var zeroTime time.Time

type View interface {
	Show()
	Hide()

	ShowRate(rate core.QSORate)
}

func (c *Counter) SetView(view View) {
	if view == nil {
		c.view = new(nullView)
		return
	}
	c.view = view
	c.view.ShowRate(c.QSORate)
}

func (c *Counter) Show() {
	c.view.Show()
	c.view.ShowRate(c.QSORate)
}

func (c *Counter) Hide() {
	c.view.Hide()
}

func (c *Counter) Notify(listener interface{}) {
	c.listeners = append(c.listeners, listener)
}

func (c *Counter) Clear() {
	c.lastHourQSOs.Clear()
	c.lastQSOTime = zeroTime

	c.LastHourRate = 0
	c.Last5MinRate = 0
	c.QSOsPerHours = make(core.QSOsPerHours)
}

func (c *Counter) Refresh() {
	now := time.Now()
	c.lastHourQSOs.RemoveBefore(now.Add(-1 * time.Hour))
	c.LastHourRate = core.QSOsPerHour(c.lastHourQSOs.Length())
	c.Last5MinRate = core.QSOsPerHour(c.lastHourQSOs.LengthAfter(now.Add(-5*time.Minute)) * 12)
	if c.lastQSOTime.IsZero() {
		c.SinceLastQSO = 0
	} else {
		c.SinceLastQSO = now.Sub(c.lastQSOTime)
	}
	c.emitRateUpdated(c.QSORate)
}

func (c *Counter) emitRateUpdated(rate core.QSORate) {
	c.view.ShowRate(rate)
	for _, listener := range c.listeners {
		if rateUpdatedListener, ok := listener.(RateUpdatedListener); ok {
			rateUpdatedListener.RateUpdated(rate)
		}
	}
}

func (c *Counter) Add(qso core.QSO) {
	if c.lastQSOTime.Before(qso.Time) {
		c.lastQSOTime = qso.Time
	}
	c.lastHourQSOs.Add(qso)

	hour := core.HourOf(qso.Time)
	qsosPerHour := c.QSOsPerHours[hour]
	c.QSOsPerHours[hour] = qsosPerHour + 1

	c.Refresh()
}

func (c *Counter) Update(oldQSO, newQSO core.QSO) {
	if oldQSO.Time == newQSO.Time {
		return
	}
	c.lastHourQSOs.RemoveQSO(oldQSO)
	c.lastHourQSOs.Add(newQSO)

	oldHour := core.HourOf(oldQSO.Time)
	if qsosPerOldHour, ok := c.QSOsPerHours[oldHour]; ok {
		c.QSOsPerHours[oldHour] = qsosPerOldHour - 1
	}
	newHour := core.HourOf(newQSO.Time)
	if qsosPerNewHour, ok := c.QSOsPerHours[newHour]; ok {
		c.QSOsPerHours[newHour] = qsosPerNewHour - 1
	}

	c.Refresh()
}

type qsoList struct {
	first *qsoListEntry
	last  *qsoListEntry
}

func (f *qsoList) Empty() bool {
	return f.first == nil
}

func (f *qsoList) Clear() {
	f.first = nil
	f.last = nil
}

func (f *qsoList) Add(qso core.QSO) {
	entry := &qsoListEntry{QSO: qso}
	if f.first == nil {
		f.first = entry
		f.last = entry
		return
	}

	for current := f.last; current != nil; current = current.Previous {
		if current.QSO.Time.Before(entry.QSO.Time) {
			entry.Next = current.Next
			if entry.Next != nil {
				entry.Next.Previous = entry
			}
			entry.Previous = current
			current.Next = entry
			if f.last == current {
				f.last = entry
			}
			return
		}
	}

	if f.first != nil {
		f.first.Previous = entry
	}
	entry.Next = f.first
	f.first = entry
}

func (f *qsoList) forward(do func(e *qsoListEntry) bool) {
	for e := f.first; e != nil; e = e.Next {
		if !do(e) {
			return
		}
	}
}

func (f *qsoList) backward(do func(e *qsoListEntry) bool) {
	for e := f.last; e != nil; e = e.Previous {
		if !do(e) {
			return
		}
	}
}

func (f *qsoList) remove(e *qsoListEntry) {
	if e.Previous != nil {
		e.Previous.Next = e.Next
	}
	if e.Next != nil {
		e.Next.Previous = e.Previous
	}
	if f.first == e {
		f.first = e.Next
	}
	if f.last == e {
		f.last = e.Previous
	}
}

func (f *qsoList) RemoveBefore(t time.Time) {
	f.forward(func(e *qsoListEntry) bool {
		if e.QSO.Time.Before(t) {
			f.remove(e)
			return true
		}
		return false
	})
}

func (f *qsoList) RemoveQSO(qso core.QSO) {
	if f.first == nil {
		return
	}

	f.forward(func(e *qsoListEntry) bool {
		if e.QSO.MyNumber == qso.MyNumber {
			f.remove(e)
		}
		return true
	})
}

func (f *qsoList) Length() int {
	length := 0
	f.backward(func(e *qsoListEntry) bool {
		length++
		return true
	})
	return length
}

func (f *qsoList) LengthAfter(t time.Time) int {
	length := 0
	f.backward(func(e *qsoListEntry) bool {
		if e.QSO.Time.After(t) {
			length++
			return true
		}
		return false
	})
	return length
}

type qsoListEntry struct {
	QSO      core.QSO
	Previous *qsoListEntry
	Next     *qsoListEntry
}

func (e qsoListEntry) String() string {
	return fmt.Sprintf("qso %d", e.QSO.MyNumber)
}

type nullView struct{}

func (v *nullView) Show()                      {}
func (v *nullView) Hide()                      {}
func (v *nullView) ShowRate(rate core.QSORate) {}
