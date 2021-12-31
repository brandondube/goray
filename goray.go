package goray

import (
	"math"
	"sync"
)

const (
	REFLECT int = 1
	REFRACT int = 2
	STOP    int = 3
)

type Ray struct {
	P Vec3
	S Vec3
}

type SagNormalFunc func(float64, float64) (float64, Vec3)

func AdvanceRay(P, S Vec3, s float64) Vec3 {
	return AddVec3(ScaleVec3(S, s), P)
}

// NewtonRaphsonIntersect returns the point P and surface normal vector N at which
// the ray and surface described by FFP intersect
func NewtonRaphsonIntersect(P1, S Vec3, FFp SagNormalFunc, s1, eps float64, maxiter int) (Vec3, Vec3) {
	var (
		sj    float64 = s1
		sjp1  float64
		delta float64
		Pj    Vec3
		sag   float64
		Fpj   float64
		N     Vec3
	)
	for i := 0; i < maxiter; i++ {
		// P1 + sk * S
		Pj = AdvanceRay(P1, S, sj)
		sag, N = FFp(Pj[0], Pj[1])
		Fpj = DotVec3(N, S)
		sag = Pj[2] - sag // Z-F
		sjp1 = sj - sag/Fpj
		delta = math.Abs(sjp1 - sj)
		if delta < eps {
			break
		}
		sj = sjp1
	}
	return Pj, N
}

func Intersect(P0, S Vec3, FFp SagNormalFunc, eps float64, maxiter int) (Vec3, Vec3) {
	// move to Z=0
	Z0 := P0[2]
	m := S[2]
	s0 := -Z0 / m
	P1 := AdvanceRay(P0, S, s0)
	return NewtonRaphsonIntersect(P1, S, FFp, 0, eps, maxiter)
}

func TransformToLocalCoords(XYZ, P, S Vec3, R *Mat3) (Vec3, Vec3) {
	XYZ2 := SubVec3(XYZ, P)
	if R != nil {
		XYZ2 = Mat3Vec3Prod(*R, XYZ2)
		S = Mat3Vec3Prod(*R, S)
	}
	return XYZ2, S
}

func Reflect(S, N Vec3) Vec3 {
	Nnorm := SumSqVec3(N)
	cosI := DotVec3(S, N) / Nnorm
	return SubVec3(S, ScaleVec3(N, -2*cosI))
}

func RaytraceNoAlloc(prescription []Surface, P, S Vec3, wvl, nAmbient float64, niterIntersect int, Pout, Sout []Vec3) {
	var (
		nsurf      = len(prescription)
		Pj    Vec3 = P
		Sj    Vec3 = S
		N     Vec3
		P0    Vec3
		Pjp1  Vec3
		Sjp1  Vec3
		surf  Surface
	)
	Pout[0] = P
	Sout[0] = S
	for j := 0; j < nsurf; j++ {
		surf = prescription[j]
		// S&M step 1
		P0, Sj = TransformToLocalCoords(Pj, surf.Origin, Sj, surf.R)
		// S&M step 2
		Pj, N = Intersect(P0, Sj, surf.Geom.SagNormal, 1e-14, niterIntersect)
		if surf.Typ == REFLECT {
			Sjp1 = Reflect(Sj, N)
		}
		Pjp1, Sjp1 = TransformToLocalCoords(Pj, ScaleVec3(surf.Origin, -1), Sj, surf.R)
		Pout[j+1] = Pjp1
		Sout[j+1] = Sjp1
		Pj, Sj = Pjp1, Sjp1
	}
}

func Raytrace(prescription []Surface, P, S Vec3, wvl, nAmbient float64, niterIntersect int) ([]Vec3, []Vec3) {
	nsurf := len(prescription)
	Pout := make([]Vec3, nsurf+1)
	Sout := make([]Vec3, nsurf+1)
	RaytraceNoAlloc(prescription, P, S, wvl, nAmbient, niterIntersect, Pout, Sout)
	return Pout, Sout
}

// BlockRaytraceNoAlloc is simply a loop over P, S, Pout, Sout to trace multiple
// rays.  It's used to help implement massively parallel raytracing
func BlockRaytraceNoAlloc(prescription []Surface, Ps, Ss []Vec3, wvl, nAmbient float64, niterIntersect int, Pout, Sout [][]Vec3) {
	nrays := len(Ps)
	for i := 0; i < nrays; i++ {
		RaytraceNoAlloc(prescription, Ps[i], Ss[i], wvl, nAmbient, niterIntersect, Pout[i], Sout[i])
	}
}

type startStop struct {
	start, stop int
}

func ParallelRaytrace(prescription []Surface, Ps, Ss []Vec3, wvl, nAmbient float64, niterIntersect, nthreads int) ([][]Vec3, [][]Vec3) {
	nrays := len(Ps)
	lenPout := len(prescription) + 1
	Pout := make([][]Vec3, nrays)
	Sout := make([][]Vec3, nrays)
	startStops := make([]startStop, 0, nthreads)
	raysPerThread := nrays / nthreads
	raysForLastThread := nrays % nthreads
	if raysForLastThread == 0 {
		raysForLastThread = raysPerThread
	}
	start := 0
	for i := 0; i < nthreads-1; i++ {
		startStops = append(startStops, startStop{start, start + raysPerThread})
		start += raysPerThread
	}
	startStops = append(startStops, startStop{start, start + raysForLastThread})
	var wg sync.WaitGroup
	for i := 0; i < nthreads; i++ {
		wg.Add(1) // add and done on each iter so that we know all threads launched
		go func(i int) {
			defer wg.Done()
			ss := startStops[i]
			P := Ps[ss.start:ss.stop]
			S := Ss[ss.start:ss.stop]
			PP := Pout[ss.start:ss.stop]
			SS := Sout[ss.start:ss.stop]
			for j := 0; j < len(PP); j++ {
				PP[j] = make([]Vec3, lenPout)
				SS[j] = make([]Vec3, lenPout)
			}
			BlockRaytraceNoAlloc(prescription, P, S, wvl, nAmbient, niterIntersect, PP, SS)
		}(i)
	}
	wg.Wait()
	return Pout, Sout
}
