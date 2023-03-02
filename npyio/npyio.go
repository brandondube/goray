package npyio

import (
	"fmt"
	"unsafe"
)

const (
	c128 = "complex128"
	f64  = "float64"

	szf64  = 8
	szc128 = 16
)

// ErrorWrongDtype is produced when a NumpyArray is converted to a native Go
// slice of the wrong type
type ErrorWrongDtype struct {
	expected string
	got      string
}

func (e ErrorWrongDtype) Error() string {
	return fmt.Sprintf("wrong dtype: expected %s, got %s", e.expected, e.got)
}

// ErrorWrongNdim is produced when a NumpyArray is converted to a native Go
// slice of the wrong number of dimensions ([]float64 vs [][]float64, e.g.)
type ErrorWrongNdim struct {
	expected int
	got      int
}

func (e ErrorWrongNdim) Error() string {
	return fmt.Sprintf("wrong ndim: expected %d, got %d", e.expected, e.got)
}

// NumpyArray is a serialization-friendly representation for data going to
// or coming from Numpy.
//
// Typ is a string of the form int8, uint16, float64, complex128, etc
// Shape contains the number of elements along each axis, in row-major order
// data is assumed to be contiguous in memory, and the stride in elements
// equal to the shape along a given axis (i.e., no padding or extra alignment)
//
// Data is the raw memory for the array, to be converted with zero-copy to
// Go types via .VecTYPE, .MatTYPE, such as VecComplex128.
type NumpyArray struct {
	Typ   string `msgpack:"type"`
	Shape []int  `msgpack:"shape"`
	Data  []byte `msgpack:"data"`
}

// VecComplex128 returns a copy-free view of the data as []complex128,
// if the shape and type are correct
func (m NumpyArray) VecComplex128() ([]complex128, error) {
	if m.Typ != c128 {
		return nil, ErrorWrongDtype{expected: c128, got: m.Typ}
	}
	if len(m.Shape) != 1 {
		return nil, ErrorWrongNdim{expected: 1, got: len(m.Shape)}
	}
	slc := unsafe.Slice((*complex128)(unsafe.Pointer(&m.Data[0])), m.Shape[0])
	return slc, nil
}

// VecFloat64 returns a copy-free view of the data as []float64,
// if the shape and type are correct
func (m NumpyArray) VecFloat64() ([]float64, error) {
	if m.Typ != f64 {
		return nil, ErrorWrongDtype{expected: f64, got: m.Typ}
	}
	if len(m.Shape) != 1 {
		return nil, ErrorWrongNdim{expected: 1, got: len(m.Shape)}
	}
	slc := unsafe.Slice((*float64)(unsafe.Pointer(&m.Data[0])), m.Shape[0])
	return slc, nil
}

// MatFloat64 returns a copy-free view of the data as [][]float64,
// if the shape and type are correct
//
// including contiguity of data; ret[i+1][0] and ret[i][j-1] are adjacent in memory
func (m NumpyArray) MatFloat64() ([][]float64, error) {
	if m.Typ != f64 {
		return nil, ErrorWrongDtype{expected: f64, got: m.Typ}
	}
	if len(m.Shape) != 2 {
		return nil, ErrorWrongNdim{expected: 2, got: len(m.Shape)}
	}
	Nrow := m.Shape[0]
	Ncol := m.Shape[1]
	out := make([][]float64, Nrow)
	offset := 0
	for i := 0; i < Nrow; i++ {
		ptr := (*float64)(unsafe.Pointer(&m.Data[offset]))
		out[i] = unsafe.Slice(ptr, Ncol)
		offset += szf64 * Ncol
	}
	return out, nil
}

// NumpyArrayFromF64Slice converts a []float64 to a NumpyArray
func NumpyArrayFromF64Slice(slc []float64) NumpyArray {
	return NumpyArray{
		Typ:   f64,
		Shape: []int{len(slc)},
		Data:  unsafe.Slice((*byte)(unsafe.Pointer(&slc[0])), len(slc)*szf64),
	}
}

// NumpyArrayFromC128Slice converts a []complex128 to a NumpyArray
func NumpyArrayFromC128Slice(slc []complex128) NumpyArray {
	return NumpyArray{
		Typ:   c128,
		Shape: []int{len(slc)},
		Data:  unsafe.Slice((*byte)(unsafe.Pointer(&slc[0])), len(slc)*szc128),
	}
}
