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

// CreateSchedule makes a new schedule
func CreateSchedule(start time.Time, end time.Time, elim int) Schedule {
	return Schedule{
		startTime:            start,
		endTime:              end,
		numberOfEliminations: elim,
	}
}

// GetPhase returns the current phase of the schedule
func (sch Schedule) GetPhase() Phase {
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
		/* ints should be converted to Duration before multiplying (?)
		https://stackoverflow.com/questions/17573190/how-to-multiply-duration-by-integer
		*/
		eliminationPeriod := sch.endTime.Sub(sch.startTime) / time.Duration(sch.numberOfEliminations)

		timeSinceStart := now.Sub(sch.startTime)
		return int(timeSinceStart / eliminationPeriod)
	}
	return sch.numberOfEliminations
}

// TriggerChangeOccurs uses a channel to indicate when the gamestate should be rechecked
/* c<-true occurs every time the game starts or an elimination occurs. On the very last
elimination, the c<-false is sent instead. This marks the game end and should be taken
as a sign to close the channel. If any of these events have already passed nothing is
sent with the expection of the final c<-false, which is sent immediately."
*/
func (sch Schedule) TriggerChangeOccurs(c chan bool) {
	now := time.Now()

	eliminationPeriod := sch.endTime.Sub(sch.startTime) / time.Duration(sch.numberOfEliminations)

	tic := now.Sub(sch.startTime)
	// This should tick on both start and end
	for i := 0; i <= sch.numberOfEliminations; i++ {
		if i == sch.numberOfEliminations { // if this is the last elimination..

			go func(tic time.Duration, c chan bool) {
				time.Sleep(tic)
				c <- false // ..close the channel
			}(tic, c)

		} else if tic >= 0 {

			// Create a process which waits for an amount of time,
			//  then sends true and returns
			go func(tic time.Duration, c chan bool) {
				time.Sleep(tic)
				c <- true
			}(tic, c)
		}
		tic += eliminationPeriod
	}
}
