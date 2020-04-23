package shapes

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
)

type Sprite struct {
	nora.Transform
	mesh nora.Mesh

	sum float64
}

func NewSprite() *Sprite {
	mat := nora.NewMaterial(shader.TEX_2D)

	s := &Sprite{
		mesh: *nora.NewMesh(mat),
	}
	s.ClearTransform()

	/* counter-clockwise
	   3 - 2
	   | / |
	   0 - 1
	*/

	vertices := []float32{
		/*xy*/ 0, 0 /*uv*/, 0, 0, // 0
		/*xy*/ 1, 0 /*uv*/, 1, 0, // 1
		/*xy*/ 1, 1 /*uv*/, 1, 1, // 2

		/*xy*/ 1, 1 /*uv*/, 1, 1, // 2
		/*xy*/ 0, 1 /*uv*/, 0, 1, // 3
		/*xy*/ 0, 0 /*uv*/, 0, 0, // 0
	}
	s.mesh.SetVertexData(6, vertices, nil, gl.TRIANGLES, []string{"position", "texCoord"}, nora.InterleavedBuffer)
	return s
}

func (m *Sprite) SetTexture(texKey nora.TextureKey) {
	m.mesh.Material().AddTextureBinding("sampler", texKey)
}

func (m *Sprite) Destroy() {
	m.mesh.Destroy()
}

func (m *Sprite) Draw(renderState *nora.RenderState) {
	m.mesh.TransDraw(renderState, m.GetTransform())
}
