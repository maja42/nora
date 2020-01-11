package nora

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Transformations
//	Handling is similar to SFML: https://www.sfml-dev.org/tutorials/2.5/graphics-transform.php
//	Models embed the "Transform" type. This represents the parent-to-local transformation.
//	Users can define their own models by embedding a Transform themselves or, if not all
//	transformation functions should be exported, by using it as a member and only exposing
//  specific functionality.
//
//	If the functionality provided by Transform is not sufficient, it is also possible to work with
//	a raw mgl32.Mat type and perform the operations manually.
//
// Drawing
//	Each model implements the Draw()-method.
//	In the easiest scenario, the model applies the transform-matrix to the passed render state and
//	tells the render target to draw one or more meshes.
//	The transform-matrix can be retrieved from the before mentioned Transform, or by managing
// 	a custom mgl32.Mat4 type. It's also possible to skip that step if the model has no local
//  transformation (it shares it's parent transformation matrix).
//
// Object hierarchies (scene graph)
//	The scene graph itself only contains the root models and has no notion of hierarchies.
//	Those are managed by the models themselves.
//	Models apply their own transformation matrix and then delegate drawing to their children.
//  This produces a temporary matrix stack during rendering, and children don't need to know
//	their parent (if they have one).
//
//  Note that the resulting world->model transform can also be cached, but this is in the responsibility
//  of the models.
//

// Transform manages a transformation matrix that can be positioned
// as well as rotated and scaled around an arbitrary pivot point.
// Offers functionality to easily modify any of those properties independently.
type Transform struct {
	origin   mgl32.Vec3
	position mgl32.Vec3
	rotation mgl32.Quat
	scaling  mgl32.Vec3

	transform      mgl32.Mat4
	transformDirty bool
	inverse        mgl32.Mat4
	inverseDirty   bool
}

// NewTransform creates a new identity-transformation.
func NewTransform() *Transform {
	t := &Transform{}
	t.ClearTransform()
	return t
}

// ClearTransform resets the transformation.
// Can also be used for initialization instead of NewTransform().
func (t *Transform) ClearTransform() {
	t.origin = [3]float32{0, 0, 0}
	t.position = [3]float32{0, 0, 0}
	t.rotation = mgl32.QuatIdent()
	t.scaling = [3]float32{1, 1, 1}

	t.transform = mgl32.Ident4()
	t.transformDirty = false
	t.inverse = mgl32.Ident4()
	t.inverseDirty = false
}

// SetPosition sets the 3D object position.
func (t *Transform) SetPosition(pos mgl32.Vec3) {
	t.position = pos
	t.transformDirty, t.inverseDirty = true, true
}

// SetPositionXY sets the 2D object position.
func (t *Transform) SetPositionXY(x, y float32) {
	t.position[0], t.position[1] = x, y
	t.transformDirty, t.inverseDirty = true, true
}

// SetPositionZ sets the object's Z position.
func (t *Transform) SetPositionZ(z float32) {
	t.position[2] = z
	t.transformDirty, t.inverseDirty = true, true
}

// SetPositionXYZ sets the 3D object position.
func (t *Transform) SetPositionXYZ(x, y, z float32) {
	t.position = [3]float32{x, y, z}
	t.transformDirty, t.inverseDirty = true, true
}

// Move translates the object in 3D space.
func (t *Transform) Move(movement mgl32.Vec3) {
	t.origin.Add(movement)
	t.transformDirty, t.inverseDirty = true, true
}

// MoveXY translates the object in 2D space.
func (t *Transform) MoveXY(x, y float32) {
	t.position[0] += x
	t.position[1] += y
	t.transformDirty, t.inverseDirty = true, true
}

// MoveXYZ translates the object in 3D space.
func (t *Transform) MoveXYZ(x, y, z float32) {
	t.position[0] += x
	t.position[1] += y
	t.position[2] += z
	t.transformDirty, t.inverseDirty = true, true
}

// GetPosition returns the 3D object position.
func (t *Transform) GetPosition() mgl32.Vec3 {
	return t.position
}

// SetRotation sets the object's 3D rotation.
func (t *Transform) SetRotation(rot mgl32.Quat) {
	t.rotation = rot
	t.transformDirty, t.inverseDirty = true, true
}

// SetAxisRotation sets the object's rotation around a 3D axis.
func (t *Transform) SetAxisRotation(radians float32, axis mgl32.Vec3) {
	t.rotation = mgl32.QuatRotate(radians, axis)
	t.transformDirty, t.inverseDirty = true, true
}

// SetRotationX sets the object's rotation around the x-axis. Resets all other rotations.
func (t *Transform) SetRotationX(radians float32) {
	t.rotation = mgl32.QuatRotate(radians, mgl32.Vec3([3]float32{1, 0, 0}))
	t.transformDirty, t.inverseDirty = true, true
}

// SetRotationY sets the object's rotation around the y-axis. Resets all other rotations.
func (t *Transform) SetRotationY(radians float32) {
	t.rotation = mgl32.QuatRotate(radians, mgl32.Vec3([3]float32{0, 1, 0}))
	t.transformDirty, t.inverseDirty = true, true
}

// SetRotationZ sets the object's rotation around the z-axis. Resets all other rotations.
// Used for 2D rotations.
func (t *Transform) SetRotationZ(radians float32) {
	t.rotation = mgl32.QuatRotate(radians, mgl32.Vec3([3]float32{0, 0, 1}))
	t.transformDirty, t.inverseDirty = true, true
}

// RotateX rotates the object along its x-axis.
func (t *Transform) RotateX(radians float32) {
	radians /= 2
	bx := float32(math.Sin(float64(radians)))
	bw := float32(math.Cos(float64(radians)))

	t.rotation.V, t.rotation.W = [3]float32{
		t.rotation.V[0]*bw + t.rotation.W*bx,
		t.rotation.V[1]*bw + t.rotation.V[2]*bx,
		t.rotation.V[2]*bw - t.rotation.V[1]*bx,
	}, t.rotation.W*bw-t.rotation.V[0]*bx

	t.transformDirty, t.inverseDirty = true, true
}

// RotateY rotates the object along its y-axis.
func (t *Transform) RotateY(radians float32) {
	radians /= 2
	by := float32(math.Sin(float64(radians)))
	bw := float32(math.Cos(float64(radians)))

	t.rotation.V, t.rotation.W = [3]float32{
		t.rotation.V[0]*bw - t.rotation.V[2]*by,
		t.rotation.V[1]*bw + t.rotation.W*by,
		t.rotation.V[2]*bw + t.rotation.V[0]*by,
	}, t.rotation.W*bw-t.rotation.V[1]*by

	t.transformDirty, t.inverseDirty = true, true
}

// RotateZ rotates the object along its z-axis.
// Used for 2D rotations.
func (t *Transform) RotateZ(radians float32) {
	radians /= 2
	bz := float32(math.Sin(float64(radians)))
	bw := float32(math.Cos(float64(radians)))

	t.rotation.V, t.rotation.W = [3]float32{
		t.rotation.V[0]*bw + t.rotation.V[1]*bz,
		t.rotation.V[1]*bw - t.rotation.V[0]*bz,
		t.rotation.V[2]*bw + t.rotation.W*bz,
	}, t.rotation.W*bw-t.rotation.V[2]*bz

	t.transformDirty, t.inverseDirty = true, true
}

// GetAxisRotation returns the object's rotation as an rotation axis ang angle in radians.
func (t *Transform) GetAxisRotation() (mgl32.Vec3, float32) {
	// not supported by mgl32 --> custom implementation (based on http://glmatrix.net/)
	rad := float32(math.Acos(float64(t.rotation.W))) * 2
	s := float32(math.Sin(float64(rad / 2)))

	if s != 0 {
		s = 1 / s
		return t.rotation.V.Mul(s), rad
	} else {
		// If s is zero, return any axis (no rotation - axis does not matter)
		return [3]float32{1, 0, 0}, rad
	}
}

// GetRotation returns the object's rotation as a quaternion.
func (t *Transform) GetRotation() mgl32.Quat {
	return t.rotation
}

// SetUniformScale sets a uniform scaling for all 3 dimensions of the object.
func (t *Transform) SetUniformScale(scale float32) {
	t.scaling = [3]float32{scale, scale, scale}
	t.transformDirty, t.inverseDirty = true, true
}

// SetScale sets the object's 3D scaling.
func (t *Transform) SetScale(scale mgl32.Vec3) {
	t.scaling = scale
	t.transformDirty, t.inverseDirty = true, true
}

// SetScaleXY sets the object's 2D scaling. The z-scaling will be 1.
func (t *Transform) SetScaleXY(x, y float32) {
	t.scaling = [3]float32{x, y, 1}
	t.transformDirty, t.inverseDirty = true, true
}

// SetScaleXYZ sets the object's 3D scaling.
func (t *Transform) SetScaleXYZ(x, y, z float32) {
	t.scaling = [3]float32{x, y, z}
	t.transformDirty, t.inverseDirty = true, true
}

// UniformScale scales the object uniformly in all 3 dimensions.
func (t *Transform) UniformScale(scale float32) {
	t.scaling = t.scaling.Mul(scale)
	t.transformDirty, t.inverseDirty = true, true
}

// Scale scales the object in 3 dimensions.
func (t *Transform) Scale(scale mgl32.Vec3) {
	t.scaling[0] *= scale[0]
	t.scaling[1] *= scale[1]
	t.scaling[2] *= scale[2]
	t.transformDirty, t.inverseDirty = true, true
}

// ScaleXY scales the object in 2 dimensions.
func (t *Transform) ScaleXY(x, y float32) {
	t.scaling[0] *= x
	t.scaling[1] *= y
	t.transformDirty, t.inverseDirty = true, true
}

// ScaleXYZ scales the object in 3 dimensions.
func (t *Transform) ScaleXYZ(x, y, z float32) {
	t.scaling[0] *= x
	t.scaling[1] *= y
	t.scaling[2] *= z
	t.transformDirty, t.inverseDirty = true, true
}

// GetScale returns the object's 3D scaling
func (t *Transform) GetScale() mgl32.Vec3 {
	return t.scaling
}

// SetOrigin sets the object's pivot point for scaling and rotations.
func (t *Transform) SetOrigin(origin mgl32.Vec3) {
	t.origin = origin
	t.transformDirty, t.inverseDirty = true, true
}

// SetOriginXYZ sets the object's pivot point for scaling and rotations.
func (t *Transform) SetOriginXYZ(x, y, z float32) {
	t.origin = [3]float32{x, y, z}
	t.transformDirty, t.inverseDirty = true, true
}

// GetOrigin returns the object's pivot point for scaling and rotations.
func (t *Transform) GetOrigin() mgl32.Vec3 {
	return t.origin
}

// GetTransform returns the object's transformation matrix.
func (t *Transform) GetTransform() mgl32.Mat4 {
	if t.transformDirty {
		t.transform = FromRotationTranslationScaleOrigin(t.rotation, t.position, t.scaling, t.origin)
		t.transformDirty = false
	}
	return t.transform
}

// GetInverseTransform returns the object's inverse transformation matrix.
func (t *Transform) GetInverseTransform() mgl32.Mat4 {
	if t.inverseDirty {
		t.inverse = t.GetTransform().Inv()
		t.inverseDirty = false
	}
	return t.inverse
}

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
