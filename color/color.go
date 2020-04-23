package color

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"math"
)

var fMaxUint16 = float32(math.MaxUint16)

// Color represents a normalized [0.0, 1.0] float32 RGBA color.
type Color struct {
	R, G, B, A float32
}

// RGBA implements the color.Color interface.
func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R * fMaxUint16)
	g = uint32(c.G * fMaxUint16)
	b = uint32(c.B * fMaxUint16)
	a = uint32(c.A * fMaxUint16)
	return
}

func colorModel(c color.Color) color.Color {
	if _, ok := c.(Color); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	return Color{
		R: float32(r) / fMaxUint16,
		G: float32(g) / fMaxUint16,
		B: float32(b) / fMaxUint16,
		A: float32(a) / fMaxUint16,
	}
}

// ColorModel represents the graphics color model (i.e. normalized 32-bit
// floating point values RGBA color).
var ColorModel = color.ModelFunc(colorModel)

var Transparent = Color{0.0, 0.0, 0.0, 0.0}
var Black = Color{0.0, 0.0, 0.0, 1.0}
var White = Color{1.0, 1.0, 1.0, 1.0}

var Red = Color{1.0, 0.0, 0.0, 1.0}
var Green = Color{0.0, 1.0, 0.0, 1.0}
var Blue = Color{0.0, 0.0, 1.0, 1.0}

var Yellow = Color{1.0, 1.0, 0.0, 1.0}
var Cyan = Color{0.0, 1.0, 1.0, 1.0}
var Magenta = Color{1.0, 0.0, 1.0, 1.0}

// Hex returns a color from a hex-string.
// The hex string can be upper- or lower case.
// Supports 6- and 8-character strings (alpha is optional), as well as 3- and 4-character short-strings
func Hex(s string) (Color, error) {
	orig := s
	if len(s) == 3 || len(s) == 4 {
		alpha := "FF"
		if len(s) == 4 {
			alpha = string(s[3]) + string(s[3])
		}
		s = string(s[0]) + string(s[0]) +
			string(s[1]) + string(s[1]) +
			string(s[2]) + string(s[2]) + alpha
	}
	if len(s) == 6 {
		s += "FF"
	}

	if len(s) != 8 {
		return Black, fmt.Errorf("invalid hex color %q", orig)
	}
	val, err := hex.DecodeString(s)
	if err != nil {
		return Black, fmt.Errorf("invalid hex color: %s", err)
	}
	return Color{
		R: float32(val[0]) / 255,
		G: float32(val[1]) / 255,
		B: float32(val[2]) / 255,
		A: float32(val[3]) / 255,
	}, nil
}

// RGB returns a color with 100% opacity
func RGB(r, g, b float32) Color {
	return Color{
		R: r,
		G: g,
		B: b,
		A: 1,
	}
}

// Gray returns a color with the given brightness and 100% opacity
func Gray(brightness float32) Color {
	return Color{
		R: brightness,
		G: brightness,
		B: brightness,
		A: 1,
	}
}
