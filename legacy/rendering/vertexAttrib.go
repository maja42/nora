package rendering

import (
	"golang.org/x/mobile/gl"
)

type vaTypeProperties struct {
	components uint8
	compType   gl.Enum
}

var vaTypePropertyMapping = map[gl.Enum]vaTypeProperties{
	gl.FLOAT:      {1, gl.FLOAT},
	gl.FLOAT_VEC2: {2, gl.FLOAT},
	gl.FLOAT_VEC3: {3, gl.FLOAT},
	gl.FLOAT_VEC4: {4, gl.FLOAT},
	// Other vertex attribute types (int, ...) are currently not supported
}

var shaderTypeStringMapping = map[gl.Enum]string{
	// Currently only contains OpenGL ES 2.0 types.
	// TODO: should naming only be present in debug-builds?
	gl.FLOAT: "float",
	gl.BOOL:  "bool",

	gl.FLOAT_VEC2: "fvec2",
	gl.FLOAT_VEC3: "fvec3",
	gl.FLOAT_VEC4: "fvec4",

	gl.INT_VEC2: "ivec2",
	gl.INT_VEC3: "ivec3",
	gl.INT_VEC4: "ivec4",

	gl.BOOL_VEC2: "bvec2",
	gl.BOOL_VEC3: "bvec3",
	gl.BOOL_VEC4: "bvec4",

	gl.FLOAT_MAT2: "fmat2",
	gl.FLOAT_MAT3: "fmat3",
	gl.FLOAT_MAT4: "fmat4",

	gl.SAMPLER_2D:   "sampler2D",
	gl.SAMPLER_CUBE: "samplerCube",
}

//func typToString(e gl.Enum) string {
//	s, ok := shaderTypeStringMapping[e]
//	if !ok {
//		return fmt.Sprintf("<%X>", e)
//	}
//	return s
//}
