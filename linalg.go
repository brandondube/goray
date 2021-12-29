package goray

type Vec3 [3]float64

type Mat3 [3][3]float64

func SumSqVec3(a Vec3) float64 {
	out := 0.
	for i := 0; i < 3; i++ {
		out += a[i] * a[i]
	}
	return out
}

func ScaleVec3(a Vec3, s float64) Vec3 {
	var out Vec3
	for i := 0; i < 3; i++ {
		out[i] = a[i] * s
	}
	return out
}

func AddVec3(a, b Vec3) Vec3 {
	var out Vec3
	for i := 0; i < 3; i++ {
		out[i] = a[i] + b[i]
	}
	return out
}

func SubVec3(a, b Vec3) Vec3 {
	var out Vec3
	for i := 0; i < 3; i++ {
		out[i] = a[i] - b[i]
	}
	return out
}

func DotVec3(a, b Vec3) float64 {
	out := 0.
	for i := 0; i < 3; i++ {
		out += a[i] * b[i]
	}
	return out
}

func Mat3Vec3Prod(A Mat3, b Vec3) Vec3 {
	var out Vec3
	for i := 0; i < 3; i++ {
		tmp := 0.0
		for j := 0; j < 3; j++ {
			tmp += A[i][j] * b[j]
		}
		out[i] = tmp
	}
	return out
}
