package limiter

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	nbJobs := 16
	nbConc := 4
	jobTime := time.Millisecond * 50

	start := time.Now()
	limit := New(nbConc)
	if limit.Cap() != nbConc {
		t.Fatalf("expected Cap() %d, got %d", nbConc, limit.Cap())
	}
	if limit.Available() != nbConc {
		t.Fatalf("expected Available() %d, got %d", nbConc, limit.Available())
	}
	if limit.Running() != 0 {
		t.Fatalf("expected Running() 0, got %d", limit.Running())
	}

	for range nbJobs {
		limit.Execute(func() {
			time.Sleep(jobTime)
		})
	}
	limit.Wait()
	duration := time.Since(start)

	expected := time.Duration(jobTime) * time.Duration(nbJobs) / time.Duration(nbConc)
	if duration < expected || duration > expected*2 {
		t.Fatalf("expected duration around %v, got %v", expected, duration)
	}
}

func TestNewDefaults(t *testing.T) {
	limit := New(0)
	if limit.Cap() != 1 {
		t.Fatalf("expected cap 1 for New(0), got %d", limit.Cap())
	}
	limitNegative := New(-5)
	if limitNegative.Cap() != 1 {
		t.Fatalf("expected cap 1 for New(-5), got %d", limitNegative.Cap())
	}
}

func TestTryExecute(t *testing.T) {
	limit := New(2)

	started := make(chan struct{})
	release := make(chan struct{})

	ok1 := limit.TryExecute(func() {
		close(started)
		<-release
	})
	if !ok1 {
		t.Fatalf("expected first TryExecute to succeed")
	}

	<-started

	// Second slot
	ok2 := limit.TryExecute(func() {
		<-release
	})
	if !ok2 {
		t.Fatalf("expected second TryExecute to succeed")
	}

	// Third slot should fail immediately
	ok3 := limit.TryExecute(func() {
		t.Errorf("should not run")
	})
	if ok3 {
		t.Fatalf("expected third TryExecute to fail")
	}

	close(release)
	limit.Wait()
}

func TestExecuteContext(t *testing.T) {
	limit := New(1)

	release := make(chan struct{})
	limit.Execute(func() {
		<-release
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := limit.ExecuteContext(ctx, func() {
		t.Errorf("should not run")
	})
	if err == nil || err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}

	close(release)
	limit.Wait()

	// Use a fresh limiter instance to test successful ExecuteContext since Wait() consumes tickets
	limit2 := New(1)
	executed := false
	err = limit2.ExecuteContext(context.Background(), func() {
		executed = true
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	limit2.Wait()
	if !executed {
		t.Fatalf("expected job to execute")
	}
}

func TestPanicSafety(t *testing.T) {
	limit := New(1)

	var recovered any
	done := make(chan struct{})
	limit.Execute(func() {
		defer close(done)
		defer func() {
			recovered = recover()
		}()
		panic("boom")
	})

	<-done

	if recovered != "boom" {
		t.Fatalf("expected panic to be caught, got %v", recovered)
	}

	// Verify slot is returned despite panic
	limit.Wait()
}

func TestConcurrencyStress(t *testing.T) {
	const maxConcurrency = 8
	const totalJobs = 200

	limit := New(maxConcurrency)
	var active atomic.Int64
	var maxObserved atomic.Int64
	var completed atomic.Int64

	for range totalJobs {
		limit.Execute(func() {
			cur := active.Add(1)
			for {
				oldMax := maxObserved.Load()
				if cur <= oldMax || maxObserved.CompareAndSwap(oldMax, cur) {
					break
				}
			}
			time.Sleep(200 * time.Microsecond)
			active.Add(-1)
			completed.Add(1)
		})
	}

	limit.Wait()

	if completed.Load() != totalJobs {
		t.Fatalf("expected %d completed jobs, got %d", totalJobs, completed.Load())
	}
	if maxObserved.Load() > int64(maxConcurrency) {
		t.Fatalf("observed concurrency %d exceeded limit %d", maxObserved.Load(), maxConcurrency)
	}
}

// ====================================================================
// Standard Benchmarks
// ====================================================================

func BenchmarkExecuteWait(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		limit := New(8)
		for range 100 {
			limit.Execute(func() {})
		}
		limit.Wait()
	}
}

func BenchmarkTryExecute(b *testing.B) {
	limit := New(100)
	var wg sync.WaitGroup
	b.ReportAllocs()
	for b.Loop() {
		wg.Add(1)
		if !limit.TryExecute(func() {
			wg.Done()
		}) {
			wg.Done()
		}
	}
	wg.Wait()
}
