package nora

import (
	"sync"
	"time"

	"go.uber.org/atomic"
)

// JobID uniquely represents an update job.
type JobID struct {
	uint64
}

// Update jobs are called right before rendering
type UpdateJob func(elapsed time.Duration)

// JobSystem manages update jobs that are called before rendering, typically each frame.
type JobSystem struct {
	m          sync.Mutex
	idSeq      atomic.Uint64
	updateJobs map[JobID]UpdateJob
}

// newJobSystem returns a new, empty store for jobs.
func newJobSystem() JobSystem {
	return JobSystem{
		updateJobs: make(map[JobID]UpdateJob),
	}
}

// Add a new update job
func (s *JobSystem) Add(job UpdateJob) JobID {
	id := JobID{s.idSeq.Inc()}

	s.m.Lock()
	defer s.m.Unlock()

	s.updateJobs[id] = job
	return id
}

// AddTimed adds a new update job that is executed with the given interval.
// If the time interval is reached multiple times within a single frame, the job is only executed once.
func (s *JobSystem) AddTimed(interval time.Duration, job UpdateJob) JobID {
	elapsed := time.Duration(0)
	return s.Add(func(duration time.Duration) {
		elapsed += duration
		if elapsed < interval {
			return
		}
		job(elapsed)
		elapsed = 0
	})
}

// AddFixed adds a new update job that is executed with the given interval.
// If the time interval is reached multiple times within a single frame, the job is executed as often as needed.
// The job always receives the interval instead of the elapsed time.
func (s *JobSystem) AddFixed(interval time.Duration, job UpdateJob) JobID {
	elapsed := time.Duration(0)
	return s.Add(func(duration time.Duration) {
		elapsed += duration
		for ; elapsed >= interval; elapsed -= interval {
			job(interval)
		}
	})
}

// Once adds a new update job that is executed in the next frame and removed automatically afterwards
func (s *JobSystem) Once(job UpdateJob) JobID {
	id := JobID{s.idSeq.Inc()}

	s.m.Lock()
	defer s.m.Unlock()
	s.updateJobs[id] = func(elapsed time.Duration) {
		job(elapsed)
		s.Remove(id)
	}
	return id
}

// Remove stops/removes an update job
func (s *JobSystem) Remove(jobID JobID) {
	s.m.Lock()
	defer s.m.Unlock()

	delete(s.updateJobs, jobID)
}

// RemoveAll stops/removes all update jobs
func (s *JobSystem) RemoveAll() {
	s.m.Lock()
	defer s.m.Unlock()

	s.updateJobs = make(map[JobID]UpdateJob)
}

// Executes all jobs with the given arguments.
func (s *JobSystem) run(elapsed time.Duration) {
	// No lock needed (iteration is safe).
	// If jobs are added/removed from within jobs, it's unspecified if the function will be executed in this frame, or the next one.

	// TODO: run in parallel
	// We don't need any ordering. If jobs should be ordered, they are combined into a single job and spawn go-routines.
	// Alternatively, I can add a function so that jobs can spawn new jobs for the current frame only
	for _, job := range s.updateJobs {
		job(elapsed)
	}
}
