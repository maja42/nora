package rendering

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-gl/mathgl/mgl32"

	tassert "github.com/stretchr/testify/assert"
)

func TestTransformable_GetAxisRotation(t *testing.T) {
	tests := []struct {
		radians float32
		axis    mgl32.Vec3
	}{
		{1.5, [3]float32{1, 0, 0}},
		{1.5, [3]float32{0, 1, 0}},
		{1.5, [3]float32{0, 0, 1}},

		{0.1, [3]float32{1, 0, 0}},
		{2.6, [3]float32{1, 0, 0}},
		{216, [3]float32{1, 0, 0}},

		{1.5, [3]float32{1, 1, 0}},
	}

	for idx, tt := range tests {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			tf := NewTransform()

			tf.SetAxisRotation(tt.radians, tt.axis)
			axis, rad := tf.GetAxisRotation()

			normalized := float32(math.Mod(float64(tt.radians), math.Pi*2))
			tassert.InDelta(t, normalized, rad, 0.01)

			tassert.InDelta(t, tt.axis[0], axis[0], 0.01)
			tassert.InDelta(t, tt.axis[1], axis[1], 0.01)
			tassert.InDelta(t, tt.axis[2], axis[2], 0.01)
		})
	}
}
