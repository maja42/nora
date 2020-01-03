package font

import (
	"github.com/go-gl/mathgl/mgl32"
)

type Font struct {
	Family string
	Style  string
	Size   int
	Chars  map[rune]Char

	Ascender  int
	Descender int
	Height    int

	Texture string
}

type Char struct {
	Width  int
	Offset mgl32.Vec2
	Pos    mgl32.Vec2
	Size   mgl32.Vec2
}
