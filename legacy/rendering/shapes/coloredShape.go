package shapes

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/maja42/logicat/rendering"
	"github.com/maja42/logicat/rendering/resources/shader"
	"golang.org/x/mobile/gl"
)

type StaticShape struct {
	rendering.AttachableModel
	rendering.Transform
	mesh rendering.Mesh

	sum float64
}

func NewStaticShape(ctx *rendering.Context) *StaticShape {
	mat := rendering.NewMaterial(shader.COL_2D)
	mat.Uniform4f("fragColor", 1, 0.5, 0.5, 1.0)

	s := &StaticShape{
		mesh: *rendering.NewMesh(ctx, mat),
	}
	s.Clear()

	vertices := []float32{
		/*xyz*/ 0, 0,
		/*xyz*/ 1, 0,
		/*xyz*/ 1, 1,
		/*xyz*/ 0, 1,
	}
	indices := []uint16{
		0, 1, 2,
		2, 3, 0,
	}
	s.mesh.SetVertexData(vertices, indices, gl.TRIANGLES, []string{"position"}, rendering.InterleavedBuffer)
	return s
}

func (m *StaticShape) Destroy() {
	m.mesh.Destroy()
}

func (m *StaticShape) Update(d time.Duration) {
	m.RotateZ(mgl32.DegToRad(1))
	m.RotateX(mgl32.DegToRad(1))
	m.RotateY(mgl32.DegToRad(1))

	m.sum += d.Seconds()

	m.SetPositionXY(1.5+float32(math.Sin(m.sum)/2), 0.5)
}

func (m *StaticShape) Draw(renderTarget rendering.RenderTarget, renderState *rendering.RenderState) {
	renderState.TransformStack.RightMul(m.GetTransform())
	renderTarget.Draw(&m.mesh, renderState)
}
