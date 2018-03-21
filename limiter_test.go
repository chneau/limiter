package limiter

import (
	"testing"
	"time"
)

// TestExample should take 4 seconds
func TestTime(t *testing.T) {
	limit := New(4)
	for i := 0; i < 16; i++ {
		limit.Execute(func() {
			time.Sleep(time.Second)
		})
	}
	limit.Wait()
}
