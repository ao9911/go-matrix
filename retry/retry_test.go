package retry

import (
	"testing"
	"time"

	"github.com/ao9911/go-matrix/backoff"
	"github.com/ao9911/go-matrix/util/xtime"
	"github.com/pkg/errors"
)

// go test -v -test.run TestRetrier_Do
func TestRetrier_Do(t *testing.T) {
	bo := backoff.NewConstantBackoff(xtime.Duration(100 * time.Millisecond))
	err := NewRetrier(bo).Do(HelloDo, 5)
	t.Log(err)
}

func HelloDo() (err error) {
	err = errors.New("retry testing")
	return
}
