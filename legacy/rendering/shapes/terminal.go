package shapes

//
//import (
//	"github.com/maja42/logicat/rendering"
//	"github.com/maja42/logicat/rendering/color"
//	"github.com/maja42/logicat/rendering/resources/shader"
//	"golang.org/x/mobile/gl"
//)
//
//type Terminal struct {
//	rendering.AttachableModel
//	rendering.Transform
//
//	font *rendering.Font
//	mesh rendering.Mesh
//
//	//dirty  bool
//	width, height int
//	buffer        buffer
//
//	color color.Color
//}
//
//type buffer [] /*y*/ [] /*x*/ rune // TODO: use 1D-buffer and calculate index instead.
//
//func NewTerminal(ctx *rendering.Context, font *rendering.Font, width, height int) *Terminal {
//	mat := rendering.NewMaterial(shader.TEX_2D)
//	//mat.Uniform4f("fragColor", 1, 1, 1, 1.0)
//
//	// TODO: use 1D-buffer and calculate index instead.
//	buf := make([][]rune, height, height)
//	for y := 0; y < height; y++ {
//		buf[y] = make([]rune, width, width)
//	}
//
//	terminal := &Terminal{
//		font:   font,
//		mesh:   *rendering.NewMesh(ctx, mat),
//		width:  width,
//		height: height,
//		buffer: buf,
//		color:  color.White,
//	}
//	terminal.Clear()
//	terminal.initGeometry()
//	return terminal
//}
//
//func (t *Terminal) Destroy() {
//	t.mesh.Destroy()
//}
//
//func (t *Terminal) Write(x, y int, txt string) {
//	if y >= t.height {
//		return
//	}
//	src := []rune(txt)
//	copy(t.buffer[y][x:], src)
//}
//
////func (t *Terminal) Write(x, y int, txt string) {
////	alloc := 1 + (y - cap(t.buffer))
////	if alloc > 0 { // not enough space
////		t.buffer = t.buffer[:cap(t.buffer)]
////		t.buffer = append(t.buffer, make([][]rune, alloc)...)
////	}
////	if len(t.buffer) <= y {
////		t.buffer = t.buffer[:y+1]
////	}
////
////	src := []rune(txt)
////	lastX := x + len(src)
////	line := t.buffer[y]
////
////	alloc = 1 + (lastX - cap(line))
////	if alloc > 0 { // not enough space
////		line = line[:cap(line)]
////		line = append(line, make([]rune, alloc)...)
////	}
////	if len(line) <= lastX {
////		line = line[:lastX+1]
////	}
////
////	n := copy(line[x:], src)
////	if n != len(src) {
////		panic("ERR") // TODO: put assertions into subpackage and use it from here
////	}
////
////	t.dirty = true
////}
//
//func (t *Terminal) initGeometry() {
//	/*
//		0 - 1 - 2 - 3
//		| / | / | / |
//		4 - 5 - 6 - 7
//	*/
//	vtxCntX := t.width + 1
//	vtxCntY := t.height + 1
//	vertices := make([]float32, vtxCntX*vtxCntY*2) // (x, y) per vertex
//	for y := 0; y <= t.height; y++ {
//		for x := 0; x <= t.width; x++ {
//			vtx := (y*vtxCntX + x) * 2
//			vertices[vtx] = float32(x)
//			vertices[vtx+1] = float32(y)
//		}
//	}
//
//	// TODO: Triangle-Strip instead of discrete triangles!
//
//	// triangles = counter clock wise
//	indices := make([]uint16, 0, t.width*t.height*2*3)
//	for y := 0; y < t.height; y++ {
//		for x := 0; x < t.width; x++ {
//			topLeftVtx := uint16(y*vtxCntX + x)
//			topRightVtx := topLeftVtx + 1
//			bottomLeftVtx := topLeftVtx + uint16(vtxCntX)
//			bottomRightVtx := bottomLeftVtx + 1
//
//			// top-left triangle
//			indices = append(indices, topLeftVtx, bottomLeftVtx, topRightVtx)
//			// bottom-right triangle
//			indices = append(indices, bottomLeftVtx, bottomRightVtx, topRightVtx)
//		}
//	}
//
//	t.mesh.SetVertexData(vertices, indices, gl.TRIANGLES, []string{"position"}, rendering.InterleavedBuffer)
//}
//
//func (t *Terminal) Color() color.Color {
//	return t.color
//}
//
//func (t *Terminal) SetColor(c color.Color) {
//	t.color = c
//	//t.dirty = true
//}
