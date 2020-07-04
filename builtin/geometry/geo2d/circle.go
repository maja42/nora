package geo2d

import (
	"math"

	"github.com/maja42/gl"
	"github.com/maja42/nora"
	"github.com/maja42/nora/assert"
	"github.com/maja42/vmath"
	"github.com/maja42/vmath/math32"
)

// Circle creates a circle with the given number of edges.
func Circle(position vmath.Vec2f, radius float32, segments int) *nora.Geometry {
	return CircleSegment(position, radius, segments, 0, math.Pi*2)
}

// Ellipsis creates an ellipsis with the given number of edges.
func Ellipsis(position vmath.Vec2f, radius vmath.Vec2f, segments int) *nora.Geometry {
	return EllipsisSegment(position, radius, segments, 0, math.Pi*2)
}

// CircleSector creates a circle segment with the given number of edges on its rounded side.
func CircleSector(position vmath.Vec2f, radius float32, segments int, fromRad, totalRad float32) *nora.Geometry {
	return EllipsisSector(position, vmath.Vec2f{radius, radius}, segments, fromRad, totalRad)
}

// EllipsisSector creates an ellipsis segment with the given number of edges on its rounded side.
func EllipsisSector(position vmath.Vec2f, radius vmath.Vec2f, segments int, fromRad, totalRad float32) *nora.Geometry {
	assert.True(segments >= 2, "Ellipses/Circles need at least 2 segments")
	assert.True(totalRad > 0, "Ellipses/Circles should cover more than 0°")
	assert.True(totalRad <= 2*math.Pi, "Ellipses/Circles should not cover more than 360°")

	stride := totalRad / float32(segments) // angle between to segment points

	points := segments + 2 // +1 for center
	vertices := make([]float32, points*2)

	// copy center position
	copy(vertices[0:], []float32{
		position[0], position[1],
	})

	// To generate a triangle-strip, we need to alternately create vectors from the front and back
	var angle float32
	frontAngle, backAngle := 0, points-2

	for i := 1; i < points; i++ {
		if i%2 == 0 {
			angle = fromRad + float32(frontAngle)*stride
			frontAngle++
		} else {
			angle = fromRad + float32(backAngle)*stride
			backAngle--
		}
		copy(vertices[i*2:], []float32{
			position[0] + math32.Sin(angle)*radius[0],
			position[1] + math32.Cos(angle)*radius[1],
		})
	}
	return nora.NewGeometry(points, vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.InterleavedBuffer)
}

// CircleSegment creates a circle segment with the given number of edges on its rounded side.
func CircleSegment(position vmath.Vec2f, radius float32, segments int, fromRad, totalRad float32) *nora.Geometry {
	return EllipsisSegment(position, vmath.Vec2f{radius, radius}, segments, fromRad, totalRad)
}

// EllipsisSegment creates an ellipsis segment with the given number of edges on its rounded side.
func EllipsisSegment(position vmath.Vec2f, radius vmath.Vec2f, segments int, fromRad, totalRad float32) *nora.Geometry {
	assert.True(segments >= 2, "Ellipses/Circles need at least 2 segments")
	assert.True(totalRad > 0, "Ellipses/Circles should cover more than 0°")
	assert.True(totalRad <= 2*math.Pi, "Ellipses/Circles should not cover more than 360°")

	stride := totalRad / float32(segments) // angle between to segment points

	points := segments + 1
	vertices := make([]float32, points*2)

	// To generate a triangle-strip, we need to alternately create vectors from the front and back
	var angle float32
	frontAngle, backAngle := 0, points-1

	for i := 0; i < points; i++ {
		targetIdx := i
		if i%2 == 0 {
			angle = fromRad + float32(frontAngle)*stride
			frontAngle++
		} else {
			angle = fromRad + float32(backAngle)*stride
			backAngle--
		}
		copy(vertices[targetIdx*2:], []float32{
			position[0] + math32.Sin(angle)*radius[0],
			position[1] + math32.Cos(angle)*radius[1],
		})
	}
	return nora.NewGeometry(points, vertices, nil, gl.TRIANGLE_STRIP, []string{"position"}, nora.InterleavedBuffer)
}
