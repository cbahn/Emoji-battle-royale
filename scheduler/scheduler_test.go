package scheduler

import (
	//	"fmt"

	"testing"
	"time"
)

func TestPhase(t *testing.T) {

	now := time.Now()

	unstarted := Schedule{
		startTime:            now.Add(1 * time.Hour),
		endTime:              now.Add(2 * time.Hour),
		numberOfEliminations: 3,
	}

	if unstarted.GetPhase() != Before {
		t.Errorf("unstarted.getPhase() test failed")
	}

	started := Schedule{
		startTime:            now.Add(-1 * time.Hour),
		endTime:              now.Add(1 * time.Hour),
		numberOfEliminations: 3,
	}

	if started.GetPhase() != During {
		t.Errorf("unstarted.getPhase() test failed")
	}

	ended := Schedule{
		startTime:            now.Add(-2 * time.Hour),
		endTime:              now.Add(-1 * time.Hour),
		numberOfEliminations: 3,
	}

	if ended.GetPhase() != After {
		t.Errorf("ended.getPhase() test failed")
	}
}

func TestGetEliminations(t *testing.T) {
	now := time.Now()

	/* each row takes the form:
	{start hour, end hour, number eliminations, expeced value}
	*/
	testData := [][]int{
		{-1, 1, 1, 0},
		{1, 2, 5, 0},
		{-2, -1, 5, 5},
		{-1, 1, 5, 2},
		{-1000, 1, 7, 6},
		{-20, 57, 93, 24},
	}

	for i, d := range testData {
		sch := Schedule{
			startTime:            now.Add(time.Duration(d[0]) * time.Hour),
			endTime:              now.Add(time.Duration(d[1]) * time.Hour),
			numberOfEliminations: d[2],
		}

		if sch.getEliminations() != d[3] {
			t.Errorf("Test[%d] expected %d, got %d", i, d[3], sch.getEliminations())
		}
	}
}

/*
// TestTriggers isn't really a good test because it requires real-time to complete it
// I'm not sure of a good way to redo the
func TestTriggers(t *testing.T) {
	now := time.Now()
	sch := Schedule{
		startTime:            now.Add(1 * time.Second),
		endTime:              now.Add(11 * time.Second),
		numberOfEliminations: 17,
	}

	c := make(chan bool, 100)

	sch.TriggerChangeOccurs(c)

	chanOpen := true
	for chanOpen {
		chanOpen = <-c
		fmt.Printf("-- %s\n", time.Now().Format(time.RFC3339))
	}
}
*/
