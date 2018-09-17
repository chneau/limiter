package limiter

// Limiter ...
type Limiter chan struct{}

// New instanciates a new Limiter
// limit: the max number of goroutines running at a time
func New(limit int) Limiter {
	if limit <= 0 {
		limit = 1
	}
	c := make(Limiter, limit)
	for i := 0; i < limit; i++ {
		c <- struct{}{}
	}
	return c
}

// Execute will queue jobs, thanks to how channels work
// job: the function you want to be run on this limiter
func (c Limiter) Execute(job func()) {
	ticket := <-c
	go func() {
		job()
		c <- ticket
	}()
}

// Wait waits that all jobs are done
// Wait have to be called only once per instance
func (c Limiter) Wait() {
	for i := 0; i < cap(c); i++ {
		<-c
	}
}
