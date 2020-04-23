package nora

import (
	"strings"
	"time"
)

func indent(s string) string {
	if len(s) == 0 {
		return s
	}
	return "    " + strings.Replace(s, "\n", "\n    ", -1)
}

// UpdateFunc receives the time elapsed since the last frame and performs some work.
type UpdateFunc func(elapsed time.Duration)

// TimedUpdate wraps the given function and only executes it with the given interval.
// If the time interval is reached multiple times within a single call (frame), the function is only executed once.
func TimedUpdate(interval time.Duration, job UpdateFunc) UpdateFunc {
	elapsed := time.Duration(0)
	return func(duration time.Duration) {
		elapsed += duration
		if elapsed < interval {
			return
		}
		job(elapsed)
		elapsed = 0
	}
}

// FixedUpdate wraps the given function and only executes it with the given interval.
// If the time interval is reached multiple times within a single call (frame), the function is executed as often as needed.
// The function always receives the interval instead of the elapsed time as its parameter.
func FixedUpdate(interval time.Duration, job UpdateFunc) UpdateFunc {
	elapsed := time.Duration(0)
	return func(duration time.Duration) {
		elapsed += duration
		for ; elapsed >= interval; elapsed -= interval {
			job(interval)
		}
	}
}
