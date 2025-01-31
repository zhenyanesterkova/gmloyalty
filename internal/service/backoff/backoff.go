package backoff

import (
	"time"
)

const (
	defaultStep = 2
)

type Backoff struct {
	minInt     time.Duration
	maxInt     time.Duration
	maxAttempt int
	attemptNum int
	nextDelay  time.Duration
}

func New(minInt, maxInt time.Duration, maxAttempt int) *Backoff {
	return &Backoff{
		minInt:     minInt,
		maxInt:     maxInt,
		maxAttempt: maxAttempt,
		nextDelay:  minInt,
	}
}

const Stop time.Duration = -1

func (b *Backoff) Next() time.Duration {
	if b.attemptNum >= b.maxAttempt {
		return Stop
	}
	b.attemptNum++
	delay := min(b.nextDelay, b.maxInt)
	b.nextDelay += defaultStep * time.Second
	return delay
}

func (b *Backoff) Reset() {
	b.attemptNum = 0
	b.nextDelay = b.minInt
}
