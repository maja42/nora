package rendering

import (
	"golang.org/x/mobile/gl"
)

// Texture unit that is used when loading new textures and configuring them.
// It is ignored and skipped by the textureTargets.
//const configTexTarget gl.Enum = gl.TEXTURE0 // TODO: CHECK if texture properties are defined per-texture or per-texture-target. I think the later --> Remove code in texture!

// resolveTexFunc turns a texture key into the currently loaded texture and it's ID.
// The ID can be used to determine if the texture changes (was reloaded) since it was last resolved.
type resolveTexFunc func(textureKey TextureKey) (texID, *Texture)

// textureTargets is responsible for binding and unbinding textures to texture targets.
// It tries to minimize the number of binding changes.
type textureTargets struct {
	ctx     *Context
	resolve resolveTexFunc

	maxTexTargets    int
	unusedTexTargets []int
}

func NewTextureTargets(ctx *Context, resolver resolveTexFunc) *textureTargets {
	t := &textureTargets{
		ctx:           ctx,
		resolve:       resolver,
		maxTexTargets: ctx.GetInteger(gl.MAX_TEXTURE_IMAGE_UNITS),
	}
	t.unusedTexTargets = make([]int, t.maxTexTargets)
	for i := 0; i < t.maxTexTargets; i++ {
		// reverse order: We take unused targets from the right, and we want to use lower texture-targets first
		t.unusedTexTargets[i] = t.maxTexTargets - 1 - i
	}
	return t
}

func (t *textureTargets) Bind(samplerLoc gl.Uniform, textureKey TextureKey) {
	texID, texture := t.resolve(textureKey)
	if texture == nil { // unknown / not-loaded texture
		t.ctx.Uniform1i(samplerLoc, 0) // unbind anything
		assertFail("Texture %q is not loaded", textureKey)
		return
	}
	_ = texID // TODO: actually use the ID to re-bind the texture if it was modified!

	if texture.target < 0 {
		texTarget := t.unusedTextureTarget()
		logger.Debugf("Binding texture %q to target %d", textureKey, texTarget)
		texture.target = texTarget
		t.ctx.ActiveTexture(gl.Enum(gl.TEXTURE0 + texTarget))
		t.ctx.BindTexture(gl.TEXTURE_2D, texture.tex)
	}

	t.ctx.Uniform1i(samplerLoc, texture.target)
}

func (t *textureTargets) Unbind() {

}

func (t *textureTargets) unusedTextureTarget() int {
	idx := len(t.unusedTexTargets) - 1
	if idx < 0 {
		assertFail("Unable to find texture target. All %d texture targets are in use.", t.maxTexTargets)
		return 0
	}

	unused := t.unusedTexTargets[idx]
	t.unusedTexTargets = t.unusedTexTargets[:idx]
	return unused
}
