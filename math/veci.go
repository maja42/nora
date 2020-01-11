// Based on github.com/go-gl/mathgl/mgl32
package math

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

type Vec2i [2]int
type Vec3i [3]int
type Vec4i [4]int

func (v Vec2i) String() string {
	return fmt.Sprintf("Vec2i[%d x %d]", v[0], v[1])
}
func (v Vec3i) String() string {
	return fmt.Sprintf("Vec3i[%d x %d]", v[0], v[1], v[2])
}
func (v Vec4i) String() string {
	return fmt.Sprintf("Vec4i[%d x %d]", v[0], v[1], v[2], v[3])
}

// Vec3 constructs a 3-dimensional vector by appending the given coordinates.
func (v Vec2i) Vec3(z int) Vec3i {
	return Vec3i{v[0], v[1], z}
}

// Vec4 constructs a 4-dimensional vector by appending the given coordinates.
func (v Vec2i) Vec4(z, w int) Vec4i {
	return Vec4i{v[0], v[1], z, w}
}

// Vec4 constructs a 4-dimensional vector by appending the given coordinates.
func (v Vec3i) Vec4(w int) Vec4i {
	return Vec4i{v[0], v[1], v[2], w}
}

// Vec2 constructs a 2-dimensional vector by discarding coordinates.
func (v Vec3i) Vec2() Vec2i {
	return Vec2i{v[0], v[1]}
}

// Vec2 constructs a 2-dimensional vector by discarding coordinates.
func (v Vec4i) Vec2() Vec2i {
	return Vec2i{v[0], v[1]}
}

// Vec3 constructs a 3-dimensional vector by discarding coordinates.
func (v Vec4i) Vec3() Vec3i {
	return Vec3i{v[0], v[1], v[2]}
}

// Vecf constructs a 2-dimensional float vector.
func (v Vec2i) Vecf() mgl32.Vec2 {
	return mgl32.Vec2{float32(v[0]), float32(v[1])}
}

// Vecf constructs a 3-dimensional float vector.
func (v Vec3i) Vecf() mgl32.Vec3 {
	return mgl32.Vec3{float32(v[0]), float32(v[1]), float32(v[2])}
}

// Vecf constructs a 4-dimensional float vector.
func (v Vec4i) Vecf() mgl32.Vec4 {
	return mgl32.Vec4{float32(v[0]), float32(v[1]), float32(v[2]), float32(v[3])}
}

// Elem extracts the elements of the vector for direct value assignment.
func (v Vec2i) Elem() (x, y int) {
	return v[0], v[1]
}

// Elem extracts the elements of the vector for direct value assignment.
func (v Vec3i) Elem() (x, y, z int) {
	return v[0], v[1], v[2]
}

// Elem extracts the elements of the vector for direct value assignment.
func (v Vec4i) Elem() (x, y, z, w int) {
	return v[0], v[1], v[2], v[3]
}

// Cross is the vector cross product. This operation is only defined on 3D
// vectors. It is equivalent to Vec3{v1[1]*v2[2]-v1[2]*v2[1],
// v1[2]*v2[0]-v1[0]*v2[2], v1[0]*v2[1] - v1[1]*v2[0]}. Another interpretation
// is that it's the vector whose magnitude is |v1||v2|sin(theta) where theta is
// the angle between v1 and v2.
//
// The cross product is most often used for finding surface normals. The cross
// product of vectors will generate a vector that is perpendicular to the plane
// they form.
//
// Technically, a generalized cross product exists as an "(N-1)ary" operation
// (that is, the 4D cross product requires 3 4D vectors). But the binary 3D (and
// 7D) cross product is the most important. It can be considered the area of a
// parallelogram with sides v1 and v2.
//
// Like the dot product, the cross product is roughly a measure of
// directionality. Two normalized perpendicular vectors will return a vector
// with a magnitude of 1.0 or -1.0 and two parallel vectors will return a vector
// with magnitude 0.0. The cross product is "anticommutative" meaning
// v1.Cross(v2) = -v2.Cross(v1), this property can be useful to know when
// finding normals, as taking the wrong cross product can lead to the opposite
// normal of the one you want.
func (v1 Vec3i) Cross(v2 Vec3i) Vec3i {
	return Vec3i{v1[1]*v2[2] - v1[2]*v2[1], v1[2]*v2[0] - v1[0]*v2[2], v1[0]*v2[1] - v1[1]*v2[0]}
}

// Add performs element-wise addition between two vectors. It is equivalent to iterating
// over every element of v1 and adding the corresponding element of v2 to it.
func (v1 Vec2i) Add(v2 Vec2i) Vec2i {
	return Vec2i{v1[0] + v2[0], v1[1] + v2[1]}
}

// Sub performs element-wise subtraction between two vectors. It is equivalent to iterating
// over every element of v1 and subtracting the corresponding element of v2 from it.
func (v1 Vec2i) Sub(v2 Vec2i) Vec2i {
	return Vec2i{v1[0] - v2[0], v1[1] - v2[1]}
}

// Mul performs a scalar multiplication between the vector and some constant value
// c. This is equivalent to iterating over every vector element and multiplying by c.
func (v1 Vec2i) Mul(c int) Vec2i {
	return Vec2i{v1[0] * c, v1[1] * c}
}

// Mul performs a scalar multiplication between the vector and some constant value
// c. This is equivalent to iterating over every vector element and multiplying by c.
func (v1 Vec2i) Mulf(c float32) mgl32.Vec2 {
	return mgl32.Vec2{float32(v1[0]) * c, float32(v1[1]) * c}
}

// Dot returns the dot product of this vector with another. There are multiple ways
// to describe this value. One is the multiplication of their lengths and cos(theta) where
// theta is the angle between the vectors: v1.v2 = |v1||v2|cos(theta).
//
// The other (and what is actually done) is the sum of the element-wise multiplication of all
// elements. So for instance, two Vec3s would yield v1.x * v2.x + v1.y * v2.y + v1.z * v2.z.
//
// This means that the dot product of a vector and itself is the square of its Len (within
// the bounds of floating points error).
//
// The dot product is roughly a measure of how closely two vectors are to pointing in the same
// direction. If both vectors are normalized, the value will be -1 for opposite pointing,
// one for same pointing, and 0 for perpendicular vectors.
func (v1 Vec2i) Dot(v2 Vec2i) int {
	return v1[0]*v2[0] + v1[1]*v2[1]
}

// Len returns the vector's length. Note that this is NOT the dimension of
// the vector (len(v)), but the mathematical length. This is equivalent to the square
// root of the sum of the squares of all elements. E.G. for a Vec2i it's
// math.Hypot(v[0], v[1]).
func (v1 Vec2i) Len() float32 {
	return float32(math.Hypot(float64(v1[0]), float64(v1[1])))

}

// LenSqr returns the vector's square length. This is equivalent to the sum of the squares of all elements.
func (v1 Vec2i) LenSqr() int {
	return v1[0]*v1[0] + v1[1]*v1[1]
}

// X is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec2i) X() int {
	return v[0]
}

// Y is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec2i) Y() int {
	return v[1]
}

// Add performs element-wise addition between two vectors. It is equivalent to iterating
// over every element of v1 and adding the corresponding element of v2 to it.
func (v1 Vec3i) Add(v2 Vec3i) Vec3i {
	return Vec3i{v1[0] + v2[0], v1[1] + v2[1], v1[2] + v2[2]}
}

// Sub performs element-wise subtraction between two vectors. It is equivalent to iterating
// over every element of v1 and subtracting the corresponding element of v2 from it.
func (v1 Vec3i) Sub(v2 Vec3i) Vec3i {
	return Vec3i{v1[0] - v2[0], v1[1] - v2[1], v1[2] - v2[2]}
}

// Mul performs a scalar multiplication between the vector and some constant value
// c. This is equivalent to iterating over every vector element and multiplying by c.
func (v1 Vec3i) Mul(c int) Vec3i {
	return Vec3i{v1[0] * c, v1[1] * c, v1[2] * c}
}

// Mul performs a scalar multiplication between the vector and some constant value
// c. This is equivalent to iterating over every vector element and multiplying by c.
func (v1 Vec3i) Mulf(c float32) mgl32.Vec3 {
	return mgl32.Vec3{float32(v1[0]) * c, float32(v1[1]) * c, float32(v1[2]) * c}
}

// Dot returns the dot product of this vector with another. There are multiple ways
// to describe this value. One is the multiplication of their lengths and cos(theta) where
// theta is the angle between the vectors: v1.v2 = |v1||v2|cos(theta).
//
// The other (and what is actually done) is the sum of the element-wise multiplication of all
// elements. So for instance, two Vec3s would yield v1.x * v2.x + v1.y * v2.y + v1.z * v2.z.
//
// This means that the dot product of a vector and itself is the square of its Len (within
// the bounds of floating points error).
//
// The dot product is roughly a measure of how closely two vectors are to pointing in the same
// direction. If both vectors are normalized, the value will be -1 for opposite pointing,
// one for same pointing, and 0 for perpendicular vectors.
func (v1 Vec3i) Dot(v2 Vec3i) int {
	return v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2]
}

// Len returns the vector's length. Note that this is NOT the dimension of
// the vector (len(v)), but the mathematical length. This is equivalent to the square
// root of the sum of the squares of all elements. E.G. for a Vec2i it's
// math.Hypot(v[0], v[1]).
func (v1 Vec3i) Len() float32 {
	return float32(math.Sqrt(float64(v1[0]*v1[0] + v1[1]*v1[1] + v1[2]*v1[2])))
}

// LenSqr returns the vector's square length. This is equivalent to the sum of the squares of all elements.
func (v1 Vec3i) LenSqr() int {
	return v1[0]*v1[0] + v1[1]*v1[1] + v1[2]*v1[2]
}

// X is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec3i) X() int {
	return v[0]
}

// Y is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec3i) Y() int {
	return v[1]
}

// Z is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec3i) Z() int {
	return v[2]
}

// Add performs element-wise addition between two vectors. It is equivalent to iterating
// over every element of v1 and adding the corresponding element of v2 to it.
func (v1 Vec4i) Add(v2 Vec4i) Vec4i {
	return Vec4i{v1[0] + v2[0], v1[1] + v2[1], v1[2] + v2[2], v1[3] + v2[3]}
}

// Sub performs element-wise subtraction between two vectors. It is equivalent to iterating
// over every element of v1 and subtracting the corresponding element of v2 from it.
func (v1 Vec4i) Sub(v2 Vec4i) Vec4i {
	return Vec4i{v1[0] - v2[0], v1[1] - v2[1], v1[2] - v2[2], v1[3] - v2[3]}
}

// Mul performs a scalar multiplication between the vector and some constant value
// c. This is equivalent to iterating over every vector element and multiplying by c.
func (v1 Vec4i) Mul(c int) Vec4i {
	return Vec4i{v1[0] * c, v1[1] * c, v1[2] * c, v1[3] * c}
}

// Mul performs a scalar multiplication between the vector and some constant value
// c. This is equivalent to iterating over every vector element and multiplying by c.
func (v1 Vec4i) Mulf(c float32) mgl32.Vec4 {
	return mgl32.Vec4{float32(v1[0]) * c, float32(v1[1]) * c, float32(v1[2]) * c, float32(v1[3]) * c}
}

// Dot returns the dot product of this vector with another. There are multiple ways
// to describe this value. One is the multiplication of their lengths and cos(theta) where
// theta is the angle between the vectors: v1.v2 = |v1||v2|cos(theta).
//
// The other (and what is actually done) is the sum of the element-wise multiplication of all
// elements. So for instance, two Vec3s would yield v1.x * v2.x + v1.y * v2.y + v1.z * v2.z.
//
// This means that the dot product of a vector and itself is the square of its Len (within
// the bounds of floating points error).
//
// The dot product is roughly a measure of how closely two vectors are to pointing in the same
// direction. If both vectors are normalized, the value will be -1 for opposite pointing,
// one for same pointing, and 0 for perpendicular vectors.
func (v1 Vec4i) Dot(v2 Vec4i) int {
	return v1[0]*v2[0] + v1[1]*v2[1] + v1[2]*v2[2] + v1[3]*v2[3]
}

// Len returns the vector's length. Note that this is NOT the dimension of
// the vector (len(v)), but the mathematical length. This is equivalent to the square
// root of the sum of the squares of all elements. E.G. for a Vec2i it's
// math.Hypot(v[0], v[1]).
func (v1 Vec4i) Len() float32 {
	return float32(math.Sqrt(float64(v1[0]*v1[0] + v1[1]*v1[1] + v1[2]*v1[2] + v1[3]*v1[3])))
}

// LenSqr returns the vector's square length. This is equivalent to the sum of the squares of all elements.
func (v1 Vec4i) LenSqr() int {
	return v1[0]*v1[0] + v1[1]*v1[1] + v1[2]*v1[2] + v1[3]*v1[3]
}

// X is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec4i) X() int {
	return v[0]
}

// Y is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec4i) Y() int {
	return v[1]
}

// Z is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec4i) Z() int {
	return v[2]
}

// W is an element access func, it is equivalent to v[n] where
// n is some valid index. The mappings are XYZW (X=0, Y=1 etc). Benchmarks
// show that this is more or less as fast as direct acces, probably due to
// inlining, so use v[0] or v.X() depending on personal preference.
func (v Vec4i) W() int {
	return v[3]
}
