package limiter

import (
	"testing"
	"time"
)

// TestExample should take 4 seconds
func TestTime(t *testing.T) {
	limit := New(1)
	for i := 0; i < 4; i++ {
		limit.Execute(func() {
			time.Sleep(time.Second)
		})
	}
	limit.Wait()
}
