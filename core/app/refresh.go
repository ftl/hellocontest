package app

import "time"

func newRefreshTicker(callback func()) *refreshTicker {
	return &refreshTicker{
		callback: callback,
	}
}

type refreshTicker struct {
	callback   func()
	ticker     *time.Ticker
	stopTicker chan struct{}
}

func (t *refreshTicker) Start() {
	if t.ticker != nil {
		return
	}

	time.Sleep(tilNextSecond(time.Now()))

	t.ticker = time.NewTicker(1 * time.Second)
	t.stopTicker = make(chan struct{})
	go func() {
		for {
			select {
			case <-t.stopTicker:
				return
			case <-t.ticker.C:
				if t.callback != nil {
					t.callback()
				}
			}
		}
	}()
}

func tilNextSecond(now time.Time) time.Duration {
	currentSecond := now.Truncate(time.Second)
	return currentSecond.Add(1 * time.Second).Sub(now)
}

func (t *refreshTicker) Stop() {
	t.ticker.Stop()
	t.ticker = nil
	close(t.stopTicker)
	t.stopTicker = nil
}
