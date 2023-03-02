package thinfilm

import (
	"math"
	"math/cmplx"
)

type Mat2C [2][2]complex128

type NT struct {
	T float64    // thickness
	N complex128 // refactive index
}

type PolState int

const (
	Spol PolState = iota
	Ppol

	RadToDeg = 180 / math.Pi
	DegToRad = math.Pi / 180

	n1j = -complex(0, 1)
)

func MatMul2C(A, B Mat2C) Mat2C {
	a := A[0][0]
	b := A[0][1]
	c := A[1][0]
	d := A[1][1]

	w := B[0][0]
	x := B[0][1]
	y := B[1][0]
	z := B[1][1]
	return Mat2C{
		{a*w + b*y, a*x + b*z},
		{c*w + d*y, c*x + d*z},
	}
}

func MatScale2C(A Mat2C, s complex128) Mat2C {
	return Mat2C{
		{A[0][0] * s, A[0][1] * s},
		{A[1][0] * s, A[1][1] * s},
	}
}

func CharacteristicMatrixS(lambda, d float64, n, theta complex128) Mat2C {
	k := complex(2*math.Pi/lambda, 0) * n
	dc := complex(d, 0)
	cost := cmplx.Cos(theta)
	beta := k * dc * cost
	sinb := cmplx.Sin(beta)
	cosb := cmplx.Cos(beta)
	upperRight := n1j * sinb / (n * cost)
	lowerLeft := n1j * n * cost * sinb
	return Mat2C{
		{cosb, upperRight},
		{lowerLeft, cosb},
	}
}

func CharacteristicMatrixP(lambda, d float64, n, theta complex128) Mat2C {
	k := complex(2*math.Pi/lambda, 0) * n
	dc := complex(d, 0)
	cost := cmplx.Cos(theta)
	beta := k * dc * cost
	sinb := cmplx.Sin(beta)
	cosb := cmplx.Cos(beta)
	upperRight := n1j * sinb * cost / n
	lowerLeft := n1j * n * sinb / cost
	return Mat2C{
		{cosb, upperRight},
		{lowerLeft, cosb},
	}
}

func Totalr(M Mat2C) complex128 {
	return M[1][0] / M[0][0]
}

func Totalt(M Mat2C) complex128 {
	return 1 / M[0][0]
}

func SnellAOR(n0, n1, theta complex128) complex128 {
	kernel := n0 / n1 * cmplx.Sin(theta)
	return cmplx.Asin(kernel)
}

func MultilayerStackrt(pol PolState, lambda float64, stack []NT, aoi float64, vacAmbient bool) (complex128, complex128) {
	var (
		n0    complex128
		n1    complex128
		theta complex128
		Amat  Mat2C
		front Mat2C
		back  Mat2C
	)
	if len(stack) == 0 {
		panic("zero length stack is meaningless")
	}
	aoi = aoi * DegToRad

	if vacAmbient {
		n0 = 1
	} else {
		n0 = stack[0].N
	}
	theta = complex(aoi, 0)
	cos0 := math.Cos(aoi)
	cos0c := complex(cos0, 0)

	term1 := 1 / (2 * n0 * cos0c)
	// n0cos0 := cos0c * n0
	Amat = Mat2C{
		{1, 0},
		{0, 1},
	}
	if pol == Ppol {
		front = Mat2C{
			{n0, cos0c},
			{n0, -cos0c},
		}
		for _, nt := range stack {
			n1 = nt.N
			theta1 := SnellAOR(n0, n1, theta)
			Mj := CharacteristicMatrixP(lambda, nt.T, n1, theta1)
			Amat = MatMul2C(Amat, Mj)
			theta = theta1
			n0 = n1
		}
		cos1c := cmplx.Cos(theta)
		back = Mat2C{
			{cos1c, 0},
			{n1, 0},
		}
	} else if pol == Spol {
		n0cos0c := n0 * cos0c
		front = Mat2C{
			{n0cos0c, 1},
			{n0cos0c, -1},
		}
		for _, nt := range stack {
			n1 = nt.N
			theta1 := SnellAOR(n0, n1, theta)
			// fmt.Printf("in stack, %v -> %v rad\n", theta, theta1)
			Mj := CharacteristicMatrixS(lambda, nt.T, n1, theta1)
			Amat = MatMul2C(Amat, Mj)
			theta = theta1
			n0 = n1
		}
		cos1c := cmplx.Cos(theta)
		back = Mat2C{
			{1, 0},
			{n1 * cos1c, 0},
		}
	} else {
		panic("invalid polarization, must be either Ppol or Spol")
	}
	Amat = MatMul2C(front, Amat)
	Amat = MatMul2C(Amat, back)
	Amat = MatScale2C(Amat, term1)
	return Totalr(Amat), Totalt(Amat)
}
