package rendering

import (
	"golang.org/x/mobile/gl"
)

// Material defines how geometry is rendered.
type Material struct {
	sProgKey ShaderProgKey

	textures map[string]TextureKey

	uniformf   map[string][]float32
	uniformMat map[string][]float32
	uniformi   map[string][]int32
}

// NewMaterial creates a new material based on the given shader.
func NewMaterial(sProgKey ShaderProgKey) *Material {
	return &Material{
		sProgKey: sProgKey,
		textures: make(map[string]TextureKey),
		uniformf: make(map[string][]float32),
		uniformi: make(map[string][]int32),
	}
}

// SetShader changes the shader program.
func (m *Material) SetShader(sProgKey ShaderProgKey) {
	m.sProgKey = sProgKey
}

func (m *Material) AddTextureBinding(uniformName string, texKey TextureKey) {
	m.textures[uniformName] = texKey
}

func (m *Material) Uniform1f(uniformName string, x float32) {
	m.uniformf[uniformName] = []float32{x}
}

func (m *Material) Uniform1fv(uniformName string, v []float32) {
	m.uniformf[uniformName] = v
}

func (m *Material) Uniform1i(uniformName string, x int32) {
	m.uniformi[uniformName] = []int32{x}
}

func (m *Material) Uniform1iv(uniformName string, v []int32) {
	m.uniformi[uniformName] = v
}

func (m *Material) Uniform2f(uniformName string, x, y float32) {
	m.uniformf[uniformName] = []float32{x, y}
}

func (m *Material) Uniform2fv(uniformName string, v []float32) {
	m.uniformf[uniformName] = v
}

func (m *Material) Uniform2i(uniformName string, x, y int32) {
	m.uniformi[uniformName] = []int32{x, y}
}

func (m *Material) Uniform2iv(uniformName string, v []int32) {
	m.uniformi[uniformName] = v
}

func (m *Material) Uniform3f(uniformName string, x, y, z float32) {
	m.uniformf[uniformName] = []float32{x, y, z}
}

func (m *Material) Uniform3fv(uniformName string, v []float32) {
	m.uniformf[uniformName] = v
}

func (m *Material) Uniform3i(uniformName string, x, y, z int32) {
	m.uniformi[uniformName] = []int32{x, y, z}
}

func (m *Material) Uniform3iv(uniformName string, v []int32) {
	m.uniformi[uniformName] = v
}

func (m *Material) Uniform4f(uniformName string, x, y, z, w float32) {
	m.uniformf[uniformName] = []float32{x, y, z, w}
}

func (m *Material) Uniform4fv(uniformName string, v []float32) {
	m.uniformf[uniformName] = v
}

func (m *Material) Uniform4i(uniformName string, x, y, z, w int32) {
	m.uniformi[uniformName] = []int32{x, y, z, w}
}

func (m *Material) Uniform4iv(uniformName string, v []int32) {
	m.uniformi[uniformName] = v
}

func (m *Material) UniformMatrix2fv(uniformName string, v []float32) {
	m.uniformMat[uniformName] = v
}

func (m *Material) UniformMatrix3fv(uniformName string, v []float32) {
	m.uniformMat[uniformName] = v
}

func (m *Material) UniformMatrix4fv(uniformName string, v []float32) {
	m.uniformMat[uniformName] = v
}

// apply must only be called during sync. rendering (expects locked context)
func (m *Material) apply(glCtx gl.Context, shader *ShaderProgram, texTargets *textureTargets) {
	// The caller needs to pass the (correct) shader program based on the internal sProgKey

	for name, texKey := range m.textures {
		loc, ok := shader.getUniformLocation(name)
		if !assert(ok, "Uniform %q is not supported by shader %q", name, m.sProgKey) {
			continue // ignore uniform
		}
		texTargets.Bind(loc, texKey)
	}

	// We should probably store the glCtx.UniformXXX func-pointer alongside the uniform values (ignore the memory-overhead)
	// Once we also support 3x4 and 4x3 matrices, we need to do that anyways, because we can't infer the matrix dimensions based in the uniform size.

	for name, u := range m.uniformf {
		loc, ok := shader.getUniformLocation(name)
		if !assert(ok, "Uniform %q is not supported by shader %q", name, m.sProgKey) {
			continue // ignore uniform
		}

		switch len(u) {
		case 1:
			glCtx.Uniform1fv(loc, u)
		case 2:
			glCtx.Uniform2fv(loc, u)
		case 3:
			glCtx.Uniform3fv(loc, u)
		case 4:
			glCtx.Uniform4fv(loc, u)
		default:
			assertFail("Unknown uniform type (floats of length %d)", len(u))
		}
	}
	for name, u := range m.uniformMat {
		loc, ok := shader.getUniformLocation(name)
		if !assert(ok, "Uniform %q is not supported by shader %q", name, m.sProgKey) {
			continue // ignore uniform
		}

		switch len(u) {
		case 4:
			glCtx.UniformMatrix2fv(loc, u)
		case 9:
			glCtx.UniformMatrix3fv(loc, u)
		case 16:
			glCtx.UniformMatrix4fv(loc, u)
		default:
			assertFail("Unknown uniform type (float matrix of length %d)", len(u))
		}
	}
	for name, u := range m.uniformi {
		loc, ok := shader.getUniformLocation(name)
		if !assert(ok, "Uniform %q is not supported by shader %q", name, m.sProgKey) {
			continue // ignore uniform
		}

		switch len(u) {
		case 1:
			glCtx.Uniform1iv(loc, u)
		case 2:
			glCtx.Uniform2iv(loc, u)
		case 3:
			glCtx.Uniform3iv(loc, u)
		case 4:
			glCtx.Uniform4iv(loc, u)
		default:
			assertFail("Unknown uniform type (ints of length %d)", len(u))
		}
	}
}
