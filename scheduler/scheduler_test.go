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

	if unstarted.getPhase() != Before {
		t.Errorf("unstarted.getPhase() test failed")
	}

	started := Schedule{
		startTime:            now.Add(-1 * time.Hour),
		endTime:              now.Add(1 * time.Hour),
		numberOfEliminations: 3,
	}

	if started.getPhase() != During {
		t.Errorf("unstarted.getPhase() test failed")
	}

	ended := Schedule{
		startTime:            now.Add(-2 * time.Hour),
		endTime:              now.Add(-1 * time.Hour),
		numberOfEliminations: 3,
	}

	if ended.getPhase() != After {
		t.Errorf("ended.getPhase() test failed")
	}
}
