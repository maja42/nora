package geo2d

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/color"
	"github.com/maja42/vmath"
)

// Border returns the geometry for a 2D rectangular border.
// The border is placed exactly above the given rectangle, the borderWidth protruding to both sides.
func Border(r vmath.Rectf, borderWidth float32, color color.Color) *nora.Geometry {
	/*         x0 x1           x2  x3
		   y3  3__________________ 2
	       y2   |\ _____________/ |
		        | |7           6| |
		        | |             | |
		   y1   | |4___________5| |
		   y0  0|/_______________\|1
	*/

	r = r.Normalize()

	thickHalf := borderWidth * 0.5
	x0 := r.Min[0] - thickHalf
	x1 := r.Min[0] + thickHalf
	x2 := r.Max[0] - thickHalf
	x3 := r.Max[0] + thickHalf

	y0 := r.Min[1] - thickHalf
	y1 := r.Min[1] + thickHalf
	y2 := r.Max[1] - thickHalf
	y3 := r.Max[1] + thickHalf

	vertices := []float32{
		x0, y0, color.R, color.G, color.B, // 0
		x3, y0, color.R, color.G, color.B, // 1
		x3, y3, color.R, color.G, color.B, // 2
		x0, y3, color.R, color.G, color.B, // 3

		x1, y1, color.R, color.G, color.B, // 4
		x2, y1, color.R, color.G, color.B, // 5
		x2, y2, color.R, color.G, color.B, // 6
		x1, y2, color.R, color.G, color.B, // 7
	}
	indices := []uint16{
		0, 1, 4, 1, 5, 4, // bottom
		1, 2, 5, 2, 6, 5, // right
		2, 3, 6, 3, 7, 6, // top
		3, 0, 7, 0, 4, 7, // left
	}
	return nora.NewGeometry(8, vertices, indices, gl.TRIANGLES, []string{"position", "color"}, nora.InterleavedBuffer)
}
