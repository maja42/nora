package shapes

import (
	"github.com/maja42/logicat/rendering"
	"github.com/maja42/logicat/rendering/resources/shader"
	"golang.org/x/mobile/gl"
)

type TexturedShape struct {
	rendering.AttachableModel
	rendering.Transform
	mesh rendering.Mesh

	sum float64
}

func NewTexturedShape(ctx *rendering.Context) *TexturedShape {
	mat := rendering.NewMaterial(shader.TEX_2D)

	s := &TexturedShape{
		mesh: *rendering.NewMesh(ctx, mat),
	}
	s.Clear()

	/* counter-clockwise
	   3 - 2
	   | / |
	   0 - 1
	*/

	vertices := []float32{
		/*xyz*/ 0, 0 /*uv*/, 0, 0,
		/*xyz*/ 0.5, 0 /*uv*/, 2, 0,
		/*xyz*/ 0.5, 0.5 /*uv*/, 2, 2,
		/*xyz*/ 0, 0.5 /*uv*/, 0, 2,
	}
	indices := []uint16{
		0, 1, 2,
		2, 3, 0,
	}
	s.mesh.SetVertexData(vertices, indices, gl.TRIANGLES, []string{"position", "texCoord"}, rendering.InterleavedBuffer)
	s.mesh.SetIndexSubData(0, []uint16{0, 0, 0, 0, 0, 0})
	return s
}

func (m *TexturedShape) SetTexture(texKey rendering.TextureKey) {
	m.mesh.Material().AddTextureBinding("sampler", texKey)
}

func (m *TexturedShape) Destroy() {
	m.mesh.Destroy()
}

func (m *TexturedShape) Draw(renderTarget rendering.RenderTarget, renderState *rendering.RenderState) {
	renderState.TransformStack.RightMul(m.GetTransform())
	renderTarget.Draw(&m.mesh, renderState)
}
