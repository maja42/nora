package nora

import (
	"github.com/maja42/gl"
	"github.com/maja42/nora/assert"
)

// Geometry represents an object's geometric properties.
// Allows reading/merging/manipulating of the underlying geometry.
type Geometry struct {
	vertexCount      int
	vertices         []float32
	indices          []uint16
	primitiveType    PrimitiveType
	vertexAttributes []string
	bufferLayout     BufferLayout
}

// NewGeometry is equivalent as creating an empty geometry object and calling Set() to fill the initial data.
// It is not required to construct a Geometry object using this constructor.
func NewGeometry(vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) *Geometry {
	g := &Geometry{}
	g.Set(vertexCount, vertices, indices, primitiveType, vertexAttributes, bufferLayout)
	return g
}

// Set replaces the existing geometry with new data.
func (g *Geometry) Set(vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) {
	AssertValidGeometry("", vertexCount, vertices, indices, primitiveType, vertexAttributes)
	g.vertexCount = vertexCount
	g.vertices = vertices
	g.indices = indices
	g.primitiveType = primitiveType
	g.vertexAttributes = vertexAttributes
	g.bufferLayout = bufferLayout
}

// Append merges new geometry at the end of the current one.
func (g *Geometry) Append(vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) {
	if g.vertexCount == 0 {
		g.Set(vertexCount, vertices, indices, primitiveType, vertexAttributes, bufferLayout)
		return
	}
	AssertValidGeometry("", vertexCount, vertices, indices, primitiveType, vertexAttributes)

	assert.True(g.primitiveType == primitiveType, "Incompatible primitive type")
	assert.True(equalStringSlice(g.vertexAttributes, vertexAttributes), "Incompatible vertex attributes: %v <> %v", g.vertexAttributes, vertexAttributes)
	assert.True(g.vertexCount+vertexCount <= 0xFFFF, "Resulting geometry is not indexable by uint16")

	// Support could be added for some of the following cases:
	assert.True(g.bufferLayout == bufferLayout, "Incompatible buffer layouts")                                                          // solvable by splicing the vertex array
	assert.True(bufferLayout == InterleavedBuffer, "Unsupported buffer layout")                                                         // solvable by splicing the vertex array
	assert.True(primitiveType == gl.POINTS || primitiveType == gl.LINES || primitiveType == gl.TRIANGLES, "Unsupported primitive type") // solvable by adding degenerate triangles

	firstIdx := len(g.indices)
	g.vertices = append(g.vertices, vertices...)
	g.indices = append(g.indices, indices...)
	// offset indices:
	for i := firstIdx; i < len(g.indices); i++ {
		g.indices[i] += uint16(g.vertexCount)
	}
	g.vertexCount += vertexCount
}

// AppendGeometry merges new geometry at the end of the current one.
func (g *Geometry) AppendGeometry(other *Geometry) {
	g.Append(other.vertexCount, other.vertices, other.indices, other.primitiveType, other.vertexAttributes, other.bufferLayout)
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// AssertValidGeometry checks if the provided geometry is in-itself valid.
// sProgKey is optional. If provided, the geometry is validated against the currently loaded shader program with that key.
func AssertValidGeometry(sProgKey ShaderProgKey, vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string) {
	// vertex data must be divisible by vertex count
	indexCount := len(indices)
	vertexSize := 0
	if vertexCount > 0 {
		vertexSize = len(vertices) / vertexCount
		assert.True(float32(len(vertices))/float32(vertexCount) == float32(vertexSize), "VertexCount and vertex data does not match")
	} else {
		assert.True(len(vertices) == 0, "VertexCount and vertex data does not match. Should there be vertices?")
	}
	// there must be attributes
	assert.True(len(vertexAttributes) > 0, "There are no vertex attributes")
	// vertex size must be big enough for all the attributes
	assert.True(vertexSize >= len(vertexAttributes), "Vertex size is too small to fit all vertex attributes") // zero-size attributes don't exist

	if indexCount > 0 {
		// vertices must be indexable
		assert.True(vertexCount <= 0xFFFF, "Too many vertices to be indexed by uint16")
		// indices must not reference out-of-bounds vertices
		minIdx, maxIdx := vertexCount, -1
		assert.Func(func() bool {
			okay := true
			for _, i := range indices {
				idx := int(i)
				if idx > maxIdx {
					maxIdx = idx
				}
				if idx < minIdx {
					minIdx = idx
				}
				if idx >= vertexCount {
					okay = false
				}
			}
			return okay
		}, "Index array references out-of-bounds vertices")

		// Detect unreferenced vertices (doesn't find all, but many):
		obsolete := vertexCount - 1 - maxIdx
		assert.True(obsolete <= 0, "Obsolete vertices: the last %d vertices are not referenced", obsolete)
		assert.True(minIdx <= 0, "Obsolete vertices: the first %d vertices are not referenced", minIdx)
		assert.True(indexCount >= vertexCount, "Obsolete vertices: not all vertices are referenced") // holes in the middle
	}
	// index count must be divisible by number-of-indices-per-primitive
	assertMsg := "Index count %d is incompatible with primitive type %q"
	switch primitiveType {
	case gl.POINTS:
	case gl.LINE_STRIP:
		assert.True(indexCount >= 1, assertMsg, indexCount, primitiveType)
	case gl.LINE_LOOP:
	case gl.LINES:
		assert.True(indexCount%2 == 0, assertMsg, indexCount, primitiveType)
	case gl.TRIANGLE_STRIP:
		assert.True(indexCount >= 2, assertMsg, indexCount, primitiveType)
	case gl.TRIANGLE_FAN:
		assert.True(indexCount >= 2, assertMsg, indexCount, primitiveType)
	case gl.TRIANGLES:
		assert.True(indexCount%3 == 0, assertMsg, indexCount, primitiveType)
	default:
		assert.Fail("Unknown primitive type %q", primitiveType)
	}

	// Validate against shader program

	if sProgKey == "" { // shader unknown --> skip
		return
	}

	sProg, _ := nora.Shaders.resolve(sProgKey)
	if !assert.True(sProg != nil, "Shader %q not loaded", sProgKey) {
		return
	}

	// Check existence of attributes
	expectedVertexSize := 0
	for _, attr := range vertexAttributes {
		typ, ok := sProg.attributeTypes[attr]
		expectedVertexSize += int(vaTypePropertyMapping[typ].components)
		assert.True(ok, "Shader %q does not support vertex attribute %s", attr)
	}
	// Check missing attributes
	if len(sProg.attributeTypes) > len(vertexAttributes) {
		assert.Fail("Shader %q has %d vertex attributes. Geometry only contains %d", sProgKey, len(sProg.attributeTypes), len(vertexAttributes))
	}

	// Check size of vertex attributes
	assert.True(vertexSize == expectedVertexSize, "Shader %q has %d elements per vertex. Geometry has %d elements.", sProgKey, expectedVertexSize, vertexSize)
}
