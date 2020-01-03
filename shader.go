package nora

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/maja42/gl"
	"github.com/maja42/nora/assert"
	"github.com/sirupsen/logrus"
)

// TODO: fix exported types.

// ModelTransformUniformName contains the name of the shader uniform (mat4) that will receive the model-transform matrix when rendering
const ModelTransformUniformName = "modelTransform"

// VPMatrixUniformName contains the name of the shader uniform (mat4) that will receive the camera's view-projection-matrix when rendering
const VPMatrixUniformName = "vpMatrix"

// ShaderProgramDefinition contains all necessary information for compiling and linking a shader program
type ShaderProgramDefinition struct {
	// Path to the vertex shader
	VertexShaderPath string
	// Path to the fragment shader
	FragmentShaderPath string
}

// shaderProgram represents a GPU shader program for rendering
type shaderProgram struct {
	program gl.Program

	attributeLocations     map[string]gl.Attrib
	attributeTypes         map[string]gl.Enum // stores the underlying type of the vertex attributes
	uniformLocations       map[string]gl.Uniform
	modelTransformLocation gl.Uniform
	vpMatrixLocation       gl.Uniform
}

// newShaderProgram creates a new shader program on the GPU.
// The program can be reused by loading (compiling and linking) different shaders.
// Needs to be destroyed afterwards to free GPU resources.
// Note: The usability of shareProgram objects is limited, because they can be reloaded at any time. Use 'ShaderProgKey's instead.
func newShaderProgram() *shaderProgram {
	// TODO: use sync.Pool for shaders / loadedShaders?
	return &shaderProgram{
		program:                gl.CreateProgram(),
		modelTransformLocation: gl.Uniform{Value: -1},
		vpMatrixLocation:       gl.Uniform{Value: -1},
	}
}

func (p *shaderProgram) String() string {
	return fmt.Sprintf("ShaderProgram(%d)", p.program.Value)
}

// Load compiles and links the shader program from the given sources.
// Can be called multiple times.
func (p *shaderProgram) Load(def *ShaderProgramDefinition) error {
	logrus.Infof("Loading %s...", p)

	shaders, err := p.compileShaders(def)
	if err != nil {
		return fmt.Errorf("compile shaders: %s", err)
	}

	defer func() {
		for _, shader := range shaders {
			gl.DeleteShader(shader)
		}
	}()

	if err := p.linkShaderProgram(shaders); err != nil {
		return fmt.Errorf("link program: %s", err)
	}

	p.fetchVertexAttributes()
	p.fetchUniforms()

	assert.NoGLError("load %s", p)
	return nil
}

func (p *shaderProgram) compileShaders(def *ShaderProgramDefinition) ([]gl.Shader, error) {
	vShader, err := compileShaderFromFile(gl.VERTEX_SHADER, def.VertexShaderPath)
	if err != nil {
		return nil, fmt.Errorf("vertex shader: %s", err)
	}
	fShader, err := compileShaderFromFile(gl.FRAGMENT_SHADER, def.FragmentShaderPath)
	if err != nil {
		gl.DeleteShader(vShader)
		return nil, fmt.Errorf("fragment shader: %s", err)
	}
	return []gl.Shader{vShader, fShader}, nil
}

// compileShaderFromFile creates a new shader object on the GPU, compiled with the given glsl source file.
// Needs to be destroyed to free GPU resources.
func compileShaderFromFile(shaderType gl.Enum, path string) (gl.Shader, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return gl.Shader{}, err
	}
	name := filepath.Base(path)
	return compileShader(name, shaderType, string(source))
}

// compileShader creates a new shader object on the GPU, compiled with the given glsl source code.
// Needs to be destroyed to free GPU resources.
func compileShader(name string, shaderType gl.Enum, src string) (gl.Shader, error) {
	logrus.Debugf("Compiling shader %q...", name)

	shader := gl.CreateShader(shaderType)

	gl.ShaderSource(shader, src)
	gl.CompileShader(shader)

	if gl.GetShaderi(shader, gl.COMPILE_STATUS) == gl.FALSE {
		log := gl.GetShaderInfoLog(shader)
		gl.DeleteShader(shader)
		return gl.Shader{}, fmt.Errorf("compile failed - %q\n%s", name, indent(log))
	}

	assert.NoGLError("compile shader %q", name)
	return shader, nil
}

func (p *shaderProgram) linkShaderProgram(shaders []gl.Shader) error {
	logrus.Debugf("Linking %s...", p)

	//noinspection ALL
	for _, shader := range shaders {
		gl.AttachShader(p.program, shader)
		defer gl.DetachShader(p.program, shader)
	}

	gl.LinkProgram(p.program)
	gl.ValidateProgram(p.program)
	if gl.GetProgrami(p.program, gl.LINK_STATUS) == gl.FALSE {
		log := gl.GetProgramInfoLog(p.program)
		return fmt.Errorf("could not link shader program: %s", log)
	}
	return nil
}

func (p *shaderProgram) fetchVertexAttributes() {
	attributeCount := uint32(gl.GetProgrami(p.program, gl.ACTIVE_ATTRIBUTES))
	p.attributeLocations = make(map[string]gl.Attrib, attributeCount)
	p.attributeTypes = make(map[string]gl.Enum, attributeCount)

	for idx := uint32(0); idx < attributeCount; idx++ {
		name, _, typ := gl.GetActiveAttrib(p.program, idx)
		p.attributeLocations[name] = gl.Attrib{Value: uint(idx)}
		p.attributeTypes[name] = typ
	}
}

func (p *shaderProgram) fetchUniforms() {
	uniformCount := uint32(gl.GetProgrami(p.program, gl.ACTIVE_UNIFORMS))
	p.uniformLocations = make(map[string]gl.Uniform, uniformCount)
	p.modelTransformLocation.Value = -1
	p.vpMatrixLocation.Value = -1

	for idx := uint32(0); idx < uniformCount; idx++ {
		name, _, _ := gl.GetActiveUniform(p.program, idx) // we *could* store the type for validation purposes
		location := gl.Uniform{Value: int32(idx)}

		switch name {
		case ModelTransformUniformName:
			p.modelTransformLocation = location
		case VPMatrixUniformName:
			p.vpMatrixLocation = location
		default:
			p.uniformLocations[name] = location
		}
	}
}

func (p *shaderProgram) Use() {
	gl.UseProgram(p.program)
}

// configureVertexAttributes ensures that all given vertex attributes are either enabled or disabled
// configureVertexAttributes must only be called during sync. rendering (no context lock)
func (p *shaderProgram) configureVertexAttributes(vertexAttributes []string, enable bool) {
	// If the enabling/disabling of vertex attributes would result in a performance loss,
	// it would be possible to cache the "currently enabled attributes" in the renderState,
	// use an uint64 bitset to check which attributes need to be disabled/enabled and go from there.

	for _, name := range vertexAttributes {
		loc, ok := p.attributeLocations[name]
		iAssertTrue(ok, "Enable vertex attribute %q: not supported by %s", name, p) // cannot happen (already caught when configuring a mesh with vertices/materials)
		if enable {
			gl.EnableVertexAttribArray(loc)
		} else {
			gl.DisableVertexAttribArray(loc)
		}
	}
}

func (p *shaderProgram) getAttribLocation(attributeName string) (gl.Attrib, gl.Enum) {
	return p.attributeLocations[attributeName], p.attributeTypes[attributeName]
}

func (p *shaderProgram) getUniformLocation(uniformName string) (gl.Uniform, bool) {
	loc, ok := p.uniformLocations[uniformName]
	return loc, ok
}

func (p *shaderProgram) Destroy() {
	logrus.Debugf("Destroying %s", p)
	gl.DeleteProgram(p.program)
}
