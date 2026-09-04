package limiter

import (
	"context"
)

// Limiter represents a lightweight, channel-based concurrency limiter.
// It limits the maximum number of concurrent goroutines running at any given time.
type Limiter chan struct{}

// New instantiates a new Limiter.
// limit: the maximum number of concurrent goroutines allowed to run simultaneously.
// If limit <= 0, it defaults to 1.
func New(limit int) Limiter {
	if limit <= 0 {
		limit = 1
	}
	c := make(Limiter, limit)
	for range limit {
		c <- struct{}{}
	}
	return c
}

// Execute queues a job, blocking until a concurrency slot becomes available.
// When a slot is available, it launches the job in a new goroutine and returns immediately.
func (c Limiter) Execute(job func()) {
	ticket := <-c
	go func() {
		defer func() {
			c <- ticket
		}()
		job()
	}()
}

// ExecuteContext attempts to acquire a concurrency slot honoring context cancellation.
// If the context is canceled or exceeds its deadline before a slot becomes available,
// it returns ctx.Err() without launching the job.
// Once acquired, the job runs in a separate goroutine and returns nil.
func (c Limiter) ExecuteContext(ctx context.Context, job func()) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ticket := <-c:
		go func() {
			defer func() {
				c <- ticket
			}()
			job()
		}()
		return nil
	}
}

// TryExecute attempts to acquire an available slot without blocking.
// Returns true and executes job in a new goroutine if a slot was available,
// or returns false immediately if all concurrency slots are currently busy.
func (c Limiter) TryExecute(job func()) bool {
	select {
	case ticket := <-c:
		go func() {
			defer func() {
				c <- ticket
			}()
			job()
		}()
		return true
	default:
		return false
	}
}

// Wait blocks until all active jobs have completed.
// Wait should be called after all desired jobs have been enqueued.
// Wait must only be called once per limiter lifecycle.
func (c Limiter) Wait() {
	for range cap(c) {
		<-c
	}
}

// Cap returns the total concurrency capacity configured for this limiter.
func (c Limiter) Cap() int {
	return cap(c)
}

// Available returns the number of concurrency slots currently free.
func (c Limiter) Available() int {
	return len(c)
}

// Running returns the number of goroutines currently executing.
func (c Limiter) Running() int {
	return cap(c) - len(c)
}
