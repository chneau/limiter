# go lang goroutine concurrency limiter

## builds

[![Build Status](https://travis-ci.org/chneau/limiter.svg)](https://travis-ci.org/chneau/limiter)

## example

limit the number of concurrent go routines to 10:

```
  limit := limiter.New(10)
  for i := 0; i < 1000; i++ {
  	limit.Execute(func() {
  		// do some work
  	})
  }
  limit.Wait()
```

