package clock

import (
	"time"

	"github.com/ftl/hellocontest/core"
)

// New returns a new clock that runs on system time.
func New() core.Clock {
	return new(defaultClock)
}

// Static returns a new clock that always returns the given static value.
func Static(time time.Time) core.Clock {
	return &staticClock{
		time: time,
	}
}

type defaultClock struct{}

func (defaultClock) Now() time.Time {
	return time.Now()
}

type staticClock struct {
	time time.Time
}

func (clock staticClock) Now() time.Time {
	return clock.time
}
