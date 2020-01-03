package nora

import (
	"fmt"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/maja42/gl"
	"github.com/maja42/nora/font"
)

type Font struct {
	font.Font
	texKey  TextureKey
	texSize mgl32.Vec2
	//tex *Texture
}

// LoadFont loads a font description and the corresponding texture.
// The texture object is loaded on the GPU.
// Needs to be destroyed afterwards to free GPU resources.
func LoadFont(path, file string) (*Font, error) {
	logrus.Infof("Loading font %q...", file)
	texKey := TextureKey("font:" + file)

	xmlPath := filepath.Join(path, file)
	desc, err := font.Load(xmlPath)
	if err != nil {
		return nil, fmt.Errorf("load font description: %w", err)
	}
	logrus.Infof("Font %s (%s): size %d, %d characters", desc.Family, desc.Style, desc.Size, len(desc.Chars))

	texPath := filepath.Join(path, desc.Texture)

	// Regarding texture (hot-)reloading:
	//	  We don't support font hot-reloading, meaning that the xml description
	//    is not automatically updated. As a consequence, the xml must match the texture
	//    during application startup.
	//	  If the texture is reloaded, the size and individual characters are allowed to be modified,
	//    as long as the relative location and size of each individual rune stays unmodified.

	size, err := nora.Textures.Load(texKey, &TextureDefinition{
		Path: texPath,
		//ForbidReload: true,
		Properties: TextureProperties{
			MinFilter: gl.LINEAR,
			MagFilter: gl.LINEAR,
			WrapS:     gl.REPEAT,
			WrapT:     gl.REPEAT,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("load texture: %s", err)
	}

	return &Font{
		Font:    desc,
		texKey:  texKey,
		texSize: size,
		//tex:  tex,
	}, nil
}

func (f *Font) Destroy() {
	logrus.Debugf("Destroying %s", f)
	nora.Textures.Unload(f.texKey)
}

func (f *Font) String() string {
	return fmt.Sprintf("Font(%s/%s/%d)", f.Family, f.Style, f.Size)
}

func (f *Font) Char(r rune) (font.Char, bool) {
	char, ok := f.Chars[r]
	return char, ok
}

func (f *Font) TexCoord(r rune) (mgl32.Vec2, mgl32.Vec2) {
	char := f.Chars[r]
	size := f.texSize

	// Texture coordinates: [0, 1], starting on the bootom left (y=inverted)
	var tl, br mgl32.Vec2
	tl[0] = float32(char.Pos[0]) / size[0]
	tl[1] = 1 - (float32(char.Pos[1]) / size[1])
	br[0] = tl[0] + float32(char.Size[0])/size[0]
	br[1] = tl[1] - float32(char.Size[1])/size[1]
	return tl, br
}

func (f *Font) TextureKey() TextureKey {
	return f.texKey
}
