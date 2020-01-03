package rendering

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/maja42/logicat/rendering/hotreload"
	"go.uber.org/atomic"
)

type loadedShader struct {
	id      sProgID // id.generation is incremented every time the program is modified
	program *ShaderProgram

	// for shader hot-reloading:
	intermediateProgram *ShaderProgram
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

type loadedTexture struct {
	id      texID
	texture *Texture

	// for texture hot-reloading:
	//forbidReload        bool // ff true, this texture must not be reloaded, because other resources depend on it/refer to it
	intermediateTexture *Texture
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

type JobID struct {
	uint64
}

type UpdateJob func(elapsed time.Duration)

type SceneGraph struct {
	ctx *Context

	sProgMutex     sync.RWMutex
	shaderPrograms map[ShaderProgKey]loadedShader
	shaderWatch    *hotreload.Watcher

	texMutex     sync.RWMutex
	textures     map[TextureKey]loadedTexture
	textureWatch *hotreload.Watcher

	cancelHotReloading context.CancelFunc

	modelLock  sync.RWMutex
	modelIDSeq atomic.Uint64
	models     map[modelID]Model

	jobLock    sync.Mutex
	jobIDSeq   atomic.Uint64
	updateJobs map[JobID]UpdateJob
}

// NewSceneGraph creates a new, empty scene graph
func NewSceneGraph(ctx *Context) *SceneGraph {
	return &SceneGraph{
		ctx:            ctx,
		shaderPrograms: make(map[ShaderProgKey]loadedShader),
		shaderWatch:    hotreload.NewWatcher(),

		textures:     make(map[TextureKey]loadedTexture),
		textureWatch: hotreload.NewWatcher(),

		models:     make(map[modelID]Model),
		updateJobs: make(map[JobID]UpdateJob),
	}
}

func (s *SceneGraph) Destroy() {
	s.cancelHotReloading()

	s.RemoveAllUpdateJobs()
	s.DetachAndDestroyAll()

	s.UnloadAllTextures()
	s.UnloadAllShaderPrograms()
}

// LoadShaderPrograms loads multiple shader programs.
func (s *SceneGraph) LoadShaderPrograms(defs map[ShaderProgKey]ShaderProgramDefinition) error {
	logger.Infof("Preparing %d shader programs...", len(defs))

	for key, def := range defs {
		defCopy := def
		if err := s.LoadShaderProgram(key, &defCopy); err != nil {
			return err
		}
	}
	return nil
}

// LoadShaderProgram loads a single shader program.
// Replaces any existing shader.
func (s *SceneGraph) LoadShaderProgram(key ShaderProgKey, def *ShaderProgramDefinition) error {
	s.sProgMutex.Lock()
	defer s.sProgMutex.Unlock()

	def.VertexShaderPath = filepath.Clean(def.VertexShaderPath)
	def.FragmentShaderPath = filepath.Clean(def.FragmentShaderPath)

	if loadedShader, ok := s.shaderPrograms[key]; ok {
		logger.Debugf("Shader %q is already loaded. Replacing it...", key)
		loadedShader.definition = def
		return s.reloadShaderProgram(key)
	}

	id := newShaderProgID()
	program := NewShaderProgram(s.ctx)
	if err := program.Load(def); err != nil {
		program.Destroy()
		return err
	}

	s.shaderPrograms[key] = loadedShader{
		id:         id,
		program:    program,
		definition: def,
	}

	s.shaderWatch.Add(def.VertexShaderPath, key)
	s.shaderWatch.Add(def.FragmentShaderPath, key)
	return nil
}

func (s *SceneGraph) LoadTexture(key TextureKey, def *TextureDefinition) (mgl32.Vec2, error) {
	def.Path = filepath.Clean(def.Path)

	if loadedTexture, ok := s.textures[key]; ok {
		//if loadedTexture.forbidReload {
		//	return fmt.Errorf("texture %q is already loaded and cannot be replaced", key)
		//}
		logger.Debugf("Texture %q is already loaded. Replacing it...", key)
		loadedTexture.definition = def
		return s.reloadTexture(key)
	}

	id := newTexID()
	tex := NewTexture(s.ctx)
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

	s.textureWatch.Add(def.Path, key)
	return tex.size, nil
}

// StartHotReloading monitors the filesystem and reloads shader programs if their source files are modified.
// Runs until the given context is canceled.
func (s *SceneGraph) StartHotReloading(ctx context.Context) error {
	ctx, s.cancelHotReloading = context.WithCancel(ctx)
	logger.Info("Starting shader and texture hot-reloading")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := s.shaderWatch.Watch(ctx, func(key interface{}) {
			sProgKey := key.(ShaderProgKey)
			err := s.ReloadShaderProgram(sProgKey)
			if assert(err == nil, "Hot-Reload of %q failed: %s", sProgKey, err) {
				logger.Debugf("Hot-reload of %q succeeded", sProgKey)
			}
		})
		if err != nil {
			logger.Warn("Failed to start shart hot-reloading: %s", err)
		}
	}()

	go func() {
		defer wg.Done()
		err := s.textureWatch.Watch(ctx, func(key interface{}) {
			texKey := key.(TextureKey)
			_, err := s.ReloadTexture(texKey)
			if assert(err == nil, "Hot-Reload of %q failed: %s", texKey, err) {
				logger.Debugf("Hot-reload of %q succeeded", texKey)
			}
		})
		if err != nil {
			logger.Warn("Failed to start texture hot-reloading: %s", err)
		}
	}()

	wg.Wait()
	return nil
}

// ReloadShaderProgram hot-reloads the given shader from the filesystem once.
func (s *SceneGraph) ReloadShaderProgram(key ShaderProgKey) error {
	s.sProgMutex.Lock()
	defer s.sProgMutex.Unlock()
	return s.reloadShaderProgram(key)
}

func (s *SceneGraph) reloadShaderProgram(key ShaderProgKey) error {
	loadedShader, ok := s.shaderPrograms[key]
	if !ok {
		return fmt.Errorf("shader %q is not loaded", key)
	}
	if loadedShader.intermediateProgram == nil {
		loadedShader.intermediateProgram = NewShaderProgram(s.ctx)
	}

	err := loadedShader.intermediateProgram.Load(loadedShader.definition)
	if err != nil {
		return fmt.Errorf("hot-reload shader %q: %s", key, err)
	}

	logger.Infof("Replacing shader %q...", key)
	loadedShader.program, loadedShader.intermediateProgram = loadedShader.intermediateProgram, loadedShader.program
	loadedShader.id.generation = loadedShader.id.generation + 1
	s.shaderPrograms[key] = loadedShader
	return nil
}

// ReloadTexture hot-reloads the given texture from the filesystem once.
func (s *SceneGraph) ReloadTexture(key TextureKey) (mgl32.Vec2, error) {
	s.texMutex.Lock()
	defer s.texMutex.Unlock()
	return s.reloadTexture(key)
}

func (s *SceneGraph) reloadTexture(key TextureKey) (mgl32.Vec2, error) {
	loadedTexture, ok := s.textures[key]
	if !ok {
		return mgl32.Vec2{}, fmt.Errorf("texture %q is not loaded", key)
	}
	if loadedTexture.intermediateTexture == nil {
		loadedTexture.intermediateTexture = NewTexture(s.ctx)
	}

	err := loadedTexture.intermediateTexture.Load(loadedTexture.definition.Path, loadedTexture.definition.Properties)
	if err != nil {
		return mgl32.Vec2{}, fmt.Errorf("hot-reload texture %q: %s", key, err)
	}

	logger.Infof("Replacing texture %q...", key)
	loadedTexture.texture, loadedTexture.intermediateTexture = loadedTexture.intermediateTexture, loadedTexture.texture
	loadedTexture.id.generation = loadedTexture.id.generation + 1
	s.textures[key] = loadedTexture
	return loadedTexture.texture.Size(), nil
}

// UnloadAllShaderPrograms unloads all shader programs
func (s *SceneGraph) UnloadAllShaderPrograms() {
	s.sProgMutex.Lock()
	defer s.sProgMutex.Unlock()
	for key := range s.shaderPrograms {
		s.unloadShaderProgram(key)
	}
}

// UnloadAllTextures unloads all textures
func (s *SceneGraph) UnloadAllTextures() {
	s.texMutex.Lock()
	defer s.texMutex.Unlock()
	for key := range s.textures {
		s.unloadTexture(key)
	}
}

// UnloadShaderProgram unloads a single program
func (s *SceneGraph) UnloadShaderProgram(key ShaderProgKey) {
	s.sProgMutex.Lock()
	defer s.sProgMutex.Unlock()
	s.unloadShaderProgram(key)
}

func (s *SceneGraph) unloadShaderProgram(key ShaderProgKey) {
	loadedProgram, ok := s.shaderPrograms[key]
	if !ok {
		logger.Warnf("Unload: Shader %q is not loaded", key)
		return
	}

	err := s.shaderWatch.Remove(loadedProgram.definition.VertexShaderPath, key)
	iAssert(err == nil, "Failed to un-watch: %s", err)
	err = s.shaderWatch.Remove(loadedProgram.definition.FragmentShaderPath, key)
	iAssert(err == nil, "Failed to un-watch: %s", err)

	loadedProgram.program.Destroy()
	if loadedProgram.intermediateProgram != nil {
		loadedProgram.intermediateProgram.Destroy()
	}
	delete(s.shaderPrograms, key)
}

// UnloadTexture unloads a single texture
func (s *SceneGraph) UnloadTexture(key TextureKey) {
	s.texMutex.Lock()
	defer s.texMutex.Unlock()
	s.unloadTexture(key)
}

func (s *SceneGraph) unloadTexture(key TextureKey) {
	loadedTexture, ok := s.textures[key]
	if !ok {
		logger.Warnf("Unload: Texture %q is not loaded", key)
		return
	}

	err := s.textureWatch.Remove(loadedTexture.definition.Path, key)
	iAssert(err == nil, "Failed to un-watch: %s", err)

	loadedTexture.texture.Destroy()
	if loadedTexture.intermediateTexture != nil {
		loadedTexture.intermediateTexture.Destroy()
	}
	delete(s.textures, key)
}

func (s *SceneGraph) resolveTexture(key TextureKey) (texID, *Texture) {
	s.texMutex.RLock()
	defer s.texMutex.RUnlock()
	texture, ok := s.textures[key]
	if !ok {
		return texID{}, nil
	}
	return texture.id, texture.texture
}

// getShaderProgram returns the (loaded) shader program and its ID.
// Returns nil if the program was not loaded yet,
func (s *SceneGraph) getShaderProgram(key ShaderProgKey) (*ShaderProgram, sProgID) {
	s.sProgMutex.RLock()
	defer s.sProgMutex.RUnlock()

	loadedProg := s.shaderPrograms[key]
	return loadedProg.program, loadedProg.id
}

// Attach a model.
// If the model implements Updateable, the update job is added automatically
// Fails if the model is already attached to a SceneGraph.
func (s *SceneGraph) Attach(model Model) error {
	id := modelID{s.modelIDSeq.Inc()}

	if !model.attach(id) {
		return fmt.Errorf("already attached")
	}

	s.modelLock.Lock()
	defer s.modelLock.Unlock()
	s.models[id] = model

	up, ok := model.(Updateable)
	if ok {
		model.setJobID(s.AddUpdateJob(up.Update))
	}
	return nil
}

// Detach a model.
// Can be attached again later.
func (s *SceneGraph) Detach(model Model) {
	id := model.detach()
	assert(id.uint64 != 0, "Model was not attached to any SceneGraph")

	s.modelLock.Lock()
	defer s.modelLock.Unlock()
	delete(s.models, id)

	jobID := model.getJobID()
	if jobID.uint64 != 0 {
		s.RemoveUpdateJob(jobID)
		model.setJobID(JobID{0})
	}
}

// DetachAll detaches all models.
func (s *SceneGraph) DetachAll() {
	s.modelLock.Lock()
	defer s.modelLock.Unlock()
	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	s.detachAll()
	s.models = make(map[modelID]Model)
}

// Detaches a model and calls its destroy function.
func (s *SceneGraph) DetachAndDestroy(model Model) {
	s.Detach(model)
	model.Destroy()
}

// Detaches all models and calls their destroy functions.
func (s *SceneGraph) DetachAndDestroyAll() {
	s.modelLock.Lock()
	defer s.modelLock.Unlock()
	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	s.detachAll()
	for _, model := range s.models {
		model.Destroy()
	}
	s.models = make(map[modelID]Model)
}

func (s *SceneGraph) detachAll() {
	for id, model := range s.models {
		mid := model.detach()
		iAssert(id == mid, "model contained a different attachment ID than the scene graph") // cannot happen

		jobID := model.getJobID()
		if jobID.uint64 != 0 {
			delete(s.updateJobs, jobID)
			//s.RemoveUpdateJob(jobID)
		}
	}
}

// Models returns a list with all attached models.
func (s *SceneGraph) Models() []Model {
	s.modelLock.RLock()
	defer s.modelLock.RUnlock()

	modelList := make([]Model, 0, len(s.models))
	for _, model := range s.models {
		modelList = append(modelList, model)
	}
	return modelList
}

func (s *SceneGraph) borrowModels() map[modelID]Model {
	s.modelLock.RLock()
	return s.models
}

func (s *SceneGraph) returnModels() {
	s.modelLock.RUnlock()
}

// AddUpdateJob adds a new update job
func (s *SceneGraph) AddUpdateJob(job UpdateJob) JobID {
	id := JobID{s.jobIDSeq.Inc()}

	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	s.updateJobs[id] = job
	return id
}

// AddTimedUpdateJob adds a new update job that is executed with the given interval
func (s *SceneGraph) AddTimedUpdateJob(interval time.Duration, job UpdateJob) JobID {
	elapsed := time.Duration(0)
	timedJob := func(duration time.Duration) {
		elapsed += duration
		if elapsed < interval {
			return
		}
		job(elapsed)
		elapsed = 0
	}
	return s.AddUpdateJob(timedJob)
}

// RemoveUpdateJob stops/removes an update job
func (s *SceneGraph) RemoveUpdateJob(jobID JobID) {
	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	delete(s.updateJobs, jobID)
}

// RemoveAllUpdateJobs stops/removes all update jobs
func (s *SceneGraph) RemoveAllUpdateJobs() {
	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	s.updateJobs = make(map[JobID]UpdateJob)
}

// Executes all jobs with the given arguments.
func (s *SceneGraph) RunUpdateJobs(elapsed time.Duration) {
	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	// TODO: run in parallel
	for _, job := range s.updateJobs {
		job(elapsed)
	}
}
