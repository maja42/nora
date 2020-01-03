package nora

import (
	"time"
)

const fpsAVGFrameCount = 64 // must be a power of two

// FPSCounter tracks the frame number, elapsed time since the last frame and the average frame rate.
type FPSCounter struct {
	frame        uint64
	lastFrame    time.Time
	frameTimes   [fpsAVGFrameCount]float64
	frameTimeIdx int
	frameSum     float64
}

func NewFPSCounter() FPSCounter {
	return FPSCounter{
		frame:        0,
		lastFrame:    time.Now(),
		frameTimes:   [fpsAVGFrameCount]float64{},
		frameTimeIdx: 0,
		frameSum:     0,
	}
}

func (f *FPSCounter) NextFrame() (uint64, time.Duration, float32) {
	f.frame++

	now := time.Now()
	duration := now.Sub(f.lastFrame)

	f.frameSum -= f.frameTimes[f.frameTimeIdx]
	f.frameTimes[f.frameTimeIdx] = duration.Seconds()
	f.frameSum += f.frameTimes[f.frameTimeIdx]
	f.frameTimeIdx = (f.frameTimeIdx + 1) & (fpsAVGFrameCount - 1)
	f.lastFrame = now

	if f.frame >= fpsAVGFrameCount {
		return f.frame, duration, 1 / float32(f.frameSum/fpsAVGFrameCount)
	}
	return f.frame, duration, 0
}

func (f *FPSCounter) LastFrame() time.Time {
	return f.lastFrame
}
