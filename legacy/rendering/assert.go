package rendering

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/mobile/gl"
)

// Source: https://www.khronos.org/registry/OpenGL-Refpages/gl4/html/glGetError.xhtml
var errMapping = map[gl.Enum]string{
	gl.NO_ERROR:                      "No error has been recorded.",
	gl.INVALID_ENUM:                  "An unacceptable value is specified for an enumerated argument. The offending command is ignored and has no other side effect than to set the error flag.",
	gl.INVALID_VALUE:                 "A numeric argument is out of range. The offending command is ignored and has no other side effect than to set the error flag.",
	gl.INVALID_OPERATION:             "The specified operation is not allowed in the current state. The offending command is ignored and has no other side effect than to set the error flag.",
	gl.INVALID_FRAMEBUFFER_OPERATION: "The command is trying to render to or read from the framebuffer while the currently bound framebuffer is not framebuffer complete. The offending command is ignored and has no other side effect than to set the error flag.",
	gl.OUT_OF_MEMORY:                 "There is not enough memory left to execute the command. The state of the GL is undefined, except for the state of the error flags, after this error is recorded.",
	//gl.STACK_UNDERFLOW:               "An attempt has been made to perform an operation that would cause an internal stack to underflow.",
	//gl.STACK_OVERFLOW:                "An attempt has been made to perform an operation that would cause an internal stack to overflow.",
}

// GetGLError checks the current gl error flag and returns an error if available
func GetGLError(glCtx gl.Context) error {
	if glError := glCtx.GetError(); glError != gl.NO_ERROR {
		errStr, ok := errMapping[glError]
		if !ok {
			errStr = "Error " + strconv.Itoa(int(glError))
		}
		return errors.New(errStr)
	}
	return nil
}

func iAssertNoGLError(glCtx gl.Context, format string, args ...interface{}) bool {
	err := GetGLError(glCtx)
	if err == nil {
		return true
	}
	return assert(err == nil, "%s", "[OpenGL error] "+fmt.Sprintf(format, args...)+"\n\t"+
		err.Error())
}

// iAssert is equivalent to assert, but indicates a problem that should not happen even for invalid API usage
func iAssert(t bool, format string, args ...interface{}) bool {
	return assert(t, "[Internal error] "+format, args...)
}

// iAssertFunc is equivalent to assertFunc, but indicates a problem that should not happen even for invalid API usage
func iAssertFunc(f func() bool, format string, args ...interface{}) bool {
	return assertFunc(f, "[Internal error] "+format, args...)
}

func assert(t bool, format string, args ...interface{}) bool {
	if !t {
		logger.Errorf(format, args...)
	}
	return t
}

func assertFail(format string, args ...interface{}) {
	assert(false, format, args...)
}

func assertFunc(f func() bool, format string, args ...interface{}) bool {
	return assert(f(), "[Internal error] "+format, args...)
}

//func assertEqual(a, b interface{}, format string, args ...interface{}) bool {
//	return assert(a == b, format, args...)
//}

//func assert(_ bool, _ string, _ ...interface{}) bool {
//	return true
//}
