package nora

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/maja42/nora/assert"
	"github.com/maja42/nora/hotreload"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

type loadedShader struct {
	id      sProgID // id.generation is incremented every time the program is modified
	program *shaderProgram

	// for shader hot-reloading:
	intermediateProgram *shaderProgram
	definition          *ShaderProgramDefinition
}

// sProgID uniquely identifies a loaded shader program
type sProgID struct {
	id         uint32
	generation uint32
}

var sProgIDSeq atomic.Uint32

func newShaderProgID() sProgID {
	return sProgID{
		id:         sProgIDSeq.Inc(),
		generation: 0,
	}
}

// ShaderProgKey is used to connect shader programs with materials.
type ShaderProgKey string

// ShaderStore manages shader programs.
type ShaderStore struct {
	m                  sync.RWMutex
	cancelHotReloading context.CancelFunc

	shaderPrograms map[ShaderProgKey]loadedShader
	fsWatcher      *hotreload.Watcher
}

// newShaderStore creates a new, empty store for shader programs.
func newShaderStore() ShaderStore {
	return ShaderStore{
		shaderPrograms: make(map[ShaderProgKey]loadedShader),
		fsWatcher:      hotreload.NewWatcher(),
	}
}

// Destroy stops shader hot-reloading (if started) and unloads all shaders.
func (s *ShaderStore) Destroy() {
	s.cancelHotReloading()
	s.UnloadAll()
}

// StartHotReloading monitors the filesystem and reloads shader programs if their source files are modified.
// Blocks until hot-reloading is stopped due to context cancellation or because of a shutdown (destroy).
func (s *ShaderStore) StartHotReloading(ctx context.Context) error {
	s.m.Lock()
	if s.cancelHotReloading != nil {
		s.m.Unlock()
		return errors.New("hot-reloading is already enabled")
	}
	ctx, s.cancelHotReloading = context.WithCancel(ctx)
	s.m.Unlock()

	logrus.Info("Starting shader hot-reloading")

	err := s.fsWatcher.Watch(ctx, func(key interface{}) {
		sProgKey := key.(ShaderProgKey)
		err := s.Reload(sProgKey)
		if assert.True(err == nil, "Hot-Reload of %q failed: %s", sProgKey, err) {
			logrus.Debugf("Hot-reload of %q succeeded", sProgKey)
		}
	})
	if err != nil {
		logrus.Warnf("Failed to perform shader hot-reloading: %s", err)
	}
	return nil
}

// LoadAll loads multiple shader programs.
func (s *ShaderStore) LoadAll(defs map[ShaderProgKey]ShaderProgramDefinition) error {
	logrus.Infof("Preparing %d shader programs...", len(defs))

	for key, def := range defs {
		defCopy := def
		if err := s.Load(key, &defCopy); err != nil {
			return err
		}
	}
	return nil
}

// Load loads a single shader program.
// Replaces any existing shader.
func (s *ShaderStore) Load(key ShaderProgKey, def *ShaderProgramDefinition) error {
	s.m.Lock()
	defer s.m.Unlock()

	def.VertexShaderPath = filepath.Clean(def.VertexShaderPath)
	def.FragmentShaderPath = filepath.Clean(def.FragmentShaderPath)

	if loadedShader, ok := s.shaderPrograms[key]; ok {
		logrus.Debugf("Shader %q is already loaded. Replacing it...", key)
		s.m.Lock()
		defer s.m.Unlock()
		loadedShader.definition = def
		return s.reload(key)
	}

	id := newShaderProgID()
	program := newShaderProgram()
	if err := program.Load(def); err != nil {
		program.Destroy()
		return err
	}

	s.shaderPrograms[key] = loadedShader{
		id:         id,
		program:    program,
		definition: def,
	}

	s.fsWatcher.Add(def.VertexShaderPath, key)
	s.fsWatcher.Add(def.FragmentShaderPath, key)
	return nil
}

// Reload hot-reloads the given shader from the filesystem once.
func (s *ShaderStore) Reload(key ShaderProgKey) error {
	s.m.Lock()
	defer s.m.Unlock()
	return s.reload(key)
}

func (s *ShaderStore) reload(key ShaderProgKey) error {
	loadedShader, ok := s.shaderPrograms[key]
	if !ok {
		return fmt.Errorf("shader %q is not loaded", key)
	}
	if loadedShader.intermediateProgram == nil {
		loadedShader.intermediateProgram = newShaderProgram()
	}

	err := loadedShader.intermediateProgram.Load(loadedShader.definition)
	if err != nil {
		return fmt.Errorf("hot-reload shader %q: %s", key, err)
	}

	logrus.Infof("Replacing shader %q...", key)
	loadedShader.program, loadedShader.intermediateProgram = loadedShader.intermediateProgram, loadedShader.program
	loadedShader.id.generation = loadedShader.id.generation + 1
	s.shaderPrograms[key] = loadedShader
	return nil
}

// UnloadAll unloads all shader programs
func (s *ShaderStore) UnloadAll() {
	s.m.Lock()
	defer s.m.Unlock()
	for key := range s.shaderPrograms {
		s.unload(key)
	}
}

// Unload unloads a single program
func (s *ShaderStore) Unload(key ShaderProgKey) {
	s.m.Lock()
	defer s.m.Unlock()
	s.unload(key)
}

func (s *ShaderStore) unload(key ShaderProgKey) {
	loadedProgram, ok := s.shaderPrograms[key]
	if !ok {
		logrus.Warnf("Unload: Shader %q is not loaded", key)
		return
	}

	err := s.fsWatcher.Remove(loadedProgram.definition.VertexShaderPath, key)
	iAssertTrue(err == nil, "Failed to un-watch shader: %s", err)
	err = s.fsWatcher.Remove(loadedProgram.definition.FragmentShaderPath, key)
	iAssertTrue(err == nil, "Failed to un-watch shader: %s", err)

	loadedProgram.program.Destroy()
	if loadedProgram.intermediateProgram != nil {
		loadedProgram.intermediateProgram.Destroy()
	}
	delete(s.shaderPrograms, key)
}

// resolve returns the (loaded) shader program and its ID.
// Returns nil if the program was not loaded yet.
func (s *ShaderStore) resolve(key ShaderProgKey) (*shaderProgram, sProgID) {
	s.m.RLock()
	defer s.m.RUnlock()

	loadedProg := s.shaderPrograms[key]
	return loadedProg.program, loadedProg.id
}

// Count returns the number of loaded shaders.
// Intermediate shaders needed for hot-reloading are not counted.
func (s *ShaderStore) Count() int {
	return len(s.shaderPrograms)
}
