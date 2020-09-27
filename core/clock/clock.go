package clock

import (
	"time"
)

// New returns a new clock that runs on system time.
func New() *DefaultClock {
	return new(DefaultClock)
}

// Static returns a new clock that always returns the given static value.
func Static(time time.Time) *StaticClock {
	return &StaticClock{
		time: time,
	}
}

type DefaultClock struct{}

func (*DefaultClock) Now() time.Time {
	return time.Now()
}

type StaticClock struct {
	time time.Time
}

func (c *StaticClock) Now() time.Time {
	return c.time
}
