# goray

Goray is a repository craeted to evaluate how fast Go can perform raytracing.  It contains a specialized linear algebra package for the 3-vectors and square 3-matricies needed for real ray tracing, as well as a basic implementation of planar and conic geometry.  The code wasn't extensively checked to be free of bugs, but was written in around half an hour, based on [prysm](https://github.com/brandondube/prysm)'s experimental raytrace module.

Overall, the performance of Go in this regime is quite good, and the type system lends itself reasonably well to the task.  The summary benchmark result, to trace a two-surface conic+planar geometry:

```
goos: darwin
goarch: amd64
pkg: github.com/brandondube/goray
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkRayTrace-12    	 3195540	       363.9 ns/op	     160 B/op	       2 allocs/op
PASS
ok  	github.com/brandondube/goray	1.643s
```

This works out to about 180 nanoseconds per surface, or 5.5 million ray-surfaces per second (per core).  Go'e exceptional concurrency story means it's likely that >100M ray-surfaces/sec can be achieved on manycore platforms.

For individual rays, this is a few hundred times faster than prysm's python code (50usec/raysurf).  Batch numpy calculations mean prysm does about 2.75M raysurf/sec, asymptotically which is within about a factor of two of this code.

I don't particularly intend to develop this further, but it was an interesting exercise.
