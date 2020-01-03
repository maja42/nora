package shapes

import (
	"time"

	"github.com/maja42/logicat/rendering"
	"github.com/maja42/logicat/rendering/color"
	"github.com/maja42/logicat/rendering/resources/shader"
	"golang.org/x/mobile/gl"
)

type Cube struct {
	rendering.AttachableModel
	rendering.Transform
	mesh  rendering.Mesh
	dirty bool

	size  float32
	color color.Color
}

func NewCube(ctx *rendering.Context, size float32) *Cube {
	mat := rendering.NewMaterial(shader.COL_NORM_3D)
	mat.Uniform4f("fragColor", 1.5, 0.5, 1.5, 1.0)

	cube := &Cube{
		mesh:  *rendering.NewMesh(ctx, mat),
		dirty: true,
		size:  size,
		color: color.White,
	}
	cube.Clear()
	return cube
}

func (m *Cube) Destroy() {
	m.mesh.Destroy()
}

func (m *Cube) Size() float32 {
	return m.size
}

func (m *Cube) SetSize(s float32) {
	m.size = s
	m.dirty = true
}

func (m *Cube) Color() color.Color {
	return m.color
}

func (m *Cube) SetColor(c color.Color) {
	m.color = c
	m.dirty = true
}

func (m *Cube) Update(time.Duration) {
	if !m.dirty {
		return
	}
	m.dirty = false

	s := m.size

	vertices := []float32{ // triangles = counter clock wise
		// front:
		/*xyz*/ 0, 0, 0 /*norm*/, 0, 0, -1,
		/*xyz*/ s, 0, 0 /*norm*/, 0, 0, -1,
		/*xyz*/ s, s, 0 /*norm*/, 0, 0, -1,
		/*xyz*/ 0, s, 0 /*norm*/, 0, 0, -1,
		// right
		/*xyz*/ s, 0, 0 /*norm*/, 1, 0, 0,
		/*xyz*/ s, 0, s /*norm*/, 1, 0, 0,
		/*xyz*/ s, s, s /*norm*/, 1, 0, 0,
		/*xyz*/ s, s, 0 /*norm*/, 1, 0, 0,
		// left
		/*xyz*/ 0, 0, s /*norm*/, -1, 0, 0,
		/*xyz*/ 0, 0, 0 /*norm*/, -1, 0, 0,
		/*xyz*/ 0, s, 0 /*norm*/, -1, 0, 0,
		/*xyz*/ 0, s, s /*norm*/, -1, 0, 0,
		// back
		/*xyz*/ s, 0, s /*norm*/, 0, 0, 1,
		/*xyz*/ 0, 0, s /*norm*/, 0, 0, 1,
		/*xyz*/ 0, s, s /*norm*/, 0, 0, 1,
		/*xyz*/ s, s, s /*norm*/, 0, 0, 1,
		// top
		/*xyz*/ 0, s, 0 /*norm*/, 0, 1, 0,
		/*xyz*/ s, s, 0 /*norm*/, 0, 1, 0,
		/*xyz*/ s, s, s /*norm*/, 0, 1, 0,
		/*xyz*/ 0, s, s /*norm*/, 0, 1, 0,
		// bottom
		/*xyz*/ 0, 0, 0 /*norm*/, 0, -1, 0,
		/*xyz*/ s, 0, 0 /*norm*/, 0, -1, 0,
		/*xyz*/ s, 0, s /*norm*/, 0, -1, 0,
		/*xyz*/ 0, 0, s /*norm*/, 0, -1, 0,
	}
	indices := []uint16{
		0, 1, 2, 0, 2, 3, // front
		4, 5, 6, 4, 6, 7, // right
		8, 9, 10, 8, 10, 11, // left
		12, 13, 14, 12, 14, 15, // back
		16, 17, 18, 16, 18, 19, // top
		20, 22, 21, 20, 23, 22, // bottom
	}
	m.mesh.SetVertexData(vertices, indices, gl.TRIANGLES, []string{"position", "normal"}, rendering.InterleavedBuffer)
}

func (m *Cube) Draw(renderTarget rendering.RenderTarget, renderState *rendering.RenderState) {
	renderState.TransformStack.RightMul(m.GetTransform())
	renderTarget.Draw(&m.mesh, renderState)
}
