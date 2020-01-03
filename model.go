package nora

import (
	"sync/atomic"
	"time"
)

// Updateable can be implemented by models
type Updateable interface {
	Update(elapsed time.Duration)
}

type modelID struct {
	uint64
}

type AttachableModel struct {
	id    modelID // atomic access only
	jobID JobID   // atomic access only
}

func (a *AttachableModel) attach(id modelID) bool {
	return atomic.CompareAndSwapUint64(&a.id.uint64, 0, id.uint64)
}

func (a *AttachableModel) detach() modelID {
	return modelID{atomic.SwapUint64(&a.id.uint64, 0)}
}

func (a *AttachableModel) setJobID(jobID JobID) {
	oldID := atomic.SwapUint64(&a.jobID.uint64, jobID.uint64)
	iAssertTrue(jobID.uint64 == 0 || oldID == 0, "There was already a job ID")
}

func (a *AttachableModel) getJobID() JobID {
	return a.jobID
}

// Model represents a drawable object in the world.
// 	- Responsible for applying transformations before drawing
// 	- Responsible for invoking draw() on all underlying meshes.
// Models can optionally implement Updateable.
type Model interface {
	// all models need to embed 'AttachableModel'
	attach(id modelID) bool
	detach() modelID
	setJobID(jobID JobID)
	getJobID() JobID

	Draw(*RenderState)
	Destroy()
}
