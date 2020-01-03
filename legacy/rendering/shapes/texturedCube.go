package shapes

import (
	"time"

	"github.com/maja42/logicat/rendering/color"

	"github.com/maja42/logicat/rendering/resources/shader"

	"github.com/maja42/logicat/rendering"
	"golang.org/x/mobile/gl"
)

type TexturedCube struct {
	rendering.AttachableModel
	rendering.Transform
	mesh  rendering.Mesh
	dirty bool

	size  float32
	color color.Color
}

func NewTexturedCube(ctx *rendering.Context, size float32) *TexturedCube {
	mat := rendering.NewMaterial(shader.COL_TEX_NORM_3D)
	mat.Uniform4f("fragColor", 1, 1, 1, 1.0)

	cube := &TexturedCube{
		mesh:  *rendering.NewMesh(ctx, mat),
		dirty: true,
		size:  size,
		color: color.White,
	}
	cube.Clear()
	return cube
}

func (m *TexturedCube) Destroy() {
	m.mesh.Destroy()
}

func (m *TexturedCube) Size() float32 {
	return m.size
}

func (m *TexturedCube) SetSize(s float32) {
	m.size = s
	m.dirty = true
}

func (m *TexturedCube) Color() color.Color {
	return m.color
}

func (m *TexturedCube) SetColor(c color.Color) {
	m.color = c
	m.dirty = true
}

func (m *TexturedCube) SetTexture(texKey rendering.TextureKey) {
	m.mesh.Material().AddTextureBinding("sampler", texKey)
}

//
//set texture(textureID) {
//this._texture = textureID;
//this.modelAsset.addTextureBinding("uSampler", textureID);
//}
//
//get texture() {
//return this._texture;
//}

func (m *TexturedCube) Update(_ time.Duration) {
	if !m.dirty {
		return
	}
	m.dirty = false

	s := m.size
	vertices := []float32{
		// front:
		/*xyz*/ 0, 0, 0 /*uv*/, 0, 0 /*norm*/, 0, 0, -1,
		/*xyz*/ s, 0, 0 /*uv*/, 1, 0 /*norm*/, 0, 0, -1,
		/*xyz*/ s, s, 0 /*uv*/, 1, 1 /*norm*/, 0, 0, -1,
		/*xyz*/ 0, s, 0 /*uv*/, 0, 1 /*norm*/, 0, 0, -1,
		// right
		/*xyz*/ s, 0, 0 /*uv*/, 0, 0 /*norm*/, 1, 0, 0,
		/*xyz*/ s, 0, s /*uv*/, 1, 0 /*norm*/, 1, 0, 0,
		/*xyz*/ s, s, s /*uv*/, 1, 1 /*norm*/, 1, 0, 0,
		/*xyz*/ s, s, 0 /*uv*/, 0, 1 /*norm*/, 1, 0, 0,
		// left
		/*xyz*/ 0, 0, s /*uv*/, 0, 0 /*norm*/, -1, 0, 0,
		/*xyz*/ 0, 0, 0 /*uv*/, 1, 0 /*norm*/, -1, 0, 0,
		/*xyz*/ 0, s, 0 /*uv*/, 1, 1 /*norm*/, -1, 0, 0,
		/*xyz*/ 0, s, s /*uv*/, 0, 1 /*norm*/, -1, 0, 0,
		// back
		/*xyz*/ s, 0, s /*uv*/, 0, 0 /*norm*/, 0, 0, 1,
		/*xyz*/ 0, 0, s /*uv*/, 1, 0 /*norm*/, 0, 0, 1,
		/*xyz*/ 0, s, s /*uv*/, 1, 1 /*norm*/, 0, 0, 1,
		/*xyz*/ s, s, s /*uv*/, 0, 1 /*norm*/, 0, 0, 1,
		// top
		/*xyz*/ 0, s, 0 /*uv*/, 0, 0 /*norm*/, 0, 1, 0,
		/*xyz*/ s, s, 0 /*uv*/, 1, 0 /*norm*/, 0, 1, 0,
		/*xyz*/ s, s, s /*uv*/, 1, 1 /*norm*/, 0, 1, 0,
		/*xyz*/ 0, s, s /*uv*/, 0, 1 /*norm*/, 0, 1, 0,
		// bottom
		/*xyz*/ 0, 0, 0 /*uv*/, 0, 0 /*norm*/, 0, -1, 0,
		/*xyz*/ s, 0, 0 /*uv*/, 1, 0 /*norm*/, 0, -1, 0,
		/*xyz*/ s, 0, s /*uv*/, 1, 1 /*norm*/, 0, -1, 0,
		/*xyz*/ 0, 0, s /*uv*/, 0, 1 /*norm*/, 0, -1, 0,
	}
	indices := []uint16{
		0, 1, 2, 0, 2, 3, // front
		4, 5, 6, 4, 6, 7, // right
		8, 9, 10, 8, 10, 11, // left
		12, 13, 14, 12, 14, 15, // back
		16, 17, 18, 16, 18, 19, // top
		20, 22, 21, 20, 23, 22, // bottom
	}
	m.mesh.SetVertexData(vertices, indices, gl.TRIANGLES, []string{"position", "texCoord", "normal"}, rendering.InterleavedBuffer)
}

func (m *TexturedCube) Draw(renderTarget rendering.RenderTarget, renderState *rendering.RenderState) {
	renderState.TransformStack.RightMul(m.GetTransform())
	renderTarget.Draw(&m.mesh, renderState)
}
