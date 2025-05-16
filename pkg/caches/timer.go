package caches

import "time"

func FutureFunc(milliseconds int64, f func()) *Timer {
	future := time.Now().Add(time.Duration(milliseconds) * time.Millisecond)
	return &Timer{
		Expiration: future.Unix(),
		timer:      time.AfterFunc(time.Until(future), f),
	}
}

type Timer struct {
	timer      *time.Timer
	Expiration int64 `json:"expiration"` // Unix timestamp
}

func (t *Timer) Stop() bool {
	return t.timer.Stop()
}
