package shapes

import (
	"fmt"

	"golang.org/x/mobile/gl"

	"github.com/maja42/logicat/rendering"
	"github.com/maja42/logicat/rendering/color"
	"github.com/maja42/logicat/rendering/resources/shader"
)

type Text struct {
	rendering.AttachableModel
	rendering.Transform

	font *rendering.Font
	mesh rendering.Mesh

	text          []rune
	width, height int // calculated

	color color.Color
}

func NewText(ctx *rendering.Context, font *rendering.Font, text string) *Text {
	mat := rendering.NewMaterial(shader.COL_TEX_2D)

	txt := &Text{
		font:  font,
		mesh:  *rendering.NewMesh(ctx, mat),
		text:  []rune(text),
		color: color.White,
	}
	txt.Clear()
	txt.update()

	txt.mesh.Material().AddTextureBinding("sampler", font.TextureKey())
	txt.SetColor(color.White)
	return txt
}

func (m *Text) Destroy() {
	m.mesh.Destroy()
}

func (m *Text) update() {
	f := m.font

	vertices := make([]float32, len(m.text)*4*4) // each rune requires 4 vertices; (x, y, u, v) per vertex
	indices := make([]uint16, len(m.text)*6)     // each rune requires 2 triangles

	var origin float32   // X
	var baseline float32 // Y

	vtx := uint16(0)
	idx := 0

	for _, r := range m.text {
		if r == '\n' {
			origin = 0
			baseline -= float32(f.Height)
		}
		c, ok := f.Char(r)
		if !ok {
			//error(ok, "Font %s does not contain symbol for rune %v", f, r) // TODO: use assert package --> move into dedicated sub-package!
			continue
		}

		/* counter-clockwise
		   3 - 2
		   | / |
		   0 - 1
		*/

		xl := origin + c.Offset[0]
		xr := xl + c.Size[0]

		yt := baseline + c.Offset[1]
		yb := yt - c.Size[1]
		tl, br := f.TexCoord(r)

		copy(vertices[vtx*4:], []float32{
			/*xy*/ xl, yb /*uv*/, tl[0], br[1],
			/*xy*/ xr, yb /*uv*/, br[0], br[1],
			/*xy*/ xr, yt /*uv*/, br[0], tl[1],
			/*xy*/ xl, yt /*uv*/, tl[0], tl[1],
		})
		copy(indices[idx:], []uint16{
			vtx, vtx + 1, vtx + 2,
			vtx + 2, vtx + 3, vtx,
		})

		origin += float32(c.Width)
		vtx += 4
		idx += 6
	}
	// remove unprintable characters (missing runes, new lines):
	vertices = vertices[:vtx*4]
	indices = indices[:idx]

	//indices = []uint16{
	//	0*4 + 0, 0*4 + 1, 0*4 + 2, 0*4 + 2, 0*4 + 3, 0*4 + 0,
	//	1*4 + 0, 1*4 + 1, 1*4 + 2, 1*4 + 2, 1*4 + 3, 1*4 + 0,
	//	2*4 + 0, 2*4 + 1, 2*4 + 2, 2*4 + 2, 2*4 + 3, 2*4 + 0,
	//	3*4 + 0, 3*4 + 1, 3*4 + 2, 3*4 + 2, 3*4 + 3, 3*4 + 0,
	//	4*4 + 0, 4*4 + 1, 4*4 + 2, 4*4 + 2, 4*4 + 3, 4*4 + 0,
	//	5*4 + 0, 5*4 + 1, 5*4 + 2, 5*4 + 2, 5*4 + 3, 5*4 + 0,
	//	//6*4 + 0, 6*4 + 1, 6*4 + 2, 6*4 + 2, 6*4 + 3, 6*4 + 0,
	//}
	m.mesh.SetVertexData(vertices, indices, gl.TRIANGLES, []string{"position", "texCoord"}, rendering.InterleavedBuffer)
}

func (m *Text) Color() color.Color {
	return m.color
}

func (m *Text) SetColor(c color.Color) {
	m.color = c
	m.mesh.Material().Uniform4f("color", c.R, c.G, c.B, c.A)
}

func (m *Text) Draw(renderTarget rendering.RenderTarget, renderState *rendering.RenderState) {
	renderState.TransformStack.RightMul(m.GetTransform())
	renderTarget.Draw(&m.mesh, renderState)
}

func (m *Text) String() string {
	return fmt.Sprintf("Text(%q/%s)", string(m.text), m.font)
}
