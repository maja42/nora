package nora

import (
	"github.com/maja42/vmath"
)

// Transformations
//	Handling is similar to SFML: https://www.sfml-dev.org/tutorials/2.5/graphics-transform.php
//	Models embed the "Transform" type. This represents the parent-to-local transformation.
//	Users can define their own models by embedding a Transform themselves or, if not all
//	transformation functions should be exported, by using it as a member and only exposing
//  specific functionality.
//
//	If the functionality provided by Transform is not sufficient, it is also possible to work with
//	a raw vmath.Mat4f type and perform the operations manually.
//
// Drawing
//	Each model implements the Draw()-method.
//	In the easiest scenario, the model applies the transform-matrix to the passed render state (matrix stack)
//	and	tells the render target to draw one or more meshes.
//	The transform-matrix can be retrieved from the before mentioned Transform, or by managing
// 	a custom vmath.Mat4f type. It's also possible to skip that step if the model has no local
//  transformation (it shares its parent's transformation matrix).
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
	origin   vmath.Vec3f
	position vmath.Vec3f
	rotation vmath.Quat
	scaling  vmath.Vec3f

	transform      vmath.Mat4f
	transformDirty bool
	inverse        vmath.Mat4f
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
	t.origin = vmath.Vec3f{0, 0, 0}
	t.position = vmath.Vec3f{0, 0, 0}
	t.rotation = vmath.IdentQuat()
	t.scaling = vmath.Vec3f{1, 1, 1}

	t.transform = vmath.Ident4f()
	t.transformDirty = false
	t.inverse = vmath.Ident4f()
	t.inverseDirty = false
}

// SetPosition sets the 3D object position.
func (t *Transform) SetPosition(pos vmath.Vec3f) {
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
	t.position = vmath.Vec3f{x, y, z}
	t.transformDirty, t.inverseDirty = true, true
}

// Move translates the object in 3D space.
func (t *Transform) Move(movement vmath.Vec3f) {
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
func (t *Transform) GetPosition() vmath.Vec3f {
	return t.position
}

// SetRotation sets the object's 3D rotation.
func (t *Transform) SetRotation(rot vmath.Quat) {
	t.rotation = rot
	t.transformDirty, t.inverseDirty = true, true
}

// SetAxisRotation sets the object's rotation around a 3D axis.
func (t *Transform) SetAxisRotation(rad float32, axis vmath.Vec3f) {
	t.rotation = vmath.QuatFromAxisAngle(axis, rad)
	t.transformDirty, t.inverseDirty = true, true
}

// SetRotationX sets the object's rotation around the x-axis. Resets all other rotations.
func (t *Transform) SetRotationX(rad float32) {
	t.rotation = vmath.QuatFromAxisAngle(vmath.Vec3f{1, 0, 0}, rad)
	t.transformDirty, t.inverseDirty = true, true
}

// SetRotationY sets the object's rotation around the y-axis. Resets all other rotations.
func (t *Transform) SetRotationY(rad float32) {
	t.rotation = vmath.QuatFromAxisAngle(vmath.Vec3f{0, 1, 0}, rad)
	t.transformDirty, t.inverseDirty = true, true
}

// SetRotationZ sets the object's rotation around the z-axis. Resets all other rotations.
// Used for 2D rotations.
func (t *Transform) SetRotationZ(rad float32) {
	t.rotation = vmath.QuatFromAxisAngle(vmath.Vec3f{0, 0, 1}, rad)
	t.transformDirty, t.inverseDirty = true, true
}

// RotateX rotates the object along its x-axis.
func (t *Transform) RotateX(rad float32) {
	t.rotation.RotateX(rad)
	t.transformDirty, t.inverseDirty = true, true
}

// RotateY rotates the object along its y-axis.
func (t *Transform) RotateY(rad float32) {
	t.rotation.RotateY(rad)
	t.transformDirty, t.inverseDirty = true, true
}

// RotateZ rotates the object along its z-axis.
// Used for 2D rotations.
func (t *Transform) RotateZ(rad float32) {
	t.rotation.RotateZ(rad)
	t.transformDirty, t.inverseDirty = true, true
}

// GetAxisRotation returns the object's rotation as an rotation axis ang angle in radians.
func (t *Transform) GetAxisRotation() (vmath.Vec3f, float32) {
	return t.rotation.AxisRotation()
}

// GetRotation returns the object's rotation as a quaternion.
func (t *Transform) GetRotation() vmath.Quat {
	return t.rotation
}

// SetUniformScale sets a uniform scaling for all 3 dimensions of the object.
func (t *Transform) SetUniformScale(scale float32) {
	t.scaling = vmath.Vec3f{scale, scale, scale}
	t.transformDirty, t.inverseDirty = true, true
}

// SetScale sets the object's 3D scaling.
func (t *Transform) SetScale(scale vmath.Vec3f) {
	t.scaling = scale
	t.transformDirty, t.inverseDirty = true, true
}

// SetScaleXY sets the object's 2D scaling. The z-scaling will be 1.
func (t *Transform) SetScaleXY(x, y float32) {
	t.scaling = vmath.Vec3f{x, y, 1}
	t.transformDirty, t.inverseDirty = true, true
}

// SetScaleXYZ sets the object's 3D scaling.
func (t *Transform) SetScaleXYZ(x, y, z float32) {
	t.scaling = vmath.Vec3f{x, y, z}
	t.transformDirty, t.inverseDirty = true, true
}

// UniformScale scales the object uniformly in all 3 dimensions.
func (t *Transform) UniformScale(scale float32) {
	t.scaling = t.scaling.MulScalar(scale)
	t.transformDirty, t.inverseDirty = true, true
}

// Scale scales the object in 3 dimensions.
func (t *Transform) Scale(scale vmath.Vec3f) {
	t.scaling.Mul(scale)
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
func (t *Transform) GetScale() vmath.Vec3f {
	return t.scaling
}

// SetOrigin sets the object's pivot point for scaling and rotations.
func (t *Transform) SetOrigin(origin vmath.Vec3f) {
	t.origin = origin
	t.transformDirty, t.inverseDirty = true, true
}

// SetOriginXYZ sets the object's pivot point for scaling and rotations.
func (t *Transform) SetOriginXYZ(x, y, z float32) {
	t.origin = vmath.Vec3f{x, y, z}
	t.transformDirty, t.inverseDirty = true, true
}

// GetOrigin returns the object's pivot point for scaling and rotations.
func (t *Transform) GetOrigin() vmath.Vec3f {
	return t.origin
}

// GetTransform returns the object's transformation matrix.
func (t *Transform) GetTransform() vmath.Mat4f {
	if t.transformDirty {
		t.transform = vmath.Mat4fFromRotationTranslationScaleOrigin(t.rotation, t.position, t.scaling, t.origin)
		t.transformDirty = false
	}
	return t.transform
}

// GetInverseTransform returns the object's inverse transformation matrix.
func (t *Transform) GetInverseTransform() vmath.Mat4f {
	if t.inverseDirty {
		t.inverse, _ = t.GetTransform().Inverse()
		t.inverseDirty = false
	}
	return t.inverse
}
