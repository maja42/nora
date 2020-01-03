package nora

import (
	"github.com/sirupsen/logrus"
)

// iAssertTrue is equivalent to assert.True, but indicates an internal problem that should not happen even for invalid API usage.
func iAssertTrue(t bool, format string, args ...interface{}) bool {
	if !t {
		logrus.Errorf("[Internal error] "+format, args...)
	}
	return t
}

// iAssertFail is equivalent to assert.Fail, but indicates an internal problem that should not happen even for invalid API usage.
func iAssertFail(format string, args ...interface{}) {
	iAssertTrue(false, format, args...)
}

// iAssertFunc is equivalent to assert.Func, but indicates an internal problem that should not happen even for invalid API usage.
func iAssertFunc(f func() bool, format string, args ...interface{}) bool {
	return iAssertTrue(f(), format, args...)
}
