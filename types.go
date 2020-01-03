package nora

import (
	"fmt"

	"github.com/maja42/gl"
)

type PrimitiveType gl.Enum
type BufferLayout uint8

const (
	InterleavedBuffer BufferLayout = iota // eg. <pos, rgb> <pos, rgb> ...
	CompactBuffer                         // eg. <pos, pos> <rgb, rgb>
)

func (p PrimitiveType) String() string {
	switch p {
	case gl.POINTS:
		return "Points"
	case gl.LINE_STRIP:
		return "LineStrip"
	case gl.LINE_LOOP:
		return "LineLoop"
	case gl.LINES:
		return "Lines"
	case gl.TRIANGLE_STRIP:
		return "TriangleStrip"
	case gl.TRIANGLE_FAN:
		return "TriangleFan"
	case gl.TRIANGLES:
		return "Triangles"
	}
	return fmt.Sprintf("PrimitiveType(0x%x)", gl.Enum(p))
}

func (b BufferLayout) String() string {
	switch b {
	case InterleavedBuffer:
		return "InterleavedBuffer"
	case CompactBuffer:
		return "CompactBuffer"
	}
	return fmt.Sprintf("BufferLayout(%d)", int(b))
}
