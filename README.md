# goray

Goray is a repository craeted to evaluate how fast Go can perform raytracing.  It contains a specialized linear algebra package for the 3-vectors and square 3-matricies needed for real ray tracing, as well as a basic implementation of planar and conic geometry.  The code wasn't extensively checked to be free of bugs, but was written in around half an hour, based on [prysm](https://github.com/brandondube/prysm)'s experimental raytrace module.

Overall, the performance of Go in this regime is quite good, and the type system lends itself reasonably well to the task.  The summary benchmark result, to trace a two-surface conic+planar geometry:

```
goos: windows
goarch: amd64
pkg: github.com/brandondube/goray
cpu: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz
BenchmarkRayTrace-8                              3663687               326.4 ns/op           160 B/op          2 allocs/op
BenchmarkRaytraceNoAlloc-8                       4386572               275.8 ns/op             0 B/op          0 allocs/op
BenchmarkParallelRaytrace1Thread1Mray-8                4         278530150 ns/op             208 B/op          3 allocs/op
BenchmarkParallelRaytrace2Thread1Mray-8                8         141101462 ns/op             504 B/op          5 allocs/op
BenchmarkParallelRaytrace3Thread1Mray-8               12          94838375 ns/op             808 B/op          7 allocs/op
BenchmarkParallelRaytrace4Thread1Mray-8               15          71159467 ns/op            1372 B/op         10 allocs/op
BenchmarkParallelRaytrace5Thread1Mray-8               20          58274915 ns/op            1526 B/op         12 allocs/op
BenchmarkParallelRaytrace6Thread1Mray-8               24          48336904 ns/op            1480 B/op         13 allocs/op
BenchmarkParallelRaytrace7Thread1Mray-8               24          41895525 ns/op            2036 B/op         16 allocs/op
BenchmarkParallelRaytrace8Thread1Mray-8               31          41124787 ns/op            2611 B/op         19 allocs/op
```

This works out to about 150 nanoseconds per surface, or 6.6 million ray-surfaces per second (per core).  Using 8 cores results in an equivalent sequential speed of about 21 nanoseconds per ray-surface, 48 million ray-surfaces per second.  That number is roughly half the speed of Code V.

I suspect commercial raytracers specialize the surface intersection operation, which is about 75% of the runtime, for low-complexity surfaces like spheres.  The surface sag and normal are evaluted three times for Newton's method, and so to an order of magnitude 50% of the runtime can be shaved off by a closed form surface intersection function.

Adjusting the `oneM` constant in the test file lets you quickly change the batch size.  At 1,000 total rays (125 rays per thread) the overhead of spawning the goroutines is almost doubling the runtime (i.e., instead of 41 ns per ray, ~72 ns per ray).

I don't particularly intend to develop this further, but it was an interesting exercise.
