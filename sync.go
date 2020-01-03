package nora

import (
	"sync"

	"github.com/maja42/gl"
	"github.com/maja42/nora/assert"
)

// TODO: locking should be done by the gl library:
// All operations on a bound buffer should be enqueued to the render thread at once

type glSync struct {
	arrayBufferLock        sync.Mutex
	elementArrayBufferLock sync.Mutex
}

// lockBuffer locks a buffer target.
// Allows exclusive use until the buffer is unlocked.
func (s *glSync) lockBuffer(target gl.Enum) {
	switch target {
	case gl.ARRAY_BUFFER:
		s.arrayBufferLock.Lock()
	case gl.ELEMENT_ARRAY_BUFFER:
		s.elementArrayBufferLock.Lock()
	default:
		assert.Fail("Unknown buffer target %v", target)
	}
}

// unlockBuffer unlocks a buffer target.
func (s *glSync) unlockBuffer(target gl.Enum) {
	switch target {
	case gl.ARRAY_BUFFER:
		s.arrayBufferLock.Unlock()
	case gl.ELEMENT_ARRAY_BUFFER:
		s.elementArrayBufferLock.Unlock()
	default:
		assert.Fail("Unknown buffer target %v", target)
	}
}
