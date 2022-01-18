package goray

import "testing"

// 395 manual unroll
// 401 rolled = 6/401 = 1.5%, who cares
// 343 ns for f32 = 15% faster (not nothing, but not the end of the world)
// better memory bandwidth scaling though...

func BenchmarkRayTrace(b *testing.B) {
	// ~330 nsec/op (5.4M raysurf/sec)
	const c = -0.05
	const k = -1
	geo := Conic{c, k}
	Surf1 := Surface{Typ: REFLECT, Origin: Vec3{0, 0, 5}, Geom: geo}
	Surf2 := Surface{Typ: STOP, Origin: Vec3{0, 0, 1/c/2 + 5}, Geom: Plane{}}
	prescription := []Surface{Surf1, Surf2}
	P0 := Vec3{0, 1, 2}
	S0 := Vec3{0, 0, 1}
	for i := 0; i < b.N; i++ {
		Raytrace(prescription, P0, S0, .6328, 1, 100)
	}
}

func BenchmarkRaytraceNoAlloc(b *testing.B) {
	// ~230 nsec/op (7.7M raysurf/sec)
	const c = -0.05
	const k = -1
	geo := Conic{c, k}
	Surf1 := Surface{Typ: REFLECT, Origin: Vec3{0, 0, 5}, Geom: geo}
	Surf2 := Surface{Typ: STOP, Origin: Vec3{0, 0, 1/c/2 + 5}, Geom: Plane{}}
	prescription := []Surface{Surf1, Surf2}
	P0 := Vec3{0, 1, 2}
	S0 := Vec3{0, 0, 1}
	Pout := make([]Vec3, len(prescription)+1)
	Sout := make([]Vec3, len(prescription)+1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RaytraceNoAlloc(prescription, P0, S0, .6328, 1, 100, Pout, Sout)
	}
}

func _benchmarkParallelRaytraceVarThreads(nthreads, nrays int, b *testing.B) {
	// make the Ps and Ss
	P := Vec3{0, 1, 0} // 1 mm rise
	S := Vec3{0, 0, 1} // propagate in the Z dir
	const c = -0.05
	const k = -1
	geo := Conic{c, k}
	Surf1 := Surface{Typ: REFLECT, Origin: Vec3{0, 0, 5}, Geom: geo}
	Surf2 := Surface{Typ: STOP, Origin: Vec3{0, 0, 1/c/2 + 5}, Geom: Plane{}}
	prescription := []Surface{Surf1, Surf2}
	Ps := make([]Vec3, nrays)
	Ss := make([]Vec3, nrays)
	// it doesn't matter that they're all the same ray for purposes
	// of the benchmark
	for i := 0; i < nrays; i++ {
		Ps[i] = P
		Ss[i] = S
	}
	Pout, Sout := AllocateOutputSpace(len(prescription), nrays)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParallelRaytrace(prescription, Ps, Ss, .6328, 1, 100, nthreads, Pout, Sout)
	}
}

const oneM = 1e6

func BenchmarkParallelRaytrace1Thread1Mray(b *testing.B) {
	_benchmarkParallelRaytraceVarThreads(1, oneM, b)
}

func BenchmarkParallelRaytrace2Thread1Mray(b *testing.B) {
	_benchmarkParallelRaytraceVarThreads(2, oneM, b)
}

func BenchmarkParallelRaytrace3Thread1Mray(b *testing.B) {
	_benchmarkParallelRaytraceVarThreads(3, oneM, b)
}

func BenchmarkParallelRaytrace4Thread1Mray(b *testing.B) {
	_benchmarkParallelRaytraceVarThreads(4, oneM, b)
}

func BenchmarkParallelRaytrace5Thread1Mray(b *testing.B) {
	_benchmarkParallelRaytraceVarThreads(5, oneM, b)
}

func BenchmarkParallelRaytrace6Thread1Mray(b *testing.B) {
	_benchmarkParallelRaytraceVarThreads(6, oneM, b)
}

func BenchmarkParallelRaytrace7Thread1Mray(b *testing.B) {
	_benchmarkParallelRaytraceVarThreads(7, oneM, b)
}

func BenchmarkParallelRaytrace8Thread1Mray(b *testing.B) {
	_benchmarkParallelRaytraceVarThreads(8, oneM, b)
}
