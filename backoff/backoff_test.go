package backoff

import (
	"testing"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
)

// go test -v -test.run TestExponentialBackoff_Next
func TestExponentialBackoff_Next(t *testing.T) {
	// Create an instance of ExponentialBackoff with initial timeout of 100ms, max timeout of 1s, and exponent factor of 2.0
	backoff := NewExponentialBackoff(
		100*time.Millisecond,
		1*time.Second,
		2.0,
	)

	for retry := 0; retry <= 5; retry++ {
		interval := backoff.Next(retry)

		t.Logf("retry=%d interval=%v", retry, interval)
		time.Sleep(interval)
	}
}

// go test -v -test.run TestConstantBackoff_Next
func TestConstantBackoff_Next(t *testing.T) {
	backoff := NewConstantBackoff(
		xtime.Duration(time.Millisecond),
	)

	for retry := 0; retry <= 5; retry++ {
		interval := backoff.Next(retry)

		t.Logf("retry=%d interval=%v", retry, interval)
		time.Sleep(interval)
	}
}
