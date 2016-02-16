package safetime

import (
	"sync"
	"time"
)

type Time struct {
	sync.Mutex
	tm time.Time
}

func Since(t Time) time.Duration {
	return time.Now().Sub(t.tm)
}

func (t *Time) SetNow() {
	t.Lock()
	defer t.Unlock()
	t.tm = time.Now()
}

type Timer struct {
	sync.Mutex
	time.Timer
}

func NewTimer(d time.Duration) *Timer {
	return &Timer{sync.Mutex{}, *time.NewTimer(d)}
}

func (tmr *Timer) Reset(d time.Duration) bool {
	tmr.Lock()
	defer tmr.Unlock()
	return tmr.Reset(d)
}

func (tmr *Timer) Stop() bool {
	tmr.Lock()
	defer tmr.Unlock()
	return tmr.Stop()
}
