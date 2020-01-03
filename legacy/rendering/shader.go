package rendering

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/maja42/logicat/utils"
	"golang.org/x/mobile/gl"
)

// TODO: Fix exported types of this file

const ModelTransformUniformName = "modelTransform"
const VPMatrixUniformName = "vpMatrix"

// ShaderProgKey is used to connect shader programs and materials/meshes.
type ShaderProgKey string

// ShaderProgramDefinition contains all necessary information for compiling and linking a shader program
type ShaderProgramDefinition struct {
	VertexShaderPath   string
	FragmentShaderPath string
}

// ShaderProgram represents a GPU shader program for rendering
type ShaderProgram struct {
	ctx     gl.Context
	program gl.Program

	attributeLocations     map[string]gl.Attrib
	attributeTypes         map[string]gl.Enum // stores the underlying type of the vertex attributes
	uniformLocations       map[string]gl.Uniform
	modelTransformLocation gl.Uniform
	vpMatrixLocation       gl.Uniform
}

// NewShaderProgram creates a new shader program on the GPU.
// The program can be reused by loading (compiling and linking) different shaders.
// Needs to be destroyed afterwards to free GPU resources.
// Note: The usability of shareProgram objects is limited, because they can be reloaded at any time. Use 'ShaderProgKey's instead!
func NewShaderProgram(glCtx gl.Context) *ShaderProgram {
	// TODO: use sync.Pool for shaders / loadedShaders?
	return &ShaderProgram{
		ctx:                    glCtx,
		program:                glCtx.CreateProgram(),
		modelTransformLocation: gl.Uniform{Value: -1},
		vpMatrixLocation:       gl.Uniform{Value: -1},
	}
}

func (p *ShaderProgram) String() string {
	return fmt.Sprintf("ShaderProgram(%d)", p.program.Value)
}

// Load compiles and links the shader program from the given sources.
// Can be called multiple times.
func (p *ShaderProgram) Load(def *ShaderProgramDefinition) error {
	logger.Infof("Loading %s...", p)

	shaders, err := p.compileShaders(def)
	if err != nil {
		return fmt.Errorf("compile shaders: %s", err)
	}

	defer func() {
		for _, shader := range shaders {
			p.ctx.DeleteShader(shader)
		}
	}()

	if err := p.linkShaderProgram(shaders); err != nil {
		return fmt.Errorf("link program: %s", err)
	}

	p.fetchVertexAttributes()
	p.fetchUniforms()

	iAssertNoGLError(p.ctx, "load %s", p)
	return nil
}

func (p *ShaderProgram) compileShaders(def *ShaderProgramDefinition) ([]gl.Shader, error) {
	vShader, err := compileShaderFromFile(p.ctx, gl.VERTEX_SHADER, def.VertexShaderPath)
	if err != nil {
		return nil, fmt.Errorf("vertex shader: %s", err)
	}
	fShader, err := compileShaderFromFile(p.ctx, gl.FRAGMENT_SHADER, def.FragmentShaderPath)
	if err != nil {
		p.ctx.DeleteShader(vShader)
		return nil, fmt.Errorf("fragment shader: %s", err)
	}
	return []gl.Shader{vShader, fShader}, nil
}

// compileShaderFromFile creates a new shader object on the GPU, compiled with the given glsl source file.
// Needs to be destroyed to free GPU resources.
func compileShaderFromFile(glCtx gl.Context, shaderType gl.Enum, path string) (gl.Shader, error) {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return gl.Shader{}, err
	}
	name := filepath.Base(path)
	return compileShader(glCtx, name, shaderType, string(source))
}

// compileShader creates a new shader object on the GPU, compiled with the given glsl source code.
// Needs to be destroyed to free GPU resources.
func compileShader(glCtx gl.Context, name string, shaderType gl.Enum, src string) (gl.Shader, error) {
	logger.Debugf("Compiling shader %q...", name)

	shader := glCtx.CreateShader(shaderType)

	glCtx.ShaderSource(shader, src)
	glCtx.CompileShader(shader)

	if glCtx.GetShaderi(shader, gl.COMPILE_STATUS) == gl.FALSE {
		log := glCtx.GetShaderInfoLog(shader)
		glCtx.DeleteShader(shader)
		return gl.Shader{}, fmt.Errorf("compile failed - %q\n%s", name, utils.Indent(log))
	}

	iAssertNoGLError(glCtx, "compile shader %q", name)
	return shader, nil
}

func (p *ShaderProgram) linkShaderProgram(shaders []gl.Shader) error {
	ctx := p.ctx
	logger.Debugf("Linking %s...", p)

	//noinspection ALL
	for _, shader := range shaders {
		ctx.AttachShader(p.program, shader)
		defer ctx.DetachShader(p.program, shader)
	}

	ctx.LinkProgram(p.program)
	ctx.ValidateProgram(p.program)
	if ctx.GetProgrami(p.program, gl.LINK_STATUS) == gl.FALSE {
		log := ctx.GetProgramInfoLog(p.program)
		return fmt.Errorf("could not link shader program: %s", log)
	}
	return nil
}

func (p *ShaderProgram) fetchVertexAttributes() {
	ctx := p.ctx

	attributeCount := uint32(ctx.GetProgrami(p.program, gl.ACTIVE_ATTRIBUTES))
	p.attributeLocations = make(map[string]gl.Attrib, attributeCount)
	p.attributeTypes = make(map[string]gl.Enum, attributeCount)

	for idx := uint32(0); idx < attributeCount; idx++ {
		name, _, typ := ctx.GetActiveAttrib(p.program, idx)
		p.attributeLocations[name] = gl.Attrib{Value: uint(idx)}
		p.attributeTypes[name] = typ
	}
}

func (p *ShaderProgram) fetchUniforms() {
	ctx := p.ctx

	uniformCount := uint32(ctx.GetProgrami(p.program, gl.ACTIVE_UNIFORMS))
	p.uniformLocations = make(map[string]gl.Uniform, uniformCount)
	p.modelTransformLocation.Value = -1
	p.vpMatrixLocation.Value = -1

	for idx := uint32(0); idx < uniformCount; idx++ {
		name, _, _ := ctx.GetActiveUniform(p.program, idx) // we *could* store the type for validation purposes
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

func (p *ShaderProgram) Use() {
	p.ctx.UseProgram(p.program)
}

// configureVertexAttributes ensures that all given vertex attributes are either enabled or disabled
// configureVertexAttributes must only be called during sync. rendering (no context lock)
func (p *ShaderProgram) configureVertexAttributes(vertexAttributes []string, enable bool) {
	// If the enabling/disabling of vertex attributes would result in a performance loss,
	// it would be possible to cache the "currently enabled attributes" in the renderState,
	// use an uint64 bitset to check which attributes need to be disabled/enabled and go from there.

	ctx := p.ctx
	for _, name := range vertexAttributes {
		loc, ok := p.attributeLocations[name]
		iAssert(ok, "Enable vertex attribute %q: not supported by %s", name, p) // cannot happen (already caught when configuring a mesh with vertices/materials)
		if enable {
			ctx.EnableVertexAttribArray(loc)
		} else {
			ctx.DisableVertexAttribArray(loc)
		}
	}
}

func (p *ShaderProgram) getAttribLocation(attributeName string) (gl.Attrib, gl.Enum) {
	return p.attributeLocations[attributeName], p.attributeTypes[attributeName]
}

func (p *ShaderProgram) getUniformLocation(uniformName string) (gl.Uniform, bool) {
	loc, ok := p.uniformLocations[uniformName]
	return loc, ok
}

func (p *ShaderProgram) Destroy() {
	logger.Debugf("Destroying %s", p)
	p.ctx.DeleteProgram(p.program)
}
