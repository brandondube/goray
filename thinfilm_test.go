package goray

import (
	"math"
	"math/cmplx"
	"sync"
	"testing"
)

func approxEqual(a, b, tol float64) bool {
	diff := math.Abs(a - b)
	return diff <= tol
}

func TestThinFilmMgF2Monolayer(t *testing.T) {
	const (
		lambda  = .587725
		n_C7980 = 1.458461
		n_MgF2  = 1.3698
		n_CeF3  = complex(1.6290, 0.0034836)
		n_ZrO2  = 2.1588
	)
	stack := []NT{
		{N: n_MgF2, T: .15},
		{N: n_C7980, T: 10_000},
	}
	r, _ := MultilayerStackrt(Ppol, lambda, stack, 0, true)
	R := cmplx.Abs(r)
	R *= R
	if !approxEqual(R, 0.022, 0.001) {
		t.Fatalf("Reflectance of MgF2 on glass was %f, expected 0.022", R)
	}
}

func TestThinFilmMultilayerAR(t *testing.T) {
	const (
		lambda  = .587725
		n_C7980 = 1.458461
		n_MgF2  = 1.3698
		n_CeF3  = complex(1.6290, 0.0034836)
		n_ZrO2  = 2.1588
	)
	stack := []NT{
		{N: n_MgF2, T: lambda / 4},
		{N: n_ZrO2, T: lambda / 2},
		{N: n_CeF3, T: lambda / 4},
		{N: n_C7980, T: 10_000},
	}
	r, _ := MultilayerStackrt(Ppol, lambda, stack, 0, true)
	R := cmplx.Abs(r)
	R *= R
	if !approxEqual(R, 0.0024, 0.001) {
		t.Fatalf("Reflectance of 3-layer AR coating on glass was %f, expected 0.0024", R)
	}
}

func BenchmarkThinFilmMultilayerAR(b *testing.B) {
	const (
		lambda  = .587725
		n_C7980 = 1.458461
		n_MgF2  = 1.3698
		n_CeF3  = complex(1.6290, 0.0034836)
		n_ZrO2  = 2.1588
	)
	stack := []NT{
		{N: n_MgF2, T: lambda / 4},
		{N: n_ZrO2, T: lambda / 2},
		{N: n_CeF3, T: lambda / 4},
		{N: n_C7980, T: 10_000},
	}
	var r complex128
	for i := 0; i < b.N; i++ {
		r, _ = MultilayerStackrt(Ppol, lambda, stack, 0, true)
	}
	_ = r
}

func BenchmarkThinFilmMultilayerARParallel8Core(b *testing.B) {
	const (
		lambda  = .587725
		n_C7980 = 1.458461
		n_MgF2  = 1.3698
		n_CeF3  = complex(1.6290, 0.0034836)
		n_ZrO2  = 2.1588

		NTHREADS = 8
		cases    = 100 * 100
	)
	stack := []NT{
		{N: n_MgF2, T: lambda / 4},
		{N: n_ZrO2, T: lambda / 2},
		{N: n_CeF3, T: lambda / 4},
		{N: n_C7980, T: 10_000},
	}
	var r complex128
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for i := 0; i < NTHREADS; i++ {
			wg.Add(1)
			go func() {
				for j := 0; j < (cases / NTHREADS); j++ {
					MultilayerStackrt(Ppol, lambda, stack, 0, true)
				}
				wg.Done()
			}()
		}
		wg.Wait()
		// r, _ = MultilayerStackrt(Ppol, lambda, stack, 0, true)
	}
	_ = r
}
