package main

import "math"

//Vec3 is a 3D vector
type Vec3 [3]float64

func (v Vec3) Magnitude() float64 {
	return math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
}

func (v Vec3) ScalarMult(c float64) Vec3 {
	return Vec3{v[0] * c, v[1] * c, v[2] * c}
}

func (v Vec3) Unit() Vec3 {
	return v.ScalarMult(1 / v.Magnitude())
}

func (v Vec3) X() float64 {
	return v[0]
}
func (v Vec3) Y() float64 {
	return v[1]
}
func (v Vec3) Z() float64 {
	return v[2]
}

func (v Vec3) Add(v2 Vec3) Vec3 {
	return Vec3{
		v[0] + v2[0],
		v[1] + v2[1],
		v[2] + v2[2],
	}
}

func (v Vec3) DotP(v2 Vec3) float64 {
	return v[0]*v2[0] + v[1]*v2[1] + v[2]*v2[2]
}

func (v Vec3) CrossP(v2 Vec3) Vec3 {
	return Vec3{
		v[1]*v2[2] - v[2]*v2[1],
		v[2]*v2[0] - v[0]*v2[2],
		v[0]*v2[1] - v[1]*v2[0],
	}
}

func (v Vec3) CosAng(v2 Vec3) float64 {
	return v.DotP(v2) / (v.Magnitude() * v2.Magnitude())
}

func (v Vec3) Component(axis Vec3) float64 {
	return axis.Unit().DotP(v)
}

var I = Vec3{1, 0, 0}
var J = Vec3{0, 1, 0}
var K = Vec3{0, 0, 1}

type mat3 [3][3]float64

func (m mat3) multiplyVec(v Vec3) Vec3 {
	return Vec3{
		Vec3(m[0]).DotP(v),
		Vec3(m[1]).DotP(v),
		Vec3(m[2]).DotP(v),
	}
}
