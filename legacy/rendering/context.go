package rendering

import (
	"math"
	"sync"
	"unsafe"

	"golang.org/x/mobile/gl"
)

// Context provides access to the sceneGraph and the underlying rendering device
type Context struct {
	gl.Context
	sync.Mutex // protects stateful OpenGL commands from executing concurrently
	// TODO: different locks for buffers and textures!
	scene SceneGraph
}

func (c Context) BufferDataFloat32(buffer gl.Enum, data []float32, usage gl.Enum) {
	c.BufferData(buffer, float32asBytes(data), usage)
}

func (c Context) BufferDataUint16(buffer gl.Enum, data []uint16, usage gl.Enum) {
	c.BufferData(buffer, uint16asBytes(data), usage)
}

func (c Context) BufferSubDataFloat32(buffer gl.Enum, byteOffset int, data []float32) {
	c.BufferSubData(buffer, byteOffset, float32asBytes(data))
}

func (c Context) BufferSubDataUint16(buffer gl.Enum, byteOffset int, data []uint16) {
	c.BufferSubData(buffer, byteOffset, uint16asBytes(data))
}

// float32asBytes returns the byte representation of a float32 array
// performs an unsafe cast that depends on go implementation details
// super-fast (constant time)
// Note: This is a workaround due to the API of x/mobile/gl. The raw OpenGL command would not require this cast.
func float32asBytes(data []float32) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	bsize := 4 * len(data)
	fptr := unsafe.Pointer(&(data[0])) // pointer to first byte of data
	barr := (*[1]byte)(fptr)           // interpret as pointer to byte-array
	bslice := (*barr)[:]               // convert to byte slice with len == cap == 1

	overwriteSliceBounds(&bslice, bsize)
	return bslice
}

// float32asBytesSafe returns the byte representation of a float32 array
// all bytes are copied into a separate buffer - this operation is safe and slow
func float32asBytesSafe(data []float32) []byte {
	b := make([]byte, 4*len(data))
	for i, v := range data {
		u := math.Float32bits(v)
		b[4*i+0] = byte(u >> 0)
		b[4*i+1] = byte(u >> 8)
		b[4*i+2] = byte(u >> 16)
		b[4*i+3] = byte(u >> 24)
	}
	return b
}

// uint16asBytes returns the byte representation of an uint16 array
// performs an unsafe cast that depends on go implementation details
// super-fast (constant time)
// Note: This is a workaround due to the API of x/mobile/gl. The raw OpenGL command would not require this cast.
func uint16asBytes(data []uint16) []byte {
	if len(data) == 0 {
		return []byte{}
	}
	bsize := 2 * len(data)
	fptr := unsafe.Pointer(&(data[0])) // pointer to first byte of data
	barr := (*[1]byte)(fptr)           // interpret as pointer to byte-array
	bslice := (*barr)[:]               // convert to byte slice with len == cap == 1

	overwriteSliceBounds(&bslice, bsize)
	return bslice
}

// uint16asBytesSafe returns the byte representation of an uint16 array
// all bytes are copied into a separate buffer - this operation is safe and slow
func uint16asBytesSafe(data []uint16) []byte {
	b := make([]byte, 2*len(data))
	for i, v := range data {
		b[2*i+0] = byte(v >> 0)
		b[2*i+1] = byte(v >> 8)
	}
	return b
}

func overwriteSliceBounds(s *[]byte, bounds int) {
	addr := unsafe.Pointer(s) // pointer to slice structure

	// Capture the address where the length and cap size is stored
	// WARNING: This is fragile, depending on a go-internal structure.
	lenAddr := uintptr(addr) + uintptr(8)
	capAddr := uintptr(addr) + uintptr(16)
	// Create pointers to the length and cap size
	lenPtr := (*int)(unsafe.Pointer(lenAddr))
	capPtr := (*int)(unsafe.Pointer(capAddr))

	// the next changes can corrupt data of the original data type
	// my tests are doing fine though - if something breaks, add the possibility to fix the original data type after usage
	*lenPtr = bounds
	*capPtr = bounds
}
