package nora

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/maja42/nora/assert"
)

// Camera defines the world's area that is visible on the screen.
type Camera interface {
	// Matrix returns the view-projection matrix and its change-counter.
	// The change-counter is incremented every time camera properties are modified.
	Matrix() (mgl32.Mat4, int)
}

type OrthoCamera struct {
	pos           mgl32.Vec2
	orthoSizeHalf mgl32.Vec2
	aspectRatio   float32

	nearPlane float32
	farPlane  float32

	vpMatrix        mgl32.Mat4
	inverseVMMatrix mgl32.Mat4

	dirtyCount int
}

// NewOrthoCamera creates a new orthogonal camera.
// By default, shows everything within [-1, +1].
func NewOrthoCamera() *OrthoCamera {
	cam := &OrthoCamera{
		pos:           [2]float32{0, 0},
		orthoSizeHalf: [2]float32{1, 1},
		aspectRatio:   1,

		nearPlane: 1,
		farPlane:  -1,
	}
	cam.updateVPMatrices()
	return cam
}

// Position returns the camera's world position (center).
func (c *OrthoCamera) Position() mgl32.Vec2 {
	return c.pos
}

// OrthoWidth returns the horizontally visible space in world coordinates.
func (c *OrthoCamera) OrthoWidth() float32 {
	return c.orthoSizeHalf[0] * 2
}

// OrthoHeight returns the vertically visible space in world coordinates.
func (c *OrthoCamera) OrthoHeight() float32 {
	return c.orthoSizeHalf[1] * 2
}

// OrthoSize returns the visible space in world coordinates.
func (c *OrthoCamera) OrthoSize() mgl32.Vec2 {
	return c.orthoSizeHalf.Mul(2)
}

// AspectRatio returns the camera's aspect ratio.
func (c *OrthoCamera) AspectRatio() float32 {
	return c.aspectRatio
}

// Left returns the position of camera's left plane in world coordinates.
func (c *OrthoCamera) Left() float32 {
	return c.pos[0] - c.orthoSizeHalf[0]
}

// Top returns the position of camera's top plane in world coordinates.
func (c *OrthoCamera) Top() float32 {
	return c.pos[1] + c.orthoSizeHalf[1]
}

// Right returns the position of camera's right plane in world coordinates.
func (c *OrthoCamera) Right() float32 {
	return c.pos[0] + c.orthoSizeHalf[0]
}

// Bottom returns the position of camera's bottom plane in world coordinates.
func (c *OrthoCamera) Bottom() float32 {
	return c.pos[1] - c.orthoSizeHalf[1]
}

// SetPosition sets the camera center's world position.
func (c *OrthoCamera) SetPosition(vec mgl32.Vec2) {
	c.pos = vec
	c.updateVPMatrices()
}

// SetPositionXY sets the camera's world position.
func (c *OrthoCamera) SetPositionXY(x, y float32) {
	c.pos = [2]float32{x, y}
	c.updateVPMatrices()
}

// Move translates the camera into the opposite direction.
func (c *OrthoCamera) Move(vec mgl32.Vec2) {
	c.pos = c.pos.Sub(vec) // camera moves to opposite direction
	c.updateVPMatrices()
}

// MoveXY translates the camera into the opposite direction.
func (c *OrthoCamera) MoveXY(x, y float32) {
	c.pos[0] -= x // camera moves to opposite direction
	c.pos[1] -= y
	c.updateVPMatrices()
}

// SetTrueOrthoWidth sets the camera's ortho width and updates the aspect ratio.
// Does not change the ortho height - distorts the image.
func (c *OrthoCamera) SetTrueOrthoWidth(width float32) {
	if !assert.True(width > 0, "Invalid ortho width %f (must be >0)", width) {
		return
	}
	c.orthoSizeHalf[0] = width / 2
	c.aspectRatio = c.orthoSizeHalf[0] / c.orthoSizeHalf[1]
	c.updateVPMatrices()
}

// SetTrueOrthoHeight sets the camera's ortho height and updates the aspect ratio.
// Does not change the ortho width - distorts the image.
func (c *OrthoCamera) SetTrueOrthoHeight(height float32) {
	if !assert.True(height > 0, "Invalid ortho height %f (must be >0)", height) {
		return
	}
	c.orthoSizeHalf[1] = height / 2
	c.aspectRatio = c.orthoSizeHalf[0] / c.orthoSizeHalf[1]
	c.updateVPMatrices()
}

// SetOrthoWidth updates the camera zoom/scale to achieve the given ortho width.
// Does not affect the aspect ratio (ortho height is also updated).
func (c *OrthoCamera) SetOrthoWidth(width float32) {
	if !assert.True(width > 0, "Invalid ortho width %f (must be >0)", width) {
		return
	}
	c.orthoSizeHalf[0] = width / 2
	c.orthoSizeHalf[1] = c.orthoSizeHalf[0] / c.aspectRatio
	c.updateVPMatrices()
}

// SetOrthoHeight updates the camera zoom/scale to achieve the given ortho height.
// Does not affect the aspect ratio (ortho width is also updated).
func (c *OrthoCamera) SetOrthoHeight(height float32) {
	if !assert.True(height > 0, "Invalid ortho height %f (must be >0)", height) {
		return
	}

	orthoWidth := height * c.aspectRatio
	c.SetOrthoWidth(orthoWidth)
}

// SetAspectRatio changes the camera's aspect ratio.
// If keepOrthoHeight is true, the ortho height is kept; otherwise the ortho width is kept.
func (c *OrthoCamera) SetAspectRatio(aspectRatio float32, keepOrthoHeight bool) {
	c.aspectRatio = aspectRatio
	if keepOrthoHeight { // adjust width
		c.orthoSizeHalf[0] = c.orthoSizeHalf[1] * c.aspectRatio
	} else { // adjust height
		c.orthoSizeHalf[1] = c.orthoSizeHalf[0] / c.aspectRatio
	}
	c.updateVPMatrices()
}

// FocusArea focuses the given area by ensuring that everything is visible.
// If the aspect ratio of the given area is different than the ratio of the camera,
// the camera zooms out, unveiling additional space to the viewer.
func (c *OrthoCamera) FocusArea(bottomLeft mgl32.Vec2, size mgl32.Vec2) {
	center := size.Mul(0.5).Add(bottomLeft)

	visAreaAspectRatio := size[0] / size[1]

	c.pos = center
	if visAreaAspectRatio > c.aspectRatio {
		c.SetOrthoWidth(size[0])
	} else {
		c.SetOrthoHeight(size[1])
	}
}

func (c *OrthoCamera) updateVPMatrices() {
	c.vpMatrix = mgl32.Ortho(c.Left(), c.Right(), c.Bottom(), c.Top(), c.nearPlane, c.farPlane)
	c.inverseVMMatrix = c.vpMatrix.Inv()
	c.dirtyCount++
}

// ClipSpaceToWorldSpace converts 2D clip space [-1, +1] into 2D world coordinates.
func (c *OrthoCamera) ClipSpaceToWorldSpace(clipSpace mgl32.Vec2) mgl32.Vec2 {
	homogeneous := clipSpace.Vec4(0, 1)
	worldSpace := c.inverseVMMatrix.Mul4x1(homogeneous)
	return worldSpace.Vec2()
}

// WorldSpaceToClipSpace converts 2D world coordinates into 2D clip space [-1, +1].
func (c *OrthoCamera) WorldSpaceToClipSpace(worldSpace mgl32.Vec2) mgl32.Vec2 {
	homogeneous := worldSpace.Vec4(0, 1)
	clipSpace := c.vpMatrix.Mul4x1(homogeneous)
	return clipSpace.Vec2()
}

// Matrix returns the view-projection matrix and its change-counter.
// The change-counter must be incremented every time camera properties are modified.
func (c *OrthoCamera) Matrix() (mgl32.Mat4, int) {
	return c.vpMatrix, c.dirtyCount
}
