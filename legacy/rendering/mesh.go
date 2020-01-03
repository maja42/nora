package rendering

import (
	"fmt"

	"golang.org/x/mobile/gl"
)

// Mesh is the (only) low-level construct for rendering geometry.
// A mesh contains geometry and has a material.
//
// Meshes have no notion of transformations.
// The higher level "models" can contain (multiple) meshes for drawing and take care for transformations.
type Mesh struct {
	ctx      *Context
	material *Material

	// Geometry:
	vbo gl.Buffer // vertex buffer object
	ibo gl.Buffer // index buffer object

	primitiveType PrimitiveType
	bufferLayout  BufferLayout
	vboSize       int // in bytes

	vertexAttributes []string

	vtxShaderProgID sProgID // shader program for which vertex information was calculated
	vertexSize      int     // in bytes
	vertexCount     int

	indexCount     int
	primitiveCount int
}

// NewMesh creates a new mesh with the given material
func NewMesh(ctx *Context, mat *Material) *Mesh {
	// TODO: use sync.Pool?
	return &Mesh{
		ctx:      ctx,
		material: mat,
		vbo:      ctx.CreateBuffer(),
	}
}

// Destroy deletes all resources associated with the mesh
func (m *Mesh) Destroy() {
	m.ctx.DeleteBuffer(m.vbo)
	m.ctx.DeleteBuffer(m.ibo)
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

// Material returns the underlying material
func (m *Mesh) Material() *Material {
	return m.material
}

// SetMaterial changes the underlying material
func (m *Mesh) SetMaterial(mat *Material) {
	m.material = mat
}

// SetVertexData defines the mesh's geometry.
//	- vertices			Array of raw vertex data
//	- indices 			Optional array of indices. If provided, indexed drawing is performed instead of array drawing.
//	- primitiveType 	The type of primitives that is drawn.
//  - vertexAttributes 	The (ordered) set of attributes within the vertices.
//  - bufferLayout      How vertices are laid out within the vertex array.
func (m *Mesh) SetVertexData(vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) {
	m.vtxShaderProgID = sProgID{} // outdated
	if indices == nil {
		m.prepareIBO(false)
		m.indexCount = len(vertices)
	} else {
		m.prepareIBO(true)
		m.indexCount = len(indices)
	}

	m.primitiveType = primitiveType
	m.bufferLayout = bufferLayout
	m.vboSize = len(vertices) * 4
	m.vertexAttributes = vertexAttributes

	m.calcVertexData()
	m.primitiveCount = m.determinePrimitiveCount(m.indexCount, primitiveType)

	ctx := m.ctx
	usage := gl.Enum(gl.STATIC_DRAW)

	ctx.Lock()
	defer ctx.Unlock()

	ctx.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	ctx.BufferDataFloat32(gl.ARRAY_BUFFER, vertices, usage)

	if indices != nil {
		ctx.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ibo)
		ctx.BufferDataUint16(gl.ELEMENT_ARRAY_BUFFER, indices, usage)
	}
}

func (m *Mesh) getShaderProgram() (*ShaderProgram, sProgID) {
	key := m.material.sProgKey
	return m.ctx.scene.getShaderProgram(key)
}

func (m *Mesh) prepareIBO(required bool) {
	hasIBO := m.ibo.Value != 0
	if required && !hasIBO {
		m.ibo = m.ctx.CreateBuffer()
	}
	if !required && hasIBO {
		m.ctx.DeleteBuffer(m.ibo)
		m.ibo.Value = 0
	}
}

func (m *Mesh) calcVertexData() {
	sProg, id := m.getShaderProgram()
	if !assert(sProg != nil, "Shader %q not loaded", m.material.sProgKey) {
		return
	}
	if id == m.vtxShaderProgID { // already up-to-date
		return
	}
	m.vtxShaderProgID = id

	components := 0
	for _, vaName := range m.vertexAttributes {
		_, typ := sProg.getAttribLocation(vaName)
		assert(typ != 0, "Attribute %q is not supported by shader %q", vaName, sProg)

		components += int(vaTypePropertyMapping[typ].components)
	}

	m.vertexSize = components * 4
	if m.vboSize == 0 {
		m.vertexCount = 0
	} else {
		assert(m.vertexSize > 0, "Corrupt vertex data (no attributes)")
		m.vertexCount = m.vboSize / m.vertexSize
	}

	// can happen if the vertex attribute type expected by the shader does not match the assumptions of the caller:
	assert(m.vertexCount*m.vertexSize == m.vboSize, "Invalid vertex data: size does not match vertex attribute expectations")
}

func (m *Mesh) determinePrimitiveCount(indexCount int, primitiveType PrimitiveType) int {
	assertMsg := "Index count %d is incompatible with primitive type %q"

	switch primitiveType {
	case gl.POINTS:
		return indexCount
	case gl.LINE_STRIP:
		assert(indexCount >= 1, assertMsg, indexCount, primitiveType)
		return indexCount - 1
	case gl.LINE_LOOP:
		return indexCount
	case gl.LINES:
		assert(indexCount%2 == 0, assertMsg, indexCount, primitiveType)
		return indexCount / 2
	case gl.TRIANGLE_STRIP:
		assert(indexCount >= 2, assertMsg, indexCount, primitiveType)
		return indexCount - 2
	case gl.TRIANGLE_FAN:
		assert(indexCount >= 2, assertMsg, indexCount, primitiveType)
		return indexCount - 2
	case gl.TRIANGLES:
		assert(indexCount%3 == 0, assertMsg, indexCount, primitiveType)
		return indexCount / 3
	}
	assertFail("Unknown primitive type %q", primitiveType)
	return 0
}

// SetVertexSubData changes parts of the underlying vertex buffer.
// 	- vertexOffset		Vertex offset.
//  - vertices			Underlying vertex data that will overwrite existing buffers
// Cannot change the underlying vertex buffer size.
func (m *Mesh) SetVertexSubData(vertexOffset int, vertices []float32) {
	m.calcVertexData()
	assert(vertexOffset >= 0 && vertexOffset < m.vertexCount, "Invalid vertex offset (out of range)")
	assert((len(vertices)*4)%(m.vertexSize) == 0, "Invalid vertex data size")

	ctx := m.ctx
	ctx.Lock()
	defer ctx.Unlock()
	ctx.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	ctx.BufferSubDataFloat32(gl.ARRAY_BUFFER, vertexOffset*m.vertexSize, vertices)
}

// SetIndexSubData changes parts of the underlying index buffer.
// 	- indexOffset		Index offset.
//  - indices			Underlying index data that will overwrite existing buffers
// Cannot change the underlying index buffer size.
func (m *Mesh) SetIndexSubData(indexOffset int, indices []uint16) {
	m.calcVertexData()
	assert(m.ibo.Value != 0, "The mesh does not use indexed drawing")
	assert(indexOffset >= 0 && indexOffset < m.indexCount, "Invalid index offset (out of range)")

	ctx := m.ctx
	ctx.Lock()
	defer ctx.Unlock()
	ctx.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.ibo)
	ctx.BufferSubDataUint16(gl.ELEMENT_ARRAY_BUFFER, indexOffset*2, indices)
}

// ClearVertexData clears the underlying buffers.
// Rendering this mesh will not draw anything.
func (m *Mesh) ClearVertexData() {
	m.SetVertexData([]float32{}, nil, gl.TRIANGLES, []string{}, InterleavedBuffer)
}

// draw must only be called during sync. rendering (no context lock)
// the renderer needs to apply the shader and material required by the mesh.
func (m *Mesh) draw(renderState *RenderState) {
	if m.indexCount == 0 {
		return
	}
	ctx := m.ctx

	if renderState.sProgID != m.vtxShaderProgID {
		m.calcVertexData()
	}

	sProg := renderState.sProg

	sProg.configureVertexAttributes(m.vertexAttributes, true)
	defer sProg.configureVertexAttributes(m.vertexAttributes, false)

	ctx.BindBuffer(gl.ARRAY_BUFFER, m.vbo)
	if m.bufferLayout == InterleavedBuffer {
		m.configureInterleavedVertexAttributes(sProg)
	} else {
		m.configureCompactVertexAttributes(sProg)
	}

	if m.ibo.Value != 0 {
		ctx.DrawElements(gl.Enum(m.primitiveType), m.indexCount, gl.UNSIGNED_SHORT, 0)
	} else {
		ctx.DrawArrays(gl.Enum(m.primitiveType), 0, m.indexCount)
	}

	renderState.totalDrawCalls++
	renderState.totalPrimitives += m.primitiveCount
}

func (m *Mesh) configureInterleavedVertexAttributes(sProg *ShaderProgram) {
	ctx := m.ctx

	offset := 0
	stride := m.vertexSize
	for _, vaName := range m.vertexAttributes {
		loc, typ := sProg.getAttribLocation(vaName)
		if !assert(typ != 0, "Unsupported vertex attribute %q", vaName) {
			continue
		}

		typProps := vaTypePropertyMapping[typ]
		ctx.VertexAttribPointer(loc, int(typProps.components), typProps.compType, false, stride, offset)
		offset += int(typProps.components) * 4
	}
}

func (m *Mesh) configureCompactVertexAttributes(sProg *ShaderProgram) {
	ctx := m.ctx

	vertexCount := m.vertexCount
	component := 0
	for _, vaName := range m.vertexAttributes {
		loc, typ := sProg.getAttribLocation(vaName)
		if !assert(typ != 0, "Unsupported vertex attribute %q", vaName) {
			continue
		}

		typProps := vaTypePropertyMapping[typ]
		stride := int(typProps.components) * 4
		ctx.VertexAttribPointer(loc, int(typProps.components), typProps.compType, false, stride, vertexCount*component*4)
		component += int(typProps.components)
	}
}
