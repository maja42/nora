package nora

import (
	"fmt"

	"github.com/maja42/gl"
	"github.com/maja42/nora/assert"
	"github.com/maja42/vmath"
)

// Mesh is the (only) low-level construct for rendering geometry, besides ReadableMesh.
// A mesh contains geometry and has a material.
//
// Meshes have no notion of transformations.
// The higher level "models" can contain (multiple) meshes for drawing and take care for transformations.
type Mesh struct {
	material *Material

	// Geometry:
	vbo gl.Buffer // vertex buffer object
	ibo gl.Buffer // index buffer object

	primitiveType PrimitiveType
	bufferLayout  BufferLayout
	vboSize       int // in bytes

	vertexAttributes []string

	vertexCount int
	vertexSize  int // in bytes

	indexCount     int
	primitiveCount int
}

// NewMesh creates a new mesh with the given material
func NewMesh(mat *Material) *Mesh {
	// Use sync.Pool for better performance?
	return &Mesh{
		material: mat,
		vbo:      gl.CreateBuffer(),
	}
}

// Destroy deletes all resources associated with the mesh
func (m *Mesh) Destroy() {
	gl.DeleteBuffer(m.vbo)
	gl.DeleteBuffer(m.ibo)
}

// Material returns the underlying material
func (m *Mesh) Material() *Material {
	return m.material
}

// SetMaterial changes the underlying material
func (m *Mesh) SetMaterial(mat *Material) {
	m.material = mat
}

// SetVertexData is equivalent to SetGeometry and defines the mesh's geometry.
//	- vertexCount       Number of vertices
//	- vertices			Array of raw vertex data
//	- indices 			Optional array of indices. If provided, indexed drawing is performed instead of array drawing.
//	- primitiveType 	The type of primitives that is drawn.
//  - vertexAttributes 	The (ordered) set of attributes within the vertices.
//  - bufferLayout      How vertices are laid out within the vertex array.
func (m *Mesh) SetVertexData(vertexCount int, vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) {
	if len(indices) == 0 {
		m.prepareIBO(false)
		m.indexCount = vertexCount
	} else {
		m.prepareIBO(true)
		m.indexCount = len(indices)
	}
	m.vertexCount = vertexCount
	if vertexCount > 0 {
		m.vertexSize = (len(vertices) / vertexCount) * 4
	}
	m.primitiveType = primitiveType
	m.bufferLayout = bufferLayout
	m.vboSize = len(vertices) * 4
	m.vertexAttributes = vertexAttributes
	m.primitiveCount = m.determinePrimitiveCount(m.indexCount, primitiveType)

	AssertValidGeometry(m.material.sProgKey, vertexCount, vertices, indices, primitiveType, vertexAttributes)

	usage := gl.Enum(gl.STATIC_DRAW)

	engine.lockBuffer(gl.ARRAY_BUFFER)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferDataFloat32(gl.ARRAY_BUFFER, vertices, usage)
	engine.unlockBuffer(gl.ARRAY_BUFFER)

	if len(indices) > 0 {
		engine.lockBuffer(gl.ELEMENT_ARRAY_BUFFER)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ibo)
		gl.BufferDataUint16(gl.ELEMENT_ARRAY_BUFFER, indices, usage)
		engine.unlockBuffer(gl.ELEMENT_ARRAY_BUFFER)
	}
}

// SetGeometry is equivalent to SetVertexData and defines the mesh's geometry.
func (m *Mesh) SetGeometry(geom *Geometry) {
	m.SetVertexData(geom.vertexCount, geom.vertices, geom.indices, geom.primitiveType, geom.vertexAttributes, geom.bufferLayout)
}

func (m *Mesh) prepareIBO(required bool) {
	hasIBO := m.ibo.Value != 0
	if required && !hasIBO {
		m.ibo = gl.CreateBuffer()
	}
	if !required && hasIBO {
		gl.DeleteBuffer(m.ibo)
		m.ibo.Value = 0
	}
}

func (m *Mesh) determinePrimitiveCount(indexCount int, primitiveType PrimitiveType) int {
	assertMsg := "Index count %d is incompatible with primitive type %q"

	switch primitiveType {
	case gl.POINTS:
		return indexCount
	case gl.LINE_STRIP:
		assert.True(indexCount >= 1, assertMsg, indexCount, primitiveType)
		return indexCount - 1
	case gl.LINE_LOOP:
		return indexCount
	case gl.LINES:
		assert.True(indexCount%2 == 0, assertMsg, indexCount, primitiveType)
		return indexCount / 2
	case gl.TRIANGLE_STRIP:
		assert.True(indexCount >= 2, assertMsg, indexCount, primitiveType)
		return indexCount - 2
	case gl.TRIANGLE_FAN:
		assert.True(indexCount >= 2, assertMsg, indexCount, primitiveType)
		return indexCount - 2
	case gl.TRIANGLES:
		assert.True(indexCount%3 == 0, assertMsg, indexCount, primitiveType)
		return indexCount / 3
	}
	assert.Fail("Unknown primitive type %q", primitiveType)
	return 0
}

// SetVertexSubData changes parts of the underlying vertex buffer.
// 	- vertexOffset		Vertex offset.
//  - vertices			Underlying vertex data that will overwrite existing buffers
// Cannot change the underlying vertex buffer size.
func (m *Mesh) SetVertexSubData(vertexOffset int, vertices []float32) {
	assert.True(vertexOffset >= 0 && vertexOffset < m.vertexCount, "Invalid vertex offset (out of range)")
	assert.True(m.vertexSize > 0 && ((len(vertices)*4)%(m.vertexSize) == 0), "Invalid vertex data size")

	engine.glSync.lockBuffer(gl.ARRAY_BUFFER)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	gl.BufferSubDataFloat32(gl.ARRAY_BUFFER, vertexOffset*m.vertexSize, vertices)
	engine.glSync.unlockBuffer(gl.ARRAY_BUFFER)
}

// SetIndexSubData changes parts of the underlying index buffer.
// 	- indexOffset		Index offset.
//  - indices			Underlying index data that will overwrite existing buffers
// Cannot change the underlying index buffer size.
func (m *Mesh) SetIndexSubData(indexOffset int, indices []uint16) {
	assert.True(m.ibo.Value != 0, "The mesh does not use indexed drawing")
	assert.True(indexOffset >= 0 && indexOffset < m.indexCount, "Invalid index offset (out of range)")

	// TODO: write assertion that checks that indices don't reference out-of-bounds vertices

	engine.glSync.lockBuffer(gl.ELEMENT_ARRAY_BUFFER)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ibo)
	gl.BufferSubDataUint16(gl.ELEMENT_ARRAY_BUFFER, indexOffset, indices)
	engine.glSync.unlockBuffer(gl.ELEMENT_ARRAY_BUFFER)
}

// ClearVertexData clears the underlying buffers.
// Rendering this mesh will not draw anything.
func (m *Mesh) ClearVertexData() {
	m.SetVertexData(0, []float32{}, nil, gl.TRIANGLES, []string{}, InterleavedBuffer)
}

// TransDraw temporarily applies a model transformation to the matrix stack for rendering the mesh.
// Utility method.
func (m *Mesh) TransDraw(renderState *RenderState, modelTransform vmath.Mat4f) {
	renderState.TransformStack.Push()
	renderState.TransformStack.MulRight(modelTransform)
	m.Draw(renderState)
	renderState.TransformStack.Pop()
}

// Draw renders the mesh.
// The required material (shader, textures, uniforms) are applied and the buffers are bound for rendering.
func (m *Mesh) Draw(renderState *RenderState) {
	if m.indexCount == 0 {
		return
	}

	sProg := renderState.applyMaterial(m.material)
	if sProg == nil { // shader is not loaded
		return
	}

	sProg.configureVertexAttributes(m.vertexAttributes, true)
	defer sProg.configureVertexAttributes(m.vertexAttributes, false)

	engine.glSync.lockBuffer(gl.ARRAY_BUFFER)

	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	if m.bufferLayout == InterleavedBuffer {
		m.configureInterleavedVertexAttributes(sProg)
	} else {
		m.configureCompactVertexAttributes(sProg)
	}

	if m.ibo.Value != 0 {
		engine.glSync.lockBuffer(gl.ELEMENT_ARRAY_BUFFER)
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ibo)
		gl.DrawElements(gl.Enum(m.primitiveType), m.indexCount, gl.UNSIGNED_SHORT, 0)
		engine.glSync.unlockBuffer(gl.ELEMENT_ARRAY_BUFFER)
	} else {
		gl.DrawArrays(gl.Enum(m.primitiveType), 0, m.indexCount)
	}
	engine.glSync.unlockBuffer(gl.ARRAY_BUFFER)

	renderState.totalDrawCalls++
	renderState.totalPrimitives += m.primitiveCount
}

func (m *Mesh) configureInterleavedVertexAttributes(sProg *shaderProgram) {
	offset := 0
	stride := m.vertexSize
	for _, vaName := range m.vertexAttributes {
		loc, typ := sProg.getAttribLocation(vaName)
		if !assert.True(typ != 0, "Unsupported vertex attribute %q", vaName) {
			continue
		}

		typProps := vaTypePropertyMapping[typ]
		gl.VertexAttribPointer(loc, int(typProps.components), typProps.compType, false, stride, offset)
		offset += int(typProps.components) * 4
	}
}

func (m *Mesh) configureCompactVertexAttributes(sProg *shaderProgram) {
	vertexCount := m.vertexCount
	component := 0
	for _, vaName := range m.vertexAttributes {
		loc, typ := sProg.getAttribLocation(vaName)
		if !assert.True(typ != 0, "Unsupported vertex attribute %q", vaName) {
			continue
		}

		typProps := vaTypePropertyMapping[typ]
		stride := int(typProps.components) * 4
		gl.VertexAttribPointer(loc, int(typProps.components), typProps.compType, false, stride, vertexCount*component*4)
		component += int(typProps.components)
	}
}

// Info returns the number of primitives, vertices and indices of the mesh.
func (m *Mesh) Info() (int, int, int) {
	return m.primitiveCount,
		m.vertexCount,
		m.indexCount
}

func (m *Mesh) String() string {
	return fmt.Sprintf(""+
		"Primitives  %d (%s, %s)\n"+
		"Vertices    %d\n"+
		"Indices     %d\n"+
		"Vertex size %d (%d attributes, %d bytes total)\n"+
		"VBO size    %d bytes",
		m.primitiveCount, m.primitiveType.String(), m.bufferLayout.String(),
		m.vertexCount,
		m.indexCount,
		m.vertexSize/4, len(m.vertexAttributes), m.vertexSize,
		m.vboSize)
}
