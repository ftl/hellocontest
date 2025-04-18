package script

import (
	"fmt"
	"time"
)

type Clock struct {
	offset time.Duration
}

func (c *Clock) Now() time.Time {
	return time.Now().Add(c.offset)
}

func (c *Clock) Set(now time.Time) {
	c.offset = now.Sub(time.Now())
}

func (c *Clock) SetFromRFC3339(s string) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(fmt.Errorf("cannot set script clock: %w", err))
	}
	c.Set(t)
}

func (c *Clock) SetMinute(minute int) {
	now := c.Now()
	offset := time.Duration(minute-now.Minute()) * time.Minute
	c.Add(offset)
}

func (c *Clock) Add(offset time.Duration) {
	c.offset += offset
}

func (c *Clock) Reset() {
	c.offset = 0
}
