package internal

import (
	"sync/atomic"
	"time"
)

// SpeedCounter count the speed
type SpeedCounter struct {
	total     int64
	lastCount int64
	lastTime  time.Time
	Lock
}

// NewSpeedCounter ...
func NewSpeedCounter() *SpeedCounter {
	return &SpeedCounter{
		total:     0,
		lastCount: 0,
		lastTime:  time.Now(),
	}
}

// Count ...
func (p *SpeedCounter) Count() int64 {
	return atomic.AddInt64(&p.total, 1)
}

// Total ...
func (p *SpeedCounter) Total() int64 {
	return atomic.LoadInt64(&p.total)
}

// Calculate ...
func (p *SpeedCounter) Calculate(
	now time.Time,
) (speed int64, duration time.Duration) {
	p.DoWithLock(func() {
		deltaCount := atomic.LoadInt64(&p.total) - p.lastCount
		deltaTime := now.Sub(p.lastTime)

		if deltaTime <= 0 {
			speed = 0
			duration = 0
		} else if deltaCount < 0 {
			speed = 0
			duration = 0
		} else {
			speed = (deltaCount * int64(time.Second)) / int64(deltaTime)
			duration = deltaTime
			p.lastCount += deltaCount
			p.lastTime = now
		}
	})

	return
}
