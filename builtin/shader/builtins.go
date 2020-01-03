package shader

import (
	"os"

	"github.com/maja42/nora"
)

const (
	RGB_2D          nora.ShaderProgKey = "rgb-2D"
	COL_2D          nora.ShaderProgKey = "col-2D"
	TEX_2D          nora.ShaderProgKey = "tex-2D"
	COL_TEX_2D      nora.ShaderProgKey = "col-tex-2D"
	RGB_TEX_2D      nora.ShaderProgKey = "rgb-tex-2D"
	RGB_3D          nora.ShaderProgKey = "rgb-3D"
	TEX_3D          nora.ShaderProgKey = "tex-3D"
	COL_3D          nora.ShaderProgKey = "col-3D"
	COL_NORM_3D     nora.ShaderProgKey = "col-norm-3D"
	COL_TEX_NORM_3D nora.ShaderProgKey = "col-tex-norm-3D"
)

// Builtins returns all built-in shader programs
var Builtins = func(shaderLocation string) map[nora.ShaderProgKey]nora.ShaderProgramDefinition {
	shaderLocation += string(os.PathSeparator)
	return map[nora.ShaderProgKey]nora.ShaderProgramDefinition{
		COL_2D: {
			VertexShaderPath:   shaderLocation + "2d.vs.glsl",
			FragmentShaderPath: shaderLocation + "col.fs.glsl",
		},
		RGB_2D: {
			VertexShaderPath:   shaderLocation + "2d-rgb.vs.glsl",
			FragmentShaderPath: shaderLocation + "rgb.fs.glsl",
		},
		TEX_2D: {
			VertexShaderPath:   shaderLocation + "2d-tex.vs.glsl",
			FragmentShaderPath: shaderLocation + "tex.fs.glsl",
		},
		COL_TEX_2D: {
			VertexShaderPath:   shaderLocation + "2d-tex.vs.glsl",
			FragmentShaderPath: shaderLocation + "col-tex.fs.glsl",
		},
		RGB_TEX_2D: {
			VertexShaderPath:   shaderLocation + "2d-rgb-tex.vs.glsl",
			FragmentShaderPath: shaderLocation + "rgb-tex.fs.glsl",
		},
		//RGB_3D: {
		//	VertexShaderPath:   shaderLocation + "3d-rgb.vs.glsl",
		//	FragmentShaderPath: shaderLocation + "rgb.fs.glsl",
		//	VertexAttributes:   rendering.VA_Pos3D | rendering.VA_RGB,
		//	Uniforms:           []string{"modelTransform", "vpMatrix"},
		//},
		//TEX_3D: {
		//	VertexShaderPath:   shaderLocation + "3d-tex.vs.glsl",
		//	FragmentShaderPath: shaderLocation + "tex.fs.glsl",
		//	VertexAttributes:   rendering.VA_Pos3D | rendering.VA_TexUV,
		//	Uniforms:           []string{"modelTransform", "vpMatrix", "uSampler"},
		//},
		//COL_3D: {
		//	VertexShaderPath:   shaderLocation + "3d-col.vs.glsl",
		//	FragmentShaderPath: shaderLocation + "rgba.fs.glsl",
		//	VertexAttributes:   rendering.VA_Pos3D,
		//	Uniforms:           []string{"modelTransform", "vpMatrix", "fragColor"},
		//},
		COL_NORM_3D: {
			VertexShaderPath:   shaderLocation + "3d-col-norm.vs.glsl",
			FragmentShaderPath: shaderLocation + "rgba.fs.glsl",
		},
		COL_TEX_NORM_3D: {
			VertexShaderPath:   shaderLocation + "3d-col-tex-norm.vs.glsl",
			FragmentShaderPath: shaderLocation + "rgba-tex.fs.glsl",
		},
	}
}
