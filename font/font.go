package font

import (
	"github.com/maja42/nora/math"
)

type Font struct {
	Family    string
	Style     string
	Size      int
	Monospace bool // true if all characters have the same width
	Chars     map[rune]Char

	Ascender  int
	Descender int
	Height    int

	Texture string
}

type Char struct {
	Width  int
	Offset math.Vec2i
	Pos    math.Vec2i
	Size   math.Vec2i
}

// AvgWidth returns the average width across all characters
func (f *Font) AvgWidth() float32 {
	if len(f.Chars) == 0 {
		return 0
	}
	width := 0
	for _, c := range f.Chars {
		width += c.Width
	}
	return float32(width) / float32(len(f.Chars))
}

// Runes returns a list with all runes in this font
func (f *Font) Runes() []rune {
	runes := make([]rune, 0, len(f.Chars))
	for r, _ := range f.Chars {
		runes = append(runes, r)
	}
	return runes
}
