package nora

import (
	"fmt"
	"sync"

	"github.com/maja42/nora/assert"

	"go.uber.org/atomic"
)

// Scene contains all models to render.
type Scene struct {
	m          sync.RWMutex
	modelIDSeq atomic.Uint64
	models     map[modelID]Model

	jobs *JobSystem
}

func newScene(jobs *JobSystem) Scene {
	return Scene{
		models: make(map[modelID]Model),
		jobs:   jobs,
	}
}

func (s *Scene) Destroy() {
	s.DetachAndDestroyAll()
}

// Attach a model.
// If the model implements Updateable, the update job is added automatically.
// Fails if the model is already attached to a scene.
func (s *Scene) Attach(model Model) error {
	id := modelID{s.modelIDSeq.Inc()}

	if !model.attach(id) {
		return fmt.Errorf("already attached")
	}

	s.m.Lock()
	defer s.m.Unlock()
	s.models[id] = model

	up, ok := model.(Updateable)
	if ok {
		model.setJobID(s.jobs.Add(up.Update))
	}
	return nil
}

// Detach a model.
// Can be attached again later.
func (s *Scene) Detach(model Model) {
	id := model.detach()
	assert.True(id.uint64 != 0, "Model was not attached to any scene")

	s.m.Lock()
	defer s.m.Unlock()
	delete(s.models, id)

	jobID := model.getJobID()
	if jobID.uint64 != 0 {
		s.jobs.Remove(jobID)
		model.setJobID(JobID{0})
	}
}

// DetachAll detaches all models.
func (s *Scene) DetachAll() {
	s.m.Lock()
	defer s.m.Unlock()
	s.jobs.m.Lock()
	defer s.jobs.m.Unlock()

	s.detachAll()
	s.models = make(map[modelID]Model)
}

// Detaches a model and calls its destroy function.
func (s *Scene) DetachAndDestroy(model Model) {
	s.Detach(model)
	model.Destroy()
}

// Detaches all models and calls their destroy functions.
func (s *Scene) DetachAndDestroyAll() {
	s.m.Lock()
	defer s.m.Unlock()
	s.jobs.m.Lock()
	defer s.jobs.m.Unlock()

	s.detachAll()
	for _, model := range s.models {
		model.Destroy()
	}
	s.models = make(map[modelID]Model)
}

func (s *Scene) detachAll() {
	for id, model := range s.models {
		mid := model.detach()
		iAssertTrue(id == mid, "model contained a different attachment ID than the scene")

		jobID := model.getJobID()
		if jobID.uint64 != 0 {
			delete(s.jobs.updateJobs, jobID)
		}
	}
}

// Models returns a list with all attached models.
func (s *Scene) Models() []Model {
	s.m.RLock()
	defer s.m.RUnlock()

	modelList := make([]Model, 0, len(s.models))
	for _, model := range s.models {
		modelList = append(modelList, model)
	}
	return modelList
}

func (s *Scene) borrowModels() (map[modelID]Model, func()) {
	s.m.RLock()
	return s.models, s.m.RUnlock
}
