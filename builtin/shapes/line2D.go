package shapes

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
)

type LineJoint int

const (
	MitterJoint = LineJoint(iota)
	BevelJoint
)

type Line2D struct {
	nora.AttachableModel
	nora.Transform
	mesh nora.Mesh

	thickness float32
	lineJoint LineJoint
	loop      bool
	color     color.Color

	points []mgl32.Vec2
	dirty  bool
}

func NewLine2D(thickness float32, lineJoint LineJoint, loop bool) *Line2D {
	mat := nora.NewMaterial(shader.COL_2D)

	s := &Line2D{
		mesh: *nora.NewMesh(mat),
	}
	s.Clear()
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
func (m *Line2D) Points() []mgl32.Vec2 {
	return m.points
}

// AddPoints appends additional waypoints to the line
func (m *Line2D) AddPoints(p ...mgl32.Vec2) {
	m.points = append(m.points, p...)
	m.dirty = true
}

// ClearPoints removes all points.
func (m *Line2D) ClearPoints() {
	m.points = nil
	m.dirty = true
}

// Color returns the line's color.
func (m *Line2D) Color() color.Color {
	return m.color
}

// SetColor changes the line color.
func (m *Line2D) SetColor(c color.Color) {
	m.color = c
	m.mesh.Material().Uniform4f("fragColor", c.R, c.G, c.B, c.A)
}

func (m *Line2D) Update(elapsed time.Duration) {
	if !m.dirty {
		return
	}
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

		line := mgl32.Vec2{end[0] - start[0], end[1] - start[1]}
		norm := mgl32.Vec2{-line[1], line[0]}
		norm = norm.Normalize()
		norm = norm.Mul(m.thickness / 2)

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

	m.mesh.SetVertexData(vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.CompactBuffer)
}

func (m *Line2D) updateMiter() {
	pointCnt := len(m.points)
	if pointCnt < 2 {
		m.mesh.ClearVertexData()
		return
	}

	vertexCount := pointCnt * 2
	if m.loop {
		vertexCount += 2
	}
	vertices := make([]float32, vertexCount*2)

	var p0, p1, p2 mgl32.Vec2
	var v1, v2, v3 mgl32.Vec2

	for i := 0; i < pointCnt; i++ {
		// Get the two segments shared by the current point
		if i == 0 {
			p0 = m.points[pointCnt-1]
		} else {
			p0 = m.points[i-1]
		}
		p1 = m.points[i]
		if i == pointCnt-1 {
			p2 = m.points[0]
		} else {
			p2 = m.points[i+1]
		}

		// Compute their normal
		v1 = computeNormal(p0, p1)
		v2 = computeNormal(p1, p2)

		// Combine the normals
		factor := 1 + (v1[0]*v2[0] + v1[1]*v2[1])
		v3 = v1.Add(v2)
		v3 = v3.Mul(m.thickness / 2 / factor)

		// calculate the final vertices
		v1 = p1.Add(v3)
		v2 = p1.Sub(v3)

		copy(vertices[i*4:], []float32{
			v1[0], v1[1],
			v2[0], v2[1],
		})
	}

	if m.loop {
		copy(vertices[pointCnt*4:], []float32{
			vertices[0], vertices[1],
			vertices[2], vertices[3],
		})
	}
	m.mesh.SetVertexData(vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.CompactBuffer)
}

func computeNormal(from, to mgl32.Vec2) mgl32.Vec2 {
	norm := mgl32.Vec2{from[1] - to[1], to[0] - from[0]}
	return norm.Normalize()
}

func (m *Line2D) Draw(renderState *nora.RenderState) {
	renderState.TransformStack.RightMul(m.GetTransform())
	m.mesh.Draw(renderState)
}
