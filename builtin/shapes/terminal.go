package shapes

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/assert"
	"github.com/maja42/nora/builtin/shader"
	"github.com/maja42/nora/color"
	"github.com/maja42/nora/math"
	"github.com/sirupsen/logrus"
)

type Terminal struct {
	nora.AttachableModel
	nora.Transform
	mesh nora.Mesh

	font        *nora.Font
	charSize    math.Vec2i
	lineSpacing int

	size math.Vec2i
	text []rune
}

func NewTerminal(font *nora.Font, size math.Vec2i, lineSpacing float32) *Terminal {
	if !font.Monospace {
		logrus.Warn("No monospace font (%v)", font)
	}

	mat := nora.NewMaterial(shader.COL_TEX_2D)
	mat.Uniform4fColor("color", color.White)
	mat.AddTextureBinding("sampler", font.TextureKey())

	t := &Terminal{
		mesh:        *nora.NewMesh(mat),
		font:        font,
		charSize:    math.Vec2i{int(font.AvgWidth()), font.Height},
		lineSpacing: int(float32(font.Height) * lineSpacing),
		size:        size,
	}
	t.ClearTransform()

	// font characters have unique offsets and dimensions, they are not "blocked".
	// therefore, every character needs vertices with unique offsets and texture coordinates.
	// this can be improved by having font textures with each character consuming the same amount of space.
	// Alternatively, text data could also be stored in uniform buffers (no WebGL support) or textures.
	//
	// Each character: counter-clockwise
	//	 3 - 2
	//	 | / |
	//	 0 - 1

	characters := size[0] * size[1]
	vertexCnt := characters * 4
	primitives := characters * 2

	vertices := make([]float32, vertexCnt*4) // x,y,u,v
	indices := make([]uint16, primitives*3)
	assert.True(len(indices) <= 0xFFFF, "terminal too big for 16bit indices")

	idx := 0
	for y := 0; y < size[1]; y++ {
		for x := 0; x < size[0]; x++ {
			vIdx := t.vtxIndex(math.Vec2i{x, y})
			copy(indices[idx:], []uint16{
				vIdx, vIdx + 1, vIdx + 2,
				vIdx + 2, vIdx + 3, vIdx,
			})
			idx += 6
		}
	}

	t.text = make([]rune, characters)

	t.mesh.SetVertexData(vertexCnt, vertices, indices, gl.TRIANGLES, []string{"position", "texCoord"}, nora.InterleavedBuffer)
	return t
}

// Returns the size of an individual character in model-space
func (t *Terminal) CharSize() math.Vec2i {
	return t.charSize
}

// Returns the size of the whole terminal in model-space
func (t *Terminal) Size() math.Vec2i {
	return math.Vec2i{t.charSize[0] * t.size[0], t.lineSpacing * t.size[1]}
}

// vtxIndex returns the vertex index for the given character position. Ignores vertex components/size
func (t *Terminal) vtxIndex(pos math.Vec2i) uint16 {
	verticesPerChar := 4
	posIdx := pos[0] + pos[1]*t.size[0]
	assert.True(posIdx < t.size[0]*t.size[1], "posIdx out-of-range: %d <> %v in %v", posIdx, pos, t.size)

	return uint16(posIdx * verticesPerChar)
}

// CharPos returns the position of a character in model-space
func (t *Terminal) CharPos(pos math.Vec2i) math.Vec2i {
	// Character positions are measured from their bottom-left corner.
	// The terminal (and the runes placed within) have their origin in the top-left.
	return math.Vec2i{pos[0] * t.charSize[0], -(pos[1] + 1) * t.lineSpacing}
}

func (t *Terminal) Destroy() {
	t.mesh.Destroy()
}

func (t *Terminal) Draw(renderState *nora.RenderState) {
	renderState.TransformStack.RightMul(t.GetTransform())
	t.mesh.Draw(renderState)
}

func (t *Terminal) SetRune(pos math.Vec2i, r rune) {
	runeIdx := pos[0] + pos[1]*t.size[0]
	t.text[runeIdx] = r

	f := t.font
	c, ok := f.Char(r)
	if !assert.True(ok, "Font %s does not contain symbol for rune %s (%v)", f, string(r), r) {
		r = ' '
	}

	cpos := t.CharPos(pos)
	xl := float32(cpos[0] + c.Offset[0])
	xr := xl + float32(c.Size[0])

	yt := float32(cpos[1] + c.Offset[1])
	yb := yt - float32(c.Size[1])
	tl, br := f.TexCoord(r)

	vtxData := []float32{
		/*xy*/ xl, yb /*uv*/, tl[0], br[1],
		/*xy*/ xr, yb /*uv*/, br[0], br[1],
		/*xy*/ xr, yt /*uv*/, br[0], tl[1],
		/*xy*/ xl, yt /*uv*/, tl[0], tl[1],
	}
	t.mesh.SetVertexSubData(int(t.vtxIndex(pos)), vtxData)
}
