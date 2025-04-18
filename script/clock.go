package script

import (
	"fmt"
	"log"
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
	log.Printf("Clock offset set to %v, current time is %v", c.offset, c.Now())
}

func (c *Clock) SetFromRFC3339(s string) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(fmt.Errorf("cannot set script clock: %w", err))
	}
	c.Set(t)
}

func (c *Clock) Add(offset time.Duration) {
	c.offset += offset
	log.Printf("Clock offset adjusted to %v, current time is %v", c.offset, c.Now())
}

func (c *Clock) Reset() {
	c.offset = 0
	log.Printf("Clock offset reset, current time is %v", c.Now())
}
