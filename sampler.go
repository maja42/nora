package nora

import (
	"sync"

	"github.com/maja42/gl"
	"github.com/maja42/nora/assert"
	"github.com/sirupsen/logrus"
)

type texBinding struct {
	texture texID
	sampler int
}

// samplerManager is responsible for binding and unbinding textures to texture targets (=samplers).
// It tries to minimize the number of binding changes.
type samplerManager struct {
	m          sync.Mutex
	maxSampler int

	activeBindings []bool
	texBinding     map[TextureKey]texBinding

	textureStore *TextureStore
}

func newSamplerManager(textureStore *TextureStore) samplerManager {
	maxSampler := gl.GetInteger(gl.MAX_TEXTURE_IMAGE_UNITS)
	t := samplerManager{
		maxSampler:     maxSampler,
		activeBindings: make([]bool, maxSampler),
		texBinding:     make(map[TextureKey]texBinding, maxSampler),

		textureStore: textureStore,
	}
	return t
}

func (t *samplerManager) unusedSampler() int {
	for tex, act := range t.activeBindings {
		if !act {
			return tex
		}
	}
	assert.Fail("Unable to bind texture. All %d samplers are in use.", t.maxSampler)
	return -1
}

func (t *samplerManager) bind(samplerLoc gl.Uniform, textureKey TextureKey) {
	texID, texture := t.textureStore.resolve(textureKey)
	if texture == nil { // unknown / not-loaded texture
		gl.Uniform1i(samplerLoc, 0) // unbind anything
		assert.Fail("Texture %q is not loaded", textureKey)
		return
	}

	t.m.Lock()
	defer t.m.Unlock()

	binding, ok := t.texBinding[textureKey]
	if !ok { // not bound yet
		sampler := t.unusedSampler()
		if sampler < 0 {
			gl.Uniform1i(samplerLoc, 0) // unbind anything
			return
		}

		t.activeBindings[sampler] = true
		t.texBinding[textureKey] = texBinding{
			texture: texID,
			sampler: sampler,
		}

		logrus.Debugf("Binding texture %q (%s) to samplerManager %d", textureKey, texture, sampler)

		gl.ActiveTexture(gl.Enum(gl.TEXTURE0 + sampler))
		gl.BindTexture(gl.TEXTURE_2D, texture.tex)
	} else if binding.texture != texID { // The texture behind the textureKey was reloaded
		gl.ActiveTexture(gl.Enum(gl.TEXTURE0 + binding.sampler))
		gl.BindTexture(gl.TEXTURE_2D, texture.tex)
		binding.texture = texID
	}
	gl.Uniform1i(samplerLoc, binding.sampler)
}

func (t *samplerManager) unbind(textureKey TextureKey) {
	t.m.Lock()
	defer t.m.Unlock()

	binding, ok := t.texBinding[textureKey]
	if !assert.True(ok, "Texture %q is not bound to any samplerManager", textureKey) {
		return
	}

	gl.ActiveTexture(gl.Enum(gl.TEXTURE0 + binding.sampler))
	gl.BindTexture(gl.TEXTURE_2D, gl.Texture{Value: 0})

	delete(t.texBinding, textureKey)
	t.activeBindings[binding.sampler] = false
}

//func (t *samplerManager) configureTarget(fn func()) {
//	t.m.Lock()
//	defer t.m.Unlock()
//
//	texTarget := t.unusedSampler()
//	if texTarget < 0 {
//		return
//	}
//
//	logrus.Debugf("Configuring texture on samplerManager %d", textureKey, texture, texTarget)
//	gl.ActiveTexture(gl.Enum(gl.TEXTURE0 + texTarget))
//	fn()
//	gl.BindTexture(gl.TEXTURE_2D, gl.Texture{Value: 0})
//}
