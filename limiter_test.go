package limiter

import (
	"log"
	"testing"
	"time"
)

// TestExample should take a bit more than 400ms (less than 440ms too)
func TestTime(t *testing.T) {
	// params
	nbJobs := 16
	nbConc := 4
	jobTime := time.Millisecond * 100

	start := time.Now()
	limit := New(nbConc)
	for i := 0; i < nbJobs; i++ {
		limit.Execute(func() {
			time.Sleep(jobTime)
		})
	}
	limit.Wait()
	duration := time.Now().Sub(start)
	expected := time.Duration(jobTime) * time.Duration(nbJobs) / time.Duration(nbConc)
	expectedTenPercent := time.Duration(float64(expected) * 1.1)
	if duration > expected && duration < expectedTenPercent {
		log.Println(expectedTenPercent, " > ", duration, " > ", expected)
	} else {
		t.FailNow()
	}
}
