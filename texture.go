package nora

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png"
	"os"

	"github.com/maja42/nora/assert"
	"github.com/sirupsen/logrus"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/maja42/gl"
)

// TODO: fix exported types.

// TextureKey is used to connect textures with materials.
type TextureKey string

// TextureDefinition contains all necessary information for loading and configuring a texture
type TextureDefinition struct {
	Path string
	//ForbidReload bool // If true, the texture must not be reloaded from the filesystem, because other resources refer to it/depend on it
	Properties TextureProperties
}

// TextureProperties specifies properties of a 2D texture.
// (cube maps and 3D textures are currently not supported)
type TextureProperties struct {
	MinFilter gl.Enum
	// Texture magnification filter
	//   gl.LINEAR, gl.NEAREST (additional options might be available depending on the used OpenGL version)
	MagFilter gl.Enum
	// Wrapping function for texture coordinate s
	//   gl.REPEAT, gl.CLAMP_TO_EDGE, gl.MIRRORED_REPEAT
	WrapS gl.Enum
	// Wrapping function for texture coordinate t
	//   gl.REPEAT, gl.CLAMP_TO_EDGE, gl.MIRRORED_REPEAT
	WrapT gl.Enum
}

// Texture represents a GPU texture object for rendering
type texture struct {
	tex  gl.Texture // Note: golang textures have their origin in the top-left corner
	size mgl32.Vec2
}

// NewTexture creates a new texture object on the GPU.
// The object can be reused by loading different textures.
// Needs to be destroyed afterwards to free GPU resources.
// Note: The usability of texture objects is limited, because they can be reloaded at any time. Use 'TextureKey's instead!
func newTexture() *texture {
	return &texture{
		tex:  gl.CreateTexture(),
		size: mgl32.Vec2{0, 0},
	}
}

func (t *texture) String() string {
	return fmt.Sprintf("Texture(%d)", t.tex.Value)
}

func (t *texture) Load(path string, properties TextureProperties) error {
	logrus.Infof("Loading texture %q into %s...", path, t)

	imgFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open texture file %q: %v", path, err)
	}
	img, format, err := image.Decode(imgFile)
	imgFile.Close()
	if err != nil {
		return fmt.Errorf("decode texture file %q: %v", path, err)
	}

	// The image.Image interface does not provide access to the raw texel data of the texture.
	// TODO: Check if "img" is a well-known type (RGB, RGBA, Alpha, ...) and avoid the copy + format conversion
	// 		 If the type is not well-known, either don't support it, access the "Pix"-field via reflection (if available), or make a to-RGBA copy.
	//		 Right now, it's highly inefficient to create a copy, because we delete the data anyways after uploading it to the GPU

	rgba := image.NewRGBA(img.Bounds())
	bounds := rgba.Bounds()
	if bounds.Min.X != 0 || bounds.Min.Y != 0 {
		return fmt.Errorf("texture bounds of %s don't start at zero", t)
	}

	logrus.Debugf("Texture format of %s is %q. Size: %dx%d", t, format, bounds.Max.X, bounds.Max.Y)

	draw.Draw(rgba, bounds, img, image.Point{0, 0}, draw.Src) // Copy / convert image data

	t.size = mgl32.Vec2{float32(bounds.Max.X), float32(bounds.Max.Y)}

	// TODO: Not sure if I need to set an active texture...
	gl.BindTexture(gl.TEXTURE_2D, t.tex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, bounds.Max.X, bounds.Max.Y, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, int(properties.MagFilter))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, int(properties.MinFilter))

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, int(properties.WrapS))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, int(properties.WrapT))

	gl.GenerateMipmap(gl.TEXTURE_2D)

	assert.NoGLError("load %s", t)
	return nil
}

func (t *texture) Destroy() {
	logrus.Debugf("Destroying %s", t)
	gl.DeleteTexture(t.tex)
}

// Size returns the dimensions of the underlying texture
// If no texture is loaded, (0,0) is returned.
func (t *texture) Size() mgl32.Vec2 {
	return t.size
}
