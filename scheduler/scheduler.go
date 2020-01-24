package scheduler

import (
	"time"
)

/* Eliminations occur at regurlar intervals,
with a time before the first but not after the last.

Start    elim1    elim2    elim3+End
  |--------|--------|--------|
*/

// Schedule tracks the times of the start, end, and progress
type Schedule struct {
	startTime            time.Time
	endTime              time.Time
	numberOfEliminations int
}

// Phase is an Enum
type Phase int

const (
	// Before my linter really likes comments
	Before Phase = iota
	// During ..
	During
	// After ..
	After
) // Golang Enum notation is weird

func (sch Schedule) getPhase() Phase {
	now := time.Now()

	if now.Before(sch.startTime) {
		return Before
	} else if now.Before(sch.endTime) {
		return During
	}
	return After
}

func (sch Schedule) getEliminations() int {
	now := time.Now()

	if now.Before(sch.startTime) {
		return 0
	} else if now.Before(sch.endTime) {
		// Dividing these time durations is confusing
		// See https://stackoverflow.com/questions/54777109/dividing-a-time-duration-in-golang
		eliminationPeriod := sch.endTime.Sub(sch.startTime) / time.Duration(sch.numberOfEliminations)

		timeSinceStart := now.Sub(sch.startTime)
		return int(timeSinceStart / eliminationPeriod)
	}
	return sch.numberOfEliminations
}
