package rendering

import "github.com/go-gl/mathgl/mgl32"

// FromRotationTranslationScaleOrigin creates a matrix from a quaternion rotation, vector translation and vector scale,
// rotating and scaling around the given origin.
//
// Based on http://glmatrix.net/
func FromRotationTranslationScaleOrigin(rotation mgl32.Quat, translation, scaling, origin mgl32.Vec3) mgl32.Mat4 {
	x, y, z, w := rotation.V[0], rotation.V[1], rotation.V[2], rotation.W
	x2 := x + x
	y2 := y + y
	z2 := z + z

	xx := x * x2
	xy := x * y2
	xz := x * z2
	yy := y * y2
	yz := y * z2
	zz := z * z2
	wx := w * x2
	wy := w * y2
	wz := w * z2

	sx := scaling[0]
	sy := scaling[1]
	sz := scaling[2]

	ox := origin[0]
	oy := origin[1]
	oz := origin[2]

	var mat = [16]float32{}
	mat[0] = (1 - (yy + zz)) * sx
	mat[1] = (xy + wz) * sx
	mat[2] = (xz - wy) * sx
	mat[3] = 0
	mat[4] = (xy - wz) * sy
	mat[5] = (1 - (xx + zz)) * sy
	mat[6] = (yz + wx) * sy
	mat[7] = 0
	mat[8] = (xz + wy) * sz
	mat[9] = (yz - wx) * sz
	mat[10] = (1 - (xx + yy)) * sz
	mat[11] = 0
	mat[12] = translation[0] + ox - (mat[0]*ox + mat[4]*oy + mat[8]*oz)
	mat[13] = translation[1] + oy - (mat[1]*ox + mat[5]*oy + mat[9]*oz)
	mat[14] = translation[2] + oz - (mat[2]*ox + mat[6]*oy + mat[10]*oz)
	mat[15] = 1
	return mat
}
