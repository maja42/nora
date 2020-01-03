package rendering

import (
	"encoding/binary"
	"math/rand"
	"testing"
	"unsafe"

	tassert "github.com/stretchr/testify/assert"
)

func Test_float32asBytes(t *testing.T) {
	for _, count := range []int{0, 1, 1024, 10000} {
		data := makeTestFloats(count)
		flen, fcap := len(data), cap(data)

		bytes1 := float32asBytesSafe(data)
		bytes2 := float32asBytes(data)

		tassert.Equal(t, flen, len(data))
		tassert.Equal(t, fcap, cap(data))

		tassert.Equal(t, flen*4, len(bytes2))
		tassert.Equal(t, flen*4, cap(bytes2))

		for i := 0; i < count*4; i++ {
			tassert.Equal(t, bytes1[i], bytes2[i])
		}
	}
}

func Test_uint16asBytes(t *testing.T) {
	for _, count := range []int{0, 1, 1024, 10000} {
		data := makeTestUint16s(count)
		flen, fcap := len(data), cap(data)

		bytes1 := uint16asBytesSafe(data)
		bytes2 := uint16asBytes(data)

		tassert.Equal(t, flen, len(data))
		tassert.Equal(t, fcap, cap(data))

		tassert.Equal(t, flen*2, len(bytes2))
		tassert.Equal(t, flen*2, cap(bytes2))

		for i := 0; i < count*2; i++ {
			tassert.Equal(t, bytes1[i], bytes2[i])
		}
	}
}

func BenchmarkConvertToBytes(b *testing.B) {
	floats := makeTestFloats(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = float32asBytesSafe(floats)
	}
}

func BenchmarkInterpretAsBytes(b *testing.B) {
	floats := makeTestFloats(100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = float32asBytes(floats)
	}
}

var nativeEndianness binary.ByteOrder

func init() {
	var i int32 = 0x01020304
	u := unsafe.Pointer(&i)
	pb := (*byte)(u)
	b := *pb
	if b == 0x04 {
		nativeEndianness = binary.LittleEndian
	} else {
		nativeEndianness = binary.BigEndian
	}
}

func Test_endianness(t *testing.T) {
	// float32asBytes depends on the platform's endianness
	tassert.Equal(t, binary.LittleEndian, nativeEndianness, "Native platform has different endianness than OpenGL")
}

func makeTestFloats(count int) []float32 {
	v := make([]float32, count)
	for i := 0; i < count; i++ {
		v[i] = rand.Float32()
	}
	return v
}

func makeTestUint16s(count int) []uint16 {
	v := make([]uint16, count)
	for i := 0; i < count; i++ {
		v[i] = uint16(rand.Uint32() & 0xFFFF)
	}
	return v
}
