package caches

import "time"

func FutureFunc(milliseconds int64, f func()) *Timer {
	return futureTimer(milliseconds, func(*Timer) { f() })
}

// futureTimer is like FutureFunc but passes the Timer to f, so the callback can
// identify which timer fired without racing on a captured variable.
func futureTimer(milliseconds int64, f func(*Timer)) *Timer {
	future := time.Now().Add(time.Duration(milliseconds) * time.Millisecond)
	t := &Timer{Expiration: future.Unix()}
	t.timer = time.AfterFunc(time.Until(future), func() { f(t) })
	return t
}

type Timer struct {
	timer      *time.Timer
	Expiration int64 `json:"expiration"` // Unix timestamp
}

func (t *Timer) Stop() bool {
	return t.timer.Stop()
}
