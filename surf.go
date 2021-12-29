package goray

import "math"

type Conic struct {
	C, K float64
}

func (con Conic) SagNormal(x, y float64) (float64, Vec3) {
	c := con.C
	k := con.K
	csq := c * c
	rsq := x*x + y*y
	// sag
	num := c * rsq
	kernel := 1 - (1+k)*csq*rsq
	phi := math.Sqrt(kernel)
	den := 1 + phi
	sag := num / den

	cByPhi := c / phi
	var N Vec3
	N[2] = 1
	N[0] = -x * cByPhi
	N[1] = -y * cByPhi
	return sag, N
}

type Plane struct{}

func (p Plane) SagNormal(x, y float64) (float64, Vec3) {
	return 0, Vec3{0, 0, 1}
}

type Geometry interface {
	SagNormal(float64, float64) (float64, Vec3)
}

type Glass interface {
	N(float64) float64
}

type Surface struct {
	Typ    int
	Origin Vec3
	Geom   Geometry
	Glas   Glass
	R      *Mat3
}
