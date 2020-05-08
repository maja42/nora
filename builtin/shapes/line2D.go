package shapes

import (
	"github.com/maja42/nora"
	"github.com/maja42/nora/assert"
	"github.com/maja42/nora/builtin/geometry/geo2d"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
	"github.com/maja42/vmath"
)

type Line2D struct {
	nora.Transform
	mesh nora.Mesh

	thickness float32
	lineJoint geo2d.LineJoint
	loop      bool
	color     color.Color

	points []vmath.Vec2f
	dirty  bool
}

func NewLine2D(thickness float32, lineJoint geo2d.LineJoint, loop bool) *Line2D {
	mat := nora.NewMaterial(shader.COL_2D)

	s := &Line2D{
		mesh: *nora.NewMesh(mat),
	}
	s.ClearTransform()
	s.SetColor(color.White)
	s.SetProperties(thickness, lineJoint, loop)
	return s
}

func (m *Line2D) Destroy() {
	m.mesh.Destroy()
}

// Properties returns the line's thickness, the joint type and if it's a loop.
func (m *Line2D) Properties() (float32, geo2d.LineJoint, bool) {
	return m.thickness, m.lineJoint, m.loop
}

// SetProperties changes the thickness, joint type and if the line loops back.
func (m *Line2D) SetProperties(thickness float32, lineJoint geo2d.LineJoint, loop bool) {
	m.thickness = thickness
	m.lineJoint = lineJoint
	m.loop = loop
	m.dirty = true
}

// Points returns all waypoints of the line.
// The caller must not modify the returned data.
func (m *Line2D) Points() []vmath.Vec2f {
	return m.points
}

// Point returns the waypoint with the given index.
// Panics if the index is out of bounds.
func (m *Line2D) Point(idx int) vmath.Vec2f {
	assert.True(idx >= 0 && idx < len(m.points), "index out of range")
	return m.points[idx]
}

// SetPoint changes the position of an existing waypoint.
// Panics if the index is out of bounds.
func (m *Line2D) SetPoint(idx int, p vmath.Vec2f) {
	assert.True(idx >= 0 && idx < len(m.points), "index out of range")
	m.points[idx] = p
	m.dirty = true
}

// AddPoints appends additional waypoints to the line.
func (m *Line2D) AddPoints(p ...vmath.Vec2f) {
	m.points = append(m.points, p...)
	m.dirty = true
}

// AddPointsAtFront appends additional waypoints to the beginning of the line.
// The order of the given points stays the same
func (m *Line2D) AddPointsAtFront(p ...vmath.Vec2f) {
	m.points = append(p, m.points...)
	m.dirty = true
}

// AddPointAtIdx adds an additional waypoint at any position of the line.
func (m *Line2D) AddPointAtIdx(idx int, p vmath.Vec2f) {
	assert.True(idx >= 0 && idx < len(m.points), "index out of range")

	// make sure there's enough space:
	m.points = append(m.points, vmath.Vec2f{})
	copy(m.points[idx+1:], m.points[idx:len(m.points)-1])
	m.points[idx] = p
	m.dirty = true
}

// RemovePoint removes a point from the line.
func (m *Line2D) RemovePoint(idx int) bool {
	if idx < 0 || idx >= len(m.points) {
		return false
	}
	m.points = append(m.points[:idx], m.points[idx+1:]...)
	m.dirty = true
	return true
}

// RemoveLastPoints removes a last n points from the line.
// Returns the number of points actually removed.
func (m *Line2D) RemoveLastPoints(count int) int {
	count = vmath.Mini(count, len(m.points))
	m.points = m.points[:len(m.points)-count]
	m.dirty = true
	return count
}

// Reverse reverses the line.
// This flips all positions.
func (m *Line2D) Reverse() {
	p := m.points

	left := 0
	right := len(p) - 1
	for ; left < right; left, right = left+1, right-1 {
		p[left], p[right] = p[right], p[left]
	}
	m.dirty = true
}

// ClearPoints removes all points.
func (m *Line2D) ClearPoints() {
	m.points = nil
	m.dirty = true
}

func (m *Line2D) Length() int {
	return len(m.points)
}

// Color returns the line's color.
func (m *Line2D) Color() color.Color {
	return m.color
}

// SetColor changes the line color.
func (m *Line2D) SetColor(c color.Color) {
	m.color = c
	m.mesh.Material().Uniform4fColor("fragColor", c)
}

func (m *Line2D) Draw(renderState *nora.RenderState) {
	if m.dirty {
		m.update()
	}
	renderState.TransformStack.PushMulRight(m.GetTransform())
	m.mesh.Draw(renderState)
	renderState.TransformStack.Pop()
}

func (m *Line2D) update() {
	geometry := geo2d.Line(m.points, m.loop, m.thickness, m.lineJoint)
	m.mesh.SetGeometry(geometry)
	m.dirty = false
}
