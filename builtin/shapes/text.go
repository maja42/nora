package shapes

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/maja42/nora/assert"

	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
)

type Text struct {
	nora.AttachableModel
	nora.Transform

	font       *nora.Font
	tabWidth   int // in characters
	tabWidthPt float32
	mesh       nora.Mesh

	text   []rune
	bounds mgl32.Vec2 // calculated

	color color.Color
}

func NewText(font *nora.Font, text string) *Text {
	mat := nora.NewMaterial(shader.COL_TEX_2D)
	mat.AddTextureBinding("sampler", font.TextureKey())

	txt := &Text{
		font:       font,
		tabWidth:   4,
		tabWidthPt: 4 * font.AvgWidth(),
		mesh:       *nora.NewMesh(mat),
		text:       []rune(text),
		color:      color.White,
	}
	txt.Clear()
	txt.update()

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
		if r == '\r' {
			continue
		}
		if r == '\n' {
			origin = 0
			baseline -= float32(f.Height)
			continue
		}
		if r == '\t' {
			origin += m.tabWidthPt
			continue
		}

		c, ok := f.Char(r)
		if !assert.True(ok, "Font %s does not contain symbol for rune %s (%v)", f, string(r), r) {
			continue
		}

		/* counter-clockwise
		   3 - 2
		   | / |
		   0 - 1
		*/

		xl := origin + float32(c.Offset[0])
		xr := xl + float32(c.Size[0])

		yt := baseline + float32(c.Offset[1])
		yb := yt - float32(c.Size[1])
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
	// remove unprintable characters (missing runes, new lines, ...):
	vertices = vertices[:vtx*4]
	indices = indices[:idx]

	vertexCount := int(vtx)

	m.mesh.SetVertexData(vertexCount, vertices, indices, gl.TRIANGLES, []string{"position", "texCoord"}, nora.InterleavedBuffer)
	m.bounds = mgl32.Vec2{origin, -baseline + float32(f.Height)}
}

func (m *Text) Set(text string) {
	m.text = []rune(text)
	m.update()
}

func (m *Text) SetTabWidth(tabWidth int) {
	m.tabWidth = tabWidth
	m.tabWidthPt = float32(tabWidth) * m.font.AvgWidth()
	m.update()
}

func (m *Text) TabWidth() int {
	return m.tabWidth
}

func (m *Text) Color() color.Color {
	return m.color
}

func (m *Text) SetColor(c color.Color) {
	m.color = c
	m.mesh.Material().Uniform4fColor("color", c)
}

func (m *Text) Draw(renderState *nora.RenderState) {
	renderState.TransformStack.RightMul(m.GetTransform())
	m.mesh.Draw(renderState)
}

func (m *Text) String() string {
	return fmt.Sprintf("Text(%q/%s)", string(m.text), m.font)
}
