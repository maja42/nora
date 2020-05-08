package geo2d

import (
	"math"

	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/assert"
	"github.com/maja42/vmath"
)

type LineJoint int

const (
	MitterJoint = LineJoint(iota)
	BevelJoint
)

type LineCap struct {
	v int
}

var (
	FlatLineCap     = LineCap{0}
	SquareLineCap   = LineCap{-1} // Extends the line by thickness/2
	TriangleLineCap = LineCap{2}  // Same as a RoundLineCap with 2 segments
)

func RoundLineCap(segments int) LineCap {
	assert.True(segments >= 2, "rounded line caps need at least 2 segments")
	return LineCap{v: segments}
}

// LineStrip creates a line following the given points.
func LineStrip(points []vmath.Vec2f, loop bool, thickness float32, lineJoint LineJoint, startCap, endCap LineCap) *nora.Geometry {
	pointCnt := len(points)
	if pointCnt <= 2 {
		loop = false
	}
	if pointCnt < 2 || loop {
		startCap, endCap = FlatLineCap, FlatLineCap
	}

	var geo *nora.Geometry
	switch lineJoint {
	case BevelJoint:
		geo = bevelJointLine(points, loop, thickness, startCap == SquareLineCap, endCap == SquareLineCap)
	case MitterJoint:
		geo = miterJointLine(points, loop, thickness)
	default:
		assert.Fail("unknown line joint")
		return nil
	}

	if startCap.v > 0 {
		frontAngle := points[0].Sub(points[1]).FlatAngle()
		geo = CircleSegment(points[0], thickness/2, startCap.v, -frontAngle, math.Pi).
			AppendGeometry(geo)
	}
	if endCap.v > 0 {
		backAngle := points[pointCnt-1].Sub(points[pointCnt-2]).FlatAngle()
		geo.AppendGeometry(CircleSegment(points[len(points)-1], thickness/2, endCap.v, -backAngle, math.Pi))
	}
	return geo
}

func bevelJointLine(points []vmath.Vec2f, loop bool, thickness float32, squareCapStart, squareCapEnd bool) *nora.Geometry {
	pointCnt := len(points)
	if pointCnt < 2 {
		return &nora.Geometry{}
	}

	lineSegments := pointCnt - 1
	if loop {
		lineSegments++
	}

	vertexCount := lineSegments * 4
	if loop {
		vertexCount += 2
	}

	vertices := make([]float32, vertexCount*2)

	for l := 0; l < lineSegments; l++ {
		start := points[l]
		end := points[(l+1)%pointCnt]

		line := vmath.Vec2f{end[0] - start[0], end[1] - start[1]}

		if l == 0 && squareCapStart {
			cap := line.Normalize().MulScalar(thickness / 2)
			start = start.Sub(cap)
		}
		if l == lineSegments-1 && squareCapEnd {
			cap := line.Normalize().MulScalar(thickness / 2)
			end = end.Add(cap)
		}

		norm := vmath.Vec2f{-line[1], line[0]}
		norm = norm.Normalize()
		norm = norm.MulScalar(thickness / 2)

		copy(vertices[l*4*2:], []float32{
			start[0] + norm[0], start[1] + norm[1],
			start[0] - norm[0], start[1] - norm[1],
			end[0] + norm[0], end[1] + norm[1],
			end[0] - norm[0], end[1] - norm[1],
		})
	}

	if loop {
		copy(vertices[lineSegments*4*2:], []float32{
			vertices[0], vertices[1],
			vertices[2], vertices[3],
		})
	}
	return nora.NewGeometry(vertexCount, vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.CompactBuffer)
}

func miterJointLine(points []vmath.Vec2f, loop bool, thickness float32) *nora.Geometry {
	pointCnt := len(points)
	if pointCnt < 2 {
		return &nora.Geometry{}
	}

	if pointCnt == 2 {
		loop = false
	}

	vertexCount := pointCnt * 2
	if loop {
		vertexCount += 2
	}
	vertices := make([]float32, vertexCount*2)

	var p0, p1, p2 vmath.Vec2f
	var v1, v2, v3 vmath.Vec2f
	var factor float32

	for i := 0; i < pointCnt; i++ {
		p1 = points[i]

		if i == 0 && !loop { // start piece; compute normal to first element
			v3 = computeNormal(p1, points[1])
			factor = 1
		} else if i == pointCnt-1 && !loop { // end piece; compute normal to last segment
			v3 = computeNormal(points[i-1], p1)
			factor = 1
		} else {
			// Get the two segments shared by the current point
			if i == 0 {
				p0 = points[pointCnt-1]
			} else {
				p0 = points[i-1]
			}

			if i == pointCnt-1 {
				p2 = points[0]
			} else {
				p2 = points[i+1]
			}

			// Compute their normal
			v1 = computeNormal(p0, p1)
			v2 = computeNormal(p1, p2)

			// Combine the normals
			factor = 1 + (v1[0]*v2[0] + v1[1]*v2[1])
			v3 = v1.Add(v2)
		}

		v3 = v3.MulScalar(thickness / 2 / factor) // adjust thickness

		// calculate the final vertices
		v1 = p1.Add(v3)
		v2 = p1.Sub(v3)

		copy(vertices[i*4:], []float32{
			v1[0], v1[1],
			v2[0], v2[1],
		})
	}

	if loop {
		copy(vertices[pointCnt*4:], []float32{
			vertices[0], vertices[1],
			vertices[2], vertices[3],
		})
	}
	return nora.NewGeometry(vertexCount, vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.CompactBuffer)
}

func computeNormal(from, to vmath.Vec2f) vmath.Vec2f {
	// short of to.Sub(from).NormalVec(true).Normalize()
	norm := vmath.Vec2f{from[1] - to[1], to[0] - from[0]}
	return norm.Normalize()
}
