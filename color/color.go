package color

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"math"

	"github.com/maja42/vmath"
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

// Grayscale turns the color into grayscale.
func (c Color) Grayscale() Color {
	_, _, l := c.HSL()
	return HSL(0, 0, l)
}

// WithBrightness returns the color with its luminosity set to a specific level (0-1).
func (c Color) WithBrightness(luminosity float32) Color {
	h, s, _ := c.HSL()
	return HSL(h, s, luminosity)
}

// HSL converts the RGB color into hue, saturation, luminosity (0-1).
func (c Color) HSL() (float32, float32, float32) {
	// Based on: https://github.com/gerow/go-color

	max := vmath.Max(vmath.Max(c.R, c.G), c.B)
	min := vmath.Min(vmath.Min(c.R, c.G), c.B)

	// luminosity is the average of the max and min rgb color intensities.
	luminosity := (max + min) / 2

	// saturation
	delta := max - min
	if delta == 0 { // gray
		return 0, 0, luminosity
	}

	var saturation float32
	if luminosity < 0.5 {
		saturation = delta / (max + min)
	} else {
		saturation = delta / (2 - max - min)
	}

	// hue
	var hue float32
	r2 := (((max - c.R) / 6) + (delta / 2)) / delta
	g2 := (((max - c.G) / 6) + (delta / 2)) / delta
	b2 := (((max - c.B) / 6) + (delta / 2)) / delta
	switch {
	case c.R == max:
		hue = b2 - g2
	case c.G == max:
		hue = (1.0 / 3.0) + r2 - b2
	case c.B == max:
		hue = (2.0 / 3.0) + g2 - r2
	}

	// fix wraparounds
	switch {
	case hue < 0:
		hue += 1
	case hue > 1:
		hue -= 1
	}
	return hue, saturation, luminosity
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

// MustHex returns a color from a hex-string, according to Hex().
// If the input string is invalid, the function panics.
func MustHex(s string) Color {
	c, err := Hex(s)
	if err != nil {
		panic(err)
	}
	return c
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

// HSL returns a color based on hue, saturation, luminosity with 100% opacity
func HSL(hue, saturation, luminosity float32) Color {
	// Based on: https://github.com/gerow/go-color

	if saturation == 0 { // gray
		return Color{luminosity, luminosity, luminosity, 1}
	}

	var v1, v2 float32
	if luminosity < 0.5 {
		v2 = luminosity * (1 + saturation)
	} else {
		v2 = (luminosity + saturation) - (saturation * luminosity)
	}

	v1 = 2*luminosity - v2

	r := hueToRGB(v1, v2, hue+(1.0/3.0))
	g := hueToRGB(v1, v2, hue)
	b := hueToRGB(v1, v2, hue-(1.0/3.0))

	return Color{r, g, b, 1}
}

func hueToRGB(v1, v2, h float32) float32 {
	if h < 0 {
		h += 1
	}
	if h > 1 {
		h -= 1
	}
	switch {
	case 6*h < 1:
		return (v1 + (v2-v1)*6*h)
	case 2*h < 1:
		return v2
	case 3*h < 2:
		return v1 + (v2-v1)*((2.0/3.0)-h)*6
	}
	return v1
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
