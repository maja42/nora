package nora

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/maja42/nora/assert"

	"github.com/sirupsen/logrus"

	"github.com/go-gl/mathgl/mgl32"

	"go.uber.org/atomic"

	"github.com/maja42/nora/hotreload"
)

type loadedTexture struct {
	id      texID
	texture *texture

	// for texture hot-reloading:
	//forbidReload        bool // if true, this texture must not be reloaded, because other resources depend on it/refer to it
	intermediateTexture *texture
	definition          *TextureDefinition
}

// texID uniquely identifies a loaded texture
// If a texture with the same textureKey is loaded/unloaded/reloaded, the texID always changes
type texID struct {
	id         uint32 // first valid texture has id 1
	generation uint32
}

var texIDSeq atomic.Uint32

func newTexID() texID {
	return texID{
		id:         texIDSeq.Inc(),
		generation: 0,
	}
}

// TextureStore manages textures.
type TextureStore struct {
	m                  sync.RWMutex
	cancelHotReloading context.CancelFunc

	textures  map[TextureKey]loadedTexture
	fsWatcher *hotreload.Watcher
}

// newTextureStore creates a new, empty store for textures.
func newTextureStore() TextureStore {
	return TextureStore{
		textures:  make(map[TextureKey]loadedTexture),
		fsWatcher: hotreload.NewWatcher(),
	}
}

// Destroy stops texture hot-reloading (if started) and unloads all textures.
func (s *TextureStore) Destroy() {
	s.cancelHotReloading()
	s.UnloadAll()
}

// StartHotReloading monitors the filesystem and reloads textures if their source files are modified.
// Runs until the given context is canceled.
func (s *TextureStore) StartHotReloading(ctx context.Context) error {
	s.m.Lock()
	if s.cancelHotReloading != nil {
		s.m.Unlock()
		return errors.New("hot-reloading is already enabled")
	}
	ctx, s.cancelHotReloading = context.WithCancel(ctx)
	s.m.Unlock()

	logrus.Info("Starting texture hot-reloading")

	err := s.fsWatcher.Watch(ctx, func(key interface{}) {
		texKey := key.(TextureKey)
		_, err := s.Reload(texKey)
		if assert.True(err == nil, "Hot-Reload of %q failed: %s", texKey, err) {
			logrus.Debugf("Hot-reload of %q succeeded", texKey)
		}
	})
	if err != nil {
		logrus.Warn("Failed to perform texture hot-reloading: %s", err)
	}
	return nil
}

func (s *TextureStore) Load(key TextureKey, def *TextureDefinition) (mgl32.Vec2, error) {
	def.Path = filepath.Clean(def.Path)

	if loadedTexture, ok := s.textures[key]; ok {
		//if loadedTexture.forbidReload {
		//	return fmt.Errorf("texture %q is already loaded and cannot be replaced", key)
		//}
		logrus.Debugf("Texture %q is already loaded. Replacing it...", key)
		s.m.Lock()
		defer s.m.Unlock()
		loadedTexture.definition = def
		return s.reloadTexture(key)
	}

	id := newTexID()
	tex := newTexture()
	if err := tex.Load(def.Path, def.Properties); err != nil {
		tex.Destroy()
		return mgl32.Vec2{}, err
	}

	s.textures[key] = loadedTexture{
		id:      id,
		texture: tex,
		//forbidReload: def.ForbidReload,
		definition: def,
	}

	s.fsWatcher.Add(def.Path, key)
	return tex.size, nil
}

// Reload hot-reloads the given texture from the filesystem once.
func (s *TextureStore) Reload(key TextureKey) (mgl32.Vec2, error) {
	s.m.Lock()
	defer s.m.Unlock()
	return s.reloadTexture(key)
}

func (s *TextureStore) reloadTexture(key TextureKey) (mgl32.Vec2, error) {
	loadedTexture, ok := s.textures[key]
	if !ok {
		return mgl32.Vec2{}, fmt.Errorf("texture %q is not loaded", key)
	}
	if loadedTexture.intermediateTexture == nil {
		loadedTexture.intermediateTexture = newTexture()
	}

	err := loadedTexture.intermediateTexture.Load(loadedTexture.definition.Path, loadedTexture.definition.Properties)
	if err != nil {
		return mgl32.Vec2{}, fmt.Errorf("hot-reload texture %q: %s", key, err)
	}

	logrus.Infof("Replacing texture %q...", key)
	loadedTexture.texture, loadedTexture.intermediateTexture = loadedTexture.intermediateTexture, loadedTexture.texture
	loadedTexture.id.generation = loadedTexture.id.generation + 1
	s.textures[key] = loadedTexture
	return loadedTexture.texture.Size(), nil
}

// UnloadAll unloads all textures
func (s *TextureStore) UnloadAll() {
	s.m.Lock()
	defer s.m.Unlock()
	for key := range s.textures {
		s.unload(key)
	}
}

// Unload unloads a single texture
func (s *TextureStore) Unload(key TextureKey) {
	s.m.Lock()
	defer s.m.Unlock()
	s.unload(key)
}

func (s *TextureStore) unload(key TextureKey) {
	loadedTexture, ok := s.textures[key]
	if !ok {
		logrus.Warnf("Unload: Texture %q is not loaded", key)
		return
	}

	err := s.fsWatcher.Remove(loadedTexture.definition.Path, key)
	iAssertTrue(err == nil, "Failed to un-watch texture: %s", err)

	loadedTexture.texture.Destroy()
	if loadedTexture.intermediateTexture != nil {
		loadedTexture.intermediateTexture.Destroy()
	}
	delete(s.textures, key)
}

// resolve returns the (loaded) texture and its ID.
// Returns nil if the texture was not loaded yet.
func (s *TextureStore) resolve(key TextureKey) (texID, *texture) {
	s.m.RLock()
	defer s.m.RUnlock()
	texture, ok := s.textures[key]
	if !ok {
		return texID{}, nil
	}
	return texture.id, texture.texture
}

// Definition returns the texture definition with the given key.
// If the texture is not loaded, an empty definition is returned.
func (s *TextureStore) Definition(key TextureKey) TextureDefinition {
	return *s.textures[key].definition
}
