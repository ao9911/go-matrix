package backoff

import (
	"math"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
)

// Backoff interface defines contract for backoff strategies
type Backoff interface {
	Next(retry int) time.Duration
}

type constantBackoff struct {
	backoffInterval xtime.Duration
}

// NewConstantBackoff returns an instance of ConstantBackoff
func NewConstantBackoff(backoffInterval xtime.Duration) Backoff {
	return &constantBackoff{backoffInterval: backoffInterval}
}

// Next returns next time for retrying operation with constant strategy
func (cb *constantBackoff) Next(retry int) time.Duration {
	if retry <= 0 {
		return 0
	}
	return time.Duration(cb.backoffInterval)
}

type exponentialBackoff struct {
	exponentFactor float64
	initialTimeout float64
	maxTimeout     float64
}

// NewExponentialBackoff returns an instance of ExponentialBackoff
func NewExponentialBackoff(initialTimeout, maxTimeout time.Duration, exponentFactor float64) Backoff {
	return &exponentialBackoff{
		exponentFactor: exponentFactor,
		initialTimeout: float64(initialTimeout / time.Millisecond),
		maxTimeout:     float64(maxTimeout / time.Millisecond),
	}
}

// Next returns next time for retrying operation with exponential strategy
func (eb *exponentialBackoff) Next(retry int) time.Duration {
	if retry <= 0 {
		return 0 * time.Millisecond
	}
	next := eb.initialTimeout * math.Pow(eb.exponentFactor, float64(retry))
	if next > eb.maxTimeout {
		next = eb.maxTimeout
	}
	return time.Duration(next) * time.Millisecond
}
