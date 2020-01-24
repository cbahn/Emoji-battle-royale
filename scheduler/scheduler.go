package scheduler

import (
	"time"
)

/* Eliminations occur at regurlar intervals,
with a time before the first but not after the last.

Start    elim1    elim2    elim3+End
  |--------|--------|--------|
*/

type Schedule struct {
	startTime            time.Time
	endTime              time.Time
	numberOfEliminations int
}

type Phase int

const (
	Before Phase = iota
	During
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
