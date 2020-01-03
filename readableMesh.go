package nora

// Note: OpenGL supports glGetBufferSubData to read back data from the GPU when needed.
// This is not supported in OpenGL ES though.

// ReadableMesh is a mesh whose vertex data can be read at any time.
// This means that vertex- and index data is not only stored in GPU-, but also in CPU-memory.
type ReadableMesh struct {
	Mesh
	vertices []float32
	indices  []uint16
}

func (m *ReadableMesh) SetVertexData(vertices []float32, indices []uint16, primitiveType PrimitiveType, vertexAttributes []string, bufferLayout BufferLayout) {
	m.Mesh.SetVertexData(vertices, indices, primitiveType, vertexAttributes, bufferLayout)

	if len(vertices) == 0 {
		m.vertices = nil
	} else { // reuse existing slice
		m.vertices = m.vertices[:0]
		m.vertices = append(m.vertices, vertices...)
	}

	if len(indices) == 0 {
		m.indices = nil
	} else { // reuse existing slice
		m.indices = m.indices[:0]
		m.indices = append(m.indices, indices...)
	}
}

func (m *ReadableMesh) SetVertexSubData(vertexOffset int, vertices []float32) {
	m.Mesh.SetVertexSubData(vertexOffset, vertices)
	copy(m.vertices[vertexOffset*m.vertexSize:], vertices)
}

func (m *ReadableMesh) SetIndexSubData(indexOffset int, indices []uint16) {
	m.Mesh.SetIndexSubData(indexOffset, indices)
	copy(m.indices[indexOffset:], indices)
}

func (m *ReadableMesh) ClearVertexData() {
	m.Mesh.ClearVertexData()
	m.vertices = nil
	m.indices = nil
}
