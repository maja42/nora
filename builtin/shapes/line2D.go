package shapes

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
	"github.com/maja42/vmath"
)

type LineJoint int

const (
	MitterJoint = LineJoint(iota)
	BevelJoint
)

type Line2D struct {
	nora.Transform
	mesh nora.Mesh

	thickness float32
	lineJoint LineJoint
	loop      bool
	color     color.Color

	points []vmath.Vec2f
	dirty  bool
}

func NewLine2D(thickness float32, lineJoint LineJoint, loop bool) *Line2D {
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
func (m *Line2D) Properties() (float32, LineJoint, bool) {
	return m.thickness, m.lineJoint, m.loop
}

// SetProperties changes the thickness, joint type and if the line loops back.
func (m *Line2D) SetProperties(thickness float32, lineJoint LineJoint, loop bool) {
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
	return m.points[idx]
}

// AddPoints appends additional waypoints to the line.
func (m *Line2D) AddPoints(p ...vmath.Vec2f) {
	m.points = append(m.points, p...)
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
	switch m.lineJoint {
	case BevelJoint:
		m.updateBevel()
	case MitterJoint:
		m.updateMiter()
	}
	m.dirty = false
}

func (m *Line2D) updateBevel() {
	pointCnt := len(m.points)
	if pointCnt < 2 {
		m.mesh.ClearVertexData()
		return
	}

	lineSegments := pointCnt - 1
	if m.loop {
		lineSegments++
	}

	vertexCount := lineSegments * 4
	if m.loop {
		vertexCount += 2
	}

	vertices := make([]float32, vertexCount*2)

	for l := 0; l < lineSegments; l++ {
		start := m.points[l]
		end := m.points[(l+1)%pointCnt]

		line := vmath.Vec2f{end[0] - start[0], end[1] - start[1]}
		norm := vmath.Vec2f{-line[1], line[0]}
		norm = norm.Normalize()
		norm = norm.MulScalar(m.thickness / 2)

		copy(vertices[l*4*2:], []float32{
			start[0] + norm[0], start[1] + norm[1],
			start[0] - norm[0], start[1] - norm[1],
			end[0] + norm[0], end[1] + norm[1],
			end[0] - norm[0], end[1] - norm[1],
		})
	}

	if m.loop {
		copy(vertices[lineSegments*4*2:], []float32{
			vertices[0], vertices[1],
			vertices[2], vertices[3],
		})
	}

	m.mesh.SetVertexData(vertexCount, vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.CompactBuffer)
}

func (m *Line2D) updateMiter() {
	pointCnt := len(m.points)
	if pointCnt < 2 {
		m.mesh.ClearVertexData()
		return
	}

	loop := m.loop
	if pointCnt == 2 {
		loop = false
	}

	vertexCount := pointCnt * 2
	if loop {
		vertexCount += 2
	}
	vertices := make([]float32, vertexCount*2)

	var p0, p1, p2 vmath.Vec2f
	var v1, v2, v3 vmath.Vec2f
	var factor float32

	for i := 0; i < pointCnt; i++ {
		p1 = m.points[i]

		if i == 0 && !loop { // start piece; compute normal to first element
			v3 = computeNormal(p1, m.points[1])
			factor = 1
		} else if i == pointCnt-1 && !loop { // end piece; compute normal to last segment
			v3 = computeNormal(m.points[i-1], p1)
			factor = 1
		} else {
			// Get the two segments shared by the current point
			if i == 0 {
				p0 = m.points[pointCnt-1]
			} else {
				p0 = m.points[i-1]
			}

			if i == pointCnt-1 {
				p2 = m.points[0]
			} else {
				p2 = m.points[i+1]
			}

			// Compute their normal
			v1 = computeNormal(p0, p1)
			v2 = computeNormal(p1, p2)

			// Combine the normals
			factor = 1 + (v1[0]*v2[0] + v1[1]*v2[1])
			v3 = v1.Add(v2)
		}

		v3 = v3.MulScalar(m.thickness / 2 / factor) // adjust thickness

		// calculate the final vertices
		v1 = p1.Add(v3)
		v2 = p1.Sub(v3)

		copy(vertices[i*4:], []float32{
			v1[0], v1[1],
			v2[0], v2[1],
		})
	}

	if loop {
		copy(vertices[pointCnt*4:], []float32{
			vertices[0], vertices[1],
			vertices[2], vertices[3],
		})
	}
	m.mesh.SetVertexData(vertexCount, vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.CompactBuffer)
}

func computeNormal(from, to vmath.Vec2f) vmath.Vec2f {
	// short of to.Sub(from).NormalVec(true).Normalize()
	norm := vmath.Vec2f{from[1] - to[1], to[0] - from[0]}
	return norm.Normalize()
}
