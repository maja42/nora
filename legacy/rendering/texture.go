package rendering

import (
	"fmt"
	"image"
	"image/draw"
	"os"

	_ "image/png"

	"github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/gl"
)

// TextureKey is used to connect textures and materials/meshes.
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
type Texture struct {
	ctx    *Context
	tex    gl.Texture // Note: golang textures have their origin in the top-left corner
	target int        // GPU texture target where this texture is bound. -1 == no binding TODO: move to textureTargets?

	size mgl32.Vec2
}

// NewTexture creates a new texture object on the GPU.
// The object can be reused by loading different textures.
// Needs to be destroyed afterwards to free GPU resources.
// Note: The usability of texture objects is limited, because they can be reloaded at any time. Use 'TextureKey's instead!
func NewTexture(ctx *Context) *Texture {
	return &Texture{
		ctx:    ctx,
		tex:    ctx.CreateTexture(),
		target: -1,
		size:   mgl32.Vec2{0, 0},
	}
}

func (t *Texture) String() string {
	return fmt.Sprintf("Texture(%d)", t.tex.Value)
}

func (t *Texture) Load(path string, properties TextureProperties) error {
	logger.Infof("Loading texture %q into %s...", path, t)

	imgFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open texture file %q: %v", path, err)
	}
	img, format, err := image.Decode(imgFile)
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

	logger.Debugf("Texture format of %s is %q. Size: %dx%d", t, format, bounds.Max.X, bounds.Max.Y)

	draw.Draw(rgba, bounds, img, image.Point{0, 0}, draw.Src) // Copy / convert image data

	t.size = mgl32.Vec2{float32(bounds.Max.X), float32(bounds.Max.Y)}

	//t.ctx.scene.texTargeter.Bind()

	ctx := t.ctx
	//ctx.Lock()
	//defer ctx.Unlock()
	//ctx.ActiveTexture(configTexTarget) // TODO: this is not ideal, and not synchronized!
	ctx.BindTexture(gl.TEXTURE_2D, t.tex)
	ctx.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, bounds.Max.X, bounds.Max.Y, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	ctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, int(properties.MagFilter))
	ctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, int(properties.MinFilter))

	ctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, int(properties.WrapS))
	ctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, int(properties.WrapT))

	ctx.GenerateMipmap(gl.TEXTURE_2D)

	iAssertNoGLError(ctx, "load %s", t)
	return nil
}

func (t *Texture) Destroy() {
	logger.Debugf("Destroying %s", t)
	t.ctx.DeleteTexture(t.tex)
}

// Size returns the dimensions of the underlying texture
// If no texture is loaded, (0,0) is returned.
func (t *Texture) Size() mgl32.Vec2 {
	return t.size
}
