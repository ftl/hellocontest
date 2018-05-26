package core

import "time"

// Clock represents a source of the current time.
type Clock interface {
	Now() time.Time
}

// NewClock returns a new clock.
func NewClock() Clock {
	return defaultClock{}
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
