package clock

import (
	"time"
)

// New returns a new clock that runs on system time.
func New() *DefaultClock {
	return new(DefaultClock)
}

type DefaultClock struct{}

func (*DefaultClock) Now() time.Time {
	return time.Now()
}

// Static returns a new clock that always returns the given static value.
func Static(time time.Time) *StaticClock {
	return &StaticClock{
		time: time,
	}
}

type StaticClock struct {
	time time.Time
}

func (c *StaticClock) Now() time.Time {
	return c.time
}

func Zero() *ZeroClock {
	return &ZeroClock{}
}

type ZeroClock struct{}

func (c *ZeroClock) Now() time.Time {
	return time.Time{}
}
