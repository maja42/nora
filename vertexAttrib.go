package nora

import "github.com/maja42/gl"

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
