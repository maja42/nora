package assert

import (
	"fmt"

	"github.com/maja42/gl"
	"github.com/sirupsen/logrus"
)

func True(t bool, format string, args ...interface{}) bool {
	if !t {
		logrus.Errorf(format, args...)
	}
	return t
}
func False(t bool, format string, args ...interface{}) bool {
	True(!t, format, args...)
	return t
}

func Fail(format string, args ...interface{}) {
	True(false, format, args...)
}

func Func(f func() bool, format string, args ...interface{}) bool {
	return True(f(), format, args...)
}

func NoGLError(format string, args ...interface{}) bool {
	err := gl.CheckError()
	if err == nil {
		return true
	}
	Fail("%s", "[OpenGL error] "+fmt.Sprintf(format, args...)+"\n\t"+
		err.Error())
	return false
}
