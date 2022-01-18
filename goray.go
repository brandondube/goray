package goray

import (
	"math"
	"sync"
	"unsafe"
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

func ParallelRaytrace(prescription []Surface, Ps, Ss []Vec3, wvl, nAmbient float64, niterIntersect, nthreads int, Pout, Sout [][]Vec3) {
	nrays := len(Ps)
	chunkSize := nrays / nthreads
	low := 0
	var wg sync.WaitGroup
	for i := 0; i < nthreads; i++ {
		wg.Add(1) // add and done on each iter so that we know all threads launched
		high := low + chunkSize
		if high > nrays {
			high = nrays
		}
		go func(start, stop int) {
			defer wg.Done()
			nrays := stop - start
			P := Ps[start:stop]
			S := Ss[start:stop]
			PP := Pout[start:stop]
			SS := Sout[start:stop]
			for i := 0; i < nrays; i++ {
				RaytraceNoAlloc(prescription, P[i], S[i], wvl, nAmbient, niterIntersect, PP[i], SS[i])
			}
		}(low, high)
		low += chunkSize
	}
	wg.Wait()
	// return Pout, Sout
}

// AllocateOutputSpace creates output buffers for a raytrace
func AllocateOutputSpace(nsurfaces, nrays int) ([][]Vec3, [][]Vec3) {
	var (
		v      Vec3
		v0     = v[0]
		ptr    *Vec3
		offset int
		slc    []Vec3
	)
	const szVec = int(unsafe.Sizeof(v))
	const szElem = int(unsafe.Sizeof(v0))
	dim1 := nrays
	dim2 := nsurfaces + 1
	allocSize := dim1 * dim2 * szVec
	rawbuf1 := make([]uint64, allocSize/8)
	rawbuf2 := make([]uint64, allocSize/8)
	buf1 := unsafe.Slice((*byte)(unsafe.Pointer(&rawbuf1[0])), allocSize)
	buf2 := unsafe.Slice((*byte)(unsafe.Pointer(&rawbuf2[0])), allocSize)
	out1 := make([][]Vec3, 0, dim1)
	out2 := make([][]Vec3, 0, dim1)

	for i := 0; i < dim1; i++ {
		ptr = (*Vec3)(unsafe.Pointer(&buf1[offset]))
		slc = unsafe.Slice(ptr, dim2)
		out1 = append(out1, slc)

		ptr = (*Vec3)(unsafe.Pointer(&buf2[offset]))
		slc = unsafe.Slice(ptr, dim2)
		out2 = append(out2, slc)
		offset += dim2 * szElem

	}
	return out1, out2
}
