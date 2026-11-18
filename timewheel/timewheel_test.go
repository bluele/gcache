package timewheel

import (
	"fmt"
	"testing"
	"time"
)

func TestCallbackCanAddWhenDispatchQueueIsFull(t *testing.T) {
	tw := New(1, 1)
	const jobCount = 4
	completed := make(chan interface{}, jobCount)
	tw.SetCallback(func(job *Job) {
		// More requests than reqs can buffer: dispatch must keep serving them.
		for i := 0; i < 3; i++ {
			tw.Add(&Job{Id: fmt.Sprintf("future-%v-%d", job.Id, i), Time: time.Now().Unix() + 3600})
		}
		completed <- job.Id
	})
	for i := 0; i < jobCount; i++ {
		tw.Add(&Job{Id: i, Time: time.Now().Unix()})
	}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for i := 0; i < jobCount; i++ {
		select {
		case <-completed:
		case <-timer.C:
			t.Fatalf("completed %d of %d callbacks; dispatch blocked requests", i, jobCount)
		}
	}
	tw.Purge()
}
