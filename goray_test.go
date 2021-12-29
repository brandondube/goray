package goray

import "testing"

func BenchmarkRayTrace(b *testing.B) {
	const c = -0.05
	const k = -1
	geo := Conic{c, k}
	Surf1 := Surface{Typ: REFLECT, Origin: Vec3{0, 0, 5}, Geom: geo}
	Surf2 := Surface{Typ: -5, Origin: Vec3{0, 0, 1/c/2 + 5}, Geom: Plane{}}
	prescription := []Surface{Surf1, Surf2}
	P0 := Vec3{0, 1, 5}
	S0 := Vec3{0, 0, 1}
	for i := 0; i < b.N; i++ {
		Raytrace(prescription, P0, S0, .6328, 1)
	}
}
