package shapes

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
)

type Triangle2D struct {
	nora.AttachableModel
	nora.Transform
	mesh nora.Mesh
}

func NewTriangle2D() *Triangle2D {
	mat := nora.NewMaterial(shader.COL_2D)
	mat.Uniform4fColor("fragColor", color.White)

	s := &Triangle2D{
		mesh: *nora.NewMesh(mat),
	}
	s.ClearTransform()

	vertices := []float32{
		/*xy*/ 0, 0,
		/*xy*/ 1, 0,
		/*xy*/ 1, 1,
	}
	s.mesh.SetVertexData(3, vertices, nil, gl.TRIANGLES, []string{"position"}, nora.InterleavedBuffer)
	return s
}

func (m *Triangle2D) Destroy() {
	m.mesh.Destroy()
}

func (m *Triangle2D) SetColor(c color.Color) {
	m.mesh.Material().Uniform4fColor("fragColor", c)
}

func (m *Triangle2D) Draw(renderState *nora.RenderState) {
	renderState.TransformStack.RightMul(m.GetTransform())
	m.mesh.Draw(renderState)
}
