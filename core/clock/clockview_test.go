package clock

import (
	"testing"
	"time"
)

type fakeView struct {
	lastTime string
}

func (v *fakeView) Show()            {}
func (v *fakeView) SetTime(t string) { v.lastTime = t }

func TestRefreshFormatsUTCWithSeconds(t *testing.T) {
	// non-UTC zone to prove the controller converts to UTC
	loc := time.FixedZone("CEST", 2*60*60)
	now := time.Date(2026, 7, 24, 15, 4, 5, 0, loc)

	view := &fakeView{}
	c := NewController(Static(now), func(f func()) { f() })
	c.SetView(view)

	if view.lastTime != "13:04:05" {
		t.Errorf("expected 13:04:05, got %q", view.lastTime)
	}
}

func TestSetViewNilFallsBackToNullView(t *testing.T) {
	c := NewController(Static(time.Now()), func(f func()) { f() })
	c.SetView(nil)
	c.refresh() // must not panic
}
