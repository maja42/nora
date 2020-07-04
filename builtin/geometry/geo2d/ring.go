package geo2d

import (
	"math"

	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/assert"
	"github.com/maja42/vmath"
	"github.com/maja42/vmath/math32"
)

// Ring creates a ring with the given number of edges.
func Ring(position vmath.Vec2f, outerRadius, innerRadius float32, segments int) *nora.Geometry {
	return EllipticalRingSegment(position, vmath.Vec2f{outerRadius, outerRadius}, vmath.Vec2f{innerRadius, innerRadius}, segments, 0, math.Pi*2)
}

// EllipticalRing creates an elliptical ring with the given number of edges.
func EllipticalRing(position vmath.Vec2f, outerRadius, innerRadius vmath.Vec2f, segments int) *nora.Geometry {
	return EllipticalRingSegment(position, outerRadius, innerRadius, segments, 0, math.Pi*2)
}

// RingSegment creates a fraction of a ring with the given number of edges.
func RingSegment(position vmath.Vec2f, outerRadius, innerRadius float32, segments int, fromRad, totalRad float32) *nora.Geometry {
	return EllipticalRingSegment(position, vmath.Vec2f{outerRadius, outerRadius}, vmath.Vec2f{innerRadius, innerRadius}, segments, fromRad, totalRad)
}

// EllipticalRingSegment creates a fraction of an elliptical ring with the given number of edges.
func EllipticalRingSegment(position vmath.Vec2f, outerRadius, innerRadius vmath.Vec2f, segments int, fromRad, totalRad float32) *nora.Geometry {
	assert.True(segments >= 2, "Rings need at least 2 segments")
	assert.True(totalRad > 0, "Rings should cover more than 0°")
	assert.True(totalRad <= 2*math.Pi, "Rings should not cover more than 360°")
	assert.True(outerRadius[0] > innerRadius[0], "Outer X-radius should be bigger than inner radius")
	assert.True(outerRadius[1] > innerRadius[1], "Outer Y-radius should be bigger than inner radius")

	stride := totalRad / float32(segments) // angle between to segment points

	points := segments + 1 // In case of a full circle, the last two points are repeated
	vtxCount := points * 2
	vertices := make([]float32, vtxCount*2)

	for i := 0; i < points; i++ {
		angle := fromRad + float32(i)*stride
		// outer:
		copy(vertices[i*4:], []float32{
			position[0] + math32.Sin(angle)*outerRadius[0],
			position[1] + math32.Cos(angle)*outerRadius[1],
		})
		// inner:
		copy(vertices[i*4+2:], []float32{
			position[0] + math32.Sin(angle)*innerRadius[0],
			position[1] + math32.Cos(angle)*innerRadius[1],
		})
	}
	return nora.NewGeometry(vtxCount, vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.InterleavedBuffer)
}
